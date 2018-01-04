// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

import (
	"bytes"
	"reflect"
	"syscall"
	"unsafe"
)

var (
	dll                   *syscall.DLL
	pOpenEx               *syscall.Proc
	pClose                *syscall.Proc
	pGetDeviceInfo        *syscall.Proc
	pCreateDeviceInfoList *syscall.Proc
	pGetDeviceInfoList    *syscall.Proc
)

func init() {
	if dll, _ = syscall.LoadDLL("ftd2xx.dll"); dll != nil {
		pOpenEx, _ = dll.FindProc("FT_OpenEx")
		pClose, _ = dll.FindProc("FT_Close")
		pGetDeviceInfo, _ = dll.FindProc("FT_GetDeviceInfo")
		pCreateDeviceInfoList, _ = dll.FindProc("FT_CreateDeviceInfoList")
		pGetDeviceInfoList, _ = dll.FindProc("FT_GetDeviceInfoList")
	}
}

func openEx(arg1 uintptr, flags uint32) (Handle, error) {
	var h Handle
	if pOpenEx == nil {
		return h, MissingErr
	}
	r1, _, _ := pOpenEx.Call(arg1, uintptr(flags), uintptr(unsafe.Pointer(&h)))
	return h, toErr(int(r1))
}

func closeHandle(h Handle) error {
	if pClose == nil {
		return MissingErr
	}
	r1, _, _ := pClose.Call(uintptr(h))
	return toErr(int(r1))
}

func getDeviceInfo(h Handle, i *DevInfo) error {
	if pGetDeviceInfo == nil {
		return MissingErr
	}
	var dev uint32
	var id uint32
	var serial [16]byte
	var desc [64]byte
	r1, _, _ := pGetDeviceInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&dev)), uintptr(unsafe.Pointer(&id)), uintptr(unsafe.Pointer(&serial[0])), uintptr(unsafe.Pointer(&desc[0])), 0)
	if r1 == 0 {
		i.Type = Type(dev)
		i.ID = uint32(id)
		i.Serial = toStr(serial[:])
		i.Desc = toStr(desc[:])
	}
	return toErr(int(r1))
}

func createDeviceInfoList() (int, error) {
	if pCreateDeviceInfoList == nil {
		return 0, MissingErr
	}
	var num uint32
	r1, _, _ := pCreateDeviceInfoList.Call(uintptr(unsafe.Pointer(&num)))
	return int(num), toErr(int(r1))
}

func getDeviceInfoList(num int) ([]DevInfo, error) {
	if pGetDeviceInfoList == nil {
		return nil, MissingErr
	}
	b := make([]byte, deviceListInfoNodeSize*num)
	var actual uint32
	r1, _, _ := pGetDeviceInfoList.Call(uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&actual)))
	var out []DevInfo
	if r1 == 0 {
		l := ((*[256]deviceListInfoNode)(unsafe.Pointer(&b[0])))[:num]
		out = make([]DevInfo, 0, num)
		for _, v := range l {
			d := DevInfo{
				Type:   Type(v.Type),
				ID:     uint32(v.ID),
				LocID:  uint32(v.LocId),
				Serial: toStr(v.SerialNumber[:]),
				Desc:   toStr(v.Description[:]),
				h:      Handle(v.ftHandle),
			}
			if v.Flags&ftFlagsOpened != 0 {
				d.Opened = true
			}
			if v.Flags&ftFlagsHispeed != 0 {
				d.HiSpeed = true
			}
			out = append(out, d)
		}
	}
	return out, toErr(int(r1))
}

//

func toStr(c []byte) string {
	i := bytes.IndexByte(c, 0)
	if i != -1 {
		return string(c[:i])
	}
	return string(c)
}

const ftFlagsOpened = 1
const ftFlagsHispeed = 2

type deviceListInfoNode struct {
	Flags        uint32
	Type         uint32
	ID           uint32
	LocId        uint32
	SerialNumber [16]byte
	Description  [64]byte
	ftHandle     uintptr
}

var deviceListInfoNodeSize = int(reflect.TypeOf((*deviceListInfoNode)(nil)).Elem().Size())
