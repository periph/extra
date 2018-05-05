# extra/cmd - read-to-use executables

This directory contains directly usable tools installable via:

```
go get periph.io/x/extra/cmd/...
```

It is expected that many of these packages cannot be cross compiled easily, as
they likely leverage `cgo`.

This directory contains executables that are considered stable and complete as
per [driver lifetime
management](https://periph.io/project/#driver-lifetime-management).

You are welcome to create pull requests to add tools here or improve the
quality of executables already here. Please make sure to abide to requests in
[project/contributing/](https://periph.io/project/contributing/).

Unlike code in periph/cmd, code in extra/cmd is allowed to use `cgo`.
