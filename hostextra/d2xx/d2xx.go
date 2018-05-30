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
	"fmt"
	"strconv"

	"periph.io/x/extra/hostextra/d2xx/ftdi"
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
	t     ftdi.DevType
	venID uint16
	devID uint16
}

func (d *device) closeDev() error {
	// Not yet called.
	return toErr("Close", d.h.d2xxClose())
}

// setupCommon is the general setup for common devices.
//
// It configures the device itself, the D2XX communication
// parameters and the USB parameters. The API doesn't make a clear distinction
// between all 3.
func (d *device) setupCommon() error {
	// Device: reset the device completely so it becomes in a known state.
	// TODO(maruel): It may not be really required, especially that this likely
	// trigger spurious I/O.
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
	if err := d.setBitMode(0, bitModeReset); err != nil {
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
// mask sets which pins are inputs and outputs for bitModeCbusBitbang.
func (d *device) setBitMode(mask byte, mode bitMode) error {
	return toErr("SetBitMode", d.h.d2xxSetBitMode(mask, byte(mode)))
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

func (d *device) readEEPROM(ee *ftdi.EEPROM) error {
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
		//
		// Fill it with an empty yet valid EEPROM content. We don't want to set
		// VenID or DevID to 0! Nobody would do that, right?
		ee.Raw = make([]byte, d.t.EEPROMSize())
		hdr := ee.AsHeader()
		hdr.DeviceType = d.t
		hdr.VendorID = d.venID
		hdr.ProductID = d.devID
	}
	return nil
}

func (d *device) programEEPROM(ee *ftdi.EEPROM) error {
	// Verify that the values are set correctly.
	if len(ee.Manufacturer) > 40 {
		return errors.New("d2xx: Manufacturer is too long")
	}
	if len(ee.ManufacturerID) > 40 {
		return errors.New("d2xx: ManufacturerID is too long")
	}
	if len(ee.Desc) > 40 {
		return errors.New("d2xx: Desc is too long")
	}
	if len(ee.Serial) > 40 {
		return errors.New("d2xx: Serial is too long")
	}
	if len(ee.Manufacturer)+len(ee.Desc) > 40 {
		return errors.New("d2xx: length of Manufacturer plus Desc is too long")
	}
	if len(ee.Raw) != 0 {
		hdr := ee.AsHeader()
		if hdr == nil {
			return errors.New("d2xx: unexpected EEPROM header size")
		}
		if hdr.DeviceType != d.t {
			return errors.New("d2xx: unexpected device type set while programming EEPROM")
		}
		if hdr.VendorID != d.venID {
			return errors.New("d2xx: unexpected VenID set while programming EEPROM")
		}
		if hdr.ProductID != d.devID {
			return errors.New("d2xx: unexpected DevID set while programming EEPROM")
		}
	}
	return toErr("EEPROMWrite", d.h.d2xxEEPROMProgram(ee))
}

func (d *device) eraseEEPROM() error {
	// Will fail on FT232R and FT245R. Not checking here, the driver will report
	// an error.
	return toErr("EraseEE", d.h.d2xxEraseEE())
}

func (d *device) readUA() ([]byte, error) {
	size, e := d.h.d2xxEEUASize()
	if e != 0 {
		return nil, toErr("EEUASize", e)
	}
	if size == 0 {
		// Happens on uninitialized EEPROM.
		return nil, nil
	}
	b := make([]byte, size)
	if e := d.h.d2xxEEUARead(b); e != 0 {
		return nil, toErr("EEUARead", e)
	}
	return b, nil
}

func (d *device) writeUA(ua []byte) error {
	size, e := d.h.d2xxEEUASize()
	if e != 0 {
		return toErr("EEUASize", e)
	}
	if size == 0 {
		return errors.New("d2xx: please program EEPROM first")
	}
	if size < len(ua) {
		return fmt.Errorf("d2xx: maximum user area size is %d bytes", size)
	}
	if size != len(ua) {
		b := make([]byte, size)
		copy(b, ua)
		ua = b
	}
	if e := d.h.d2xxEEUAWrite(ua); e != 0 {
		return toErr("EEUAWrite", e)
	}
	return nil
}

func (d *device) setBaudRate(hz int64) error {
	if hz >= 1<<31 {
		return errors.New("d2xx: baud rate too high")
	}
	return toErr("SetBaudRate", d.h.d2xxSetBaudRate(uint32(hz)))
}

//

const missing = -1
const noCGO = -2

// bitMode is used by setBitMode to change the chip behavior.
type bitMode uint8

const (
	// Resets all Pins to their default value
	bitModeReset bitMode = 0x00
	// Sets the DBus to asynchronous bit-bang.
	bitModeAsyncBitbang bitMode = 0x01
	// Switch to MPSSE mode (FT2232, FT2232H, FT4232H and FT232H).
	bitModeMpsse bitMode = 0x02
	// Sets the DBus to synchronous bit-bang (FT232R, FT245R, FT2232, FT2232H,
	// FT4232H and FT232H).
	bitModeSyncBitbang bitMode = 0x04
	// Switch to MCU host bus emulation (FT2232, FT2232H, FT4232H and FT232H).
	bitModeMcuHost bitMode = 0x08
	// Switch to fast opto-isolated serial mode (FT2232, FT2232H, FT4232H and
	// FT232H).
	bitModeFastSerial bitMode = 0x10
	// Sets the CBus in 4 bits bit-bang mode (FT232R and FT232H)
	// In this case, upper nibble controls which pin is output/input, lower
	// controls which of outputs are high and low.
	bitModeCbusBitbang bitMode = 0x20
	// Single Channel Synchronous 245 FIFO mode (FT2232H and FT232H).
	bitModeSyncFifo bitMode = 0x40
)

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
	d2xxGetDeviceInfo() (ftdi.DevType, uint16, uint16, int)
	d2xxEEPROMRead(d ftdi.DevType, e *ftdi.EEPROM) int
	d2xxEEPROMProgram(e *ftdi.EEPROM) int
	d2xxEraseEE() int
	d2xxWriteEE(offset uint8, value uint16) int
	d2xxEEUASize() (int, int)
	d2xxEEUARead(ua []byte) int
	d2xxEEUAWrite(ua []byte) int
	d2xxSetChars(eventChar byte, eventEn bool, errorChar byte, errorEn bool) int
	d2xxSetUSBParameters(in, out int) int
	d2xxSetFlowControl() int
	d2xxSetTimeouts(readMS, writeMS int) int
	d2xxSetLatencyTimer(delayMS uint8) int
	d2xxSetBaudRate(hz uint32) int
	d2xxGetQueueStatus() (uint32, int)
	d2xxRead(b []byte) (int, int)
	d2xxWrite(b []byte) (int, int)
	d2xxGetBitMode() (byte, int)
	d2xxSetBitMode(mask, mode byte) int
}

// handle is a d2xx handle.
type handle uintptr

var _ d2xxHandle = handle(0)
