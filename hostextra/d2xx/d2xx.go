// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This file is the abstraction layer against the various OS specific
// implementations.
//
// It converts the int error value into error type.
//
// D2XX programmer's guide; Explains how to use the DLL provided by ftdi.
// http://www.ftdichip.com/Support/Documents/ProgramGuides/D2XX_Programmer's_Guide(FT_000071).pdf
//
// D2XX samples; http://www.ftdichip.com/Support/SoftwareExamples/CodeExamples/VC.htm
//
// There is multiple ways to access a FT232H:
//
// - Some operating systems include a limited "serial port only" driver.
// - Future Technologic Devices International Ltd provides their own private
//   source driver.
// - FTDI also provides a "serial port only" driver surnamed VCP.
// - https://www.intra2net.com/en/developer/libftdi/ is an open source driver,
//   that is acknowledged by FTDI.

package d2xx

import (
	"errors"
	"strconv"
	"unsafe"
)

// Version returns the version number of the D2xx driver currently used.
func Version() (uint8, uint8, uint8) {
	return d2xxGetLibraryVersion()
}

//

func numDevices() (int, error) {
	num, e := d2xxCreateDeviceInfoList()
	if e != 0 {
		return 0, toErr("GetNumDevices initialization failed", e)
	}
	return num, nil
}

//

func openDev(i int) (*device, error) {
	// TODO(maruel): The handle is leaked in failure paths.
	h, e := d2xxOpen(i)
	d := &device{h: h}
	if e != 0 {
		return d, toErr("Open", e)
	}
	if d.t, d.venID, d.devID, e = h.d2xxGetDeviceInfo(); e != 0 {
		return d, toErr("GetDeviceInfo", e)
	}

	// Under the hood, it calls both FT_GetDeviceInfo and FT_EEPROM_READ.
	if e := h.d2xxEEPROMRead(d); e != 0 {
		return nil, toErr("EEPROMRead", e)
	}

	// Sets up USB parameters.
	if err := d.setup(); err != nil {
		return nil, err
	}
	if err := d.flushPending(); err != nil {
		return nil, err
	}

	// Reset mode to setting in EEPROM.
	// TODO(maruel): Eventually we may want to read the state and expose it
	// instead, to not cause unwanted glitches.
	if err := d.setBitMode(0, 0); err != nil {
		return nil, nil
	}
	switch d.t {
	case ft232H, ft2232H, ft4232H: // ft2232
		if err := d.setupMPSSE(); err != nil {
			return nil, err
		}
	case ft232R:
		// Asynchronous bitbang
		if err := d.setBitMode(0, 1); err != nil {
			return nil, err
		}
	default:
	}
	return d, nil
}

// device is the low level d2xx device handle.
type device struct {
	h              handle
	t              devType
	venID          uint16
	devID          uint16
	manufacturer   string
	manufacturerID string
	desc           string
	serial         string
	eeprom         []byte
	isMPSSE        bool // if false, uses CBus bitbang
}

func (d *device) closeDev() error {
	// Not yet called.
	return toErr("Close", d.h.d2xxClose())
}

func (d *device) getI(i *Info) {
	i.Type = d.t.String()
	i.VenID = d.venID
	i.DevID = d.devID
	i.Manufacturer = d.manufacturer
	i.ManufacturerID = d.manufacturerID
	i.Desc = d.desc
	i.Serial = d.serial
	if len(d.eeprom) > 0 {
		// Only consider the device "good" if we could read the EEPROM.
		i.Opened = true
		i.EEPROM = make([]byte, len(d.eeprom))
		copy(i.EEPROM, d.eeprom)

		// Use the custom structs instead of the ones provided by the library. The
		// reason is that it had to be written for Windows anyway, and this enables
		// using a single code path everywhere.
		hdr := (*eepromHeader)(unsafe.Pointer(&d.eeprom[0]))
		i.MaxPower = uint16(hdr.MaxPower)
		i.SelfPowered = hdr.SelfPowered != 0
		i.RemoteWakeup = hdr.RemoteWakeup != 0
		i.PullDownEnable = hdr.PullDownEnable != 0
		switch d.t {
		case ft232H:
			h := (*eepromFt232h)(unsafe.Pointer(&d.eeprom[0]))
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
		case ft232R:
			h := (*eepromFt232r)(unsafe.Pointer(&d.eeprom[0]))
			i.IsHighCurrent = h.IsHighCurrent != 0
			i.UseExtOsc = h.UseExtOsc != 0
			i.InvertTXD = h.InvertTXD != 0
			i.InvertRXD = h.InvertRXD != 0
			i.InvertRTS = h.InvertRTS != 0
			i.InvertCTS = h.InvertCTS != 0
			i.InvertDTR = h.InvertDTR != 0
			i.InvertDSR = h.InvertDSR != 0
			i.InvertDCD = h.InvertDCD != 0
			i.InvertRI = h.InvertRI != 0
			i.Cbus0 = uint8(h.Cbus0)
			i.Cbus1 = uint8(h.Cbus1)
			i.Cbus2 = uint8(h.Cbus2)
			i.Cbus3 = uint8(h.Cbus3)
			i.Cbus4 = uint8(h.Cbus4)
			i.DriverType = uint8(h.DriverType)
		default:
			// TODO(maruel): Implement me!
		}
	}
}

