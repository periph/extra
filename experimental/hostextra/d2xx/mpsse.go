// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// MPSSE is Multi-Protocol Synchronous Serial Engine
//
// MPSSE basics:
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_135_MPSSE_Basics.pdf
//
// MPSSE and MCU emulation modes:
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_108_Command_Processor_for_MPSSE_and_MCU_Host_Bus_Emulation_Modes.pdf

package d2xx

import (
	"errors"
	"strconv"
	"time"

	"periph.io/x/periph/conn/gpio"
)

const (
	// TDI/TDO serial operation synchronised on clock edges.
	//
	// Combinations are:
	// - MSB or LSB first
	// - Long bitstreams or short ones
	// - Ouput, Input or both
	// - Data is sent on clock rising or falling edge
	//
	// Long streams:
	// - [1, 65536] bytes (length is sent minus one, requires 8 bits multiple)
	//
	// <op>, <LengthLow-1>, <LengthHigh-1>, <byte0>, ..., <byteN>
	dataOutMSBFBytesRise  = 0x10
	dataOutMSBFBytesFall  = 0x11
	dataInMSBFBytesRise   = 0x20
	dataInMSBFBytesFall   = 0x24
	dataIOMSBFBytesRise   = 0x31
	dataIOMSBFBytesFall   = 0x34
	dataOutLSBFBytesRise  = 0x18
	dataOutLSBFBytesFall  = 0x19
	dataInLSBFBytesRise   = 0x28
	dataInLSBFBytesFall   = 0x2C
	dataIOLSBFBytesInRise = 0x39
	dataIOLSBFBytesInFall = 0x3C
	// Short streams:
	// - [1, 8] bits
	//
	// <op>, <Length-1>, <byte>
	dataOutMSBFBitsRise  = 0x12
	dataOutMSBFBitsFall  = 0x13
	dataInMSBFBitsRise   = 0x22
	dataInMSBFBitsFall   = 0x26
	dataIOMSBFBitsRise   = 0x33
	dataIOMSBFBitsFall   = 0x36
	dataOutLSBFBitsRise  = 0x1A
	dataOutLSBFBitsFall  = 0x1B
	dataInLSBFBitsRise   = 0x2A
	dataInLSBFBitsFall   = 0x2E
	dataIOLSBFBitsInRise = 0x3B
	dataIOLSBFBitsInFall = 0x3E
	// Data line drives low when the data is 0 and tristates high on data 1. This
	// is used with I²C.
	dataTristate = 0x9E

	// TSM operation (for JTAG).
	//
	// - Send bits 6 to 0 to the TMS pin using LSB or MSB.
	// - Bit 7 is passed to TDI/DO before the first clock of TMS and is held
	//   static for the duration of TMS clocking.
	//
	// <op>, <Length>, <byte>
	tmsOutLSBFRise = 0x4A
	tmsOutLSBFFall = 0x4B
	tmsIOLSBInRise = 0x6A
	tmsIOLSBInFall = 0x6B
	// Unclear: 0x6E and 0x6F

	// GPIO operation.
	//
	// - Operates on 8 GPIOs at a time, e.g. C0~C7 or D0~D7.
	// - Direction 1 means output, 0 means input.
	//
	// <op>, <value>, <direction>
	gpioSetD  = 0x80
	gpioSetC  = 0x82
	gpioReadD = 0x81
	gpioReadC = 0x83

	// Internal loopback.
	//
	// Connects TDI and TDO together.
	internalLoopbackEnable  = 0x84
	internalLoopbackDisable = 0x85

	// Clock.
	//
	// The TCK/SK has a 50% duty cycle.
	//
	// The inactive clock state can be set via the gpioSetD command and control
	// bit 0.
	//
	// By default, the base clock is 6MHz via a 5x divisor. On
	// FT232H/FT2232H/FT4232H, the 5x divisor can be disabled.
	clock30MHz = 0x8A
	clock6MHz  = 0x8B
	// Sets clock divisor.
	//
	// The effective value depends if clock30MHz was sent or not.
	//
	// - 0(1) 6MHz / 30MHz
	// - 1(2) 3MHz / 15MHz
	// - 2(3) 2MHz / 10MHz
	// - 3(4) 1.5MHz / 7.5MHz
	// - 4(5) 1.25MHz / 6MHz
	// - ...
	// - 0xFFFF(65536) 91.553Hz / 457.763Hz
	//
	// <op>, <valueL-1>, <valueH-1>
	clockSetDivisor = 0x86
	// Uses 3 phases data clocking: data is valid on both clock edges. Needed
	// for I²C.
	clock3Phase = 0x8C
	// Uses normal 2 phases data clocking.
	clock2Phase = 0x8D
	// Enables clock even while not doing any operation. Used with JTAG.
	// Enables the clock between [1, 8] pulses.
	// <op>, <length-1>
	clockOnShort = 0x8E
	// Enables the clock between [8, 524288] pulses in 8 multiples.
	// <op>, <lengthL-1>, <lengthH-1>
	clockOnLong = 0x8F
	// Enables clock until D5 is high or low. Used with JTAG.
	clockUntilHigh = 0x94
	clockUntilLow  = 0x95
	// <op>, <lengthL-1>, <lengthH-1> in 8 multiples.
	clockUntilHighLong = 0x9C
	clockUntilLowLong  = 0x9D
	// Enables adaptive clocking. Used with JTAG.
	//
	// This causes the controller to wait for D7 signal state as an ACK.
	clockAdaptive = 0x96
	// Disables adaptive clocking.
	clockNormal = 0x97

	// CPU mode.
	//
	// Access the device registers like a memory mapped device.
	//
	// <op>, <addrLow>
	cpuReadShort = 0x90
	// <op>, <addrHi>, <addrLow>
	cpuReadFar = 0x91
	// <op>, <addrLow>, <data>
	cpuWriteShort = 0x92
	// <op>, <addrHi>, <addrLow>, <data>
	cpuWriteFar = 0x91

	// Buffer operations.
	//
	// Flush the buffer back to the host.
	flush = 0x87
	// Wait until D5 (JTAG) or I/O1 (CPU) is high. Once it is detected as
	// high, the MPSSE engine moves on to process the next instruction.
	waitHigh = 0x88
	waitLow  = 0x89
)

