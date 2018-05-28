// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !windows

package d2xx

/*
#include "ftd2xx.h"
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"
)

const disabled = false

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

func d2xxOpen(i int) (handle, int) {
	var h C.FT_HANDLE
	e := C.FT_Open(C.int(i), &h)
	if uintptr(h) == 0 && e == 0 {
		panic("unexpected")
	}
	return handle(h), int(e)
}

func (h handle) d2xxClose() int {
	return int(C.FT_Close(h.toH()))
}

func (h handle) d2xxResetDevice() int {
	return int(C.FT_ResetDevice(h.toH()))
}

func (h handle) d2xxGetDeviceInfo() (devType, uint16, uint16, int) {
	var dev C.FT_DEVICE
	var id C.DWORD
	if e := C.FT_GetDeviceInfo(h.toH(), &dev, &id, nil, nil, nil); e != 0 {
		return unknown, 0, 0, int(e)
	}
	return devType(dev), uint16(id >> 16), uint16(id), 0
}

func (h handle) d2xxEEPROMRead(t devType, ee *eeprom) int {
	var manufacturer [64]C.char
	var manufacturerID [64]C.char
	var desc [64]C.char
	var serial [64]C.char
	if l := t.eepromSize(); len(ee.raw) < l {
		ee.raw = make([]byte, t.eepromSize())
	} else if len(ee.raw) > l {
		ee.raw = ee.raw[:l]
	}
	eepromVoid := unsafe.Pointer(&ee.raw[0])
	hdr := (*eepromHeader)(eepromVoid)

	// There something odd going on here.
	//
	// On a ft232h, we observed that hdr.deviceType MUST NOT be set, but on a
	// ft232r, it MUST be set. Since we can't know in advance what we must use,
	// just try both. ¯\_(ツ)_/¯
	hdr.deviceType = t
	if e := C.FT_EEPROM_Read(h.toH(), eepromVoid, C.DWORD(len(ee.raw)), &manufacturer[0], &manufacturerID[0], &desc[0], &serial[0]); e != 0 {
		// FT_INVALID_PARAMETER
		if e == 6 {
			hdr.deviceType = 0
			e = C.FT_EEPROM_Read(h.toH(), eepromVoid, C.DWORD(len(ee.raw)), &manufacturer[0], &manufacturerID[0], &desc[0], &serial[0])
		}
		if e != 0 {
			return int(e)
		}
	}

	ee.manufacturer = C.GoString(&manufacturer[0])
	ee.manufacturerID = C.GoString(&manufacturerID[0])
	ee.desc = C.GoString(&desc[0])
	ee.serial = C.GoString(&serial[0])
	return 0
}

func (h handle) d2xxSetChars(eventChar byte, eventEn bool, errorChar byte, errorEn bool) int {
	v := C.UCHAR(0)
	if eventEn {
		v = 1
	}
	w := C.UCHAR(0)
	if errorEn {
		w = 1
	}
	return int(C.FT_SetChars(h.toH(), C.UCHAR(eventChar), v, C.UCHAR(errorChar), w))
}

func (h handle) d2xxSetUSBParameters(in, out int) int {
	return int(C.FT_SetUSBParameters(h.toH(), C.DWORD(in), C.DWORD(out)))
}

func (h handle) d2xxSetFlowControl() int {
	return int(C.FT_SetFlowControl(h.toH(), C.FT_FLOW_RTS_CTS, 0, 0))
}

func (h handle) d2xxSetTimeouts(readMS, writeMS int) int {
	return int(C.FT_SetTimeouts(h.toH(), C.DWORD(readMS), C.DWORD(writeMS)))
}

func (h handle) d2xxSetLatencyTimer(delayMS uint8) int {
	return int(C.FT_SetLatencyTimer(h.toH(), C.UCHAR(delayMS)))
}

func (h handle) d2xxGetQueueStatus() (uint32, int) {
	var v C.DWORD
	e := C.FT_GetQueueStatus(h.toH(), &v)
	return uint32(v), int(e)
}

func (h handle) d2xxRead(b []byte) (int, int) {
	var bytesRead C.DWORD
	e := C.FT_Read(h.toH(), C.LPVOID(unsafe.Pointer(&b[0])), C.DWORD(len(b)), &bytesRead)
	return int(bytesRead), int(e)
}

func (h handle) d2xxWrite(b []byte) (int, int) {
	var bytesSent C.DWORD
	e := C.FT_Write(h.toH(), C.LPVOID(unsafe.Pointer(&b[0])), C.DWORD(len(b)), &bytesSent)
	return int(bytesSent), int(e)
}

func (h handle) d2xxGetBitMode() (byte, int) {
	var s C.UCHAR
	e := C.FT_GetBitMode(h.toH(), &s)
	return uint8(s), int(e)
}

func (h handle) d2xxSetBitMode(mask, mode byte) int {
	return int(C.FT_SetBitMode(h.toH(), C.UCHAR(mask), C.UCHAR(mode)))
}

func (h handle) toH() C.FT_HANDLE {
	return C.FT_HANDLE(h)
}
