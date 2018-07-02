# periph extra!

[![mascot](https://raw.githubusercontent.com/periph/website/master/site/static/img/periph-mascot-280.png)](https://periph.io/)

Supplemental tools for [periph.io/x/periph](https://periph.io/x/periph) that
cannot live there. It may be because the tool/library:

- depends on external packages
- uses cgo

[![Go Report Card](https://goreportcard.com/badge/periph.io/x/extra)](https://goreportcard.com/report/periph.io/x/extra)
[![Coverage Status](https://codecov.io/gh/periph/extra/graph/badge.svg)](https://codecov.io/gh/periph/extra)
[![Build Status](https://travis-ci.org/periph/extra.svg)](https://travis-ci.org/periph/extra)
[![Gitter chat](https://badges.gitter.im/google/periph.png)](https://gitter.im/periph-io/Lobby)


## Install

Install the whole suite with `periphextra` enabled to have access to the full
functionality:

```
go get -u periph.io/x/extra/cmd/...
go install -tags periphextra periph.io/x/periph/cmd/...
```

See https://periph.io for more details.


## Authors

`periph` was initiated with ❤️️ and passion by [Marc-Antoine
Ruel](https://github.com/maruel). The full list of contributors is in
[AUTHORS](https://github.com/google/periph/blob/master/AUTHORS) and
[CONTRIBUTORS](https://github.com/google/periph/blob/master/CONTRIBUTORS).


## Disclaimer

This is not an official Google product (experimental or otherwise), it
is just code that happens to be owned by Google.

This project is not affiliated with the Go project.
