// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package d2xx

import (
	"errors"
	"strconv"
	"sync"
	"unsafe"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/spi"
)

// Type is the FTDI device type.
//
// The value can be "FT232H", "FT232R", etc.
//
// An empty string means the type is unknown.
type Type string

// Info is the information gathered about the connected FTDI device.
//
// The data is gathered from the USB descriptor.
type Info struct {
	// Opened is true if the device was successfully opened.
	Opened bool
	// Type is the FTDI device type.
	Type Type
	// VenID is the vendor ID from the USB descriptor information. It is expected
	// to be 0x0403 (FTDI).
	VenID uint16
	// DevID is the product ID from the USB descriptor information. It is
	// expected to be one of 0x6001, 0x6006 or 0x6010.
	DevID uint16
}

// EEPROM is the unprocessed EEPROM content.
//
// The EEPROM is in 3 parts: the 56 bytes header, the 4 strings and the rest
// which is used as an 'user area'. The size of the user area depends on the
// length of the strings. The user area content is not included in this struct.
type EEPROM struct {
	// Raw is the raw EEPROM content. It is normally around 56 bytes and excludes
	// the strings.
	Raw []byte

	// The following condition must be true: len(Manufacturer) + len(Desc) <= 40.
	Manufacturer   string
	ManufacturerID string
	Desc           string
	Serial         string
}

func (e *EEPROM) Interpret(t Type, p *ProcessedEEPROM) {
	if len(e.Raw) == 0 {
		return
	}
	// Use the custom structs instead of the ones provided by the library. The
	// reason is that it had to be written for Windows anyway, and this enables
	// using a single code path everywhere.
	hdr := (*eepromHeader)(unsafe.Pointer(&e.Raw[0]))
	p.MaxPower = uint16(hdr.MaxPower)
	p.SelfPowered = hdr.SelfPowered != 0
	p.RemoteWakeup = hdr.RemoteWakeup != 0
	p.PullDownEnable = hdr.PullDownEnable != 0
	switch t {
	case "FT232H":
		h := (*eepromFt232h)(unsafe.Pointer(&e.Raw[0]))
		p.CSlowSlew = h.ACSlowSlew != 0
		p.CSchmittInput = h.ACSchmittInput != 0
		p.CDriveCurrent = uint8(h.ACDriveCurrent)
		p.DSlowSlew = h.ADSlowSlew != 0
		p.DSchmittInput = h.ADSchmittInput != 0
		p.DDriveCurrent = uint8(h.ADDriveCurrent)
		p.Cbus0 = uint8(h.Cbus0)
		p.Cbus1 = uint8(h.Cbus1)
		p.Cbus2 = uint8(h.Cbus2)
		p.Cbus3 = uint8(h.Cbus3)
		p.Cbus4 = uint8(h.Cbus4)
		p.Cbus5 = uint8(h.Cbus5)
		p.Cbus6 = uint8(h.Cbus6)
		p.Cbus7 = uint8(h.Cbus7)
		p.Cbus8 = uint8(h.Cbus8)
		p.Cbus9 = uint8(h.Cbus9)
		p.FT1248Cpol = h.FT1248Cpol != 0
		p.FT1248Lsb = h.FT1248Lsb != 0
		p.FT1248FlowControl = h.FT1248FlowControl != 0
		p.IsFifo = h.IsFifo != 0
		p.IsFifoTar = h.IsFifoTar != 0
		p.IsFastSer = h.IsFastSer != 0
		p.IsFT1248 = h.IsFT1248 != 0
		p.PowerSaveEnable = h.PowerSaveEnable != 0
		p.DriverType = uint8(h.DriverType)
	case "FT232R":
		h := (*eepromFt232r)(unsafe.Pointer(&e.Raw[0]))
		p.IsHighCurrent = h.IsHighCurrent != 0
		p.UseExtOsc = h.UseExtOsc != 0
		p.InvertTXD = h.InvertTXD != 0
		p.InvertRXD = h.InvertRXD != 0
		p.InvertRTS = h.InvertRTS != 0
		p.InvertCTS = h.InvertCTS != 0
		p.InvertDTR = h.InvertDTR != 0
		p.InvertDSR = h.InvertDSR != 0
		p.InvertDCD = h.InvertDCD != 0
		p.InvertRI = h.InvertRI != 0
		p.Cbus0 = uint8(h.Cbus0)
		p.Cbus1 = uint8(h.Cbus1)
		p.Cbus2 = uint8(h.Cbus2)
		p.Cbus3 = uint8(h.Cbus3)
		p.Cbus4 = uint8(h.Cbus4)
		p.DriverType = uint8(h.DriverType)
	default:
		// TODO(maruel): Implement me!
	}
}

