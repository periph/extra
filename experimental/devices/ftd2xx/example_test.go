// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx_test

import (
	"fmt"
	"log"

	"periph.io/x/extra/experimental/devices/ftd2xx"
	"periph.io/x/periph/host"
)

func Example() {
	if _, err := host.Init(); err != nil {
		log.Fatal(err)
	}
	for _, d := range ftd2xx.All() {
		fmt.Printf("%s\n", d)
	}
}

func ExampleVersion() {
	// Print the ftd2xx driver version. It will be 0.0.0 if not found.
	major, minor, build := ftd2xx.Version()
	fmt.Printf("Using library %d.%d.%d\n", major, minor, build)
}
