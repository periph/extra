// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// This functionality requires MPSSE.
//
// Interfacing I²C:
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_113_FTDI_Hi_Speed_USB_To_I2C_Example.pdf
//
// Implementation based on
// http://www.ftdichip.com/Support/Documents/AppNotes/AN_255_USB%20to%20I2C%20Example%20using%20the%20FT232H%20and%20FT201X%20devices.pdf
//
// Page 18: MPSSE does not automatically support clock stretching for I²C.

package d2xx

import (
	"errors"
	"fmt"

	"periph.io/x/periph/conn"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/i2c"
	"periph.io/x/periph/conn/physic"
)

type i2cBus struct {
	f *FT232H
}

// Close stops I²C mode, returns to high speed mode, disable tri-state.
func (d *i2cBus) Close() error {
	d.f.mu.Lock()
	err := d.stopI2C()
	d.f.mu.Unlock()
	return err
}

// Duplex implements conn.Conn.
func (d *i2cBus) Duplex() conn.Duplex {
	return conn.Half
}

func (d *i2cBus) String() string {
	return d.f.String()
}

// SetSpeed implements i2c.Bus.
func (d *i2cBus) SetSpeed(f physic.Frequency) error {
	if f > 10*physic.MegaHertz {
		return fmt.Errorf("d2xx: invalid speed %s; maximum supported clock is 10MHz", f)
	}
	if f < 100*physic.Hertz {
		return fmt.Errorf("d2xx: invalid speed %s; minimum supported clock is 100Hz; did you forget to multiply by physic.KiloHertz?", f)
	}
	// TODO(maruel): Use proper mpsse command.
	clk := ((30 * physic.MegaHertz / f) - 1) * 2 / 3
	cmd := [...]byte{
		clock30MHz, byte(clk), byte(clk >> 8),
	}
	if _, err := d.f.h.write(cmd[:]); err != nil {
		return err
	}
	return d.setI2CLinesIdle()
}

// Tx implements i2c.Bus.
func (d *i2cBus) Tx(addr uint16, w, r []byte) error {
	d.f.mu.Lock()
	defer d.f.mu.Unlock()
	if err := d.setI2CStart(); err != nil {
		return err
	}
	a := [1]byte{byte(addr)}
	if err := d.writeBytes(a[:]); err != nil {
		return err
	}
	if len(w) != 0 {
		if err := d.writeBytes(w); err != nil {
			return err
		}
	}
	if len(r) != 0 {
		if err := d.readBytes(r); err != nil {
			return err
		}
	}
	if err := d.setI2CStop(); err != nil {
		return err
	}
	return d.setI2CLinesIdle()
}

// SCL implements i2c.Pins.
func (d *i2cBus) SCL() gpio.PinIO {
	return d.f.D0
}

// SDA implements i2c.Pins.
func (d *i2cBus) SDA() gpio.PinIO {
	return d.f.D1
}

func (d *i2cBus) setupI2C() error {
	// Initialize MPSSE to a known state.
	f := 100 * physic.KiloHertz
	clk := ((30 * physic.MegaHertz / f) - 1) * 2 / 3
	cmd := [...]byte{
		clock3Phase,
		dataTristate, 0x07, 0x00,
		clock30MHz, byte(clk), byte(clk >> 8),
	}
	d.f.usingI2C = true
	if _, err := d.f.h.write(cmd[:]); err != nil {
		return err
	}
	return d.setI2CLinesIdle()
}

func (d *i2cBus) stopI2C() error {
	cmd := [...]byte{
		clock2Phase,
		dataTristate, 0x00, 0x00,
		clock30MHz, 0, 0,
	}
	_, err := d.f.h.write(cmd[:])
	d.f.usingI2C = false
	return err
}

// setI2CLinesIdle sets all D0~D7 lines high.
func (d *i2cBus) setI2CLinesIdle() error {
	cmd := [...]byte{gpioSetD, 0xFF, 0xFB}
	_, err := d.f.h.write(cmd[:])
	return err
}

func (d *i2cBus) setI2CStart() error {
	// Runs the command 4 times as a way to delay execution.
	cmd := [...]byte{
		// SDA low
		gpioSetD, 0xFD, 0xFB,
		gpioSetD, 0xFD, 0xFB,
		gpioSetD, 0xFD, 0xFB,
		gpioSetD, 0xFD, 0xFB,
		// SCL low
		gpioSetD, 0xFC, 0xFB,
		gpioSetD, 0xFC, 0xFB,
		gpioSetD, 0xFC, 0xFB,
		gpioSetD, 0xFC, 0xFB,
	}
	_, err := d.f.h.write(cmd[:])
	return err
}

func (d *i2cBus) setI2CStop() error {
	// Runs the command 4 times as a way to delay execution.
	cmd := [...]byte{
		// SCL low
		gpioSetD, 0xFC, 0xFB,
		gpioSetD, 0xFC, 0xFB,
		gpioSetD, 0xFC, 0xFB,
		gpioSetD, 0xFC, 0xFB,
		// SDA low
		gpioSetD, 0xFD, 0xFB,
		gpioSetD, 0xFD, 0xFB,
		gpioSetD, 0xFD, 0xFB,
		gpioSetD, 0xFD, 0xFB,
		// SDA and SCL high
		gpioSetD, 0xFF, 0xFB,
		gpioSetD, 0xFF, 0xFB,
		gpioSetD, 0xFF, 0xFB,
		gpioSetD, 0xFF, 0xFB,
	}
	_, err := d.f.h.write(cmd[:])
	return err
}

func (d *i2cBus) writeBytes(w []byte) error {
	// TODO(maruel): Implement both with and without NAK check.
	var r [1]byte
	for i := range w {
		cmd := [...]byte{
			dataOut | dataOutFall, 0x00, 0x00, w[i],
			gpioSetD, 0xFE, 0xFb,
			dataIn | dataBit, 0x00,
			flush,
		}
		if _, err := d.f.h.write(cmd[:]); err != nil {
			return err
		}
		if _, err := d.f.h.read(r[:]); err != nil {
			return err
		}
		if r[0]&1 == 0 {
			return errors.New("got NAK")
		}
	}
	return nil
}

func (d *i2cBus) readBytes(r []byte) error {
	var ack byte
	for i := range r {
		if i == len(r)-1 {
			// NAK.
			ack = 0xFF
		}
		cmd := [...]byte{
			// Length 0 means one byte in.
			// TODO(maruel): dataBit.
			dataIn, 0x00, 0x00,
			dataOut | dataOutFall | dataBit, 0x00, ack,
			gpioSetD, 0xFE, 0xFB,
			// Force read buffer flush. This is only necessary if NAK are not ignored.
			flush,
		}
		if _, err := d.f.h.write(cmd[:]); err != nil {
			return err
		}
		// TODO(maruel): Create a buffer version.
		if _, err := d.f.h.read(r[i:1]); err != nil {
			return err
		}
	}
	return nil
}

var _ i2c.BusCloser = &i2cBus{}
var _ i2c.Pins = &i2cBus{}