func (d *device) setup() error {
	// Disable event/error characters.
	if e := d.h.d2xxSetChars(0, false, 0, false); e != 0 {
		return toErr("SetChars", e)
	}
	// Set I/O timeouts to 5 sec.
	if e := d.h.d2xxSetTimeouts(5000, 5000); e != 0 {
		return toErr("SetTimeouts", e)
	}
	// Latency timer at default 16ms.
	return toErr("SetLatencyTimer", d.h.d2xxSetLatencyTimer(16))
}

func (d *device) reset() error {
	return toErr("Reset", d.h.d2xxResetDevice())
}

func (d *device) getBitMode() (byte, error) {
	l, e := d.h.d2xxGetBitMode()
	if e != 0 {
		return 0, toErr("GetBitMode", e)
	}
	return l, nil
}

// setBitMode change the mode of operation of the device.
//
// mask sets which pins are inputs and outputs.
//
// mode can be:
//  0x00 Reset
//  0x01 Asynchronous bit bang
//  0x02 MPSSE (ft232h, ft2232, ft2232h, ft4232h)
//  0x04 Synchronous bit bang (ft232h, ft232r, ft245r, ft2232, ft2232h, ft4232h)
//  0x08 MCU host bus emulation mode (ft232h, ft2232, ft2232h, ft4232h)
//  0x10 Fast opto-isolated serial mode (ft232h, ft2232, ft2232h, ft4232h)
//  0x20 CBus bit bang mode (ft232h and ft232r)
//  0x40 Single channel synchrnous 245 fifo mode (ft232h and ft2232h)
func (d *device) setBitMode(mask, mode byte) error {
	return toErr("SetBitMode", d.h.d2xxSetBitMode(mask, mode))
}

func (d *device) flushPending() error {
	p, e := d.h.d2xxGetQueueStatus()
	if p == 0 || e != 0 {
		return toErr("FlushPending", e)
	}
	_, e = d.h.d2xxRead(make([]byte, p))
	return toErr("FlushPending", e)
}

func (d *device) read(b []byte) (int, error) {
	p, e := d.h.d2xxGetQueueStatus()
	if p == 0 || e != 0 {
		return int(p), toErr("Read", e)
	}
	v := int(p)
	if v > len(b) {
		v = len(b)
	}
	n, e := d.h.d2xxRead(b[:v])
	return n, toErr("Read", e)
}

func (d *device) write(b []byte) (int, error) {
	// Use a stronger guarantee that all bytes have been written.
	n, e := d.h.d2xxWrite(b)
	return n, toErr("Write", e)
}

// devType is the FTDI device type.
type devType uint32

const (
	ftBM devType = iota
	ftAM
	ft100AX
	unknown
	ft2232C
	ft232R
	ft2232H
	ft4232H
	ft232H
	ftXSeries
	ft4222H0
	ft4222H1_2
	ft4222H3
	ft4222Prog
	ft900
	ft930
	ftUMFTPD3A
)

func (d devType) String() string {
	switch d {
	case ftBM:
		return "ftbm"
	case ftAM:
		return "ftam"
	case ft100AX:
		return "ft100ax"
	case ft2232C:
		return "ft2232c"
	case ft232R:
		return "ft232r"
	case ft2232H:
		return "ft2232h"
	case ft4232H:
		return "ft4232h"
	case ft232H:
		return "ft232h"
	case ftXSeries:
		return "ft2NNx"
	case ft4222H0:
		return "ft4222h 0"
	case ft4222H1_2:
		return "ft4222h 1 or 2"
	case ft4222H3:
		return "ft4222h 3"
	case ft4222Prog:
		return "ft4222 prog"
	default:
		return "unknown"
	}
}

func (d devType) eepromSize() int {
	// This data was determined by tracing with a debugger.
	//
	// It must not be any other value, like 56 used on posix. ¯\_(ツ)_/¯
	switch d {
	case ft232H:
		return 44
	case ft232R:
		return 32
	default:
		// TODO(maruel): Figure out.
		return 56
	}
}

