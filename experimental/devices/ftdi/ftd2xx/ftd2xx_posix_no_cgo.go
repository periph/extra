// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !cgo
// +build !windows

package ftd2xx

// Library functions.

func getLibraryVersion() (uint8, uint8, uint8) {
	return 0, 0, 0
}

func createDeviceInfoList() (int, int) {
	return 0, noCGO
}

// Device functions.

func open(i int) (*device, int) {
	return nil, noCGO
}

func (d *device) closeHandle() int {
	return noCGO
}

func (d *device) resetDevice() int {
	return noCGO
}

func (d *device) getInfo() int {
	return noCGO
}

func (d *device) getReadPending() (int, int) {
	return 0, noCGO
}

func (d *device) doRead(b []byte) (int, int) {
	return 0, noCGO
}

func (d *device) getBits() (byte, int) {
	return 0, noCGO
}

type handle uintptr
