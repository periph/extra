// Copyright 2016 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftdi

import (
	"fmt"
	"log"
	"testing"

	"periph.io/x/periph/host"
)

func Example_All() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	for _, d := range All() {
		fmt.Printf("%s\n", d)
	}
}

func TestFT232H(t *testing.T) {
}
