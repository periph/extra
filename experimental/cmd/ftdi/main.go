// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// ftdi prints out information about the FTDI devices found on the USB bus.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"periph.io/x/extra/experimental/devices/ftdi"
	"periph.io/x/extra/experimental/devices/ftdi/ftd2xx"
	"periph.io/x/periph/host"
)

func process(d ftdi.Dev) {
	var i ftdi.Info
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

	major, minor, build := ftd2xx.Driver.Version()
	fmt.Printf("Using library %d.%d.%d\n", major, minor, build)
	if err := ftdi.RegisterDriver(&ftd2xx.Driver); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	all := ftdi.All()
	plural := ""
	if len(all) > 1 {
		plural = "s"
	}
	fmt.Printf("Found %d device%s\n", len(all), plural)
	for _, d := range all {
		process(d)
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "ftdi: %s.\n", err)
		os.Exit(1)
	}
}
