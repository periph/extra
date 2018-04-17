// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build cgo

package d2xx

// TODO(maruel): https://github.com/golang/go/issues/7211 would help target the
// optimal ARM architecture.

/*
#cgo LDFLAGS: ${SRCDIR}/linux_arm/libftd2xx.a
*/
import "C"
