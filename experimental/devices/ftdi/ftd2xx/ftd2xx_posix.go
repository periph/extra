// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build !windows

package ftd2xx

/*
#include "ftd2xx.h"

typedef union EEPROM_GENERIC {
	FT_EEPROM_HEADER   common;
	FT_EEPROM_232B     b;
	FT_EEPROM_232R     r;
	FT_EEPROM_232H     singleH;
	FT_EEPROM_2232     dual;
	FT_EEPROM_2232H    dualH;
	FT_EEPROM_4232H    quadH;
	FT_EEPROM_X_SERIES x;
} EEPROM_GENERIC;

// Is the largest of each header (56)....
DWORD EEPROM_MAX_SIZE = sizeof(EEPROM_GENERIC);
*/
import "C"
import (
	"unsafe"

	"periph.io/x/extra/experimental/devices/ftdi"
)

// Library functions.

func getLibraryVersion() (uint8, uint8, uint8) {
	var v C.DWORD
	C.FT_GetLibraryVersion(&v)
	return uint8(v >> 16), uint8(v >> 8), uint8(v)
}

func createDeviceInfoList() (int, int) {
	var num C.DWORD
	e := C.FT_CreateDeviceInfoList(&num)
	return int(num), int(e)
}

/*
func getDeviceInfoList(num int) ([]ftdi.Info, int) {
	l := make([]C.FT_DEVICE_LIST_INFO_NODE, num)
	n := C.DWORD(num)
	e := C.FT_GetDeviceInfoList(&l[0], &n)
	var out []DevInfo
	if e == 0 {
		out = make([]ftdi.Info, 0, num)
		for _, v := range l {
			d := DevInfo{
				Type:   devType(v.Type).String(),
				VenID:   uint16(v.ID >> 16),
				ProductID: uint16(v.ID),
				LocID:  uint32(v.LocId),
				Serial: C.GoString(&v.SerialNumber[0]),
				Desc:   C.GoString(&v.Description[0]),
				h:      handle(v.ftHandle),
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
	return out, int(e)
}
*/

// Device functions.

func open(i int) (handle, int) {
	var h C.FT_HANDLE
	e := C.FT_Open(C.int(i), &h)
	if uintptr(h) == 0 && e == 0 {
		panic("unexpected")
	}
	return handle(h), int(e)
}

func closeHandle(h handle) int {
	e := C.FT_Close(C.FT_HANDLE(h))
	return int(e)
}

func getInfo(h handle, i *ftdi.Info) int {
	var dev C.FT_DEVICE
	var id C.DWORD
	// TODO(maruel): When specifying serial or desc, the function fails. It's not
	// really important because we read the EEPROM instead.
	if e := C.FT_GetDeviceInfo(C.FT_HANDLE(h), &dev, &id, nil, nil, nil); e != 0 {
		return int(e)
	}
	i.Opened = true
	i.Type = devType(dev).String()
	i.VenID = uint16(id >> 16)
	i.ProductID = uint16(id)

	var manufacturer [64]C.char
	var manufacturerID [64]C.char
	var desc [64]C.char
	var serial [64]C.char
	eeprom := make([]byte, int(C.EEPROM_MAX_SIZE))
	eepromVoid := unsafe.Pointer(&eeprom[0])
	hdr := (*C.FT_EEPROM_HEADER)(eepromVoid)
	// It must not be set here, while it must be set on Windows. Probably a
	// difference between v1 and v2.
	//hdr.deviceType = dev
	if e := C.FT_EEPROM_Read(C.FT_HANDLE(h), eepromVoid, C.DWORD(len(eeprom)), &manufacturer[0], &manufacturerID[0], &desc[0], &serial[0]); e != 0 {
		return int(e)
	}
	i.MaxPower = uint16(hdr.MaxPower)
	i.SelfPowered = hdr.SelfPowered != 0
	i.RemoteWakeup = hdr.RemoteWakeup != 0
	i.PullDownEnable = hdr.PullDownEnable != 0

	switch devType(dev) {
	case ft232H:
		// TODO(maruel): Everything is empty, even when using
		// examples/EEPROM/read/eeprom-read.c.
		h := (*C.FT_EEPROM_232H)(eepromVoid)
		i.CSlowSlew = h.ACSlowSlew != 0
		i.CSchmittInput = h.ACSchmittInput != 0
		i.CDriveCurrent = uint8(h.ACDriveCurrent)
		i.DSlowSlew = h.ADSlowSlew != 0
		i.DSchmittInput = h.ADSchmittInput != 0
		i.DDriveCurrent = uint8(h.ADDriveCurrent)
		i.Cbus0 = uint8(h.Cbus0)
		i.Cbus1 = uint8(h.Cbus1)
		i.Cbus2 = uint8(h.Cbus2)
		i.Cbus3 = uint8(h.Cbus3)
		i.Cbus4 = uint8(h.Cbus4)
		i.Cbus5 = uint8(h.Cbus5)
		i.Cbus6 = uint8(h.Cbus6)
		i.Cbus7 = uint8(h.Cbus7)
		i.Cbus8 = uint8(h.Cbus8)
		i.Cbus9 = uint8(h.Cbus9)
		i.FT1248Cpol = h.FT1248Cpol != 0
		i.FT1248Lsb = h.FT1248Lsb != 0
		i.FT1248FlowControl = h.FT1248FlowControl != 0
		i.IsFifo = h.IsFifo != 0
		i.IsFifoTar = h.IsFifoTar != 0
		i.IsFastSer = h.IsFastSer != 0
		i.IsFT1248 = h.IsFT1248 != 0
		i.PowerSaveEnable = h.PowerSaveEnable != 0
		i.DriverType = uint8(h.DriverType)
	default:
	}

	i.EEPROM = eeprom

	i.Manufacturer = C.GoString(&manufacturer[0])
	i.ManufacturerID = C.GoString(&manufacturerID[0])
	i.Desc = C.GoString(&desc[0])
	i.Serial = C.GoString(&serial[0])
	return 0
}
