# gouring


[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/ii64/gouring.svg)](https://pkg.go.dev/github.com/ii64/gouring)

Low-level io uring library

```
go get github.com/ii64/gouring
```

## Example

```go
import "github.com/ii64/gouring"
import "github.com/ii64/gouring/queue"

// io_uring_setup
ring, err := gouring.New(256, nil) // default io uring setup param
if err != nil {
    // ...
}
defer ring.Close() // munmap shared memory, cleanup
var (
    ret int
    err error
)

// io_uring_register
ret, err = ring.Register(gouring.IORING_REGISTER_BUFFERS, addr, nrArg)

// io_uring_enter
ret, err = ring.Enter(toSubmit, minComplete, gouring.IORING_ENTER_GETEVENTS, nil)

// setup param
params := ring.Params()

// ring fd
fd := ring.Fd()

// Submission Queue
sq := ring.SQ()

// Completion Queue
cq := ring.CQ()

/* Using queue package */
q := queue.New(ring)
go func() {
    q.Run(func(cqe *gouring.CQEntry) {
        // cqe processing
        _ = cqe.UserData
        _ = cqe.Res
        _ = cqe.Flags
    })
}()

// buffer
data := []byte("print on stdout\n")

// get sqe
sqe := q.GetSQEntry()
sqe.UserData = 0 // identifier / event id
sqe.Opcode = gouring.IORING_OP_WRITE // op write
sqe.Fd = int32(syscall.Stdout) // fd 1
sqe.Len = uint32(len(data)) // buffer size
sqe.SetOffset(0) // fd offset
sqe.SetAddr(&data[0]) // buffer addr

// submit sqe
submitted, err := q.Submit()
if err != nil {
    // ...
}
```

### Referece
[github.com/iceber/iouring-go](https://github.com/iceber/iouring-go)