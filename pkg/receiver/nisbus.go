package receiver

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Ares566/RaspberryPi-SBUS-Golang/pkg/utils"

	"github.com/d2r2/go-logger"
	"github.com/stianeikeland/go-rpio/v4"
)

const (
	startbyte   byte   = 0x0f
	endbyte     byte   = 0x00
	mask        uint16 = 0x07ff // The maximum 11-bit channel value
	middlePoint        = 993
)

// Flags stores SBUS flags
type Flags struct {
	Ch17      bool
	Ch18      bool
	Framelost bool
	Failsafe  bool
	FrameOK   bool
}

type SBUS struct {
	port io.ReadWriteCloser

	// TODO delete example code
	stepPinY rpio.Pin
	dirPinY  rpio.Pin
	stepPinX rpio.Pin
	dirPinX  rpio.Pin

	Channels [16]uint16
	Flags    Flags
	log      logger.PackageLog
}

func (n *SBUS) ScanBUS(ctx context.Context) {

	sbusPackage := make([]byte, 0, 25)

	for {

		select {
		case <-ctx.Done():
			n.log.Info("stop ScanBUS")
			return
		default:
		}

		buf := make([]byte, 25)
		_, err := n.port.Read(buf)
		if err != nil {
			n.log.Error("Error reading data: ", err)
			continue
		}

		for _, packageByte := range buf {

			select {
			case <-ctx.Done():
				n.log.Info("stop ScanBUS")
				return
			default:
			}

			// ждем стартовый байт
			if len(sbusPackage) == 0 && packageByte != startbyte {
				continue
			}

			sbusPackage = append(sbusPackage, packageByte)

			if len(sbusPackage) == 25 {
				if packageByte != endbyte {
					n.log.Info("wrong data package: ", sbusPackage)
				}

				err = n.unmarshalFrame(sbusPackage)
				if err != nil {
					n.log.Error(err)
					n.Flags.FrameOK = false
				} else {
					//n.log.Infof("RX: %v", n.Channels)
					n.Flags.FrameOK = true
				}

				sbusPackage = make([]byte, 0, 25)
			}
		}
	}
}

func (n *SBUS) Start(ctx context.Context) {

	// get signal from SBUS
	go n.ScanBUS(ctx)

	// TODO delete example processors, add chan Channels & Flags

	// processing chanels
	// left/right chanel 1
	go n.serveX(ctx)

	// up/down chanel 2
	go n.serveY(ctx)

}

// serveX example of processing chanel 1
func (n *SBUS) serveX(ctx context.Context) {
	for {

		select {
		case <-ctx.Done():
			n.log.Info("stop serveX")
			return
		default:
		}

		if !n.Flags.FrameOK {
			continue
		}

		delta := int16(n.Channels[1] - middlePoint)
		absDelta := utils.Abs(delta)
		if absDelta > 50 {
			if delta < 0 {
				n.dirPinX.High()
			} else {
				n.dirPinX.Low()
			}
		} else {
			n.stepPinX.Low()
			continue
		}

		speed := uint16(utils.Mapping(float64(absDelta), 1, 850, 5000, 50))
		n.log.Infof("speedX = %d ", speed)
		n.log.Infof("absDeltaX = %d \n", absDelta)
		if speed < 50 {
			speed = 50
		}
		if speed > 5000 {
			speed = 5000
		}

		timing := int(utils.Mapping(float64(absDelta), 50, 5000, 10, 100))
		for i := 0; i < timing; i++ {
			n.stepPinX.High()
			time.Sleep(time.Duration(speed) * time.Microsecond)
			n.stepPinX.Low()
			time.Sleep(time.Duration(speed) * time.Microsecond)
		}
	}
}

