// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

import (
	"errors"
	"fmt"
	"unsafe"
)

// MissingErr is returned when the library ftd2xx couldn't be loaded at runtime.
var MissingErr = errors.New("ftd2xx: couldn't load driver; visit https://github.com/periph/extra/tree/master/experimental/devices/ftdi/ftd2xx")

// Handle is a device handle.
type Handle uintptr

// Close closes an Handle.
func (h Handle) Close() error {
	return closeHandle(h)
}

// GetDevInfo returns information about a specific device already opened.
func (h Handle) GetDevInfo(i *DevInfo) error {
	return getDeviceInfo(h, i)
}

// Type is the FTDI device type.
type Type int

const (
	FTBM Type = iota
	FTAM
	FT100AX
	Unknown
	FT2232C
	FT232R
	FT2232H
	FT4232H
	FT232H
	FTXSeries
	FT4222H0
	FT4222H1_2
	FT4222H3
	FT4222Prog
)

// DevInfo is generic information about a connected device.
type DevInfo struct {
	Opened  bool
	HiSpeed bool
	Type    Type
	ID      uint32
	LocID   uint32
	Serial  string
	Desc    string
	h       Handle
}

// TODO(maruel): In practice we'll probably only keep Open() by index after an
// enumeration.

// OpenBySerial opens a device via its serial number.
func OpenBySerial(serial string) (Handle, error) {
	r := []rune(serial + "\000")
	return openEx((uintptr)(unsafe.Pointer(&r[0])), openBySerialNumber)
}

// OpenByDescription opens a device via its description.
func OpenByDescription(desc string) (Handle, error) {
	r := []rune(desc + "\000")
	return openEx((uintptr)(unsafe.Pointer(&r[0])), openByDescription)
}

// OpenByLocation opens a device via its location on the USB bus.
func OpenByLocation(locID uint32) (Handle, error) {
	return openEx(uintptr(locID), openByLocation)
}

// ListDevices returns the list of enumerated FTDI devices discovered.
func ListDevices() ([]DevInfo, error) {
	num, err := createDeviceInfoList()
	if err != nil || num == 0 {
		return nil, err
	}
	// Returns barely anything unless the device was opened.
	return getDeviceInfoList(num)
}

//

// OpenEx Flags
const openBySerialNumber = 1
const openByDescription = 2
const openByLocation = 4

// Baud Rates
// ...

func toErr(e int) error {
	switch e {
	case 0: // FT_OK
		return nil
	case 1: // FT_INVALID_HANDLE
		return errors.New("ftd2xx: invalid handle")
	case 2: // FT_DEVICE_NOT_FOUND
		return errors.New("ftd2xx: device not found")
	case 3: // FT_DEVICE_NOT_OPENED
		return errors.New("ftd2xx: device not opened")
	case 4: // FT_IO_ERROR
		return errors.New("ftd2xx: I/O error")
	case 5: // FT_INSUFFICIENT_RESOURCES
		return errors.New("ftd2xx: insufficient resources")
	case 6: // FT_INVALID_PARAMETER
		return errors.New("ftd2xx: invalid parameter")
	case 7: // FT_INVALID_BAUD_RATE
		return errors.New("ftd2xx: invalid baud rate")
	case 8: // FT_DEVICE_NOT_OPENED_FOR_ERASE
		return errors.New("ftd2xx: device not opened for erase")
	case 9: // FT_DEVICE_NOT_OPENED_FOR_WRITE
		return errors.New("ftd2xx: device not opened for write")
	case 10: // FT_FAILED_TO_WRITE_DEVICE
		return errors.New("ftd2xx: failed to write device")
	case 11: // FT_EEPROM_READ_FAILED
		return errors.New("ftd2xx: eeprom read failed")
	case 12: // FT_EEPROM_WRITE_FAILED
		return errors.New("ftd2xx: eeprom write failed")
	case 13: // FT_EEPROM_ERASE_FAILED
		return errors.New("ftd2xx: eeprom erase failed")
	case 14: // FT_EEPROM_NOT_PRESENT
		return errors.New("ftd2xx: eeprom not present")
	case 15: // FT_EEPROM_NOT_PROGRAMMED
		return errors.New("ftd2xx: eeprom not programmed")
	case 16: // FT_INVALID_ARGS
		return errors.New("ftd2xx: invalid argument")
	case 17: // FT_NOT_SUPPORTED
		return errors.New("ftd2xx: not supported")
	case 18: // FT_OTHER_ERROR
		return errors.New("ftd2xx: other error")
	case 19: // FT_DEVICE_LIST_NOT_READY
		return errors.New("ftd2xx: device list not ready")
	default:
		return fmt.Errorf("ftd2xx: unknown status %d", e)
	}
}