// ProcessedEEPROM is the interpreted EEPROM content.
//
// Interpretation depends on the device and this struct us prone to change as
// new FTDI devices are supported.
type ProcessedEEPROM struct {
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
}

// Dev represents one FTDI device.
//
// There can be multiple FTDI devices connected to a host.
//
// The device may also export one or multiple of I²C, SPI buses. You need to
// either cast into the right hardware, but more simply use the i2creg / spireg
// bus/port registries.
type Dev interface {
	// conn.Resource
	String() string
	Halt() error

	// Info returns information about an opened device.
	Info(i *Info)

	// Header returns the GPIO pins exposed on the chip.
	Header() []gpio.PinIO

	// SetSpeed sets the base clock for all I/O transactions.
	SetSpeed(hz int64) error

	// EEPROM returns the EEPROM content.
	EEPROM(ee *EEPROM) error
	// WriteEEPROM updates the EEPROM. Must be used carefully.
	WriteEEPROM(ee *EEPROM) error
	// UserArea reads and return the EEPROM part that can be used to stored user
	// defined values.
	UserArea() ([]byte, error)
	// WriteUserArea updates the user area in the EEPROM.
	//
	// If the length of ua is less than the available space, is it zero extended.
	WriteUserArea(ua []byte) error
}

// TODO(maruel): JTAG, Parallel, UART.

// broken represents a device that couldn't be opened correctly.
//
// It returns an error message to help the user diagnose issues.
type broken struct {
	index int
	err   error
}

func (b *broken) String() string {
	return "broken#" + strconv.Itoa(b.index) + ": " + b.err.Error()
}

func (b *broken) Halt() error {
	return nil
}

func (b *broken) Info(i *Info) {
	i.Opened = false
}

func (b *broken) Header() []gpio.PinIO {
	return nil
}

func (b *broken) SetSpeed(hz int64) error {
	return b.err
}

func (b *broken) EEPROM(ee *EEPROM) error {
	return b.err
}

func (b *broken) WriteEEPROM(ee *EEPROM) error {
	return b.err
}

func (b *broken) UserArea() ([]byte, error) {
	return nil, b.err
}

func (b *broken) WriteUserArea(ua []byte) error {
	return b.err
}

// generic represents a generic FTDI device.
//
// It is used for the models that this package doesn't fully support yet.
type generic struct {
	// Immutable after initialization.
	index int
	h     device

	// Mutable.
	initialized bool
	// ee is a cache of the device's EEPROM content.
	//ee EEPROM
}

func (f *generic) String() string {
	return string(f.h.t.Type()) + "(" + strconv.Itoa(f.index) + ")"
}

// Halt implements conn.Resource.
//
// This halts all operations going through this device.
func (f *generic) Halt() error {
	return f.h.reset()
}

// Info returns information about an opened device.
func (f *generic) Info(i *Info) {
	i.Opened = true
	i.Type = f.h.t.Type()
	i.VenID = f.h.venID
	i.DevID = f.h.devID
}

// Header returns the GPIO pins exposed on the chip.
func (f *generic) Header() []gpio.PinIO {
	return nil
}

func (f *generic) SetSpeed(hz int64) error {
	// TODO(maruel): When using MPSEE, use the MPSEE command.
	return f.h.setBaudRate(hz)
}

func (f *generic) EEPROM(ee *EEPROM) error {
	return f.h.readEEPROM(ee)
	/*
		if f.ee.Raw == nil {
			if err := f.h.readEEPROM(&f.ee); err != nil {
				return nil
			}
			if f.ee.Raw == nil {
				// It's a fresh new device. Devices bought via Adafruit already have
				// their EEPROM programmed with Adafruit branding but devices sold by
				// CJMCU are not. Since d2xxGetDeviceInfo() above succeeded, we know the
				// device type via the USB descriptor, which is sufficient to load the
				// driver, which permits to program the EEPROM to "bootstrap" it.
				f.ee.Raw = []byte{}
			}
		}
		*ee = f.ee
		return nil
	*/
}

func (f *generic) WriteEEPROM(ee *EEPROM) error {
	return f.h.programEEPROM(ee)
}

func (f *generic) UserArea() ([]byte, error) {
	return f.h.readUA()
}

func (f *generic) WriteUserArea(ua []byte) error {
	return f.h.writeUA(ua)
}

func (f *generic) initialize() error {
	if err := f.h.initialize(); err != nil {
		return err
	}
	f.initialized = true
	return nil
}

//

