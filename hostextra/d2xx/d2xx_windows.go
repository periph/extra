// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package d2xx

import (
	"bytes"
	"syscall"
	"unsafe"
)

var disabled = true

// Library functions.

func d2xxGetLibraryVersion() (uint8, uint8, uint8) {
	var v uint32
	if pGetLibraryVersion != nil {
		pGetLibraryVersion.Call(uintptr(unsafe.Pointer(&v)))
	}
	return uint8(v >> 16), uint8(v >> 8), uint8(v)
}

func d2xxCreateDeviceInfoList() (int, int) {
	var num uint32
	r1, _, _ := pCreateDeviceInfoList.Call(uintptr(unsafe.Pointer(&num)))
	return int(num), int(r1)
}

// Device functions.

func d2xxOpen(i int) (handle, int) {
	var h handle
	r1, _, _ := pOpen.Call(uintptr(i), uintptr(unsafe.Pointer(&h)))
	return h, int(r1)
}

func (h handle) d2xxClose() int {
	r1, _, _ := pClose.Call(h.toH())
	return int(r1)
}

func (h handle) d2xxResetDevice() int {
	r1, _, _ := pResetDevice.Call(h.toH())
	return int(r1)
}

func (h handle) d2xxGetDeviceInfo() (devType, uint16, uint16, int) {
	var d devType
	var id uint32
	if r1, _, _ := pGetDeviceInfo.Call(h.toH(), uintptr(unsafe.Pointer(&d)), uintptr(unsafe.Pointer(&id)), 0, 0, 0); r1 != 0 {
		return unknown, 0, 0, int(r1)
	}
	return devType(d), uint16(id >> 16), uint16(id), 0
}

func (h handle) d2xxEEPROMRead(d *device) int {
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
	if r1, _, _ := pEEPROMRead.Call(h.toH(), uintptr(eepromVoid), uintptr(len(d.eeprom)), m, mi, de, s); r1 != 0 {
		return int(r1)
	}

	d.manufacturer = toStr(manufacturer[:])
	d.manufacturerID = toStr(manufacturerID[:])
	d.desc = toStr(desc[:])
	d.serial = toStr(serial[:])
	return 0
}

func (h handle) d2xxSetChars(eventChar byte, eventEn bool, errorChar byte, errorEn bool) int {
	v := uintptr(0)
	if eventEn {
		v = 1
	}
	w := uintptr(0)
	if errorEn {
		w = 1
	}
	r1, _, _ := pSetChars.Call(h.toH(), uintptr(eventChar), v, uintptr(errorChar), w)
	return int(r1)
}

func (h handle) d2xxSetUSBParameters(in, out int) int {
	r1, _, _ := pSetUSBParameters.Call(h.toH(), uintptr(in), uintptr(out))
	return int(r1)
}

func (h handle) d2xxSetFlowControl() int {
	// FT_FLOW_RTS_CTS
	r1, _, _ := pSetFlowControl.Call(h.toH(), 0x0100, 0, 0)
	return int(r1)
}

func (h handle) d2xxSetTimeouts(readMS, writeMS int) int {
	r1, _, _ := pSetTimeouts.Call(h.toH(), uintptr(readMS), uintptr(writeMS))
	return int(r1)
}

func (h handle) d2xxSetLatencyTimer(delayMS uint8) int {
	r1, _, _ := pSetLatencyTimer.Call(h.toH(), uintptr(delayMS))
	return int(r1)
}

func (h handle) d2xxGetQueueStatus() (uint32, int) {
	var v uint32
	r1, _, _ := pGetQueueStatus.Call(h.toH(), uintptr(unsafe.Pointer(&v)))
	return v, int(r1)
}

func (h handle) d2xxRead(b []byte) (int, int) {
	var bytesRead uint32
	r1, _, _ := pRead.Call(h.toH(), uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), uintptr(unsafe.Pointer(&bytesRead)))
	return 0, int(r1)
}

func (h handle) d2xxWrite(b []byte) (int, int) {
	var bytesSent uint32
	r1, _, _ := pWrite.Call(h.toH(), uintptr(unsafe.Pointer(&b[0])), uintptr(len(b)), uintptr(unsafe.Pointer(&bytesSent)))
	return 0, int(r1)
}

func (h handle) d2xxGetBitMode() (byte, int) {
	var s uint8
	r1, _, _ := pGetBitMode.Call(h.toH(), uintptr(unsafe.Pointer(&s)))
	return s, int(r1)
}

func (h handle) d2xxSetBitMode(mask, mode byte) int {
	r1, _, _ := pSetBitMode.Call(h.toH(), uintptr(mask), uintptr(mode))
	return int(r1)
}

func (h handle) toH() uintptr {
	return uintptr(h)
}

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
	pSetFlowControl       *syscall.Proc
	pSetLatencyTimer      *syscall.Proc
	pSetTimeouts          *syscall.Proc
	pSetUSBParameters     *syscall.Proc
	pWrite                *syscall.Proc
)

func init() {
	if dll, _ := syscall.LoadDLL("ftd2xx.dll"); dll != nil {
		// If any function is not found, disable the support.
		disabled = false
		find := func(n string) *syscall.Proc {
			s, _ := dll.FindProc(n)
			if s == nil {
				disabled = true
			}
			return s
		}
		pClose = find("FT_Close")
		pCreateDeviceInfoList = find("FT_CreateDeviceInfoList")
		pEEPROMRead = find("FT_EEPROM_Read")
		pGetBitMode = find("FT_GetBitMode")
		pGetDeviceInfo = find("FT_GetDeviceInfo")
		pGetLibraryVersion = find("FT_GetLibraryVersion")
		pGetQueueStatus = find("FT_GetQueueStatus")
		pOpen = find("FT_Open")
		pRead = find("FT_Read")
		pResetDevice = find("FT_ResetDevice")
		pSetBitMode = find("FT_SetBitMode")
		pSetChars = find("FT_SetChars")
		pSetFlowControl = find("FT_SetFlowControl")
		pSetLatencyTimer = find("FT_SetLatencyTimer")
		pSetTimeouts = find("FT_SetTimeouts")
		pSetUSBParameters = find("FT_SetUSBParameters")
		pWrite = find("FT_Write")
	}
}

func toStr(c []byte) string {
	i := bytes.IndexByte(c, 0)
	if i != -1 {
		return string(c[:i])
	}
	return string(c)
}
