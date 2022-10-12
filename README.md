# gouring

[![License: MIT][1]](LICENSE)
[![Go Reference][2]][3]


```bash
go get github.com/ii64/gouring
```
# Examples

## Write
```go
// setup
h, err := gouring.New(256, 0)
if err != nil { 
	log.Fatal("Error creating ring: ", err)
}
defer h.Close()

fd, err := unix.Open("/tmp/gouring_test", unix.O_RDWR, 0677)
if err != nil {
	log.Fatal("Error opening file: ", err)
}

sqe := h.GetSqe()
b := []byte("io_uring!\n")
gouring.PrepWrite(sqe, fd, &b[0], len(b), 0)

submitted, err := h.SubmitAndWait(1)
if err != nil { 
	log.Fatal("Error waiting ring: ", err)
}
println(submitted) // 1

var cqe *gouring.IoUringCqe
err = h.WaitCqe(&cqe) 
if err != nil {
	log.Fatal("Error waiting cqe: ", err)
} // check also EINTR

_ = cqe.UserData
_ = cqe.Res
_ = cqe.Flags
```

## Read
```go
// setup
h, err := gouring.New(256, 0)
if err != nil { 
	log.Fatal("Error creating ring: ", err)
}
defer h.Close()

fd, err := unix.Open("/tmp/gouring_test", unix.O_RDWR, 0677)
if err != nil {
	log.Fatal("Error opening file: ", err)
}

sqe := h.GetSqe()
b := make([]byte, 20)
gouring.PrepRead(sqe, fd, &b[0], len(b), 0)

submitted, err := h.SubmitAndWait(1)
if err != nil { 
	log.Fatal("Error waiting ring: ", err)
}
println(submitted) // 1

var cqe *gouring.IoUringCqe
err = h.WaitCqe(&cqe) 
if err != nil {
	log.Fatal("Error waiting cqe: ", err)
} // check also EINTR

_ = cqe.UserData
_ = cqe.Res
_ = cqe.Flags

log.Println("CQE: ", cqe)
log.Println("Buffer: ", b)
log.Println("Buffer: ", string(b))
```

## Graph

| SQPOLL | non-SQPOLL |
| ------ | ---------- |
| ![sqpoll_fig][sqpoll_fig] | ![nonsqpoll_fig][nonsqpoll_fig] |

### Reference

https://github.com/axboe/liburing

[1]: https://img.shields.io/badge/License-MIT-yellow.svg
[2]: https://pkg.go.dev/badge/github.com/ii64/gouring.svg
[3]: https://pkg.go.dev/github.com/ii64/gouring
[sqpoll_fig]: assets/sqpoll.svg
[nonsqpoll_fig]: assets/nonsqpoll.svg