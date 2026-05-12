// Package mxl provides Go bindings to the MXL (Media eXchange Layer) C SDK.
//
// MXL is a shared-memory publish/subscribe layer for exchanging uncompressed
// video, audio, and ancillary data between co-located media functions. Flows
// live as tmpfs-backed ring buffers; readers and writers communicate via
// memory mapping and futex wakeups, with no copies in the data path.
//
// This binding currently exposes the reader side of the discrete-flow API
// (video and data grains). Writers and continuous (audio) flows are stubbed
// and will be filled in once the reader path is proven end-to-end.
//
// Build requirements:
//
//   - libmxl must be installed and visible to pkg-config. Run
//     `pkg-config --cflags --libs libmxl` to verify before `go build`.
//   - cgo must be enabled (CGO_ENABLED=1, default on native builds).
//
// Typical usage:
//
//	inst, err := mxl.NewInstance("/dev/shm/mxl", "")
//	if err != nil { return err }
//	defer inst.Close()
//
//	r, err := inst.NewReader("<flow-uuid>")
//	if err != nil { return err }
//	defer r.Close()
//
//	info, _ := r.Info()
//	for {
//	    idx := mxl.CurrentIndex(info.Config.GrainRate)
//	    g, err := r.GetGrain(idx, 200*time.Millisecond)
//	    if errors.Is(err, mxl.ErrTimeout) { continue }
//	    if err != nil { return err }
//	    handle(g.Payload) // borrowed; do not retain past next read
//	}
//
// Payload slices returned by reads alias shared memory directly. They are
// valid only until the next read on the same Reader; callers that need to
// retain data must copy it (use Grain.Copy()).
package mxl