// TODO(maruel): To add:
// - FT_IoCtl
// UART:
// - FT_SetBaudRate
// - FT_SetDivisor
// - FT_SetDataCharacteristics
// - FT_SetFlowControl
// - FT_SetDtr / FT_ClrDtr / FT_SetRts / FT_ClrRts / FT_SetBreakOn FT_SetBreakOff
// - FT_SetTimeouts / FT_GetQueueStatus / FT_SetEventNotification / FT_GetStatus
// - FT_SetWaitMask / FT_WaitOnMask
// - FT_GetEventStatus
// - FT_GetModemStatus / FT_SetChars / FT_Purge
// EEPROM:
// - FT_ReadEE
// - FT_EE_Read / FT_EE_ReadEx
// - FT_WriteEE
// - FT_EraseEE
// - FT_EEPROM_Program
// - FT_EE_Program / FT_EE_ProgramEx
// - FT_EE_UASize / FT_EE_UAWrite / FT_EE_UARead
// - FT_PROGRAM_DATA
// EEPROM FT232H:
// - FT_EE_ReadConfig / FT_EE_WriteConfig
// - FT_EE_ReadECC
// - FT_GetQueueStatusEx
// - FT_ComPortIdle / FT_ComPortCancelIdle
// - FT_VendorCmdGet / FT_VendorCmdSet / FT_VendorCmdGetEx / FT_VendorCmdSetEx
// USB:
// - FT_SetLatencyTimer / FT_GetLatencyTimer
// - FT_SetUSBParameters / FT_SetDeadmanTimeout
// - FT_SetVIDPID / FT_GetVIDPID
// - FT_StopInTask / FT_RestartInTask
// - FT_SetResetPipeRetryCount / FT_ResetPort / FT_CyclePort

const missing = -1
const noCGO = -2

const (
	bitmodeReset        = 0x00 // Reset all Pins to their default value
	bitmodeAsyncBitbang = 0x01 // Asynchronous bit bang
	bitmodeMpsse        = 0x02 // MPSSE (ft2232, ft2232h, ft4232h, ft232h)
	bitmodeSyncBitbang  = 0x04 // Synchronous bit bang (ft232r, ft245r, ft2232, ft2232h, ft4232h and ft232h)
	bitmodeMcuHost      = 0x08 // MCU host bus emulation (ft2232, ft2232h, ft4232h and ft232h)
	bitmodeFastSerial   = 0x10 // Fast opto-isolated serial mode (ft2232, ft2232h, ft4232h and ft232h)
	// In this case, upper nibble controls which pin is output/input, lower
	// controls which of outputs are high and low.
	bitmodeCbusBitbang = 0x20 // CBUS bit bang (ft232r and ft232h)
	bitmodeSyncFifo    = 0x40 // Single Channel Synchronous 245 FIFO mode (ft2232h and ft232h)
)

// For FT_EE_Program with FT_PROGRAM_DATA.
const (
	ft232HCBusTristate = 0x00 // Tristate
	ft232HCBusTxled    = 0x01 // Tx LED
	ft232HCBusRxled    = 0x02 // Rx LED
	ft232HCBusTxrxled  = 0x03 // Tx and Rx LED
	ft232HCBusPwren    = 0x04 // Power Enable
	ft232HCBusSleep    = 0x05 // Sleep
	ft232HCBusDrive0   = 0x06 // Drive pin to logic 0
	ft232HCBusDrive1   = 0x07 // Drive pin to logic 1
	ft232HCBusIomode   = 0x08 // IO Mode for CBUS bit-bang
	ft232HCBusTxden    = 0x09 // Tx Data Enable
	ft232HCBusClk30    = 0x0A // 30MHz clock
	ft232HCBusClk15    = 0x0B // 15MHz clock
	ft232HCBusClk7dot5 = 0x0C // 7.5MHz clock
)

// eepromHeader is FT_EEPROM_HEADER.
type eepromHeader struct {
	deviceType     devType // FTxxxx device type to be programmed
	VendorID       uint16  // Defaults to 0x0403; can be changed.
	ProductID      uint16  // Defaults to 0x6001 for ft232h, relevant value.
	SerNumEnable   uint8   // Non-zero if serial number to be used.
	MaxPower       uint16  // 0 < MaxPower <= 500
	SelfPowered    uint8   // 0 = bus powered, 1 = self powered
	RemoteWakeup   uint8   // 0 = not capable, 1 = capable
	PullDownEnable uint8   //
}

