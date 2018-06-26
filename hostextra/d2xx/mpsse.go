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
	"fmt"
	"strconv"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
)

const (
	// TDI/TDO serial operation synchronised on clock edges.
	//
	// Long streams (default):
	// - [1, 65536] bytes (length is sent minus one, requires 8 bits multiple)
	//   <op>, <LengthLow-1>, <LengthHigh-1>, <byte0>, ..., <byteN>
	//
	// Short streams (dataBit is specified):
	// - [1, 8] bits
	//   <op>, <Length-1>, <byte>
	//
	// When both dataOut and dataIn are specified, one of dataOutFall or
	// dataInFall should be specified, at least for most sane protocols.
	//
	// Flags:
	dataOut     byte = 0x10 // Enable output, default on +VE (Rise)
	dataIn      byte = 0x20 // Enable input, default on +VE (Rise)
	dataOutFall byte = 0x01 // instead of Rise
	dataInFall  byte = 0x04 // instead of Rise
	dataLSBF    byte = 0x08 // instead of MSBF
	dataBit     byte = 0x02 // instead of Byte

	// Data line drives low when the data is 0 and tristates high on data 1. This
	// is used with I²C.
	// <op>, <ADBus pins>, <ACBus pins>
	dataTristate byte = 0x9E

	// TSM operation (for JTAG).
	//
	// - Send bits 6 to 0 to the TMS pin using LSB or MSB.
	// - Bit 7 is passed to TDI/DO before the first clock of TMS and is held
	//   static for the duration of TMS clocking.
	//
	// <op>, <Length>, <byte>
	tmsOutLSBFRise byte = 0x4A
	tmsOutLSBFFall byte = 0x4B
	tmsIOLSBInRise byte = 0x6A
	tmsIOLSBInFall byte = 0x6B
	// Unclear: 0x6E and 0x6F

	// GPIO operation.
	//
	// - Operates on 8 GPIOs at a time, e.g. C0~C7 or D0~D7.
	// - Direction 1 means output, 0 means input.
	//
	// <op>, <value>, <direction>
	gpioSetD byte = 0x80
	gpioSetC byte = 0x82
	// <op>, returns <value>
	gpioReadD byte = 0x81
	gpioReadC byte = 0x83

	// Internal loopback.
	//
	// Connects TDI and TDO together.
	internalLoopbackEnable  byte = 0x84
	internalLoopbackDisable byte = 0x85

	// Clock.
	//
	// The TCK/SK has a 50% duty cycle.
	//
	// The inactive clock state can be set via the gpioSetD command and control
	// bit 0.
	//
	// By default, the base clock is 6MHz via a 5x divisor. On
	// FT232H/FT2232H/FT4232H, the 5x divisor can be disabled.
	clock30MHz byte = 0x8A
	clock6MHz  byte = 0x8B
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
	clockSetDivisor byte = 0x86
	// Uses 3 phases data clocking: data is valid on both clock edges. Needed
	// for I²C.
	clock3Phase byte = 0x8C
	// Uses normal 2 phases data clocking.
	clock2Phase byte = 0x8D
	// Enables clock even while not doing any operation. Used with JTAG.
	// Enables the clock between [1, 8] pulses.
	// <op>, <length-1>
	clockOnShort byte = 0x8E
	// Enables the clock between [8, 524288] pulses in 8 multiples.
	// <op>, <lengthL-1>, <lengthH-1>
	clockOnLong byte = 0x8F
	// Enables clock until D5 is high or low. Used with JTAG.
	clockUntilHigh byte = 0x94
	clockUntilLow  byte = 0x95
	// <op>, <lengthL-1>, <lengthH-1> in 8 multiples.
	clockUntilHighLong byte = 0x9C
	clockUntilLowLong  byte = 0x9D
	// Enables adaptive clocking. Used with JTAG.
	//
	// This causes the controller to wait for D7 signal state as an ACK.
	clockAdaptive byte = 0x96
	// Disables adaptive clocking.
	clockNormal byte = 0x97

	// CPU mode.
	//
	// Access the device registers like a memory mapped device.
	//
	// <op>, <addrLow>
	cpuReadShort byte = 0x90
	// <op>, <addrHi>, <addrLow>
	cpuReadFar byte = 0x91
	// <op>, <addrLow>, <data>
	cpuWriteShort byte = 0x92
	// <op>, <addrHi>, <addrLow>, <data>
	cpuWriteFar byte = 0x91

	// Buffer operations.
	//
	// Flush the buffer back to the host.
	flush byte = 0x87
	// Wait until D5 (JTAG) or I/O1 (CPU) is high. Once it is detected as
	// high, the MPSSE engine moves on to process the next instruction.
	waitHigh byte = 0x88
	waitLow  byte = 0x89
)

