// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package hostextra defines the extra drivers for the host itself.
//
// The host is the machine where this code is running.
//
// Subpackages contain the drivers that are loaded automatically. Contrary to
// periph.io/x/periph/host, hostextra loads drivers that depends on either third
// party Go packages and/or on cgo.
package hostextra
