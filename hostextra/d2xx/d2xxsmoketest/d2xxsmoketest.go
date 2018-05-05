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

	"periph.io/x/extra/hostextra/d2xx"
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
		return fmt.Errorf("expected %s, got %T", *devType, dev)
	case "ft232r":
		if d, ok := dev.(*d2xx.FT232R); ok {
			return testFT232R(d)
		}
		return fmt.Errorf("expected %s, got %T", *devType, dev)
	case "":
		return errors.New("-type is required")
	default:
		return errors.New("unrecognized -type, only ft232h and ft232r are supported")
	}
}

func testFT232H(d *d2xx.FT232H) error {
	// TODO(maruel): Assert registries, connected wires?.
	return nil
}

func testFT232R(d *d2xx.FT232R) error {
	// TODO(maruel): Assert registries, connected wires?.
	return nil
}