// setupMPSSE sets the device into MPSSE mode.
//
// This requires a f232h, ft2232, ft2232h or a ft4232h.
func (d *device) setupMPSSE() error {
	// http://www.ftdichip.com/Support/Documents/AppNotes/AN_255_USB%20to%20I2C%20Example%20using%20the%20FT232H%20and%20FT201X%20devices.pdf
	// Pre-state:
	// - Write EEPROM i.IsFifo = true so the device DBus is started in tristate.

	// Enable the MPSSE controller.
	if err := d.setBitMode(0, 2); err != nil {
		return err
	}

	for _, v := range []byte{0xAA, 0xAB} {
		// Write a bad command and ensure it returned correctly.
		if _, err := d.write([]byte{v}); err != nil {
			return err
		}
		// Try for 200ms.
		var b [2]byte
		success := false
		for start := time.Now(); time.Since(start) < 200*time.Millisecond; {
			if n, err := d.read(b[:]); err != nil {
				return err
			} else if n == 0 {
				// Slow down the busy loop a little.
				time.Sleep(100 * time.Microsecond)
				continue
			}
			// 0xFA means invalid command, 0xAA is the command echoed back.
			if b[0] != 0xFA || b[1] != v {
				return toErr("SetupMPSSE", 4)
			}
			success = true
			break
		}
		if !success {
			return fmt.Errorf("d2xx: I/O failure while trying to setup MPSSE: %v", b[:])
		}
	}

	// Initialize MPSSE to a known state.
	cmd := []byte{
		clock30MHz, clockNormal, clock2Phase, internalLoopbackDisable,
		gpioSetC, 0xFF, 0x00,
		gpioSetD, 0xFF, 0x00,
	}
	if _, err := d.write(cmd); err != nil {
		return err
	}
	// Success!!
	return nil
}

//

// mpsseRegRead reads the memory mapped registers from the device.
func (d *device) mpsseRegRead(addr uint16) (byte, error) {
	// Unlike most other operations, the uint16 byte order is <hi>, <lo>.
	b := [...]byte{cpuReadFar, byte(addr >> 8), byte(addr)}
	if _, err := d.write(b[:]); err != nil {
		return 0, err
	}
	_, err := d.read(b[:1])
	return b[0], err
}

// mpsseClock sets the clock at the closest value and returns it.
func (d *device) mpsseClock(f physic.Frequency) (physic.Frequency, error) {
	clk := clock30MHz
	base := 30 * physic.MegaHertz
	div := base / f
	if div >= 65536 {
		clk = clock6MHz
		base /= 5
		div = base / f
		if div >= 65536 {
			return 0, errors.New("d2xx: clock frequency is too low")
		}
	}
	b := [...]byte{clk, clockSetDivisor, byte(div - 1), byte((div - 1) >> 8)}
	_, err := d.write(b[:])
	return base / div, err
}