func newFT232H(g generic) *FT232H {
	f := &FT232H{
		generic: g,
		cbus:    gpiosMPSSE{h: &g.h, cbus: true},
		dbus:    gpiosMPSSE{h: &g.h},
	}
	f.cbus.init()
	f.dbus.init()
	f.dbus.pins[0].dp = gpio.Float
	f.dbus.pins[2].dp = gpio.Float
	f.dbus.pins[4].dp = gpio.Float
	f.cbus.pins[7].dp = gpio.PullDown

	for i := range f.dbus.pins {
		f.hdr[i] = &f.dbus.pins[i]
	}
	for i := range f.cbus.pins {
		f.hdr[i+8] = &f.cbus.pins[i]
	}
	f.hdr[16] = &invalidPin{num: 16, n: "C8"} // , dp: gpio.PullUp
	f.hdr[17] = &invalidPin{num: 17, n: "C9"} // , dp: gpio.PullUp
	f.D0 = f.hdr[0]
	f.D1 = f.hdr[1]
	f.D2 = f.hdr[2]
	f.D3 = f.hdr[3]
	f.D4 = f.hdr[4]
	f.D5 = f.hdr[5]
	f.D6 = f.hdr[6]
	f.D7 = f.hdr[7]
	f.C0 = f.hdr[8]
	f.C1 = f.hdr[9]
	f.C2 = f.hdr[10]
	f.C3 = f.hdr[11]
	f.C4 = f.hdr[12]
	f.C5 = f.hdr[13]
	f.C6 = f.hdr[14]
	f.C7 = f.hdr[15]
	f.C8 = f.hdr[16]
	f.C9 = f.hdr[17]

	// Update state by forcing all pins as inputs.
	f.h.mpsseCBus(0, 0)
	f.h.mpsseDBus(0, 0)
	f.cbus.read()
	f.dbus.read()
	return f
}

// FT232H represents a FT232H device.
//
// It implements Dev.
//
// Each group of pins D0~D7 and C0~C7 can be changed at once in one pass via
// DBus() or CBus().
//
// This enables usage as an 8 bit parallel port.
//
// Pins C8 and C9 can only be used in 'slow' mode via EEPROM and are currently
// not implemented.
//
// Datasheet
//
// http://www.ftdichip.com/Support/Documents/DataSheets/ICs/DS_FT232H.pdf
type FT232H struct {
	generic

	D0 gpio.PinIO // Clock output
	D1 gpio.PinIO // Data out
	D2 gpio.PinIO // Data in
	D3 gpio.PinIO // Chip select
	D4 gpio.PinIO
	D5 gpio.PinIO
	D6 gpio.PinIO
	D7 gpio.PinIO
	C0 gpio.PinIO
	C1 gpio.PinIO
	C2 gpio.PinIO
	C3 gpio.PinIO
	C4 gpio.PinIO
	C5 gpio.PinIO
	C6 gpio.PinIO
	C7 gpio.PinIO
	C8 gpio.PinIO // Not implemented
	C9 gpio.PinIO // Not implemented

	mu       sync.Mutex
	usingI2C bool
	usingSPI bool
	cbus     gpiosMPSSE
	dbus     gpiosMPSSE
	i        i2cBus
	s        spiPort

	hdr [18]gpio.PinIO
}

func (f *FT232H) String() string {
	return "ft232h(" + strconv.Itoa(f.index) + ")"
}

// Header returns the GPIO pins exposed on the chip.
func (f *FT232H) Header() []gpio.PinIO {
	out := make([]gpio.PinIO, len(f.hdr))
	copy(out, f.hdr[:])
	return out
}

// CBus sets the values of C0 to C7 in the specified direction and value.
//
// 0 direction means input, 1 means output.
func (f *FT232H) CBus(direction, value byte) error {
	return f.h.mpsseCBus(direction, value)
}

// DBus sets the values of D0 to d7 in the specified direction and value.
//
// 0 direction means input, 1 means output.
//
// This function must be used to set Clock idle level.
func (f *FT232H) DBus(direction, value byte) error {
	return f.h.mpsseDBus(direction, value)
}

// CBusRead reads the values of C0 to C7.
func (f *FT232H) CBusRead() (byte, error) {
	return f.h.mpsseCBusRead()
}

// DBusRead reads the values of D0 to D7.
func (f *FT232H) DBusRead() (byte, error) {
	return f.h.mpsseDBusRead()
}

