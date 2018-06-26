// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package d2xx

import (
	"errors"
	"strconv"
	"sync"

	"periph.io/x/extra/hostextra/d2xx/ftdi"
	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
	"periph.io/x/periph/conn/spi"
)

// Info is the information gathered about the connected FTDI device.
//
// The data is gathered from the USB descriptor.
type Info struct {
	// Opened is true if the device was successfully opened.
	Opened bool
	// Type is the FTDI device type.
	//
	// The value can be "FT232H", "FT232R", etc.
	//
	// An empty string means the type is unknown.
	Type string
	// VenID is the vendor ID from the USB descriptor information. It is expected
	// to be 0x0403 (FTDI).
	VenID uint16
	// DevID is the product ID from the USB descriptor information. It is
	// expected to be one of 0x6001, 0x6006, 0x6010, 0x6014.
	DevID uint16
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
	SetSpeed(f physic.Frequency) error

	// EEPROM returns the EEPROM content.
	EEPROM(ee *ftdi.EEPROM) error
	// WriteEEPROM updates the EEPROM. Must be used carefully.
	WriteEEPROM(ee *ftdi.EEPROM) error
	// EraseEEPROM erases the EEPROM. Must be used carefully.
	EraseEEPROM() error
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

func (b *broken) SetSpeed(f physic.Frequency) error {
	return b.err
}

func (b *broken) EEPROM(ee *ftdi.EEPROM) error {
	return b.err
}

func (b *broken) WriteEEPROM(ee *ftdi.EEPROM) error {
	return b.err
}

func (b *broken) EraseEEPROM() error {
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
}

func (f *generic) String() string {
	return f.h.t.String() + "(" + strconv.Itoa(f.index) + ")"
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
	i.Type = f.h.t.String()
	i.VenID = f.h.venID
	i.DevID = f.h.devID
}

// Header returns the GPIO pins exposed on the chip.
func (f *generic) Header() []gpio.PinIO {
	return nil
}

func (f *generic) SetSpeed(freq physic.Frequency) error {
	// TODO(maruel): When using MPSEE, use the MPSEE command.
	// TODO(maruel): Doc says the actual speed is 16x, confirm.
	return f.h.setBaudRate(int64(freq / physic.Hertz))
}

func (f *generic) EEPROM(ee *ftdi.EEPROM) error {
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

func (f *generic) WriteEEPROM(ee *ftdi.EEPROM) error {
	// TODO(maruel): Compare with the cached EEPROM, and only update the
	// different values if needed so reduce the EEPROM wear.
	// f.h.h.d2xxWriteEE()
	return f.h.programEEPROM(ee)
}

func (f *generic) EraseEEPROM() error {
	return f.h.eraseEEPROM()
}

func (f *generic) UserArea() ([]byte, error) {
	return f.h.readUA()
}

func (f *generic) WriteUserArea(ua []byte) error {
	return f.h.writeUA(ua)
}

//

func newFT232H(g generic) (*FT232H, error) {
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
	// TODO(maruel): C8 and C9 can be used when their mux in the EEPROM is set to
	// ft232hCBusIOMode.
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
	if err := f.h.setupMPSSE(); err != nil {
		return nil, err
	}
	f.s.c.f = f
	f.i.f = f
	return f, nil
}

// FT232H represents a FT232H device.
//
// It implements Dev.
//
// The device can be used in a few different modes, two modes are supported:
//
// - D0~D3 as a serial protocol (MPSEE), supporting I²C and SPI (and eventually
// UART), In this mode, D4~D7 and C0~C7 can be used as synchronized GPIO.
//
// - D0~D7 as a synchronous 8 bits bit-bang port. In this mode, only a few pins
// on CBus are usable in slow mode.
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

	hdr  [18]gpio.PinIO
	cbus gpiosMPSSE
	dbus gpiosMPSSE

	mu       sync.Mutex
	usingI2C bool
	usingSPI bool
	i        i2cBus
	s        spiMPSEEPort
	// TODO(maruel): Technically speaking, a SPI port could be hacked up too in
	// sync bit-bang but there's less point when MPSEE is available.
}

// Header returns the GPIO pins exposed on the chip.
func (f *FT232H) Header() []gpio.PinIO {
	out := make([]gpio.PinIO, len(f.hdr))
	copy(out, f.hdr[:])
	return out
}

func (f *FT232H) SetSpeed(freq physic.Frequency) error {
	// TODO(maruel): When using MPSEE, use the MPSEE command. If using sync
	// bit-bang, use setBaudRate().

	// TODO(maruel): Doc says the actual speed is 16x, confirm.
	return f.h.setBaudRate(int64(freq / physic.Hertz))
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
// It uses D0, D1 and D2. This enforces the device to be in MPSEE mode.
//
// D0 is SCL. It must to be pulled up externally.
//
// D1 and D2 are used for SDA. D1 is the output using open drain, D2 is the
// input. D1 and D2 must be wired together and must be pulled up externally.
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
//
// This enforces the device to be in MPSEE mode.
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
	return &f.s, nil
}

//

func newFT232R(g generic) (*FT232R, error) {
	f := &FT232R{
		generic: g,
		dbus: [...]syncPin{
			{num: 0, n: "D0/TX"},
			{num: 1, n: "D1/RX"},
			{num: 2, n: "D2/RTS"},
			{num: 3, n: "D3/CTS"},
			{num: 4, n: "D4/DTR"},
			{num: 5, n: "D5/DSR"},
			{num: 6, n: "D6/DCD"},
			{num: 7, n: "D7/RI"},
		},
		cbus: [...]cbusPin{
			{num: 8, n: "C0", p: gpio.PullUp},
			{num: 9, n: "C1", p: gpio.PullUp},
			{num: 10, n: "C2", p: gpio.PullUp},
			{num: 11, n: "C3", p: gpio.Float},
		},
	}
	for i := range f.dbus {
		f.dbus[i].bus = f
		f.hdr[i] = &f.dbus[i]
	}
	for i := range f.cbus {
		f.cbus[i].bus = f
		f.hdr[i+8] = &f.cbus[i]
	}
	f.D0 = f.hdr[0]
	f.D1 = f.hdr[1]
	f.D2 = f.hdr[2]
	f.D3 = f.hdr[3]
	f.D4 = f.hdr[4]
	f.D5 = f.hdr[5]
	f.D6 = f.hdr[6]
	f.D7 = f.hdr[7]
	f.TX = f.hdr[0]
	f.RX = f.hdr[1]
	f.RTS = f.hdr[2]
	f.CTS = f.hdr[3]
	f.DTR = f.hdr[4]
	f.DSR = f.hdr[5]
	f.DCD = f.hdr[6]
	f.RI = f.hdr[7]
	f.C0 = f.hdr[8]
	f.C1 = f.hdr[9]
	f.C2 = f.hdr[10]
	f.C3 = f.hdr[11]

	// Default to 3MHz.
	if err := f.h.setBaudRate(3000000); err != nil {
		return nil, err
	}

	// Set all CBus pins as input.
	if err := f.h.setBitMode(0, bitModeCbusBitbang); err != nil {
		return nil, err
	}
	// And read their value.
	// TODO(maruel): Sadly this is impossible to know which pin is input or
	// output, but we could try to guess, as the call above may generate noise on
	// the line which could interfere with the device connected.
	var err error
	if f.cbusnibble, err = f.h.getBitMode(); err != nil {
		return nil, err
	}
	// Set all DBus as synchronous bitbang, everything as input.
	if err := f.h.setBitMode(0, bitModeSyncBitbang); err != nil {
		return nil, err
	}
	// And read their value.
	var b [1]byte
	if _, err := f.h.read(b[:]); err != nil {
		return nil, err
	}
	f.dvalue = b[0]
	f.s.c.f = f
	return f, nil
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
// The FT232R has 128 bytes output buffer and 256 bytes input buffer.
//
// Pin C4 can only be used in 'slow' mode via EEPROM and is currently not
// implemented.
//
// Datasheet
//
// http://www.ftdichip.com/Support/Documents/DataSheets/ICs/DS_FT232R.pdf
type FT232R struct {
	generic

	D0 gpio.PinIO
	D1 gpio.PinIO
	D2 gpio.PinIO
	D3 gpio.PinIO
	D4 gpio.PinIO
	D5 gpio.PinIO
	D6 gpio.PinIO
	D7 gpio.PinIO
	// Aliases to the Dn pins for user convenience. They point to the exact same
	// pin.
	TX  gpio.PinIO
	RX  gpio.PinIO
	RTS gpio.PinIO
	CTS gpio.PinIO
	DTR gpio.PinIO
	DSR gpio.PinIO
	DCD gpio.PinIO
	RI  gpio.PinIO

	// The CBus pins are slower to use, but can drive an high load, like a LED.
	C0 gpio.PinIO
	C1 gpio.PinIO
	C2 gpio.PinIO
	C3 gpio.PinIO

	dbus [8]syncPin
	cbus [4]cbusPin
	hdr  [12]gpio.PinIO

	// Mutable.
	mu         sync.Mutex
	usingSPI   bool
	s          spiSyncPort
	dmask      uint8 // 0 input, 1 output
	dvalue     uint8
	cbusnibble uint8 // upper nibble is I/O control, lower nibble is values.
}

// Header returns the GPIO pins exposed on the chip.
func (f *FT232R) Header() []gpio.PinIO {
	out := make([]gpio.PinIO, len(f.hdr))
	copy(out, f.hdr[:])
	return out
}

// SetDBusMask sets all D0~D7 input or output mode at once.
//
// mask is the input/output pins to use. A bit value of 0 sets the
// corresponding pin to an input, a bit value of 1 sets the corresponding pin
// to an output.
//
// It should be called before calling Tx().
func (f *FT232R) SetDBusMask(mask uint8) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.usingSPI {
		return errors.New("d2xx: already using SPI")
	}
	if mask != f.dmask {
		if err := f.h.setBitMode(mask, bitModeSyncBitbang); err != nil {
			return err
		}
		f.dmask = mask
	}
	return nil
}

// Tx does synchronized read-then-write on all the D0~D7 GPIOs.
//
// SetSpeed() determines the pace at which the I/O is done.
//
// SetDBusMask() determines which bits are interpreted in the w and r byte
// slice. w has its significant value masked by 'mask' and r has its
// significant value masked by '^mask'.
//
// Input sample is done *before* updating outputs. So r[0] is sampled before
// w[0] is used. The last w byte should be duplicated if an addition read is
// desired.
//
// On the Adafruit cable, only the first 4 bits D0(TX), D1(RX), D2(RTS) and
// D3(CTS) are connected. This is just enough to create a full duplex SPI bus!
func (f *FT232R) Tx(w, r []byte) error {
	if len(w) != len(r) {
		// TODO(maruel): Accept nil for one.
		return errors.New("d2xx: length of buffer w and r must match")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.usingSPI {
		return errors.New("d2xx: already using SPI")
	}
	// Chunk into 64 bytes chunks. That's half the buffer size of the chip.
	// TODO(maruel): Determine what's optimal.
	const chunk = 64
	for len(w) != 0 {
		c := len(w)
		if c > chunk {
			c = chunk
		}
		if _, err := f.h.write(w[:c]); err != nil {
			return err
		}
		if _, err := f.h.read(r[:c]); err != nil {
			return err
		}
		w = w[c:]
		r = w[c:]
	}
	return nil
}

// SPI returns a SPI port over the first 4 pins.
//
// It uses D0(TX), D1(RX), D2(RTS) and D3(CTS). D2(RTS) is the clock, D0(TX)
// the output (MOSI), D1(RX) is the input (MISO) and D3(CTS) is CS line.
func (f *FT232R) SPI() (spi.PortCloser, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.usingSPI {
		return nil, errors.New("d2xx: already using SPI")
	}
	// Don't mark it as being used yet. It only become used once Connect() is
	// called.
	return &f.s, nil
}

func (f *FT232R) syncBusFunc(n int) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	// TODO(maruel): Once UART is supported:
	// func := []string{"TX", "RX", "RTS", "CTS", "DTR", "DSR", "DCD", "RI"}
	// if f.usingSPI {
	//   func := []string{"SPI_MOSI", "SPI_MISO", "SPI_CLK", "SPI_CS", ...}
	// }
	mask := uint8(1 << uint(n))
	if f.dmask&mask != 0 {
		return "Out/" + gpio.Level(f.dvalue&mask != 0).String()
	}
	return "In/" + f.syncBusReadLocked(n).String()
}

func (f *FT232R) syncBusIn(n int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	// TODO(maruel): if usingSPI && n < 4.
	mask := uint8(1 << uint(n))
	if f.dmask&mask == 0 {
		// Already input.
		return nil
	}
	v := f.dmask &^ mask
	if err := f.h.setBitMode(v, bitModeSyncBitbang); err != nil {
		return err
	}
	f.dmask = v
	return nil
}

func (f *FT232R) syncBusRead(n int) gpio.Level {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.syncBusReadLocked(n)
}

func (f *FT232R) syncBusReadLocked(n int) gpio.Level {
	// In synchronous mode, to read we must write first to for a sample.
	b := [1]byte{f.dvalue}
	if _, err := f.h.write(b[:]); err != nil {
		return gpio.Low
	}
	mask := uint8(1 << uint(n))
	if _, err := f.h.read(b[:]); err != nil {
		return gpio.Low
	}
	f.dvalue = b[0]
	return f.dvalue&mask != 0
}

func (f *FT232R) syncBusOut(n int, l gpio.Level) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	mask := uint8(1 << uint(n))
	if f.dmask&mask != 1 {
		// Was input.
		v := f.dmask | mask
		if err := f.h.setBitMode(v, bitModeSyncBitbang); err != nil {
			return err
		}
		f.dmask = v
	}
	b := [1]byte{f.dvalue}
	if _, err := f.h.write(b[:]); err != nil {
		return err
	}
	f.dvalue = b[0]
	// In synchronous mode, we must read after writing to flush the buffer.
	if _, err := f.h.write(b[:]); err != nil {
		return err
	}
	return nil
}

