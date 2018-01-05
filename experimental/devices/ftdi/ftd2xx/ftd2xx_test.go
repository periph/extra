// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package ftd2xx

import (
	"fmt"
)

func Example() {
	// Print the ftd2xx driver version. It will be 0.0.0 if not found.
	major, minor, build := Driver.Version()
	fmt.Printf("Using library %d.%d.%d\n", major, minor, build)
}
