// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package ftdi implements support for various Future Technology devices like
// the FT232H USB GPIO, I²C, SPI, CAN, UART, JTAG bus.
//
// It currently leverages subpackage ftd2xx but it is designed to be ported to
// the open source library libftdi.
//
// Debian
//
// This includes Raspbian and Ubuntu.
//
// First configure cgo as explained at https://periph.io/x/extra#hdr-Debian.
//
// You need to install libusb-1.0, run:
//
//  sudo apt install libusb-1.0-0-dev
//
// MacOS
//
// First configure cgo as explained at https://periph.io/x/extra#hdr-MacOS.
//
// That's it!
//
// Windows
//
// Good news, no configuration is needed, it'll work as-is.
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
// See doc.go in
// https://github.com/periph/extra/tree/master/experimental/devices/ftdi/ftd2xx
// for more developer links.
package ftdi

// There is multiple ways to access a FT232H:
//
// - Some operating systems include a limited "serial port only" driver.
// - Future Technologic Devices International Ltd provides their own private
//   source driver.
// - FTDI also provides a "serial port only" driver surnamed VCP.
// - https://www.intra2net.com/en/developer/libftdi/ is an open source driver,
//   that is acknowledged by FTDI.
//
// Interfacing I²C:
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_113_FTDI_Hi_Speed_USB_To_I2C_Example.pdf
//
// Interfacing SPI:
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_114_FTDI_Hi_Speed_USB_To_SPI_Example.pdf
//
// Interfacing JTAG:
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_129_FTDI_Hi_Speed_USB_To_JTAG_Example.pdf
//
// Interfacing parallel port:
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_167_FT1248_Parallel_Serial_Interface_Basics.pdf
//
// MPSSE basics:
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_135_MPSSE_Basics.pdf
//
// MPSSE and MCU emulation modes:
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_108_Command_Processor_for_MPSSE_and_MCU_Host_Bus_Emulation_Modes.pdf
