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

func process(r *ftdi.Ref) {
	fmt.Printf("%#v\n", r)
	h, err := r.Open()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}
	fmt.Printf("%#v\n", h)
	var d ftd2xx.DevInfo
	if err = h.DevInfo(&d); err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	fmt.Printf("%#v\n", d)
	if err = h.Close(); err != nil {
		fmt.Printf("ERROR: %v\n", err)
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
	all := ftdi.All()
	plural := ""
	if len(all) > 1 {
		plural = "s"
	}
	fmt.Printf("Found %d device%s\n", len(all), plural)
	for _, r := range all {
		process(&r)
	}
	return nil
}

func main() {
	if err := mainImpl(); err != nil {
		fmt.Fprintf(os.Stderr, "ftdi: %s.\n", err)
		os.Exit(1)
	}
}
