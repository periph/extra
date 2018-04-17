# periph extra!

[![mascot](https://raw.githubusercontent.com/periph/website/master/site/static/img/periph-mascot-280.png)](https://periph.io/)

Supplemental tools for [periph.io/x/periph](https://periph.io/x/periph) that
cannot live in this repository. It may be because the tool/library:

- depends on external packages
- uses cgo

[![Go Report Card](https://goreportcard.com/badge/periph.io/x/extra)](https://goreportcard.com/report/periph.io/x/extra)
[![Coverage Status](https://codecov.io/gh/periph/extra/graph/badge.svg)](https://codecov.io/gh/periph/extra)
[![Build Status](https://travis-ci.org/periph/extra.svg)](https://travis-ci.org/periph/extra)
[![Gitter chat](https://badges.gitter.im/google/periph.png)](https://gitter.im/periph-io/Lobby)


## Install

```
go get -u -v periph.io/x/extra/cmd/...
```


### Debian

This includes Raspbian and Ubuntu.

You need to install pkg-config to enable cgo, run:

    sudo apt install pkg-config

### macOS

- Install [Homebrew](https://brew.sh)
- Follow instructions, it may ask to run `xcode-select -install`
- Install pkg-config with: `brew install pkg-config`

Optional: install without root with the following steps:

    mkdir -p ~/homebrew
    curl -sL https://github.com/Homebrew/brew/tarball/1.5.13 | tar xz --strip 1 -C ~/homebrew
    export PATH="$PATH:$HOME/homebrew/bin"
    echo 'export PATH="$PATH:$HOME/homebrew/bin"' >> ~/.bash_profile
    brew upgrade


### Windows

TODO. One way is to install mingw-w64.


## Authors

`periph` was initiated with ❤️️ and passion by [Marc-Antoine
Ruel](https://github.com/maruel). The full list of contributors is in
[AUTHORS](https://github.com/google/periph/blob/master/AUTHORS) and
[CONTRIBUTORS](https://github.com/google/periph/blob/master/CONTRIBUTORS).


## Disclaimer

This is not an official Google product (experimental or otherwise), it
is just code that happens to be owned by Google.

This project is not affiliated with the Go project.
