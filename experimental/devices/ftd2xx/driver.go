// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

import (
	"sync"

	"periph.io/x/periph"
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

var (
	mu  sync.Mutex
	all []Dev
)

// open opens a FTDI device.
//
// Must be called with mu held.
func open(i int) (Dev, error) {
	h, err := openH(i)
	if err != nil {
		return nil, err
	}
	var info Info
	h.getI(&info)
	g := generic{index: i, h: h, info: info}
	switch info.Type {
	case "ft232h":
		return &FT232H{
			generic: g,
			C0:      Pin{num: 0, n: "C0"},
			C1:      Pin{num: 1, n: "C1"},
			C2:      Pin{num: 2, n: "C2"},
			C3:      Pin{num: 3, n: "C3"},
			C4:      Pin{num: 4, n: "C4"},
			C5:      Pin{num: 5, n: "C5"},
			C6:      Pin{num: 6, n: "C6"},
			C7:      Pin{num: 7, n: "C7"},
			C8:      Pin{num: 8, n: "C8"},
			C9:      Pin{num: 9, n: "C9"},
			D0:      Pin{num: 10, n: "D0"},
			D1:      Pin{num: 11, n: "D1"},
			D2:      Pin{num: 12, n: "D2"},
			D3:      Pin{num: 13, n: "D3"},
			D4:      Pin{num: 14, n: "D4"},
			D5:      Pin{num: 15, n: "D5"},
			D6:      Pin{num: 16, n: "D6"},
			D7:      Pin{num: 17, n: "D7"},
		}, nil
	case "ft232r":
		return &FT232R{
			generic: g,
			TX:      Pin{num: 0, n: "TX"},
			RX:      Pin{num: 1, n: "RX"},
			RTS:     Pin{num: 2, n: "RTS"},
			CTS:     Pin{num: 3, n: "CTS"},
			DTR:     Pin{num: 4, n: "DTR"},
			DSR:     Pin{num: 5, n: "DSR"},
			DCD:     Pin{num: 6, n: "DCD"},
			RI:      Pin{num: 7, n: "R:"},
		}, nil
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
		if d, err1 := open(i); err1 == nil {
			all = append(all, d)
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
