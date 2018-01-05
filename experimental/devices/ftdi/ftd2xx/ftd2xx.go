// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

import (
	"errors"
	"fmt"

	"periph.io/x/extra/experimental/devices/ftdi"
)

// Driver implements ftdi.Driver.
var Driver driver

type driver struct {
}

// Version implements ftdi.Driver.
func (d *driver) Version() (uint8, uint8, uint8) {
	return getLibraryVersion()
}

// Open implements ftdi.Driver.
func (d *driver) Open(i int) (ftdi.Handle, error) {
	h, e := open(i)
	if e != 0 {
		return h, toErr("Open failed", e)
	}
	return h, nil
}

// NumDevices implements ftdi.Driver.
func (d *driver) NumDevices() (int, error) {
	num, e := createDeviceInfoList()
	if e != 0 {
		return 0, toErr("GetNumDevices initialization failed", e)
	}
	return num, nil
}

/*
// ListDevices implements ftdi.Driver.
func ListDevices() ([]DevInfo, error) {
	num, e := createDeviceInfoList()
	if e != 0 {
		return nil, toErr("ListDevice initialization failed", e)
	}
	// Returns barely anything unless the device was opened.
	dev, e := getDeviceInfoList(num)
	if e != 0 {
		return nil, toErr("ListDevice failed", e)
	}
	return dev, nil
}
*/

//

// handle implements ftdi.handle.
type handle uintptr

// Close implements ftdi.handle.
func (h handle) Close() error {
	return toErr("Close", closeHandle(h))
}

// GetInfo implements ftdi.handle.
//
// Under the hood, it calls both FT_GetDeviceInfo and FT_EEPROM_READ.
func (h handle) GetInfo(i *ftdi.Info) error {
	if e := getInfo(h, i); e != 0 {
		return toErr("GetInfo", e)
	}
	return nil
}

//

// devType is the FTDI device type.
type devType int

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

// TODO(maruel): To add:
// - FT_Read
// - FT_Write
// - FT_IoCtl
// - FT_ResetDevice
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
// - FT_SetBitMode / FT_GetBitMode
// - FT_SetUSBParameters / FT_SetDeadmanTimeout
// - FT_SetVIDPID / FT_GetVIDPID
// - FT_StopInTask / FT_RestartInTask
// - FT_SetResetPipeRetryCount / FT_ResetPort / FT_CyclePort

const missing = -1
const noCGO = -2

// Baud Rates
// ...

func toErr(s string, e int) error {
	switch e {
	case missing:
		// when the library ftd2xx couldn't be loaded at runtime.
		return errors.New("ftd2xx: couldn't load driver; visit https://github.com/periph/extra/tree/master/experimental/devices/ftdi/ftd2xx")
	case noCGO:
		return errors.New("ftd2xx: can't be used without cgo")
	case 0: // FT_OK
		return nil
	case 1: // FT_INVALID_HANDLE
		return fmt.Errorf("ftd2xx: %s: invalid handle", s)
	case 2: // FT_DEVICE_NOT_FOUND
		return fmt.Errorf("ftd2xx: %s: device not found", s)
	case 3: // FT_DEVICE_NOT_OPENED
		return fmt.Errorf("ftd2xx: %s: device not opened", s)
	case 4: // FT_IO_ERROR
		return fmt.Errorf("ftd2xx: %s: I/O error", s)
	case 5: // FT_INSUFFICIENT_RESOURCES
		return fmt.Errorf("ftd2xx: %s: insufficient resources", s)
	case 6: // FT_INVALID_PARAMETER
		return fmt.Errorf("ftd2xx: %s: invalid parameter", s)
	case 7: // FT_INVALID_BAUD_RATE
		return fmt.Errorf("ftd2xx: %s: invalid baud rate", s)
	case 8: // FT_DEVICE_NOT_OPENED_FOR_ERASE
		return fmt.Errorf("ftd2xx: %s: device not opened for erase", s)
	case 9: // FT_DEVICE_NOT_OPENED_FOR_WRITE
		return fmt.Errorf("ftd2xx: %s: device not opened for write", s)
	case 10: // FT_FAILED_TO_WRITE_DEVICE
		return fmt.Errorf("ftd2xx: %s: failed to write device", s)
	case 11: // FT_EEPROM_READ_FAILED
		return fmt.Errorf("ftd2xx: %s: eeprom read failed", s)
	case 12: // FT_EEPROM_WRITE_FAILED
		return fmt.Errorf("ftd2xx: %s: eeprom write failed", s)
	case 13: // FT_EEPROM_ERASE_FAILED
		return fmt.Errorf("ftd2xx: %s: eeprom erase failed", s)
	case 14: // FT_EEPROM_NOT_PRESENT
		return fmt.Errorf("ftd2xx: %s: eeprom not present", s)
	case 15: // FT_EEPROM_NOT_PROGRAMMED
		return fmt.Errorf("ftd2xx: %s: eeprom not programmed", s)
	case 16: // FT_INVALID_ARGS
		return fmt.Errorf("ftd2xx: %s: invalid argument", s)
	case 17: // FT_NOT_SUPPORTED
		return fmt.Errorf("ftd2xx: %s: not supported", s)
	case 18: // FT_OTHER_ERROR
		return fmt.Errorf("ftd2xx: %s: other error", s)
	case 19: // FT_DEVICE_LIST_NOT_READY
		return fmt.Errorf("ftd2xx: %s: device list not ready", s)
	default:
		return fmt.Errorf("ftd2xx: %s: unknown status %d", s, e)
	}
}