func (f *FT232R) cBusFunc(n int) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	fmask := uint8(0x10 << uint(n))
	vmask := uint8(1 << uint(n))
	if f.cbusnibble&fmask != 0 {
		return "Out/" + gpio.Level(f.cbusnibble&vmask != 0).String()
	}
	return "In/" + f.cBusReadLocked(n).String()
}

func (f *FT232R) cBusIn(n int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fmask := uint8(0x10 << uint(n))
	if f.cbusnibble&fmask == 0 {
		// Already input.
		return nil
	}
	v := f.cbusnibble &^ fmask
	if err := f.h.setBitMode(v, bitModeCbusBitbang); err != nil {
		return err
	}
	f.cbusnibble = v
	return nil
}

func (f *FT232R) cBusRead(n int) gpio.Level {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.cBusReadLocked(n)
}

func (f *FT232R) cBusReadLocked(n int) gpio.Level {
	v, err := f.h.getBitMode()
	if err != nil {
		return gpio.Low
	}
	f.cbusnibble = v
	vmask := uint8(1 << uint(n))
	return f.cbusnibble&vmask != 0
}

func (f *FT232R) cBusOut(n int, l gpio.Level) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	fmask := uint8(0x10 << uint(n))
	vmask := uint8(1 << uint(n))
	v := f.cbusnibble | fmask
	if l {
		v |= vmask
	} else {
		v &^= vmask
	}
	if f.cbusnibble == v {
		// Was already in the right mode.
		return nil
	}
	if err := f.h.setBitMode(v, bitModeCbusBitbang); err != nil {
		return err
	}
	f.cbusnibble = v
	return nil
}

//

var _ conn.Resource = Dev(nil)
