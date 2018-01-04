// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftdi

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"periph.io/x/extra/experimental/devices/ftdi/ftd2xx"
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

// Dev represents one FT232H device.
//
// TODO(maruel): It will eventually be generic.
//
// There can be multiple devices connected to a host.
type Dev struct {
	id int
	h  ftd2xx.Handle

	C0 Pin // 21
	C1 Pin // 25
	C2 Pin // 26
	C3 Pin // 27
	C4 Pin // 28
	C5 Pin // 29
	C6 Pin // 30
	C7 Pin // 31
	C8 Pin // 32
	C9 Pin // 33
	D0 Pin // 13
	D1 Pin // 14
	D2 Pin // 15
	D3 Pin // 16
	D4 Pin // 17
	D5 Pin // 18
	D6 Pin // 19
	D7 Pin // 20
}

func (d *Dev) String() string {
	return fmt.Sprintf("ft232h(%d)", d.id)
}

// Halt implements conn.Resource.
func (d *Dev) Halt() error {
	return nil
}

func (d *Dev) Close() error {
	err := d.h.Close()
	d.h = 0
	return err
}

// DevInfo returns information about an opened device.
func (d *Dev) DevInfo(i *ftd2xx.DevInfo) error {
	return d.h.GetDevInfo(i)
}

// Ref is a reference to a FTDI device found on the USB bus but not necessarily
// yet opened.
type Ref struct {
	ID     uint32
	LocID  uint32
	Serial string
	Desc   string
}

// Open opens a FTDI device.
func (r *Ref) Open() (*Dev, error) {
	h, err := ftd2xx.OpenByLocation(r.LocID)
	if err != nil {
		return nil, err
	}
	return &Dev{h: h}, nil
}

// All enumerates all the connected FTDI devices.
func All() []Ref {
	mu.Lock()
	defer mu.Unlock()
	l, err := ftd2xx.ListDevices()
	if err != nil {
		log.Printf("ftdi.All(): %v", err)
		return nil
	}
	out := make([]Ref, 0, len(l))
	for _, v := range l {
		out = append(out, Ref{ID: v.ID, LocID: v.LocID, Serial: v.Serial, Desc: v.Desc})
	}
	return out
}

//

var (
	mu sync.Mutex
)

var _ gpio.PinIO = &Pin{}
