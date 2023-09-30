# RaspberryPi SBUS
Realization of SBUS Protocol for Raspberry Pi written in Golang. Tested on RaspberryPi 3/4


### Prepare your Raspberry Pi 3
By default, the serial port has a login console running so there is some config updates needed to get the serial port working over GPIO. As it turns out, this interferes with the default bluetooth configuration because they both use the same GPIO interfaces.

Set up the serial port with `sudo raspi-config`. 
Turn on serial port, turn off console login

And last step is configure `/boot/config.txt`.

`> nano /boot/config.txt`
```
enable_uart=1
dtoverlay=pi3-disable-bt
```