// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// d2xx prints out information about the FTDI devices found on the USB bus.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"periph.io/x/extra/hostextra/d2xx"
	"periph.io/x/extra/hostextra/d2xx/ftdi"
	"periph.io/x/periph/host"
)

func process(d d2xx.Dev) {
	i := d2xx.Info{}
	d.Info(&i)
	fmt.Printf("  Type:           %s\n", i.Type)
	fmt.Printf("  Vendor ID:      %#04x\n", i.VenID)
	fmt.Printf("  Device ID:      %#04x\n", i.DevID)

	ee := ftdi.EEPROM{}
	if err := d.EEPROM(&ee); err == nil {
		fmt.Printf("  Manufacturer:   %s\n", ee.Manufacturer)
		fmt.Printf("  ManufacturerID: %s\n", ee.ManufacturerID)
		fmt.Printf("  Desc:           %s\n", ee.Desc)
		fmt.Printf("  Serial:         %s\n", ee.Serial)

		h := ee.AsHeader()
		fmt.Printf("  MaxPower:       %dmA\n", h.MaxPower)
		fmt.Printf("  SelfPowered:    %x\n", h.SelfPowered)
		fmt.Printf("  RemoteWakeup:   %x\n", h.RemoteWakeup)
		fmt.Printf("  PullDownEnable: %x\n", h.PullDownEnable)
		switch i.Type {
		case "FT232H":
			p := ee.AsFT232H()
			fmt.Printf("  CSlowSlew:      %d\n", p.ACSlowSlew)
			fmt.Printf("  CSchmittInput:  %d\n", p.ACSchmittInput)
			fmt.Printf("  CDriveCurrent:  %dmA\n", p.ACDriveCurrent)
			fmt.Printf("  DSlowSlew:      %d\n", p.ADSlowSlew)
			fmt.Printf("  DSchmittInput:  %d\n", p.ADSchmittInput)
			fmt.Printf("  DDriveCurrent:  %dmA\n", p.ADDriveCurrent)
			fmt.Printf("  Cbus0:          %s\n", p.Cbus0)
			fmt.Printf("  Cbus1:          %s\n", p.Cbus1)
			fmt.Printf("  Cbus2:          %s\n", p.Cbus2)
			fmt.Printf("  Cbus3:          %s\n", p.Cbus3)
			fmt.Printf("  Cbus4:          %s\n", p.Cbus4)
			fmt.Printf("  Cbus5:          %s\n", p.Cbus5)
			fmt.Printf("  Cbus6:          %s\n", p.Cbus6)
			fmt.Printf("  Cbus7:          %s\n", p.Cbus7)
			fmt.Printf("  Cbus8:          %s\n", p.Cbus8)
			fmt.Printf("  Cbus9:          %s\n", p.Cbus9)
			fmt.Printf("  FT1248Cpol:     %d\n", p.FT1248Cpol)
			fmt.Printf("  FT1248Lsb:      %d\n", p.FT1248Lsb)
			fmt.Printf("  FT1248FlowCtrl: %d\n", p.FT1248FlowControl)
			fmt.Printf("  IsFifo:         %d\n", p.IsFifo)
			fmt.Printf("  IsFifoTar:      %d\n", p.IsFifoTar)
			fmt.Printf("  IsFastSer:      %d\n", p.IsFastSer)
			fmt.Printf("  IsFT1248:       %d\n", p.IsFT1248)
			fmt.Printf("  PowerSaveEnabl: %d\n", p.PowerSaveEnable)
			fmt.Printf("  DriverType:     %d\n", p.DriverType)
		case "FT232R":
			p := ee.AsFT232R()
			fmt.Printf("  IsHighCurrent:  %d\n", p.IsHighCurrent)
			fmt.Printf("  UseExtOsc:      %d\n", p.UseExtOsc)
			fmt.Printf("  InvertTXD:      %d\n", p.InvertTXD)
			fmt.Printf("  InvertRXD:      %d\n", p.InvertRXD)
			fmt.Printf("  InvertRTS:      %d\n", p.InvertRTS)
			fmt.Printf("  InvertCTS:      %d\n", p.InvertCTS)
			fmt.Printf("  InvertDTR:      %d\n", p.InvertDTR)
			fmt.Printf("  InvertDSR:      %d\n", p.InvertDSR)
			fmt.Printf("  InvertDCD:      %d\n", p.InvertDCD)
			fmt.Printf("  InvertRI:       %d\n", p.InvertRI)
			fmt.Printf("  Cbus0:          %s\n", p.Cbus0)
			fmt.Printf("  Cbus1:          %s\n", p.Cbus1)
			fmt.Printf("  Cbus2:          %s\n", p.Cbus2)
			fmt.Printf("  Cbus3:          %s\n", p.Cbus3)
			fmt.Printf("  Cbus4:          %s\n", p.Cbus4)
			fmt.Printf("  DriverType:     %d\n", p.DriverType)
		default:
			fmt.Printf("Unknown device:   %s\n", i.Type)
		}
		log.Printf("  Raw: %x\n", ee.Raw)
	} else {
		fmt.Printf("Failed to read EEPROM: %v\n", err)
	}

	if ua, err := d.UserArea(); err != nil {
		fmt.Printf("Failed to read UserArea: %v\n", err)
	} else {
		fmt.Printf("UserArea: %x\n", ua)
	}

	hdr := d.Header()
	for _, p := range hdr {
		fmt.Printf("%s: %s\n", p, p.Function())
	}
}

func mainImpl() error {
	verbose := flag.Bool("v", false, "verbose mode")
	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)
	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}

	if _, err := host.Init(); err != nil {
		return err
	}

	major, minor, build := d2xx.Version()
	fmt.Printf("Using library %d.%d.%d\n", major, minor, build)
	all := d2xx.All()
	plural := ""
	if len(all) > 1 {
		plural = "s"
	}
	fmt.Printf("Found %d device%s\n", len(all), plural)
	for i, d := range all {
		fmt.Printf("- Device #%d\n", i)
		process(d)
		if i != len(all)-1 {
			fmt.Printf("\n")
		}
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "d2xx: %s.\n", err)
		os.Exit(1)
	}
}
