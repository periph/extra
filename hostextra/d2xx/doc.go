// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package d2xx implements support for various Future Technology devices like
// the FT232H USB GPIO, I²C, SPI, CAN, UART, JTAG bus via its D2XX driver.
//
// Debian
//
// This includes Raspbian and Ubuntu.
//
// Configure cgo as explained at https://periph.io/x/extra#hdr-Debian.
//
// Temporary: Run this command after connecting your FTDI device to temporarily
// disable linux's native driver:
//  sudo modprobe -r ftdi_so usbserial
//
// Permanent: Reconnect your device after running the following command:
//
//  cd $GOPATH/src/periph.io/x/extra/hostextra/d2xx
//  sudo cp debian/d98-ft232h.rules /etc/udev/rules.d/
//  sudo udevadm control --reload-rules
//  sudo udevadm trigger --verbose
//
// That's it!
//
// MacOS
//
// Configure cgo as explained at https://periph.io/x/extra#hdr-MacOS.
//
// Disable Apple's native FTDI driver with:
//  sudo kextunload -b com.apple.driver.AppleUSBFTDI
//
// That's it!
//
// Windows
//
// Install the driver from http://www.ftdichip.com/Drivers/D2XX.htm.
//
// Good news, no configuration is needed, it'll work as-is!
//
// Supported products
//
// http://www.ftdichip.com/Products/ICs/FT232R.htm
//
// http://www.ftdichip.com/Products/ICs/FT232H.htm
//
// Datasheets
//
// http://www.ftdichip.com/Support/Documents/DataSheets/ICs/DS_FT232R.pdf
//
// http://www.ftdichip.com/Support/Documents/DataSheets/ICs/DS_FT232H.pdf
//
// Troubleshooting
//
// See sources in
// https://github.com/periph/extra/tree/master/hostextra/d2xx
// for more developer links.
package d2xx
