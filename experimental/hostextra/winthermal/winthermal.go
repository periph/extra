// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package winthermal reads the temperature sensors via WMI on Windows.
//
// This is an incomplete work-in-progress.
package winthermal

import (
	"errors"
	"time"

	"periph.io/x/periph"
	"periph.io/x/periph/devices"
)

// Dev represents an handle to a WMI based sensor.
//
// Dev implements devices.Environmental.
type Dev struct {
	h obj
}

// Halt implements conn.Resource.
func (d *Dev) Halt() error {
	return nil
}

// Sense implements devices.Environmental.
func (d *Dev) Sense(env *devices.Environment) error {
	env.Temperature = devices.Celsius(d.h.CurrentTemperature)
	return nil
}

func (d *Dev) SenseContinuous(interval time.Duration) (<-chan devices.Environment, error) {
	return nil, errors.New("winthermal: not implemented yet")
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
var _ devices.Environmental = &Dev{}
