// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

package hostextra

import (
	_ "periph.io/x/extra/hostextra/d2xx"
	"periph.io/x/periph"
	"periph.io/x/periph/host"
)

// Init calls host.Init(), which calls periph.Init() and returns it as-is.
//
// The difference with host.Init() and periph.Init() is that hostextra.Init()
// includes more drivers, the drivers that either depend on third party
// packages or on cgo.
//
// Since host.Init() is used, all drivers in periph.io/x/periph/host are also
// automatically loaded.
func Init() (*periph.State, error) {
	return host.Init()
}
