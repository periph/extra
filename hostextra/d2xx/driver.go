// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package d2xx

import (
	"sync"

	"periph.io/x/extra/hostextra/d2xx/ftdi"
	"periph.io/x/periph"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/conn/spi/spireg"
)

// All enumerates all the connected FTDI devices.
func All() []Dev {
	mu.Lock()
	defer mu.Unlock()
	out := make([]Dev, len(all))
	copy(out, all)
	return out
}

//

var (
	mu  sync.Mutex
	all []Dev
)

// open opens a FTDI device.
//
// Must be called with mu held.
func open(i int) (Dev, error) {
	h, err := openDev(i)
	if err != nil {
		return nil, err
	}
	if err := h.setupCommon(); err != nil {
		h.closeDev()
		return nil, err
	}
	// Makes a copy of the handle.
	g := generic{index: i, h: *h}
	// Makes a copy of the generic instance.
	switch g.h.t {
	case ftdi.FT232H:
		f, err := newFT232H(g)
		if err != nil {
			h.closeDev()
			return nil, err
		}
		return f, nil
	case ftdi.FT232R:
		f, err := newFT232R(g)
		if err != nil {
			h.closeDev()
			return nil, err
		}
		return f, nil
	default:
		return &g, nil
	}
}

// registerDev registers the header and supported buses and ports in the
// relevant registries.
func registerDev(d Dev) error {
	hdr := d.Header()
	p := make([][]pin.Pin, len(hdr))
	for i := range hdr {
		p[i] = []pin.Pin{hdr[i]}
	}
	if err := pinreg.Register(d.String(), p); err != nil {
		return err
	}
	switch t := d.(type) {
	case *FT232H:
		if err := i2creg.Register(d.String(), nil, -1, t.I2C); err != nil {
			return err
		}
		if err := spireg.Register(d.String(), nil, -1, t.SPI); err != nil {
			return err
		}
	}
	return nil
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "d2xx"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) After() []string {
	return nil
}

func (d *driver) Init() (bool, error) {
	num, err := numDevices()
	if err != nil {
		return true, err
	}
	for i := 0; i < num; i++ {
		// TODO(maruel): Close the device one day. :)
		if d, err1 := open(i); err1 == nil {
			all = append(all, d)
			if err := registerDev(d); err != nil {
				return true, err
			}
		} else {
			// Create a shallow broken handle, so the user can learn how to fix the
			// problem.
			err = err1
			all = append(all, &broken{index: i, err: err})
		}
	}
	return true, err
}

func init() {
	if !disabled {
		periph.MustRegister(&driver{})
	}
}

var _ periph.Driver = &driver{}
