// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package hostextra

import "periph.io/x/periph"

// Init calls periph.Init() and returns it as-is.
//
// The difference with host.Init() and periph.Init() is that hostextra.Init()
// includes more drivers, the drivers that either depend on third party
// packages or on cgo.
func Init() (*periph.State, error) {
	return periph.Init()
}
