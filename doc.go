// Copyright 2018 The Periph Authors. All rights reserved.
// Use of this source code is governed under the Apache License, Version 2.0
// that can be found in the LICENSE file.

// Package extra is for documentation only. Explains how to setup cgo.
//
// Debian
//
// This includes Raspbian and Ubuntu.
//
// You need to install pkg-config to enable cgo, run:
//
//  sudo apt install pkg-config
//
// MacOS
//
// You can install pkg-config via Homebrew (https://brew.sh). First install
// Homebrew.
//
// Either follow the official instructions at https://brew.sh to install system
// wide, or better install without root with the following steps. No root
// needed!
//
//  mkdir -p ~/homebrew
//  curl -sL https://github.com/Homebrew/brew/tarball/1.5.13 | tar xz --strip 1 -C ~/homebrew
//  export PATH="$PATH:$HOME/homebrew/bin"
//  echo 'export PATH="$PATH:$HOME/homebrew/bin"' >> ~/.bash_profile
//  brew upgrade
//
// and follow instructions. For example it may ask to run 'xcode-select
// -install'.
//
// Then install pkgconfig:
//
//  brew install pkgconfig
package extra
