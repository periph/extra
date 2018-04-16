// Copyright 2017 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// +build cgo

package d2xx

/*
#cgo LDFLAGS: -framework CoreFoundation -framework IOKit ${SRCDIR}/darwin_amd64/libftd2xx.a
*/
import "C"
