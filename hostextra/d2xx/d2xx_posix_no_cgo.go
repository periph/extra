// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !cgo
// +build !windows

package d2xx

const disabled = true

// Library functions.

func d2xxGetLibraryVersion() (uint8, uint8, uint8) {
	return 0, 0, 0
}

func d2xxCreateDeviceInfoList() (int, int) {
	return 0, noCGO
}

// Device functions.

func d2xxOpen(i int) (*device, int) {
	return nil, noCGO
}

func (d *device) d2xxClose() int {
	return noCGO
}

func (d *device) d2xxResetDevice() int {
	return noCGO
}

func (d *device) getInfo() int {
	return noCGO
}

func (d *device) setup() int {
	return noCGO
}

func (d *device) d2xxGetQueueStatus() (uint32, int) {
	return 0, noCGO
}

func (d *device) d2xxRead(b []byte) (int, int) {
	return 0, noCGO
}

func (d *device) d2xxWrite(b []byte) (int, int) {
	return 0, noCGO
}

func (d *device) d2xxGetBitMode() (byte, int) {
	return 0, noCGO
}

func (d *device) d2xxSetBitMode(mask, mode byte) int {
	return noCGO
}

type handle uintptr
