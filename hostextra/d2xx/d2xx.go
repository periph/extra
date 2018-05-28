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
	return d, nil
}

// device is the lower level d2xx device handle, just above 'handle' which
// directly maps to D2XX function calls.
//
// device converts the int error type into Go native error and handles higher
// level functionality like reading and writing to the USB connection.
//
// The content of the struct is immutable after initialization.
type device struct {
	h     handle
	t     devType
	venID uint16
	devID uint16
}

func (d *device) closeDev() error {
	// Not yet called.
	return toErr("Close", d.h.d2xxClose())
}

func (d *device) initialize(ee *eeprom) error {
	if e := d.h.d2xxEEPROMRead(d.t, ee); e != 0 {
		// 15 == FT_EEPROM_NOT_PROGRAMMED
		if e != 15 {
			return toErr("EEPROMRead", e)
		}
		// It's a fresh new device. Devices bought via Adafruit already have
		// their EEPROM programmed with Adafruit branding but devices sold by
		// CJMCU are not. Since d2xxGetDeviceInfo() above succeeded, we know the
		// device type via the USB descriptor, which is sufficient to load the
		// driver, which permits to program the EEPROM to "bootstrap" it.
	}

	if err := d.setupCommon(); err != nil {
		return err
	}
	switch d.t {
	case ft232H, ft2232H, ft4232H: // ft2232
		if err := d.setupMPSSE(); err != nil {
			return err
		}
	case ft232R:
		// Asynchronous bitbang
		if err := d.setBitMode(0, 1); err != nil {
			return err
		}
	default:
	}
	return nil
}

// setupCommon is the general setup for common devices.
//
// It configures the device itself, the D2XX communication
// parameters and the USB parameters. The API doesn't make a clear distinction
// between all 3.
func (d *device) setupCommon() error {
	// Device: reset the device completely so it becomes in a known state.
	if err := d.reset(); err != nil {
		return err
	}
	// Driver: maximum packet size. Note that this clears any data in the buffer,
	// so it is good to do it immediately after a reset. The 'out' parameter is
	// ignored.
	if e := d.h.d2xxSetUSBParameters(65536, 0); e != 0 {
		return toErr("SetUSBParameters", e)
	}
	// Not sure: Disable event/error characters.
	if e := d.h.d2xxSetChars(0, false, 0, false); e != 0 {
		return toErr("SetChars", e)
	}
	// Driver: Set I/O timeouts to 5 sec.
	if e := d.h.d2xxSetTimeouts(5000, 5000); e != 0 {
		return toErr("SetTimeouts", e)
	}
	// Device: Latency timer at 1ms.
	if e := d.h.d2xxSetLatencyTimer(1); e != 0 {
		return toErr("SetLatencyTimer", e)
	}
	// Not sure: Turn on flow control to synchronize IN requests.
	if e := d.h.d2xxSetFlowControl(); e != 0 {
		return toErr("SetFlowControl", e)
	}
	// Device: Reset mode to setting in EEPROM.
	if err := d.setBitMode(0, 0); err != nil {
		return nil
	}
	return nil
}

// reset resets the device.
func (d *device) reset() error {
	if e := d.h.d2xxResetDevice(); e != 0 {
		return toErr("Reset", e)
	}
	// USB/driver: Flush any pending read buffer that had been sent by device
	// before it reset.
	return d.flushPending()
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

// flushPending flushes any data left in the read buffer.
func (d *device) flushPending() error {
	p, e := d.h.d2xxGetQueueStatus()
	if p == 0 || e != 0 {
		return toErr("FlushPending/GetQueueStatus", e)
	}
	_, e = d.h.d2xxRead(make([]byte, p))
	return toErr("FlushPending/Read", e)
}

func (d *device) read(b []byte) (int, error) {
	p, e := d.h.d2xxGetQueueStatus()
	if p == 0 || e != 0 {
		return int(p), toErr("Read/GetQueueStatus", e)
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

// eeprom contains the EEPROM content.
//
// The EEPROM is in 3 parts: the 56 bytes header, the 4 strings and the rest
// which is used as an 'user area'. The size of the user area depends on the
// length of the strings. Its content is not included in this struct.
type eeprom struct {
	// raw is the raw EEPROM content. It is normally around 56 bytes and excludes
	// the strings.
	raw []byte

	// The following condition must be true: len(manufacturer) + len(desc) <= 40.
	manufacturer   string
	manufacturerID string
	desc           string
	serial         string
}

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
	SerNumEnable   uint8   // Non-zero if serial number to be used
	MaxPower       uint16  // 0 < MaxPower <= 500mA
	SelfPowered    uint8   // 0 = bus powered, 1 = self powered
	RemoteWakeup   uint8   // 0 = not capable, 1 = capable
	PullDownEnable uint8   // Non zero if pull down in suspect enabled

	// ft232h specific.
	ACSlowSlew        uint8 // Non-zero if AC bus pins have slow slew
	ACSchmittInput    uint8 // Non-zero if AC bus pins are Schmitt input
	ACDriveCurrent    uint8 // Valid values are 4mA, 8mA, 12mA, 16mA
	ADSlowSlew        uint8 // Non-zero if AD bus pins have slow slew
	ADSchmittInput    uint8 // Non-zero if AD bus pins are Schmitt input
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
	IsFifo            uint8 // Non-zero if Interface is 245 FIFO
	IsFifoTar         uint8 // Non-zero if Interface is 245 FIFO CPU target
	IsFastSer         uint8 // Non-zero if Interface is Fast serial
	IsFT1248          uint8 // Non-zero if Interface is FT1248
	PowerSaveEnable   uint8 //
	DriverType        uint8 //
}

// eepromFt232r is FT_EEPROM_232R
type eepromFt232r struct {
	// eepromHeader
	deviceType     devType // FTxxxx device type to be programmed
	VendorID       uint16  // 0x0403
	ProductID      uint16  // 0x6001
	SerNumEnable   uint8   // Non-zero if serial number to be used
	MaxPower       uint16  // 0 < MaxPower <= 500mA
	SelfPowered    uint8   // 0 = bus powered, 1 = self powered
	RemoteWakeup   uint8   // 0 = not capable, 1 = capable
	PullDownEnable uint8   // Non zero if pull down in suspect enabled

	// ft232r specific.
	IsHighCurrent uint8 // High Drive I/Os; 3mA instead of 1mA (@3.3V)
	UseExtOsc     uint8 // Use external oscillator
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
	d2xxEEPROMRead(d devType, e *eeprom) int
	d2xxSetChars(eventChar byte, eventEn bool, errorChar byte, errorEn bool) int
	d2xxSetUSBParameters(in, out int) int
	d2xxSetFlowControl() int
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