// serveY example of processing chanel  2
func (n *SBUS) serveY(ctx context.Context) {
	for {

		select {
		case <-ctx.Done():
			n.log.Info("stop serveY")
			return
		default:
		}

		if !n.Flags.FrameOK {
			continue
		}

		delta := int16(n.Channels[2] - middlePoint)
		absDelta := utils.Abs(delta)
		if absDelta > 50 {
			if delta < 0 {
				n.dirPinY.High()
			} else {
				n.dirPinY.Low()
			}
		} else {
			n.stepPinY.Low()
			continue
		}

		speed := uint16(utils.Mapping(float64(absDelta), 1, 850, 5000, 50))
		n.log.Infof("speedY = %d ", speed)
		n.log.Infof("absDeltaY = %d \n", absDelta)
		if speed < 50 {
			speed = 50
		}
		if speed > 5000 {
			speed = 5000
		}

		timing := int(utils.Mapping(float64(absDelta), 50, 5000, 10, 100))
		for i := 0; i < timing; i++ {
			n.stepPinY.High()
			time.Sleep(time.Duration(speed) * time.Microsecond)
			n.stepPinY.Low()
			time.Sleep(time.Duration(speed) * time.Microsecond)
		}
	}
}

// UnmarshalFrame tries to create a Frame from a byte array
func (n *SBUS) unmarshalFrame(data []byte) (err error) {
	if data[0] != startbyte {
		err = fmt.Errorf("incorrect start byte %v", data[0])
		return
	}
	if data[24] != endbyte {
		err = fmt.Errorf("incorrect end byte %v", data[24])
		return
	}

	n.Channels[0] = (uint16(data[1]) | uint16(data[2])<<8) & mask
	n.Channels[1] = (uint16(data[2])>>3 | uint16(data[3])<<5) & mask
	n.Channels[2] = (uint16(data[3])>>6 | uint16(data[4])<<2 | uint16(data[5])<<10) & mask
	n.Channels[3] = (uint16(data[5])>>1 | uint16(data[6])<<7) & mask
	n.Channels[4] = (uint16(data[6])>>4 | uint16(data[7])<<4) & mask
	n.Channels[5] = (uint16(data[7])>>7 | uint16(data[8])<<1 | uint16(data[9])<<9) & mask
	n.Channels[6] = (uint16(data[9])>>2 | uint16(data[10])<<6) & mask
	n.Channels[7] = (uint16(data[10])>>5 | uint16(data[11])<<3) & mask
	n.Channels[8] = (uint16(data[12]) | uint16(data[13])<<8) & mask
	n.Channels[9] = (uint16(data[13])>>3 | uint16(data[14])<<5) & mask
	n.Channels[10] = (uint16(data[14])>>6 | uint16(data[15])<<2 | uint16(data[16])<<10) & mask
	n.Channels[11] = (uint16(data[16])>>1 | uint16(data[17])<<7) & mask
	n.Channels[12] = (uint16(data[17])>>4 | uint16(data[18])<<4) & mask
	n.Channels[13] = (uint16(data[18])>>7 | uint16(data[19])<<1 | uint16(data[20])<<9) & mask
	n.Channels[14] = (uint16(data[20])>>2 | uint16(data[21])<<6) & mask
	n.Channels[15] = (uint16(data[21])>>5 | uint16(data[22])<<3) & mask

	n.Flags.Failsafe = (data[23] & 0x10) != 0
	n.Flags.Framelost = (data[23] & 0x20) != 0
	n.Flags.Ch18 = (data[23] & 0x40) != 0
	n.Flags.Ch17 = (data[23] & 0x80) != 0

	return
}

func NewReceiver(port io.ReadWriteCloser, stepPinX rpio.Pin, dirPinX rpio.Pin, stepPinY rpio.Pin, dirPinY rpio.Pin, log logger.PackageLog) *SBUS {

	return &SBUS{
		port: port,

		stepPinX: stepPinX,
		dirPinX:  dirPinX,

		stepPinY: stepPinY,
		dirPinY:  dirPinY,

		log: log,
	}
}
