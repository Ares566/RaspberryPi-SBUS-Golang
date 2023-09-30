package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ares566/RaspberryPi-SBUS-Golang/pkg/receiver"

	"github.com/d2r2/go-logger"
	"github.com/jacobsa/go-serial/serial"
	"github.com/stianeikeland/go-rpio/v4"
)

var lg = logger.NewPackageLogger("main",
	//logger.DebugLevel,
	logger.InfoLevel,
)

func main() {

	defer func() {
		_ = logger.FinalizeLogger()
	}()

	ctx, cancel := context.WithCancel(context.Background())

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	defer lg.Info("goodbye")

	options := serial.OpenOptions{
		PortName:        "/dev/ttyAMA0",
		BaudRate:        100000,
		DataBits:        8,
		StopBits:        2,
		ParityMode:      serial.PARITY_EVEN,
		MinimumReadSize: 25,
	}

	port, err := serial.Open(options)
	if err != nil {
		log.Fatal("error serial port opening ", err)
	}
	defer port.Close()

	// Open and map memory to access gpio, check for errors
	if err := rpio.Open(); err != nil {
		lg.Fatal(err)
		return
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// connect 2 stepper motors to Raspberry for example
	dirPinY := rpio.Pin(23)
	stepPinY := rpio.Pin(22)

	dirPinY.Output()  // Output mode
	stepPinY.Output() // Output mode
	dirPinY.PullDown()
	stepPinY.PullDown()

	dirPinX := rpio.Pin(24)
	stepPinX := rpio.Pin(25)

	dirPinX.Output()  // Output mode
	stepPinX.Output() // Output mode
	dirPinX.PullDown()
	stepPinX.PullDown()

	//turretControl := turret.New(imu, dirPinY, stepPinY, lg)
	//turretControl.Run(ctx)

	turretControlReceiver := receiver.NewReceiver(port, stepPinX, dirPinX, stepPinY, dirPinY, lg)
	turretControlReceiver.Start(ctx)

	<-shutdown

	lg.Info("start shutdown system")

	cancel()

}
