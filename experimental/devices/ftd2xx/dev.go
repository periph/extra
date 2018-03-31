// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

import (
	"fmt"

	"periph.io/x/periph/conn"
)

// VenID is the vendor ID for official FTDI devices.
const VenID uint16 = 0x0403

// Info is the information gathered about the connected FTDI device.
//
// Some is gathered from the USB descriptor, some from the EEPROM if possible.
type Info struct {
	// Opened is true if the device was successfully opened.
	Opened bool
	// Type is the FTDI device type.
	//
	// It has the form "ft232h". An empty string means the type is unknown.
	Type string
	// USB descriptor information.
	VenID     uint16
	ProductID uint16

	// The remainder is part of EEPROM.

	Manufacturer   string
	ManufacturerID string
	Desc           string
	Serial         string

	MaxPower       uint16 // 0 < MaxPower <= 500
	SelfPowered    bool   // false if powered by the USB bus
	RemoteWakeup   bool   //
	PullDownEnable bool   // true if pull down in suspend enabled

	// FT232H specific data.
	CSlowSlew         bool  // AC bus pins have slow slew
	CSchmittInput     bool  // AC bus pins are Schmitt input
	CDriveCurrent     uint8 // valid values are 4mA, 8mA, 12mA, 16mA
	DSlowSlew         bool  // non-zero if AD bus pins have slow slew
	DSchmittInput     bool  // non-zero if AD bus pins are Schmitt input
	DDriveCurrent     uint8 // valid values are 4mA, 8mA, 12mA, 16mA
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
	FT1248Cpol        bool  // FT1248 clock polarity - clock idle high (true) or clock idle low (false)
	FT1248Lsb         bool  // FT1248 data is LSB (true), or MSB (false)
	FT1248FlowControl bool  // FT1248 flow control enable
	IsFifo            bool  // Interface is 245 FIFO
	IsFifoTar         bool  // Interface is 245 FIFO CPU target
	IsFastSer         bool  // Interface is Fast serial
	IsFT1248          bool  // Interface is FT1248
	PowerSaveEnable   bool  //
	DriverType        uint8 //

	// FT232R specific data.
	IsHighCurrent bool // If interface is high current
	UseExtOsc     bool // Use External Oscillator
	InvertTXD     bool // Invert TXD
	InvertRXD     bool // Invert RXD
	InvertRTS     bool // Invert RTS
	InvertCTS     bool // Invert CTS
	InvertDTR     bool // Invert DTR
	InvertDSR     bool // Invert DSR
	InvertDCD     bool // Invert DCD
	InvertRI      bool // Invert RI
	//Cbus0         uint8 // Cbus Mux control
	//Cbus1         uint8 // Cbus Mux control
	//Cbus2         uint8 // Cbus Mux control
	//Cbus3         uint8 // Cbus Mux control
	//Cbus4         uint8 // Cbus Mux control
	//DriverType    uint8 //

	// EEPROM is the raw EEPROM data.
	EEPROM []byte
}

// Dev represents one FTDI device.
//
// There can be multiple FTDI devices connected to a host.
type Dev interface {
	fmt.Stringer
	conn.Resource
	GetInfo(i *Info)
}

// generic represents a generic FTDI device.
//
// It is used for the models that this package doesn't fully support yet.
type generic struct {
	index int
	h     *device // it may be nil if the device couldn't be opened.
	info  Info
}

func (g *generic) String() string {
	return fmt.Sprintf("ftdi(%d)", g.index)
}

// Halt implements conn.Resource.
//
// This halts all operations going through this device.
func (g *generic) Halt() error {
	return g.h.reset()
}

// GetDevInfo returns information about an opened device.
func (g *generic) GetInfo(i *Info) {
	*i = g.info
}

// FT232H represents a FT232H device.
//
// It implemented Dev.
type FT232H struct {
	generic

	C0 Pin
	C1 Pin
	C2 Pin
	C3 Pin
	C4 Pin
	C5 Pin
	C6 Pin
	C7 Pin
	C8 Pin
	C9 Pin
	D0 Pin
	D1 Pin
	D2 Pin
	D3 Pin
	D4 Pin
	D5 Pin
	D6 Pin
	D7 Pin
}

func (f *FT232H) String() string {
	return fmt.Sprintf("ft232h(%d)", f.index)
}

// FT232R represents a FT232R device.
//
// It implemented Dev.
type FT232R struct {
	generic

	TX  Pin
	RX  Pin
	RTS Pin
	CTS Pin
	DTR Pin
	DSR Pin
	DCD Pin
	RI  Pin
}

func (f *FT232R) String() string {
	return fmt.Sprintf("ft232r(%d)", f.index)
}

//

var _ conn.Resource = Dev(nil)
