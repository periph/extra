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

func d2xxGetLibraryVersion() (uint8, uint8, uint8) {
	var v C.DWORD
	C.FT_GetLibraryVersion(&v)
	return uint8(v >> 16), uint8(v >> 8), uint8(v)
}

func d2xxCreateDeviceInfoList() (int, int) {
	var num C.DWORD
	e := C.FT_CreateDeviceInfoList(&num)
	return int(num), int(e)
}

// Device functions.

func d2xxOpen(i int) (*device, int) {
	var h C.FT_HANDLE
	e := C.FT_Open(C.int(i), &h)
	if uintptr(h) == 0 && e == 0 {
		panic("unexpected")
	}
	return &device{h: handle(h)}, int(e)
}

func (d *device) d2xxClose() int {
	e := C.FT_Close(d.toH())
	return int(e)
}

func (d *device) d2xxResetDevice() int {
	e := C.FT_ResetDevice(d.toH())
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

func (d *device) setup() int {
	// Disable event/error characters.
	if e := C.FT_SetChars(d.toH(), 0, 0, 0, 0); e != 0 {
		return int(e)
	}
	// Set I/O timeouts to 5 sec.
	if e := C.FT_SetTimeouts(d.toH(), 5000, 5000); e != 0 {
		return int(e)
	}
	// Latency timer at default 16ms.
	return int(C.FT_SetLatencyTimer(d.toH(), 16))
}

func (d *device) d2xxGetQueueStatus() (uint32, int) {
	var v C.DWORD
	e := C.FT_GetQueueStatus(d.toH(), &v)
	return uint32(v), int(e)
}

func (d *device) d2xxRead(b []byte) (int, int) {
	var bytesRead C.DWORD
	e := C.FT_Read(d.toH(), C.LPVOID(unsafe.Pointer(&b[0])), C.DWORD(len(b)), &bytesRead)
	return int(bytesRead), int(e)
}

func (d *device) d2xxWrite(b []byte) (int, int) {
	var bytesSent C.DWORD
	e := C.FT_Write(d.toH(), C.LPVOID(unsafe.Pointer(&b[0])), C.DWORD(len(b)), &bytesSent)
	return int(bytesSent), int(e)
}

func (d *device) d2xxGetBitMode() (byte, int) {
	var s C.UCHAR
	e := C.FT_GetBitMode(d.toH(), &s)
	return uint8(s), int(e)
}

func (d *device) d2xxSetBitMode(mask, mode byte) int {
	return int(C.FT_SetBitMode(d.toH(), C.UCHAR(mask), C.UCHAR(mode)))
}

func (d *device) toH() C.FT_HANDLE {
	return C.FT_HANDLE(d.h)
}

// handle is a d2xx handle.
type handle C.FT_HANDLE
