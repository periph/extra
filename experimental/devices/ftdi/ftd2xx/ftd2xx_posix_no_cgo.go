// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !cgo
// +build !windows

package ftd2xx

import (
	"errors"
)

func openEx(arg1 uintptr, flags uint32) (Handle, error) {
	return 0, errors.New("ftd2xx: can't be used without cgo")
}

func closeHandle(h Handle) error {
	return errors.New("ftd2xx: can't be used without cgo")
}

func getDeviceInfo(h Handle, i *DevInfo) error {
	return errors.New("ftd2xx: can't be used without cgo")
}

func createDeviceInfoList() (int, error) {
	return 0, errors.New("ftd2xx: can't be used without cgo")
}

func getDeviceInfoList(num int) ([]DevInfo, error) {
	return nil, errors.New("ftd2xx: can't be used without cgo")
}
