// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !cgo
// +build !windows

package d2xx

import "periph.io/x/extra/hostextra/d2xx/ftdi"

const disabled = true

// Library functions.

func d2xxGetLibraryVersion() (uint8, uint8, uint8) {
	return 0, 0, 0
}

func d2xxCreateDeviceInfoList() (int, int) {
	return 0, noCGO
}

// Device functions.

func d2xxOpen(i int) (handle, int) {
	return 0, noCGO
}

func (h handle) d2xxClose() int {
	return noCGO
}

func (h handle) d2xxResetDevice() int {
	return noCGO
}

func (h handle) d2xxGetDeviceInfo() (ftdi.DevType, uint16, uint16, int) {
	return ftdi.Unknown, 0, 0, noCGO
}

func (h handle) d2xxEEPROMRead(t ftdi.DevType, ee *ftdi.EEPROM) int {
	return noCGO
}

func (h handle) d2xxEEPROMProgram(e *ftdi.EEPROM) int {
	return noCGO
}

func (h handle) d2xxEraseEE() int {
	return noCGO
}

func (h handle) d2xxWriteEE(offset uint8, value uint16) int {
	return noCGO
}

func (h handle) d2xxEEUASize() (int, int) {
	return 0, noCGO
}

func (h handle) d2xxEEUARead(ua []byte) int {
	return noCGO
}

func (h handle) d2xxEEUAWrite(ua []byte) int {
	return noCGO
}

func (h handle) d2xxSetChars(eventChar byte, eventEn bool, errorChar byte, errorEn bool) int {
	return noCGO
}

func (h handle) d2xxSetUSBParameters(in, out int) int {
	return noCGO
}

func (h handle) d2xxSetFlowControl() int {
	return noCGO
}

func (h handle) d2xxSetTimeouts(readMS, writeMS int) int {
	return noCGO
}

func (h handle) d2xxSetLatencyTimer(delayMS uint8) int {
	return noCGO
}

func (h handle) d2xxSetBaudRate(hz uint32) int {
	return noCGO
}

func (h handle) d2xxGetQueueStatus() (uint32, int) {
	return 0, noCGO
}

func (h handle) d2xxRead(b []byte) (int, int) {
	return 0, noCGO
}

func (h handle) d2xxWrite(b []byte) (int, int) {
	return 0, noCGO
}

func (h handle) d2xxGetBitMode() (byte, int) {
	return 0, noCGO
}

func (h handle) d2xxSetBitMode(mask, mode byte) int {
	return noCGO
}
