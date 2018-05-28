// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// d2xx-program programs a FTDI device.
//
// It can either program the EEPROM or the User Area.
package main

import (
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"periph.io/x/extra/hostextra/d2xx"
	"periph.io/x/periph/host"
)

func writeEEPROM(d d2xx.Dev, manufacturer, manufacturerID, desc, serial string) error {
	ee := d2xx.EEPROM{}
	if err := d.EEPROM(&ee); err != nil {
		fmt.Printf("Failed to read EEPROM: %v\n", err)
	}
	ee.Manufacturer = manufacturer
	ee.ManufacturerID = manufacturerID
	ee.Desc = desc
	ee.Serial = serial
	log.Printf("Writing: %x", ee.Raw)
	return d.WriteEEPROM(&ee)
}

func mainImpl() error {
	verbose := flag.Bool("v", false, "verbose mode")
	manufacturer := flag.String("m", "", "manufacturer")
	manufacturerID := flag.String("mid", "", "manufacturer ID")
	desc := flag.String("d", "", "description")
	serial := flag.String("s", "", "serial")
	ua := flag.String("ua", "", "hex encoded data")

	flag.Parse()
	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}
	log.SetFlags(log.Lmicroseconds)

	if flag.NArg() != 0 {
		return errors.New("unexpected argument, try -help")
	}
	if *ua == "" {
		if *manufacturer == "" || *manufacturerID == "" || *desc == "" || *serial == "" {
			return errors.New("all of -m, -mid, -d and -s are required, or use -ua")
		}
	} else {
		if *manufacturer != "" || *manufacturerID != "" || *desc != "" || *serial != "" {
			return errors.New("all of -m, -mid, -d and -s cannot be used with -ua")
		}
	}

	if _, err := host.Init(); err != nil {
		return err
	}
	major, minor, build := d2xx.Version()
	log.Printf("Using library %d.%d.%d\n", major, minor, build)

	all := d2xx.All()
	if len(all) == 0 {
		return errors.New("found no FTDI device on the USB bus")
	}
	if len(all) > 1 {
		return fmt.Errorf("for safety reasons, plug exactly one FTDI device on the USB bus, found %d devices", len(all))
	}
	d := all[0]

	if *ua == "" {
		return writeEEPROM(d, *manufacturer, *manufacturerID, *desc, *serial)
	}
	raw, err := hex.DecodeString(*ua)
	if err != nil {
		return err
	}
	return d.WriteUserArea(raw)
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "d2xx-program: %s.\n", err)
		os.Exit(1)
	}
}
