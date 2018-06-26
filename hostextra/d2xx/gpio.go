// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Emulate independent GPIOs.

package d2xx

import (
	"errors"
	"time"

	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/physic"
)

// http://www.ftdichip.com/Support/Documents/AppNotes/AN_232R-01_Bit_Bang_Mode_Available_For_FT232R_and_Ft245R.pdf

// syncBus is the handler of a synchronous bitbang bus.
type syncBus interface {
	syncBusFunc(n int) string
	syncBusIn(n int) error
	syncBusRead(n int) gpio.Level
	syncBusOut(n int, l gpio.Level) error
}

// syncPin represents a GPIO on a synchronous bitbang bus. It is stateless.
type syncPin struct {
	n   string
	num int
	bus syncBus
}

// String implements conn.Resource.
func (s *syncPin) String() string {
	return s.n
}

// Halt implements conn.Resource.
func (s *syncPin) Halt() error {
	return nil
}

// Name implements pin.Pin.
func (s *syncPin) Name() string {
	return s.n
}

// Number implements pin.Pin.
func (s *syncPin) Number() int {
	return s.num
}

// Function implements pin.Pin.
func (s *syncPin) Function() string {
	return s.bus.syncBusFunc(s.num)
}

// In implements gpio.PinIn.
func (s *syncPin) In(pull gpio.Pull, e gpio.Edge) error {
	if e != gpio.NoEdge {
		// We could support it on D5.
		return errors.New("d2xx: edge triggering is not supported")
	}
	if pull != gpio.PullUp && pull != gpio.PullNoChange {
		// EEPROM has a PullDownEnable flag.
		return errors.New("d2xx: pull is not supported")
	}
	return s.bus.syncBusIn(s.num)
}

// Read implements gpio.PinIn.
func (s *syncPin) Read() gpio.Level {
	return s.bus.syncBusRead(s.num)
}

// WaitForEdge implements gpio.PinIn.
func (s *syncPin) WaitForEdge(t time.Duration) bool {
	return false
}

// DefaultPull implements gpio.PinIn.
func (s *syncPin) DefaultPull() gpio.Pull {
	// 200kΩ
	// http://www.ftdichip.com/Support/Documents/DataSheets/ICs/DS_FT232R.pdf
	// p. 24
	return gpio.PullUp
}

// Pull implements gpio.PinIn.
func (s *syncPin) Pull() gpio.Pull {
	return gpio.PullUp
}

// Out implements gpio.PinOut.
func (s *syncPin) Out(l gpio.Level) error {
	return s.bus.syncBusOut(s.num, l)
}

// PWM implements gpio.PinOut.
func (s *syncPin) PWM(d gpio.Duty, f physic.Frequency) error {
	return errors.New("d2xx: not implemented")
}

/*
func (s *syncPin) Drive() physic.ElectricCurrent {
	// optionally 3
	//return s.bus.ee.DDriveCurrent * physic.MilliAmpere
	return physic.MilliAmpere
}

func (s *syncPin) SlewLimit() bool {
	//return s.bus.ee.DSlowSlew
	return false
}

func (s *syncPin) Hysteresis() bool {
	//return s.bus.ee.DSchmittInput
	return true
}
*/

//

// cBus is the handler of a CBus bitbang bus.
type cBus interface {
	cBusFunc(n int) string
	cBusIn(n int) error
	cBusRead(n int) gpio.Level
	cBusOut(n int, l gpio.Level) error
}

// cbusPin represents a GPIO on a CBus bitbang bus. It is stateless.
type cbusPin struct {
	n   string
	num int
	p   gpio.Pull
	bus cBus
}

// String implements conn.Resource.
func (c *cbusPin) String() string {
	return c.n
}

// Halt implements conn.Resource.
func (c *cbusPin) Halt() error {
	return nil
}

// Name implements pin.Pin.
func (c *cbusPin) Name() string {
	return c.n
}

// Number implements pin.Pin.
func (c *cbusPin) Number() int {
	return c.num
}

// Function implements pin.Pin.
func (c *cbusPin) Function() string {
	return c.bus.cBusFunc(c.num)
}

// In implements gpio.PinIn.
func (c *cbusPin) In(pull gpio.Pull, e gpio.Edge) error {
	if e != gpio.NoEdge {
		// We could support it on D5.
		return errors.New("d2xx: edge triggering is not supported")
	}
	if pull != c.p && pull != gpio.PullNoChange {
		// EEPROM has a PullDownEnable flag.
		return errors.New("d2xx: pull is not supported")
	}
	return c.bus.cBusIn(c.num)
}

// Read implements gpio.PinIn.
func (c *cbusPin) Read() gpio.Level {
	return c.bus.cBusRead(c.num)
}

// WaitForEdge implements gpio.PinIn.
func (c *cbusPin) WaitForEdge(t time.Duration) bool {
	return false
}

// DefaultPull implements gpio.PinIn.
func (c *cbusPin) DefaultPull() gpio.Pull {
	// 200kΩ
	// http://www.ftdichip.com/Support/Documents/DataSheets/ICs/DS_FT232R.pdf
	// p. 24
	return c.p
}

// Pull implements gpio.PinIn.
func (c *cbusPin) Pull() gpio.Pull {
	return c.p
}

// Out implements gpio.PinOut.
func (c *cbusPin) Out(l gpio.Level) error {
	return c.bus.cBusOut(c.num, l)
}

// PWM implements gpio.PinOut.
func (c *cbusPin) PWM(d gpio.Duty, f physic.Frequency) error {
	return errors.New("d2xx: not implemented")
}

/*
func (c *cbusPin) Drive() physic.ElectricCurrent {
	// optionally 3
	//return c.bus.ee.CDriveCurrent * physic.MilliAmpere
	return physic.MilliAmpere
}

func (c *cbusPin) SlewLimit() bool {
	//return c.bus.ee.CSlowSlew
	return false
}

func (c *cbusPin) Hysteresis() bool {
	//return c.bus.ee.CSchmittInput
	return true
}
*/

var _ gpio.PinIO = &syncPin{}
var _ gpio.PinIO = &cbusPin{}