// eepromFt232h is FT_EEPROM_232H
type eepromFt232h struct {
	// eepromHeader
	deviceType     devType // FTxxxx device type to be programmed
	VendorID       uint16  // 0x0403
	ProductID      uint16  // 0x6001
	SerNumEnable   uint8   // non-zero if serial number to be used
	MaxPower       uint16  // 0 < MaxPower <= 500
	SelfPowered    uint8   // 0 = bus powered, 1 = self powered
	RemoteWakeup   uint8   // 0 = not capable, 1 = capable
	PullDownEnable uint8   //

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

// eepromFt232r is FT_EEPROM_232R
type eepromFt232r struct {
	// eepromHeader
	deviceType     devType // FTxxxx device type to be programmed
	VendorID       uint16  // 0x0403
	ProductID      uint16  // 0x6001
	SerNumEnable   uint8   // non-zero if serial number to be used
	MaxPower       uint16  // 0 < MaxPower <= 500
	SelfPowered    uint8   // 0 = bus powered, 1 = self powered
	RemoteWakeup   uint8   // 0 = not capable, 1 = capable
	PullDownEnable uint8   //

	// ft232r specific.
	IsHighCurrent uint8
	UseExtOsc     uint8
	InvertTXD     uint8
	InvertRXD     uint8
	InvertRTS     uint8
	InvertCTS     uint8
	InvertDTR     uint8
	InvertDSR     uint8
	InvertDCD     uint8
	InvertRI      uint8
	Cbus0         uint8 // Cbus Mux control
	Cbus1         uint8 // Cbus Mux control
	Cbus2         uint8 // Cbus Mux control
	Cbus3         uint8 // Cbus Mux control
	Cbus4         uint8 // Cbus Mux control
	DriverType    uint8 //
}

func toErr(s string, e int) error {
	msg := ""
	switch e {
	case missing:
		// when the library d2xx couldn't be loaded at runtime.
		msg = "couldn't load driver; visit https://github.com/periph/extra/tree/master/hostextra/d2xx"
	case noCGO:
		msg = "can't be used without cgo"
	case 0: // FT_OK
		return nil
	case 1: // FT_INVALID_HANDLE
		msg = "invalid handle"
	case 2: // FT_DEVICE_NOT_FOUND
		msg = "device not found; see https://github.com/periph/extra/tree/master/hostextra/d2xx for help"
	case 3: // FT_DEVICE_NOT_OPENED
		msg = "device busy; see https://github.com/periph/extra/tree/master/hostextra/d2xx for help"
	case 4: // FT_IO_ERROR
		msg = "I/O error"
	case 5: // FT_INSUFFICIENT_RESOURCES
		msg = "insufficient resources"
	case 6: // FT_INVALID_PARAMETER
		msg = "invalid parameter"
	case 7: // FT_INVALID_BAUD_RATE
		msg = "invalid baud rate"
	case 8: // FT_DEVICE_NOT_OPENED_FOR_ERASE
		msg = "device not opened for erase"
	case 9: // FT_DEVICE_NOT_OPENED_FOR_WRITE
		msg = "device not opened for write"
	case 10: // FT_FAILED_TO_WRITE_DEVICE
		msg = "failed to write device"
	case 11: // FT_EEPROM_READ_FAILED
		msg = "eeprom read failed"
	case 12: // FT_EEPROM_WRITE_FAILED
		msg = "eeprom write failed"
	case 13: // FT_EEPROM_ERASE_FAILED
		msg = "eeprom erase failed"
	case 14: // FT_EEPROM_NOT_PRESENT
		msg = "eeprom not present"
	case 15: // FT_EEPROM_NOT_PROGRAMMED
		msg = "eeprom not programmed"
	case 16: // FT_INVALID_ARGS
		msg = "invalid argument"
	case 17: // FT_NOT_SUPPORTED
		msg = "not supported"
	case 18: // FT_OTHER_ERROR
		msg = "other error"
	case 19: // FT_DEVICE_LIST_NOT_READY
		msg = "device list not ready"
	default:
		msg = "unknown status " + strconv.Itoa(e)
	}
	return errors.New("d2xx: " + s + ": " + msg)
}

// Common functions that must be implemented in addition to
// d2xxGetLibraryVersion(), d2xxCreateDeviceInfoList() and d2xxOpen().
type d2xxHandle interface {
	d2xxClose() int
	d2xxResetDevice() int
	d2xxGetDeviceInfo() (devType, uint16, uint16, int)
	d2xxEEPROMRead(d *device) int
	d2xxSetChars(eventChar byte, eventEn bool, errorChar byte, errorEn bool) int
	d2xxSetTimeouts(readMS, writeMS int) int
	d2xxSetLatencyTimer(delayMS uint8) int
	d2xxGetQueueStatus() (uint32, int)
	d2xxRead(b []byte) (int, int)
	d2xxWrite(b []byte) (int, int)
	d2xxGetBitMode() (byte, int)
	d2xxSetBitMode(mask, mode byte) int
}

// handle is a d2xx handle.
type handle uintptr

var _ d2xxHandle = handle(0)
