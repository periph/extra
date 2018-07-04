// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package d2xx

import (
	"strconv"
	"sync"

	"periph.io/x/extra/hostextra/d2xx/ftdi"
	"periph.io/x/periph"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/conn/i2c/i2creg"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
	"periph.io/x/periph/conn/spi/spireg"
)

// All enumerates all the connected FTDI devices.
func All() []Dev {
	drv.mu.Lock()
	defer drv.mu.Unlock()
	out := make([]Dev, len(drv.all))
	copy(out, drv.all)
	return out
}

//

// open opens a FTDI device.
//
// Must be called with mu held.
func open(opener func(i int) (d2xxHandle, int), i int) (Dev, error) {
	h, err := openDev(opener, i)
	if err != nil {
		return nil, err
	}
	if err := h.setupCommon(); err != nil {
		// setupCommon() takes the device in its previous state. It could be in an
		// unexpected state, so try resetting it first.
		if err := h.reset(); err != nil {
			h.closeDev()
			return nil, err
		}
		if err := h.setupCommon(); err != nil {
			h.closeDev()
			return nil, err
		}
		// The second attempt worked.
	}
	// Makes a copy of the handle.
	g := generic{index: i, h: *h, name: h.t.String() + "(" + strconv.Itoa(i) + ")"}
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
	for _, p := range hdr {
		if err := gpioreg.Register(p); err != nil {
			return err
		}
	}

	raw := make([][]pin.Pin, len(hdr))
	for i := range hdr {
		raw[i] = []pin.Pin{hdr[i]}
	}
	if err := pinreg.Register(d.String(), raw); err != nil {
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
		// TODO(maruel): UART
	case *FT232R:
		// TODO(maruel): SPI, UART
	}
	return nil
}

// driver implements periph.Driver.
type driver struct {
	mu         sync.Mutex
	all        []Dev
	d2xxOpen   func(i int) (d2xxHandle, int)
	numDevices func() (int, error)
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
	num, err := d.numDevices()
	if err != nil {
		return true, err
	}
	for i := 0; i < num; i++ {
		// TODO(maruel): Close the device one day. :)
		if dev, err1 := open(d.d2xxOpen, i); err1 == nil {
			d.all = append(d.all, dev)
			if err := registerDev(dev); err != nil {
				return true, err
			}
		} else {
			// Create a shallow broken handle, so the user can learn how to fix the
			// problem.
			//
			// TODO(maruel): On macOS with a FT232R, calling two processes in a row
			// often results in a broken device on the second process. Figure out why
			// and make it more resilient.
			err = err1
			// The serial number is not available so what can be listed is limited.
			// TODO(maruel): Add VID/PID?
			name := "broken#" + strconv.Itoa(i) + ": " + err.Error()
			d.all = append(d.all, &broken{index: i, err: err, name: name})
		}
	}
	return true, err
}

func (d *driver) reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.all = nil
	// numDevices is mocked in tests.
	d.numDevices = numDevices
	// d2xxOpen is mocked in tests.
	d.d2xxOpen = d2xxOpen

	// The d2xx can hang for up to the timeout under certain circumstances and the
	// Go profiler is not very useful to find the source, so use manual logging
	// to see where time it spent.
	//d.d2xxOpen = func(i int) (d2xxHandle, int) {
	//	h, e := d2xxOpen(i)
	//	return &d2xxLoggingHandle{h}, e
	//}
}

func init() {
	if !disabled {
		drv.reset()
		periph.MustRegister(&drv)
	}
}

var drv driver
