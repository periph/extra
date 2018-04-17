// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build cgo
// +build !darwin,!amd64
// +build !linux,!amd64
// +build !linux,!arm
// +build !windows

package d2xx

/*
#cgo LDFLAGS: -lftd2xx
*/
import "C"
