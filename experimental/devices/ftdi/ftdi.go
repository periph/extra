// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftdi

import (
	"fmt"
	"io"
	"sync"

	"periph.io/x/periph/conn"
)

// VenID is the vendor ID for official FTDI devices.
const VenID uint16 = 0x0403

// Info is the information gathered about the connected FTDI device.
//
// Some is gather from USB descriptor, some from the EEPROM if possible.
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

// Generic represents a generic FTDI device.
//
// It is used for the models that this package doesn't fully support yet.
type Generic struct {
	index int
	h     Handle // it may be nil if the device couldn't be opened.
	info  Info
}

func (g *Generic) String() string {
	return fmt.Sprintf("ftdi(%d)", g.index)
}

// Halt implements conn.Resource.
func (g *Generic) Halt() error {
	// TODO(maruel): Halt all operations going through this device.
	return nil
}

// GetDevInfo returns information about an opened device.
func (g *Generic) GetInfo(i *Info) {
	*i = g.info
}

// FT232R represents a FT232R device.
type FT232R struct {
	Generic

	TX  Pin
	RX  Pin
	RTS Pin
	CTS Pin
	DTR Pin
	DSR Pin
	DCD Pin
	RI  Pin
}

// FT232H represents a FT232H device.
type FT232H struct {
	Generic

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

// All enumerates all the connected FTDI devices.
//
// Some may not be opened; they may already be opened by another process or by
// a driver included by the operating system.
//
// See
// https://github.com/periph/extra/tree/master/experimental/devices/ftdi/ftd2xx
func All() []Dev {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Dev, len(all))
	copy(out, all)
	return out
}

// Driver is implemented by ftd2xx and eventually by libftdi and ftd3xx.
type Driver interface {
	// Version returns the major, minor and build number version of the user mode
	// library.
	Version() (uint8, uint8, uint8)
	// NumDevices returns the number of FTDI devices found on the USB buses.
	NumDevices() (int, error)
	// Open opens a device at index i.
	Open(i int) (Handle, error)
}

// Handle is implemented by ftd2xx and eventually by libftdi and ftd3xx.
type Handle interface {
	// Nobody should normally close the handle.
	io.Closer
	// GetInfo returns information about the device.
	GetInfo(i *Info) error
	// TODO(maruel): Add operations.
}

// RegisterDriver registers a driver.
//
// Normally this should be &ftd2xx.Driver.
//
// Opens all devices found that are not already busy due to another OS provided
// driver.
func RegisterDriver(d Driver) error {
	mu.Lock()
	defer mu.Unlock()
	driver = d
	num, err := driver.NumDevices()
	if err != nil {
		return err
	}
	for i := 0; i < num; i++ {
		if d, err1 := open(i); err1 == nil {
			all = append(all, d)
		} else {
			// Create a shallow generic handle.
			err = err1
			all = append(all, &Generic{index: i})
		}
	}
	return err
}

//

var (
	mu     sync.Mutex
	driver Driver
	all    []Dev
	failed error
)

// open opens a FTDI device.
//
// Must be called with mu held.
func open(i int) (Dev, error) {
	h, err := driver.Open(i)
	if err != nil {
		return nil, err
	}
	var info Info
	if err := h.GetInfo(&info); err != nil {
		return nil, err
	}
	g := Generic{index: i, h: h, info: info}
	switch info.Type {
	case "ft232h":
		return &FT232H{Generic: g}, nil
	case "ft232r":
		return &FT232R{Generic: g}, nil
	default:
		return &g, nil
	}
}
