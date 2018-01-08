// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

import (
	"bytes"
	"syscall"
	"unsafe"
)

// Library functions.

func getLibraryVersion() (uint8, uint8, uint8) {
	var v uint32
	if pGetLibraryVersion != nil {
		pGetLibraryVersion.Call(uintptr(unsafe.Pointer(&v)))
	}
	return uint8(v >> 16), uint8(v >> 8), uint8(v)
}

func createDeviceInfoList() (int, int) {
	if pCreateDeviceInfoList == nil {
		return 0, missing
	}
	var num uint32
	r1, _, _ := pCreateDeviceInfoList.Call(uintptr(unsafe.Pointer(&num)))
	return int(num), int(r1)
}

// Device functions.

func open(i int) (*device, int) {
	var h handle
	if pOpen == nil {
		return nil, missing
	}
	r1, _, _ := pOpen.Call(uintptr(i), uintptr(unsafe.Pointer(&h)))
	return &device{h: h}, int(r1)
}

func (d *device) closeHandle() int {
	if pClose == nil {
		return missing
	}
	r1, _, _ := pClose.Call(d.toH())
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
	hdr := (*eeprom_header)(eepromVoid)
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

func (d *device) getReadPending() (int, int) {
	if pGetQueueStatus == nil {
		return 0, missing
	}
	return 0, missing
}

func (d *device) doRead(b []byte) (int, int) {
	if pRead == nil {
		return 0, missing
	}
	return 0, missing
}

func (d *device) getBits() (byte, int) {
	if pGetBitMode == nil {
		return 0, missing
	}
	var s uint8
	r1, _, _ := pGetBitMode.Call(d.toH(), uintptr(unsafe.Pointer(&s)))
	return s, int(r1)
}

func (d *device) toH() uintptr {
	return uintptr(d.h)
}

// handle is a d2xx handle.
//
// TODO(maruel): Convert to type alias once go 1.9+ is required.
type handle uintptr

//

var (
	// Library functions.
	pGetLibraryVersion    *syscall.Proc
	pCreateDeviceInfoList *syscall.Proc
	//pGetDeviceInfoList    *syscall.Proc

	// Device functions.
	pOpen           *syscall.Proc
	pClose          *syscall.Proc
	pGetDeviceInfo  *syscall.Proc
	pEEPROMRead     *syscall.Proc
	pGetBitMode     *syscall.Proc
	pSetBitMode     *syscall.Proc
	pGetQueueStatus *syscall.Proc
	pRead           *syscall.Proc
)

func init() {
	if dll, _ := syscall.LoadDLL("ftd2xx.dll"); dll != nil {
		// Library functions.
		pGetLibraryVersion, _ = dll.FindProc("FT_GetLibraryVersion")
		pCreateDeviceInfoList, _ = dll.FindProc("FT_CreateDeviceInfoList")
		//pGetDeviceInfoList, _ = dll.FindProc("FT_GetDeviceInfoList")

		// Device functions.
		pOpen, _ = dll.FindProc("FT_Open")
		pClose, _ = dll.FindProc("FT_Close")
		pGetDeviceInfo, _ = dll.FindProc("FT_GetDeviceInfo")
		pEEPROMRead, _ = dll.FindProc("FT_EEPROM_Read")
		pGetBitMode, _ = dll.FindProc("FT_GetBitMode")
		pSetBitMode, _ = dll.FindProc("FT_SetBitMode")
		pGetQueueStatus, _ = dll.FindProc("FT_GetQueueStatus")
		pRead, _ = dll.FindProc("FT_Read")
	}
}

func toStr(c []byte) string {
	i := bytes.IndexByte(c, 0)
	if i != -1 {
		return string(c[:i])
	}
	return string(c)
}
