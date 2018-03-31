// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// ftd2xx prints out information about the FTDI devices found on the USB bus.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"periph.io/x/extra/experimental/devices/ftd2xx"
	"periph.io/x/periph/host"
)

func process(d ftd2xx.Dev) {
	i := ftd2xx.Info{}
	d.GetInfo(&i)
	fmt.Printf("  Type:           %s\n", i.Type)
	fmt.Printf("  Vendor ID:      %#04x\n", i.VenID)
	fmt.Printf("  Product ID:     %#04x\n", i.ProductID)
	fmt.Printf("  Manufacturer:   %s\n", i.Manufacturer)
	fmt.Printf("  ManufacturerID: %s\n", i.ManufacturerID)
	fmt.Printf("  Desc:           %s\n", i.Desc)
	fmt.Printf("  Serial:         %s\n", i.Serial)
	fmt.Printf("  MaxPower:       %dmA\n", i.MaxPower)
	fmt.Printf("  SelfPowered:    %t\n", i.SelfPowered)
	fmt.Printf("  RemoteWakeup:   %t\n", i.RemoteWakeup)
	fmt.Printf("  PullDownEnable: %t\n", i.PullDownEnable)
	switch i.Type {
	case "ft232h":
		fmt.Printf("  CSlowSlew:      %t\n", i.CSlowSlew)
		fmt.Printf("  CSchmittInput:  %t\n", i.CSchmittInput)
		fmt.Printf("  CDriveCurrent:  %dmA\n", i.CDriveCurrent)
		fmt.Printf("  DSlowSlew:      %t\n", i.DSlowSlew)
		fmt.Printf("  DSchmittInput:  %t\n", i.DSchmittInput)
		fmt.Printf("  DDriveCurrent:  %dmA\n", i.DDriveCurrent)
		fmt.Printf("  Cbus0:          %#02x\n", i.Cbus0)
		fmt.Printf("  Cbus1:          %#02x\n", i.Cbus1)
		fmt.Printf("  Cbus2:          %#02x\n", i.Cbus2)
		fmt.Printf("  Cbus3:          %#02x\n", i.Cbus3)
		fmt.Printf("  Cbus4:          %#02x\n", i.Cbus4)
		fmt.Printf("  Cbus5:          %#02x\n", i.Cbus5)
		fmt.Printf("  Cbus6:          %#02x\n", i.Cbus6)
		fmt.Printf("  Cbus7:          %#02x\n", i.Cbus7)
		fmt.Printf("  Cbus8:          %#02x\n", i.Cbus8)
		fmt.Printf("  Cbus9:          %#02x\n", i.Cbus9)
		fmt.Printf("  FT1248Cpol:     %t\n", i.FT1248Cpol)
		fmt.Printf("  FT1248Lsb:      %t\n", i.FT1248Lsb)
		fmt.Printf("  FT1248FlowCtrl: %t\n", i.FT1248FlowControl)
		fmt.Printf("  IsFifo:         %t\n", i.IsFifo)
		fmt.Printf("  IsFifoTar:      %t\n", i.IsFifoTar)
		fmt.Printf("  IsFastSer:      %t\n", i.IsFastSer)
		fmt.Printf("  IsFT1248:       %t\n", i.IsFT1248)
		fmt.Printf("  PowerSaveEnabl: %t\n", i.PowerSaveEnable)
		fmt.Printf("  DriverType:     %#02x\n", i.DriverType)
	case "ft232r":
		fmt.Printf("  IsHighCurrent:  %t\n", i.IsHighCurrent)
		fmt.Printf("  UseExtOsc:      %t\n", i.UseExtOsc)
		fmt.Printf("  InvertTXD:      %t\n", i.InvertTXD)
		fmt.Printf("  InvertRXD:      %t\n", i.InvertRXD)
		fmt.Printf("  InvertRTS:      %t\n", i.InvertRTS)
		fmt.Printf("  InvertCTS:      %t\n", i.InvertCTS)
		fmt.Printf("  InvertDTR:      %t\n", i.InvertDTR)
		fmt.Printf("  InvertDSR:      %t\n", i.InvertDSR)
		fmt.Printf("  InvertDCD:      %t\n", i.InvertDCD)
		fmt.Printf("  InvertRI:       %t\n", i.InvertRI)
		fmt.Printf("  Cbus0:          %#02x\n", i.Cbus0)
		fmt.Printf("  Cbus1:          %#02x\n", i.Cbus1)
		fmt.Printf("  Cbus2:          %#02x\n", i.Cbus2)
		fmt.Printf("  Cbus3:          %#02x\n", i.Cbus3)
		fmt.Printf("  Cbus4:          %#02x\n", i.Cbus4)
		fmt.Printf("  DriverType:     %#02x\n", i.DriverType)
	default:
	}
	log.Printf("  Full struct:\n%#v\n", i)
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

	major, minor, build := ftd2xx.Version()
	fmt.Printf("Using library %d.%d.%d\n", major, minor, build)
	all := ftd2xx.All()
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
		fmt.Fprintf(os.Stderr, "ftd2xx: %s.\n", err)
		os.Exit(1)
	}
}