// mpsseTx runs a transaction on the clock on pins D0, D1 and D2.
//
// It can only do it on a multiple of 8 bits.
func (d *device) mpsseTx(w, r []byte, ew, er gpio.Edge, lsbf bool) error {
	op := byte(0)
	if lsbf {
		op |= dataLSBF
	}
	l := len(w)
	if len(w) != 0 {
		if len(w) > 65536 {
			return errors.New("d2xx: write buffer too long; max 65536")
		}
		op |= dataOut
		if ew == gpio.FallingEdge {
			op |= dataOutFall
		}
	}
	if len(r) != 0 {
		if len(r) > 65536 {
			return errors.New("d2xx: read buffer too long; max 65536")
		}
		op |= dataIn
		if er == gpio.FallingEdge {
			op |= dataInFall
		}
		if l != 0 && len(r) != l {
			return errors.New("d2xx: mismatched buffer lengths")
		}
		l = len(r)
	}
	// The FT232H has 1Kb Tx and Rx buffers. So partial writes should be done.
	// TODO(maruel): Test.

	// flushBuffer can be useful if rbits != 0.
	cmd := []byte{op, byte(l - 1), byte((l - 1) >> 8)}
	if _, err := d.write(append(cmd, w...)); err != nil {
		return err
	}
	if len(r) != 0 {
		_, err := d.read(r)
		return err
	}
	return nil
}

// mpsseTxShort runs a transaction on the clock pins D0, D1 and D2 for a byte
// or less: between 1 and 8 bits.
func (d *device) mpsseTxShort(w byte, wbits, rbits int, ew, er gpio.Edge, lsbf bool) (byte, error) {
	op := byte(dataBit)
	if lsbf {
		op |= dataLSBF
	}
	l := wbits
	if wbits != 0 {
		if wbits > 8 {
			return 0, errors.New("d2xx: write buffer too long; max 8")
		}
		op |= dataOut
		if ew == gpio.FallingEdge {
			op |= dataOutFall
		}
	}
	if rbits != 0 {
		if rbits > 8 {
			return 0, errors.New("d2xx: read buffer too long; max 8")
		}
		op |= dataIn
		if er == gpio.FallingEdge {
			op |= dataInFall
		}
		if l != 0 && rbits != l {
			return 0, errors.New("d2xx: mismatched buffer lengths")
		}
		l = rbits
	}
	b := [3]byte{op, byte(l - 1)}
	cmd := b[:2]
	if wbits != 0 {
		cmd = append(cmd, w)
	}
	if _, err := d.write(cmd); err != nil {
		return 0, err
	}
	if rbits != 0 {
		_, err := d.read(b[:1])
		return b[0], err
	}
	return 0, nil
}

func (d *device) mpsseCBus(mask, value byte) error {
	b := [...]byte{gpioSetC, mask, value}
	_, err := d.write(b[:])
	return err
}

// mpsseDBus operates on 8 GPIOs at a time D0~D7.
//
// Direction 1 means output, 0 means input.
func (d *device) mpsseDBus(mask, value byte) error {
	b := [...]byte{gpioSetD, mask, value}
	_, err := d.write(b[:])
	return err
}

func (d *device) mpsseCBusRead() (byte, error) {
	b := [...]byte{gpioReadC}
	if _, err := d.write(b[:]); err != nil {
		return 0, err
	}
	if _, err := d.read(b[:1]); err != nil {
		return 0, err
	}
	return b[0], nil
}

func (d *device) mpsseDBusRead() (byte, error) {
	b := [...]byte{gpioReadD}
	if _, err := d.write(b[:]); err != nil {
		return 0, err
	}
	if _, err := d.read(b[:1]); err != nil {
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

// Halt implements gpio.PinIO.
func (g *gpioMPSSE) Halt() error {
	return nil
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

// DefaultPull implements gpio.PinIn.
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

// PWM implements gpio.PinOut.
func (g *gpioMPSSE) PWM(d gpio.Duty, f physic.Frequency) error {
	return errors.New("d2xx: not implemented")
}

/*
func (g *gpioMPSSE) Drive() physic.ElectricCurrent {
	//return g.a.ee.CDriveCurrent * physic.MilliAmpere
	return 2 * physic.MilliAmpere
}

func (g *gpioMPSSE) SlewLimit() bool {
	//return g.a.ee.CSlowSlew
	return false
}

func (g *gpioMPSSE) Hysteresis() bool {
	//return g.a.ee.DSchmittInput
	return true
}
*/

var _ gpio.PinIO = &gpioMPSSE{}
