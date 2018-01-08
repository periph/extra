// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !windows

package ftd2xx

/*
#include "ftd2xx.h"
*/
import "C"
import (
	"unsafe"
)

// Library functions.

func getLibraryVersion() (uint8, uint8, uint8) {
	var v C.DWORD
	C.FT_GetLibraryVersion(&v)
	return uint8(v >> 16), uint8(v >> 8), uint8(v)
}

func createDeviceInfoList() (int, int) {
	var num C.DWORD
	e := C.FT_CreateDeviceInfoList(&num)
	return int(num), int(e)
}

// Device functions.

func open(i int) (*device, int) {
	var h C.FT_HANDLE
	e := C.FT_Open(C.int(i), &h)
	if uintptr(h) == 0 && e == 0 {
		panic("unexpected")
	}
	return &device{h: handle(h)}, int(e)
}

func (d *device) closeHandle() int {
	e := C.FT_Close(d.toH())
	return int(e)
}

func (d *device) getInfo() int {
	var dev C.FT_DEVICE
	var id C.DWORD
	// TODO(maruel): When specifying serial or desc, the function fails. It's not
	// really important because we read the EEPROM instead.
	if e := C.FT_GetDeviceInfo(d.toH(), &dev, &id, nil, nil, nil); e != 0 {
		return int(e)
	}
	d.t = devType(dev)
	d.venID = uint16(id >> 16)
	d.productID = uint16(id)

	var manufacturer [64]C.char
	var manufacturerID [64]C.char
	var desc [64]C.char
	var serial [64]C.char
	d.eeprom = make([]byte, d.t.eepromSize())
	eepromVoid := unsafe.Pointer(&d.eeprom[0])
	hdr := (*eeprom_header)(eepromVoid)

	// There something odd going on here.
	//
	// On a ft232h, we observed that hdr.deviceType MUST NOT be set, but on a
	// ft232r, it MUST be set. Since we can't know in advance what we must use,
	// just try both. ¯\_(ツ)_/¯
	hdr.deviceType = d.t
	if e := C.FT_EEPROM_Read(d.toH(), eepromVoid, C.DWORD(len(d.eeprom)), &manufacturer[0], &manufacturerID[0], &desc[0], &serial[0]); e != 0 {
		// FT_INVALID_PARAMETER
		if e == 6 {
			hdr.deviceType = 0
			e = C.FT_EEPROM_Read(d.toH(), eepromVoid, C.DWORD(len(d.eeprom)), &manufacturer[0], &manufacturerID[0], &desc[0], &serial[0])
		}
		if e != 0 {
			return int(e)
		}
	}

	d.manufacturer = C.GoString(&manufacturer[0])
	d.manufacturerID = C.GoString(&manufacturerID[0])
	d.desc = C.GoString(&desc[0])
	d.serial = C.GoString(&serial[0])
	return 0
}

func (d *device) getReadPending() (int, int) {
	// C.FT_GetQueueStatus(d.toH(), &pendingBytes);
	return 0, missing
}

func (d *device) doRead(b []byte) (int, int) {
	// FT_Read(d.toH(), &b[0], len(b), &bytesRead);
	return 0, missing
}

func (d *device) getBits() (byte, int) {
	var s C.UCHAR
	e := C.FT_GetBitMode(d.toH(), &s)
	return uint8(s), int(e)
}

func (d *device) toH() C.FT_HANDLE {
	return C.FT_HANDLE(d.h)
}

// handle is a d2xx handle.
//
// TODO(maruel): Convert to type alias once go 1.9+ is required.
type handle C.FT_HANDLE
