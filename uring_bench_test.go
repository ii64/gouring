package gouring

import (
	"context"
	"runtime"
	"syscall"
	"testing"
)

func BenchmarkQueueNop(b *testing.B) {
	type opt struct {
		name    string
		entries uint32
		p       IoUringParams
	}

	ts := []opt{
		{"def-256", 256, IoUringParams{Flags: 0}},
		{"sqpoll-256-4-10000", 256, IoUringParams{Flags: IORING_SETUP_SQPOLL, SqThreadCpu: 16, SqThreadIdle: 10_000}},
	}

	consumer := func(h *IoUring, ctx context.Context, count int) {
		var cqe *IoUringCqe
		var err error
		for i := 0; i < count; i++ {
			if ctx.Err() != nil {
				return
			}
			err = h.WaitCqe(&cqe)
			if err == syscall.EINTR {
				continue // ignore INTR
			} else if err != nil {
				panic(err)
			}
			if cqe.Res < 0 {
				panic(syscall.Errno(-cqe.Res))
			}

			h.SeenCqe(cqe)
		}
	}

	for _, tc := range ts {
		b.Run(tc.name, func(b *testing.B) {
			h := testNewIoUringWithParams(b, tc.entries, &tc.p)
			defer h.Close()
			var (
				j         uint32
				sqe       *IoUringSqe
				err       error
				submitted int
			)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j = 0; j < tc.entries; j++ {
					for {
						// sqe could be nil if SQ is already full so we spin until we got one
						sqe = h.GetSqe()
						if sqe != nil {
							break
						}
						runtime.Gosched()
					}
					PrepNop(sqe)
					sqe.UserData.SetUint64(uint64(i + int(j)))
				}
				submitted, err = h.Submit()
				if err != nil {
					panic(err)
				}
				consumer(h, ctx, submitted)
			}
			b.StopTimer()
		})

	}
}
