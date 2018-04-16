// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package winthermal reads the temperature sensors via WMI on Windows.
//
// This is an incomplete work-in-progress.
package winthermal

import (
	"periph.io/x/periph"
	"periph.io/x/periph/devices"
)

type Dev struct {
	h obj
}

func (d *Dev) Sense(env *devices.Environment) error {
	env.Temperature = devices.Celsius(d.h.CurrentTemperature)
	return nil
}

//

// obj represents a MSAcpi_ThermalZoneTemperature instance. It intentionally
// leaves a lot of members out.
type obj struct {
	CurrentTemperature int
	InstanceName       string
	SamplingPeriod     int
}

type driver struct {
}

func (d *driver) String() string {
	return "winthermal"
}

func (d *driver) Prerequisites() []string {
	return nil
}

func (d *driver) Init() (bool, error) {
	return true, initWindows()
}

var _ periph.Driver = &driver{}
