// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package d2xx

import (
	"bytes"
	"syscall"
	"unsafe"
)

const disabled = false

// Library functions.

func d2xxGetLibraryVersion() (uint8, uint8, uint8) {
	var v uint32
	if pGetLibraryVersion != nil {
		pGetLibraryVersion.Call(uintptr(unsafe.Pointer(&v)))
	}
	return uint8(v >> 16), uint8(v >> 8), uint8(v)
}

func d2xxCreateDeviceInfoList() (int, int) {
	if pCreateDeviceInfoList == nil {
		return 0, missing
	}
	var num uint32
	r1, _, _ := pCreateDeviceInfoList.Call(uintptr(unsafe.Pointer(&num)))
	return int(num), int(r1)
}

// Device functions.

func d2xxOpen(i int) (*device, int) {
	var h handle
	if pOpen == nil {
		return nil, missing
	}
	r1, _, _ := pOpen.Call(uintptr(i), uintptr(unsafe.Pointer(&h)))
	return &device{h: h}, int(r1)
}

func (d *device) d2xxClose() int {
	if pClose == nil {
		return missing
	}
	r1, _, _ := pClose.Call(d.toH())
	return int(r1)
}

func (d *device) d2xxResetDevice() int {
	if pResetDevice == nil {
		return missing
	}
	r1, _, _ := pResetDevice.Call(d.toH())
	return int(r1)
}

func (d *device) getInfo() int {
	if pGetDeviceInfo == nil || pEEPROMRead == nil {
		return missing
	}
	var id uint32
	if r1, _, _ := pGetDeviceInfo.Call(d.toH(), uintptr(unsafe.Pointer(&d.t)), uintptr(unsafe.Pointer(&id)), 0, 0, 0); r1 != 0 {
		return int(r1)
	}
	d.venID = uint16(id >> 16)
	d.productID = uint16(id)

	var manufacturer [64]byte
	var manufacturerID [64]byte
	var desc [64]byte
	var serial [64]byte
	// Shortcuts.
	m := uintptr(unsafe.Pointer(&manufacturer[0]))
	mi := uintptr(unsafe.Pointer(&manufacturerID[0]))
	de := uintptr(unsafe.Pointer(&desc[0]))
	s := uintptr(unsafe.Pointer(&serial[0]))

	d.eeprom = make([]byte, d.t.eepromSize())
	eepromVoid := unsafe.Pointer(&d.eeprom[0])
	hdr := (*eepromHeader)(eepromVoid)
	// It MUST be set here. This is not always the case on posix.
	hdr.deviceType = d.t
	if r1, _, _ := pEEPROMRead.Call(d.toH(), uintptr(eepromVoid), uintptr(len(d.eeprom)), m, mi, de, s); r1 != 0 {
		return int(r1)
	}

	d.manufacturer = toStr(manufacturer[:])
	d.manufacturerID = toStr(manufacturerID[:])
	d.desc = toStr(desc[:])
	d.serial = toStr(serial[:])
	return 0
}

func (d *device) setup() int {
	if pSetChars == nil || pSetTimeouts == nil || pSetLatencyTimer == nil {
		return missing
	}
	// Disable event/error characters.
	if r1, _, _ := pSetChars.Call(d.toH(), 0, 0, 0, 0); r1 != 0 {
		return int(r1)
	}
	// Set I/O timeouts to 5 sec.
	if r1, _, _ := pSetTimeouts.Call(d.toH(), 5000, 5000); r1 != 0 {
		return int(r1)
	}
	// Latency timer at default 16ms.
	r1, _, _ := pSetLatencyTimer.Call(d.toH(), 16)
	return int(r1)
}

func (d *device) d2xxGetQueueStatus() (uint32, int) {
	if pGetQueueStatus == nil {
		return 0, missing
	}
	var v uint32
	r1, _, _ := pGetQueueStatus.Call(d.toH(), uintptr(unsafe.Pointer(&v)))
	return v, int(r1)
}

func (d *device) d2xxRead(b []byte) (int, int) {
	if pRead == nil {
		return 0, missing
	}
	var bytesRead uint32
	r1, _, _ := pRead.Call(d.toH(), uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), uintptr(unsafe.Pointer(&bytesRead)))
	return 0, int(r1)
}

func (d *device) d2xxWrite(b []byte) (int, int) {
	if pWrite == nil {
		return 0, missing
	}
	var bytesSent uint32
	r1, _, _ := pWrite.Call(d.toH(), uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), uintptr(unsafe.Pointer(&bytesSent)))
	return 0, int(r1)
}

func (d *device) d2xxGetBitMode() (byte, int) {
	if pGetBitMode == nil {
		return 0, missing
	}
	var s uint8
	r1, _, _ := pGetBitMode.Call(d.toH(), uintptr(unsafe.Pointer(&s)))
	return s, int(r1)
}

func (d *device) d2xxSetBitMode(mask, mode byte) int {
	if pSetBitMode == nil {
		return missing
	}
	r1, _, _ := pSetBitMode.Call(d.toH(), uintptr(mask), uintptr(mode))
	return int(r1)
}

func (d *device) toH() uintptr {
	return uintptr(d.h)
}

// handle is a d2xx handle.
type handle uintptr

//

var (
	pClose                *syscall.Proc
	pCreateDeviceInfoList *syscall.Proc
	pEEPROMRead           *syscall.Proc
	pGetBitMode           *syscall.Proc
	pGetDeviceInfo        *syscall.Proc
	pGetLibraryVersion    *syscall.Proc
	pGetQueueStatus       *syscall.Proc
	pOpen                 *syscall.Proc
	pRead                 *syscall.Proc
	pResetDevice          *syscall.Proc
	pSetBitMode           *syscall.Proc
	pSetChars             *syscall.Proc
	pSetLatencyTimer      *syscall.Proc
	pSetTimeouts          *syscall.Proc
	pWrite                *syscall.Proc
)

func init() {
	if dll, _ := syscall.LoadDLL("ftd2xx.dll"); dll != nil {
		pClose, _ = dll.FindProc("FT_Close")
		pCreateDeviceInfoList, _ = dll.FindProc("FT_CreateDeviceInfoList")
		pEEPROMRead, _ = dll.FindProc("FT_EEPROM_Read")
		pGetBitMode, _ = dll.FindProc("FT_GetBitMode")
		pGetDeviceInfo, _ = dll.FindProc("FT_GetDeviceInfo")
		pGetLibraryVersion, _ = dll.FindProc("FT_GetLibraryVersion")
		pGetQueueStatus, _ = dll.FindProc("FT_GetQueueStatus")
		pOpen, _ = dll.FindProc("FT_Open")
		pRead, _ = dll.FindProc("FT_Read")
		pResetDevice, _ = dll.FindProc("FT_ResetDevice")
		pSetBitMode, _ = dll.FindProc("FT_SetBitMode")
		pSetChars, _ = dll.FindProc("FT_SetChars")
		pSetLatencyTimer, _ = dll.FindProc("FT_SetLatencyTimer")
		pSetTimeouts, _ = dll.FindProc("FT_SetTimeouts")
		pWrite, _ = dll.FindProc("FT_Write")
	}
}

func toStr(c []byte) string {
	i := bytes.IndexByte(c, 0)
	if i != -1 {
		return string(c[:i])
	}
	return string(c)
}
