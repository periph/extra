// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

import (
	"sync"

	"periph.io/x/periph"
	"periph.io/x/periph/conn/pin"
	"periph.io/x/periph/conn/pin/pinreg"
)

// All enumerates all the connected FTDI devices.
//
// Some may not be opened; they may already be opened by another process or by
// a driver included by the operating system.
//
// See https://github.com/periph/extra/tree/master/experimental/devices/ftd2xx.
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
	var info Info
	h.getI(&info)
	g := generic{index: i, h: h, info: info}
	switch info.Type {
	case "ft232h":
		return newFT232H(g), nil
	case "ft232r":
		return newFT232R(g), nil
	default:
		return &g, nil
	}
}

// driver implements periph.Driver.
type driver struct {
}

func (d *driver) String() string {
	return "ftd2xx"
}

func (d *driver) Prerequisites() []string {
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
			hdr := d.Header()
			p := make([][]pin.Pin, len(hdr))
			for i := range hdr {
				p[i] = []pin.Pin{hdr[i]}
			}
			pinreg.Register(d.String(), p)
		} else {
			// Create a shallow generic handle.
			err = err1
			all = append(all, &generic{index: i})
		}
	}
	return true, err
}

func init() {
	periph.MustRegister(&driver{})
}

var _ periph.Driver = &driver{}
