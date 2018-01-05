// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !cgo
// +build !windows

package ftd2xx

import "periph.io/x/extra/experimental/devices/ftdi"

// Library functions.

func getLibraryVersion() (uint8, uint8, uint8) {
	return 0, 0, 0
}

func createDeviceInfoList() (int, int) {
	return 0, noCGO
}

/*
func getDeviceInfoList(num int) ([]ftdi.Info, int) {
	return nil, noCGO
}
*/

// Device functions.

func open(i int) (handle, int) {
	return 0, noCGO
}

func closeHandle(h handle) int {
	return noCGO
}

func getInfo(h handle, i *ftdi.Info) int {
	return noCGO
}
