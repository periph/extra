// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package d2xxsmoketest is leveraged by extra-smoketest to verify that a
// FT232H/FT232R is working as expectd.
package d2xxsmoketest

import (
	"errors"
	"flag"
	"fmt"
	"time"

	"periph.io/x/extra/hostextra/d2xx"
	"periph.io/x/periph/conn/gpio"
)

// SmokeTest is imported by extra-smoketest.
type SmokeTest struct {
}

// Name implements the SmokeTest interface.
func (s *SmokeTest) Name() string {
	return "d2xx"
}

// Description implements the SmokeTest interface.
func (s *SmokeTest) Description() string {
	return "Tests FT232H/FT232R"
}

// Run implements the SmokeTest interface.
func (s *SmokeTest) Run(f *flag.FlagSet, args []string) (err error) {
	devType := f.String("type", "", "Device type to test, i.e. ft232h or ft232r")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		f.Usage()
		return errors.New("unrecognized arguments")
	}

	all := d2xx.All()
	if len(all) != 1 {
		return fmt.Errorf("exactly one device is expected, got %d", len(all))
	}
	dev := all[0]
	switch *devType {
	case "ft232h":
		if d, ok := dev.(*d2xx.FT232H); ok {
			return testFT232H(d)
		}
	case "ft232r":
		if d, ok := dev.(*d2xx.FT232R); ok {
			return testFT232R(d)
		}
	case "":
		return errors.New("-type is required")
	default:
		return errors.New("unrecognized -type, only ft232h and ft232r are supported")
	}
	return fmt.Errorf("expected %s, got %T; %s", *devType, dev, dev.String())
}

func testFT232H(d *d2xx.FT232H) error {
	// TODO(maruel): Read EEPROM, connected wires?.
	return gpioTest(d.C7)
}

func testFT232R(d *d2xx.FT232R) error {
	// TODO(maruel): Read EEPROM, connected wires?.
	return gpioTest(d.D3)
}

// gpioTest reads and write in a tight loop to evaluate performance. This makes
// sure that the flush operation is used, vs relying on SetLatencyTimer value.
func gpioTest(p gpio.PinIO) error {
	fmt.Printf("  Testing GPIO performance:\n")
	const loops = 1000
	fmt.Printf("    %d reads:  ", loops)
	if err := p.In(gpio.PullNoChange, gpio.NoEdge); err != nil {
		return err
	}
	start := time.Now()
	for i := 0; i < loops; i++ {
		p.Read()
	}
	s := time.Since(start)
	fmt.Printf("%s; %s/op\n", s, s/loops)
	fmt.Printf("    %d writes: ", loops)
	if err := p.Out(gpio.Low); err != nil {
		return err
	}
	start = time.Now()
	for i := 0; i < loops; i++ {
		if err := p.Out(gpio.Low); err != nil {
			return err
		}
	}
	s = time.Since(start)
	fmt.Printf("%s; %s/op\n", s, s/loops)
	return nil
}