// setupMPSSE sets the device into MPSSE mode.
//
// This requires a f232h, ft2232, ft2232h or a ft4232h.
func (d *device) setupMPSSE() error {
	// http://www.ftdichip.com/Support/Documents/AppNotes/AN_255_USB%20to%20I2C%20Example%20using%20the%20FT232H%20and%20FT201X%20devices.pdf
	// Pre-state:
	// - Write EEPROM i.IsFifo = true so the device DBus is started in tristate.
	// FT_SetUSBParameters(ftHandle, 65536, 0)
	// FT_SetFlowControl

	// Enable the MPSSE controller.
	if err := d.setBitMode(0, 2); err != nil {
		return err
	}

	for _, v := range []byte{0xAA, 0xAB} {
		// Write a bad command and ensure it returned correctly.
		if _, err := d.write([]byte{v}); err != nil {
			return err
		}
		var b [2]byte
		if _, err := d.read(b[:]); err != nil {
			return err
		}
		// 0xFA means invalid command, 0xAA is the command echoed back.
		if b[0] != 0xFA || b[1] != v {
			return toErr("SetupMPSSE", 4)
			//return 4 // FT_IO_ERROR
		}
	}

	// Other I²C stuff skipped.
	_, err := d.write([]byte{clock30MHz, clockNormal, clock2Phase, internalLoopbackDisable})
	if err == nil {
		d.isMPSSE = true
	}
	if err != nil {
		return err
	}
	d.write([]byte{0x80, 0xC9, 0xFB})
	return err
}

//

// mpsseClock sets the clock at the closest value and returns it.
func (d *device) mpsseClock() (time.Duration, error) {
	// clockSetDivisor
	return 0, errors.New("d2xx: Not implemented")
}

func (d *device) mpsseTx(w, r []byte) error {
	// One of dataXXX
	return errors.New("d2xx: Not implemented")
}

func (d *device) mpsseCBus(mask, value byte) error {
	_, err := d.write([]byte{gpioSetC, mask, value})
	return err
}

func (d *device) mpsseDBus(mask, value byte) error {
	_, err := d.write([]byte{gpioSetD, mask, value})
	return err
}

func (d *device) mpsseCBusRead() (byte, error) {
	if _, err := d.write([]byte{gpioReadC}); err != nil {
		return 0, err
	}
	var b [1]byte
	if _, err := d.read(b[:]); err != nil {
		return 0, err
	}
	return b[0], nil
}

