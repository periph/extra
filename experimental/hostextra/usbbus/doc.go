// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package usbbus implements OS specific functions for conn/usb.
//
// This includes handling the connected devices on process startup and handling
// of the connection events.
//
// This package is only built with the build tag 'usb' because it causes a
// dependency on https://github.com/kylelemons/gousb. This package uses cgo that
// depends on libusb being installed. This is generally not the case by
// default, so it causes a go get failure which is really obnoxious to users.
//
// Debian
//
// This includes Raspbian and Ubuntu.
//
// First configure cgo as explained at https://periph.io/x/extra#hdr-Debian.
//
// You need to install libusb-1.0:
//
//  sudo apt install libusb-1.0-0-dev
//
// MacOS
//
// First configure cgo as explained at https://periph.io/x/extra#hdr-MacOS.
//
//  brew install libusb
//
// Windows
//
// The package is currently disabled on Windows.
package usbbus
