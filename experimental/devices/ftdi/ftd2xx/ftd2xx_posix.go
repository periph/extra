// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !windows

package ftd2xx

/*
#include "ftd2xx.h"
*/
import "C"

func openEx(arg1 uintptr, flags uint32) (Handle, error) {
	var h C.FT_HANDLE
	e := C.FT_OpenEx(C.PVOID(arg1), C.DWORD(flags), &h)
	if uintptr(h) == 0 && e == 0 {
		panic("unexpected")
	}
	return Handle(h), toErr(int(e))
}

func closeHandle(h Handle) error {
	e := C.FT_Close(C.FT_HANDLE(h))
	return toErr(int(e))
}

func getDeviceInfo(h Handle, i *DevInfo) error {
	var dev C.FT_DEVICE
	var id C.DWORD
	var serial [16]C.char
	var desc [64]C.char
	e := C.FT_GetDeviceInfo(C.FT_HANDLE(h), &dev, &id, (*C.char)(&serial[0]), (*C.char)(&desc[0]), nil)
	if e == 0 {
		i.Type = Type(dev)
		i.ID = uint32(id)
		i.Serial = C.GoString(&serial[0])
		i.Desc = C.GoString(&desc[0])
	}
	return toErr(int(e))
}

func createDeviceInfoList() (int, error) {
	var num C.DWORD
	e := C.FT_CreateDeviceInfoList(&num)
	return int(num), toErr(int(e))
}

func getDeviceInfoList(num int) ([]DevInfo, error) {
	l := make([]C.FT_DEVICE_LIST_INFO_NODE, num)
	n := C.DWORD(num)
	e := C.FT_GetDeviceInfoList(&l[0], &n)
	var out []DevInfo
	if e == 0 {
		out = make([]DevInfo, 0, num)
		for _, v := range l {
			d := DevInfo{
				Type:   Type(v.Type),
				ID:     uint32(v.ID),
				LocID:  uint32(v.LocId),
				Serial: C.GoString(&v.SerialNumber[0]),
				Desc:   C.GoString(&v.Description[0]),
				h:      Handle(v.ftHandle),
			}
			if v.Flags&C.FT_FLAGS_OPENED != 0 {
				d.Opened = true
			}
			if v.Flags&C.FT_FLAGS_HISPEED != 0 {
				d.HiSpeed = true
			}
			out = append(out, d)
		}
	}
	return out, toErr(int(e))
}
