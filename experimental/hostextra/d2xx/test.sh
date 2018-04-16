#!/bin/bash
# Copyright 2017 The Periph Authors. All rights reserved.
# Use of this source code is governed under the Apache License, Version 2.0
# that can be found in the LICENSE file.

# Builds the package on multiple OSes to confirm it builds fine.
#
# It is recommended to use the -i flag so subsequent runs are much faster.

set -eu

cd `dirname $0`

OPT=$*

function build {
  echo "Testing on $1/$2"
  GOOS=$1 GOARCH=$2 go build $OPT
}

build darwin amd64
build linux amd64
build linux arm
build linux 386
build windows amd64
