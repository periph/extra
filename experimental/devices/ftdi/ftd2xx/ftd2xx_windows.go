// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

import (
	"bytes"
	"syscall"
	"unsafe"

	"periph.io/x/extra/experimental/devices/ftdi"
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

/*
func getDeviceInfoList(num int) ([]ftdi.Info, int) {
	if pGetDeviceInfoList == nil {
		return nil, missing
	}
	b := make([]byte, deviceListInfoNodeSize*num)
	var actual uint32
	r1, _, _ := pGetDeviceInfoList.Call(uintptr(unsafe.Pointer(&b[0])), uintptr(unsafe.Pointer(&actual)))
	var out []ftdi.Info
	if r1 == 0 {
		l := ((*[256]deviceListInfoNode)(unsafe.Pointer(&b[0])))[:num]
		out = make([]ftdi.Info, 0, num)
		for _, v := range l {
			d := DevInfo{
				Type:   devType(v.Type).String(),
				ID:     uint32(v.ID),
				LocID:  uint32(v.LocId),
				Serial: toStr(v.SerialNumber[:]),
				Desc:   toStr(v.Description[:]),
				h:      handle(v.ftHandle),
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
	return out, int(r1)
}
*/

// Device functions.

func open(i int) (handle, int) {
	var h handle
	if pOpen == nil {
		return h, missing
	}
	r1, _, _ := pOpen.Call(uintptr(i), uintptr(unsafe.Pointer(&h)))
	return h, int(r1)
}

func closeHandle(h handle) int {
	if pClose == nil {
		return missing
	}
	r1, _, _ := pClose.Call(uintptr(h))
	return int(r1)
}

func getInfo(h handle, i *ftdi.Info) int {
	if pGetDeviceInfo == nil || pEEPROMRead == nil {
		return missing
	}
	var dev uint32
	var id uint32
	if r1, _, _ := pGetDeviceInfo.Call(uintptr(h), uintptr(unsafe.Pointer(&dev)), uintptr(unsafe.Pointer(&id)), 0, 0, 0); r1 != 0 {
		return int(r1)
	}
	i.Opened = true
	i.Type = devType(dev).String()
	i.VenID = uint16(id >> 16)
	i.ProductID = uint16(id)

	var manufacturer [64]byte
	var manufacturerID [64]byte
	var desc [64]byte
	var serial [64]byte
	eeprom := make([]byte, eeprom_generic_size)
	eepromVoid := unsafe.Pointer(&eeprom[0])
	hdr := (*eeprom_header)(eepromVoid)
	// It must be set here, while it must not be set on Linux. Probably a
	// difference between v1 and v2.
	hdr.deviceType = dev
	if r1, _, _ := pEEPROMRead.Call(uintptr(h), uintptr(eepromVoid), uintptr(len(eeprom)), uintptr(unsafe.Pointer(&manufacturer[0])), uintptr(unsafe.Pointer(&manufacturerID[0])), uintptr(unsafe.Pointer(&desc[0])), uintptr(unsafe.Pointer(&serial[0]))); r1 != 0 {
		return int(r1)
	}
	i.MaxPower = uint16(hdr.MaxPower)
	i.SelfPowered = hdr.SelfPowered != 0
	i.RemoteWakeup = hdr.RemoteWakeup != 0
	i.PullDownEnable = hdr.PullDownEnable != 0

	switch devType(dev) {
	case ft232H:
		h := (*eeprom_ft232h)(eepromVoid)
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

	i.Manufacturer = toStr(manufacturer[:])
	i.ManufacturerID = toStr(manufacturerID[:])
	i.Desc = toStr(desc[:])
	i.Serial = toStr(serial[:])
	return 0
}

//

var (
	// Library functions.
	pGetLibraryVersion    *syscall.Proc
	pCreateDeviceInfoList *syscall.Proc
	//pGetDeviceInfoList    *syscall.Proc

	// Device functions.
	pOpen          *syscall.Proc
	pClose         *syscall.Proc
	pGetDeviceInfo *syscall.Proc
	pEEPROMRead    *syscall.Proc
)

type eeprom_generic struct {
	i int
}

// This data was determined by glancing at the disassembled code.
//
// It must not be any other value, like 56 used on posix.
var eeprom_generic_size = 44

// eeprom_header
type eeprom_header struct {
	deviceType     uint32 // FTxxxx device type to be programmed
	VendorID       uint16 // 0x0403
	ProductID      uint16 // 0x6001
	SerNumEnable   uint8  // non-zero if serial number to be used
	MaxPower       uint16 // 0 < MaxPower <= 500
	SelfPowered    uint8  // 0 = bus powered, 1 = self powered
	RemoteWakeup   uint8  // 0 = not capable, 1 = capable
	PullDownEnable uint8  //
}

type eeprom_ft232h struct {
	// eeprom_header
	deviceType     uint32 // FTxxxx device type to be programmed
	VendorID       uint16 // 0x0403
	ProductID      uint16 // 0x6001
	SerNumEnable   uint8  // non-zero if serial number to be used
	MaxPower       uint16 // 0 < MaxPower <= 500
	SelfPowered    uint8  // 0 = bus powered, 1 = self powered
	RemoteWakeup   uint8  // 0 = not capable, 1 = capable
	PullDownEnable uint8  //

	// ft232h specific.
	ACSlowSlew        uint8 // AC bus pins have slow slew
	ACSchmittInput    uint8 // AC bus pins are Schmitt input
	ACDriveCurrent    uint8 // valid values are 4mA, 8mA, 12mA, 16mA
	ADSlowSlew        uint8 // non-zero if AD bus pins have slow slew
	ADSchmittInput    uint8 // non-zero if AD bus pins are Schmitt input
	ADDriveCurrent    uint8 // valid values are 4mA, 8mA, 12mA, 16mA
	Cbus0             uint8 // Cbus Mux control
	Cbus1             uint8 // Cbus Mux control
	Cbus2             uint8 // Cbus Mux control
	Cbus3             uint8 // Cbus Mux control
	Cbus4             uint8 // Cbus Mux control
	Cbus5             uint8 // Cbus Mux control
	Cbus6             uint8 // Cbus Mux control
	Cbus7             uint8 // Cbus Mux control
	Cbus8             uint8 // Cbus Mux control
	Cbus9             uint8 // Cbus Mux control
	FT1248Cpol        uint8 // FT1248 clock polarity - clock idle high (true) or clock idle low (false)
	FT1248Lsb         uint8 // FT1248 data is LSB (true), or MSB (false)
	FT1248FlowControl uint8 // FT1248 flow control enable
	IsFifo            uint8 // Interface is 245 FIFO
	IsFifoTar         uint8 // Interface is 245 FIFO CPU target
	IsFastSer         uint8 // Interface is Fast serial
	IsFT1248          uint8 // Interface is FT1248
	PowerSaveEnable   uint8 //
	DriverType        uint8 //
}

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
	}
}

func toStr(c []byte) string {
	i := bytes.IndexByte(c, 0)
	if i != -1 {
		return string(c[:i])
	}
	return string(c)
}

/*
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
*/
