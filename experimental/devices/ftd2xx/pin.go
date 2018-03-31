// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

import (
	"errors"
	"fmt"
	"time"

	"periph.io/x/periph/conn/gpio"
)

// Pin is a pin on a FTDI device.
//
// Pin implements gpio.PinIO.
type Pin struct {
	f   string
	n   string
	num int
}

// String implements pin.Pin.
func (p *Pin) String() string {
	return p.n
}

// Name implements pin.Pin.
func (p *Pin) Name() string {
	return p.n
}

// Number implements pin.Pin.
func (p *Pin) Number() int {
	return p.num
}

// Function implements pin.Pin.
func (p *Pin) Function() string {
	return p.f
}

// In implements gpio.PinIn.
func (p *Pin) In(pull gpio.Pull, e gpio.Edge) error {
	return errors.New("ft232h: to be implemented")
}

// Read implements gpio.PinIn.
func (p *Pin) Read() gpio.Level {
	return gpio.Low
}

// WaitForEdge implements gpio.PinIn.
func (p *Pin) WaitForEdge(t time.Duration) bool {
	return false
}

// Pull implements gpio.PinIn.
func (p *Pin) Pull() gpio.Pull {
	return gpio.PullNoChange
}

// Out implements gpio.PinOut.
func (p *Pin) Out(l gpio.Level) error {
	return errors.New("ft232h: to be implemented")
}

var _ fmt.Stringer = &Pin{}
var _ gpio.PinIO = &Pin{}