func (d *device) mpsseDBusRead() (byte, error) {
	if _, err := d.write([]byte{gpioReadD}); err != nil {
		return 0, err
	}
	var b [1]byte
	if _, err := d.read(b[:]); err != nil {
		return 0, err
	}
	return b[0], nil
}

//

// gpiosMPSSE is a slice of 8 GPIO pins driven via MPSSE.
//
// This permits keeping a cache.
type gpiosMPSSE struct {
	h    *device
	cbus bool // false if D bus

	// Cache of values
	direction byte
	value     byte

	pins [8]gpioMPSSE
}

func (g *gpiosMPSSE) init() {
	s := "D"
	if g.cbus {
		s = "C"
	}
	for i := range g.pins {
		g.pins[i].a = g
		g.pins[i].n = s + strconv.Itoa(i)
		g.pins[i].num = i
		g.pins[i].dp = gpio.PullUp
	}
}

func (g *gpiosMPSSE) in(n int) error {
	if g.h == nil {
		return errors.New("d2xx: device not open")
	}
	g.direction = g.direction & ^(1 << uint(n))
	if g.cbus {
		return g.h.mpsseCBus(g.direction, g.value)
	}
	return g.h.mpsseDBus(g.direction, g.value)
}

func (g *gpiosMPSSE) read() (byte, error) {
	if g.h == nil {
		return 0, errors.New("d2xx: device not open")
	}
	var err error
	if g.cbus {
		g.value, err = g.h.mpsseCBusRead()
	} else {
		g.value, err = g.h.mpsseDBusRead()
	}
	return g.value, err
}

func (g *gpiosMPSSE) out(n int, l gpio.Level) error {
	if g.h == nil {
		return errors.New("d2xx: device not open")
	}
	g.direction = g.direction | (1 << uint(n))
	if l {
		g.value |= 1 << uint(n)
	} else {
		g.value &^= 1 << uint(n)
	}
	if g.cbus {
		return g.h.mpsseCBus(g.direction, g.value)
	}
	return g.h.mpsseDBus(g.direction, g.value)
}

//

// gpioMPSSE is a GPIO pin on a FTDI device driven via MPSSE.
//
// gpioMPSSE implements gpio.PinIO.
type gpioMPSSE struct {
	a   *gpiosMPSSE
	n   string
	num int
	dp  gpio.Pull
}

// String implements pin.Pin.
func (g *gpioMPSSE) String() string {
	return g.n
}

// Name implements pin.Pin.
func (g *gpioMPSSE) Name() string {
	return g.n
}

// Number implements pin.Pin.
func (g *gpioMPSSE) Number() int {
	return g.num
}

// Function implements pin.Pin.
func (g *gpioMPSSE) Function() string {
	s := "Out/"
	m := byte(1 << uint(g.num))
	if g.a.direction&m == 0 {
		s = "In/"
		g.a.read()
	}
	return s + gpio.Level(g.a.value&m != 0).String()
}

// In implements gpio.PinIn.
func (g *gpioMPSSE) In(pull gpio.Pull, e gpio.Edge) error {
	if e != gpio.NoEdge {
		// We could support it on D5.
		return errors.New("d2xx: edge triggering is not supported")
	}
	if pull != gpio.Float && pull != gpio.PullNoChange {
		// In tristate, we can only pull up.
		// EEPROM has a PullDownEnable flag.
		return errors.New("d2xx: pull is not supported")
	}
	return g.a.in(g.num)
}

// Read implements gpio.PinIn.
func (g *gpioMPSSE) Read() gpio.Level {
	v, _ := g.a.read()
	return gpio.Level(v&(1<<uint(g.num)) != 0)
}

// WaitForEdge implements gpio.PinIn.
func (g *gpioMPSSE) WaitForEdge(t time.Duration) bool {
	return false
}

// DefaultPull implements gpio.PinDefaultPull.
func (g *gpioMPSSE) DefaultPull() gpio.Pull {
	return g.dp
}

// Pull implements gpio.PinIn.
func (g *gpioMPSSE) Pull() gpio.Pull {
	return gpio.PullNoChange
}

// Out implements gpio.PinOut.
func (g *gpioMPSSE) Out(l gpio.Level) error {
	return g.a.out(g.num, l)
}

var _ gpio.PinIO = &gpioMPSSE{}
