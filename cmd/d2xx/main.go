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
	"periph.io/x/periph/host"
)

func process(d d2xx.Dev) {
	i := d2xx.Info{}
	d.Info(&i)
	fmt.Printf("  Type:           %s\n", i.Type)
	fmt.Printf("  Vendor ID:      %#04x\n", i.VenID)
	fmt.Printf("  Device ID:      %#04x\n", i.DevID)
	ee := d2xx.EEPROM{}
	if err := d.EEPROM(&ee); err == nil {
		fmt.Printf("  Manufacturer:   %s\n", ee.Manufacturer)
		fmt.Printf("  ManufacturerID: %s\n", ee.ManufacturerID)
		fmt.Printf("  Desc:           %s\n", ee.Desc)
		fmt.Printf("  Serial:         %s\n", ee.Serial)

		p := d2xx.ProcessedEEPROM{}
		ee.Interpret(i.Type, &p)
		fmt.Printf("  MaxPower:       %dmA\n", p.MaxPower)
		fmt.Printf("  SelfPowered:    %t\n", p.SelfPowered)
		fmt.Printf("  RemoteWakeup:   %t\n", p.RemoteWakeup)
		fmt.Printf("  PullDownEnable: %t\n", p.PullDownEnable)
		switch i.Type {
		case "FT232H":
			fmt.Printf("  CSlowSlew:      %t\n", p.CSlowSlew)
			fmt.Printf("  CSchmittInput:  %t\n", p.CSchmittInput)
			fmt.Printf("  CDriveCurrent:  %dmA\n", p.CDriveCurrent)
			fmt.Printf("  DSlowSlew:      %t\n", p.DSlowSlew)
			fmt.Printf("  DSchmittInput:  %t\n", p.DSchmittInput)
			fmt.Printf("  DDriveCurrent:  %dmA\n", p.DDriveCurrent)
			fmt.Printf("  Cbus0:          %#02x\n", p.Cbus0)
			fmt.Printf("  Cbus1:          %#02x\n", p.Cbus1)
			fmt.Printf("  Cbus2:          %#02x\n", p.Cbus2)
			fmt.Printf("  Cbus3:          %#02x\n", p.Cbus3)
			fmt.Printf("  Cbus4:          %#02x\n", p.Cbus4)
			fmt.Printf("  Cbus5:          %#02x\n", p.Cbus5)
			fmt.Printf("  Cbus6:          %#02x\n", p.Cbus6)
			fmt.Printf("  Cbus7:          %#02x\n", p.Cbus7)
			fmt.Printf("  Cbus8:          %#02x\n", p.Cbus8)
			fmt.Printf("  Cbus9:          %#02x\n", p.Cbus9)
			fmt.Printf("  FT1248Cpol:     %t\n", p.FT1248Cpol)
			fmt.Printf("  FT1248Lsb:      %t\n", p.FT1248Lsb)
			fmt.Printf("  FT1248FlowCtrl: %t\n", p.FT1248FlowControl)
			fmt.Printf("  IsFifo:         %t\n", p.IsFifo)
			fmt.Printf("  IsFifoTar:      %t\n", p.IsFifoTar)
			fmt.Printf("  IsFastSer:      %t\n", p.IsFastSer)
			fmt.Printf("  IsFT1248:       %t\n", p.IsFT1248)
			fmt.Printf("  PowerSaveEnabl: %t\n", p.PowerSaveEnable)
			fmt.Printf("  DriverType:     %#02x\n", p.DriverType)
		case "FT232R":
			fmt.Printf("  IsHighCurrent:  %t\n", p.IsHighCurrent)
			fmt.Printf("  UseExtOsc:      %t\n", p.UseExtOsc)
			fmt.Printf("  InvertTXD:      %t\n", p.InvertTXD)
			fmt.Printf("  InvertRXD:      %t\n", p.InvertRXD)
			fmt.Printf("  InvertRTS:      %t\n", p.InvertRTS)
			fmt.Printf("  InvertCTS:      %t\n", p.InvertCTS)
			fmt.Printf("  InvertDTR:      %t\n", p.InvertDTR)
			fmt.Printf("  InvertDSR:      %t\n", p.InvertDSR)
			fmt.Printf("  InvertDCD:      %t\n", p.InvertDCD)
			fmt.Printf("  InvertRI:       %t\n", p.InvertRI)
			fmt.Printf("  Cbus0:          %#02x\n", p.Cbus0)
			fmt.Printf("  Cbus1:          %#02x\n", p.Cbus1)
			fmt.Printf("  Cbus2:          %#02x\n", p.Cbus2)
			fmt.Printf("  Cbus3:          %#02x\n", p.Cbus3)
			fmt.Printf("  Cbus4:          %#02x\n", p.Cbus4)
			fmt.Printf("  DriverType:     %#02x\n", p.DriverType)
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
