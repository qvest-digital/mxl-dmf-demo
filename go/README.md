<!--
SPDX-FileCopyrightText: 2025 Contributors to the Media eXchange Layer project.
SPDX-License-Identifier: Apache-2.0
-->

# mxl — Go bindings for libmxl

Idiomatic Go bindings to the MXL (Media eXchange Layer) C SDK. Mirrors the
public API exposed by `mxl/{mxl,flow,flowinfo,time,rational,dataformat}.h`:
discrete-grain and continuous-sample I/O, synchronization groups, instance
management, and the time/index helpers.

## Install

```sh
go get github.com/qvest-digital/mxl-dmf-demo/go@v0.1.0
```

The package lives in a `go/` subdirectory of the wider `mxl-dmf-demo`
repository. Subdirectory Go modules need a prefixed tag for module proxy
resolution, so the corresponding repo tag is `go/v0.1.0`. Until a tag is
published, pin by commit:

```sh
go get github.com/qvest-digital/mxl-dmf-demo/go@<commit-sha>
```

## Build requirements

The binding is a cgo wrapper, so libmxl must be installed on the host. Once
it is, `go build` works with no special flags in the common case.

### Step 1 — install libmxl

Build and install libmxl (see the top-level `README.md` for the C build).
The install layout is:

```
<prefix>/
  include/mxl/*.h
  lib/libmxl.so          (RUNPATH=$ORIGIN, finds libmxl-common as sibling)
  lib/libmxl-common.so
  lib/pkgconfig/libmxl.pc
```

`libmxl.so` has `RUNPATH=$ORIGIN` baked in by the build, so it locates its
own runtime dependency (`libmxl-common.so`) without any help from the
loader. `libmxl.pc` advertises `-Wl,-rpath,${libdir}` in its `Libs:` line,
so any binary linked through `pkg-config --libs libmxl` automatically
records the install prefix in its own `RUNPATH`.

The net effect: **no `CGO_LDFLAGS`, no `LD_LIBRARY_PATH`, no `-rpath`
juggling on the consumer side.**

### Step 2 — point pkg-config at libmxl (only if not in a default prefix)

If you installed libmxl to a standard prefix (`/usr/local`, `/usr`,
distribution-managed location), pkg-config finds it automatically and
**no environment variables are needed**.

If you installed to a custom prefix:

```sh
export PKG_CONFIG_PATH="<prefix>/lib/pkgconfig:$PKG_CONFIG_PATH"
pkg-config --cflags --libs libmxl    # sanity check
```

### Step 3 — `go build`

```sh
go build ./...
./my-app
```

That's it. The resulting binary has `<prefix>/lib` in its `RUNPATH` and
runs without further configuration:

```sh
$ readelf -d ./my-app | grep RUNPATH
 0x0000000000000001 (RUNPATH)   Library runpath: [<prefix>/lib]
```

### When you still need environment variables

| Situation | Variable to set |
| --- | --- |
| `pkg-config` can't find `libmxl.pc` | `PKG_CONFIG_PATH` |
| Building libmxl as static (`BUILD_SHARED_LIBS=OFF`) and want `--static` linking | `PKG_CONFIG="pkg-config --static"` |
| Relocating the binary to a host without libmxl at the recorded `RUNPATH` | `LD_LIBRARY_PATH` at runtime |
| Cross-compiling | `CC`, `CXX`, `PKG_CONFIG_SYSROOT_DIR`, `CGO_ENABLED=1` |

## Quick example

```go
package main

import (
    "errors"
    "log"
    "time"

    "github.com/qvest-digital/mxl-dmf-demo/go"
)

func main() {
    inst, err := mxl.NewInstance("/dev/shm/mxl", "")
    if err != nil {
        log.Fatal(err)
    }
    defer inst.Close()

    r, err := inst.NewReader("5fbec3b1-1b0f-417d-9059-8b94a47197ed")
    if err != nil {
        log.Fatal(err)
    }
    defer r.Close()

    info, _ := r.Info()
    rate := info.Config.Common.GrainRate
    idx := mxl.CurrentIndex(rate)

    for {
        g, err := r.GetGrain(idx, 200*time.Millisecond)
        switch {
        case errors.Is(err, mxl.ErrTimeout):
            idx = mxl.CurrentIndex(rate)
        case err != nil:
            log.Fatal(err)
        default:
            log.Printf("grain %d: %d bytes", g.Index, len(g.Payload))
            idx++
        }
    }
}
```

## Examples

Runnable programs in [examples/](examples/):

- [read-grain](examples/read-grain/main.go) — subscribe to a discrete flow
- [write-grain](examples/write-grain/main.go) — create/write a discrete flow
- [read-samples](examples/read-samples/main.go) — subscribe to a continuous flow
- [write-samples](examples/write-samples/main.go) — create/write a continuous flow
- [sync-group](examples/sync-group/main.go) — wait on multiple flows in lock-step

Build them with:

```sh
cd go && go build ./examples/...
```

## API surface

Top-level types:

| Type | Purpose |
| --- | --- |
| `Instance` | One MXL domain. Source of readers, writers, sync groups. |
| `Reader` | Subscribes to a flow; grain or sample read APIs depending on format. |
| `Writer` | Creates / opens a flow for production; grain or sample write APIs. |
| `Grain` / `GrainWriteAccess` | Read-only / mutable view onto a single grain. |
| `SamplesView` / `SamplesWriteAccess` | Read-only / mutable view onto a sample range across all channels. |
| `SyncGroup` | Synchronizes a wait across multiple readers. |
| `FlowInfo` / `FlowConfig` / `FlowRuntime` | Mirror of libmxl's flow header. |
| `Status` | Mirrors `mxlStatus`; implements `error` and round-trips via `errors.Is` / `errors.As`. |

Errors returned by every cgo call are `Status` values (or `ErrClosed` for
use-after-close); compare with `errors.Is(err, mxl.ErrTimeout)` etc.

## Memory safety

Payload slices returned by reads — and by writer-side `OpenGrain` /
`OpenSamples` — alias libmxl's shared memory directly. They are only valid
until the next read, the matching `Commit`/`Cancel`, or the
`Reader`/`Writer` being closed. Copy out (`Grain.Copy()`,
`SamplesView.CopyChannel(ch)`) to retain data.