// I2C returns an I²C bus over the AD bus.
//
// It uses D0, D1 and D2.
//
// D0 is SCL. It needs to be pulled up externally.
//
// D1 and D2 are used for SDA. D1 is the output using open drain, D2 is the
// input. D1 and D2 need to be wired together and pulled up externally.
//
// It is recommended to set the mode to ‘245 FIFO’ in the EEPROM of the FT232H.
//
// The FIFO mode is recommended because it allows the ADbus lines to start as
// tristate. If the chip starts in the default UART mode, then the ADbus lines
// will be in the default UART idle states until the application opens the port
// and configures it as MPSSE. Care should also be taken that the RD# input on
// ACBUS is not asserted in this initial state as this can cause the FIFO lines
// to drive out.
func (f *FT232H) I2C() (i2c.BusCloser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.usingI2C {
		return nil, errors.New("d2xx: already using I²C")
	}
	if f.usingSPI {
		return nil, errors.New("d2xx: already using SPI")
	}
	f.i.f = f
	if err := f.i.setupI2C(); err != nil {
		f.i.stopI2C()
		return nil, err
	}
	return &f.i, nil
}

// SPI returns a SPI port over the AD bus.
//
// It uses D0, D1, D2 and D3. D0 is the clock, D1 the output (MOSI), D2 is the
// input (MISO) and D3 is CS line.
func (f *FT232H) SPI() (spi.PortCloser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.usingI2C {
		return nil, errors.New("d2xx: already using I²C")
	}
	if f.usingSPI {
		return nil, errors.New("d2xx: already using SPI")
	}
	// Don't mark it as being used yet. It only become used once Connect() is
	// called.
	f.s.f = f
	return &f.s, nil
}

//

func newFT232R(g generic) *FT232R {
	f := &FT232R{
		generic: g,
		pins: [...]invalidPin{
			{num: 0, n: "C0"},              // dp: gpio.PullUp
			{num: 1, n: "C1"},              // dp: gpio.PullUp
			{num: 2, n: "C2"},              // dp: gpio.PullUp
			{num: 3, n: "C3"},              // dp: gpio.Float
			{num: 4, n: "C4"},              // dp: gpio.Float
			{num: 5, n: "TX", f: "UART"},   // dp: gpio.PullUp
			{num: 6, n: "RX", f: "UART"},   // dp: gpio.PullUp
			{num: 7, n: "RTS", f: "UART"},  // dp: gpio.PullUp
			{num: 8, n: "CTS", f: "UART"},  // dp: gpio.PullUp
			{num: 9, n: "DTR", f: "UART"},  // dp: gpio.PullUp
			{num: 10, n: "DSR", f: "UART"}, // dp: gpio.PullUp
			{num: 11, n: "DCD", f: "UART"}, // dp: gpio.PullUp
			{num: 12, n: "RI", f: "UART"},  // dp: gpio.PullUp
		},
	}
	for i := range f.pins {
		f.hdr[i] = &f.pins[i]
	}
	f.C0 = f.hdr[0]
	f.C1 = f.hdr[1]
	f.C2 = f.hdr[2]
	f.C3 = f.hdr[3]
	f.C4 = f.hdr[4]
	f.TX = f.hdr[5]
	f.RX = f.hdr[6]
	f.RTS = f.hdr[7]
	f.CTS = f.hdr[8]
	f.DTR = f.hdr[9]
	f.DSR = f.hdr[10]
	f.DCD = f.hdr[11]
	f.RI = f.hdr[12]
	return f
}

// FT232R represents a FT232RL/FT232RQ device.
//
// It implements Dev.
//
// Not all pins may be physically connected on the header!
//
// Adafruit's version only has the following pins connected: RX, TX, RTS and
// CTS.
//
// SparkFun's version exports all pins *except* (inexplicably) the CBus ones.
//
// Datasheet
//
// http://www.ftdichip.com/Support/Documents/DataSheets/ICs/DS_FT232R.pdf
type FT232R struct {
	generic

	C0  gpio.PinIO
	C1  gpio.PinIO
	C2  gpio.PinIO
	C3  gpio.PinIO
	C4  gpio.PinIO
	TX  gpio.PinIO // TXD
	RX  gpio.PinIO // RXD
	RTS gpio.PinIO
	CTS gpio.PinIO
	DTR gpio.PinIO
	DSR gpio.PinIO
	DCD gpio.PinIO
	RI  gpio.PinIO

	pins [13]invalidPin
	hdr  [13]gpio.PinIO
}

func (f *FT232R) String() string {
	return "ft232r(" + strconv.Itoa(f.index) + ")"
}

// Header returns the GPIO pins exposed on the chip.
func (f *FT232R) Header() []gpio.PinIO {
	out := make([]gpio.PinIO, len(f.hdr))
	copy(out, f.hdr[:])
	return out
}

//

var _ conn.Resource = Dev(nil)
