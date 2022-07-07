# gouring

[![License: MIT][1]](LICENSE)
[![Go Reference][2]][3]


```bash
go get github.com/ii64/gouring
```
## Example

```go
// setup
h, err := gouring.New(256, 0)
if err != nil { /*...*/ }
defer h.Close() 

sqe := h.GetSQE()
b := []byte("io_uring!\n")
PrepWrite(sqe, 1, &b[0], len(b), 0)

submitted, err := h.SubmitAndWait(1)
if err != nil { /*...*/ }
println(submitted) // 1

var cqe *gouring.IoUringCqe
err = h.WaitCqe(&cqe) 
if err != nil { /*...*/ } // check also EINTR

_ = cqe.UserData
_ = cqe.Res
_ = cqe.Flags
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