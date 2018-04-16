# hostextra

This directory contains host drivers that are considered stable and complete as
per [driver lifetime
management](https://periph.io/project/#driver-lifetime-management).

You are welcome to create pull requests to add drivers here or improve the
quality of drivers already here. Please make sure to abide to requests in
[project/contributing/](https://periph.io/project/contributing/).

Unlike code in [periph.io/x/periph/host](https://periph.io/x/periph/host), code
under `hostextra` is allowed to use `cgo` or depend on third party Go packages.
