package gouring

import (
	"context"
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRingQueueGetSQE(t *testing.T) {
	h := testNewIoUring(t, 256, 0)
	defer h.Close()

	assert.NotEqual(t, 0, h.RingFd)
	assert.NotEqual(t, 0, h.EnterRingFd)

	sqe := h.io_uring_get_sqe()
	assert.NotNil(t, sqe)
	fmt.Printf("%+#v\n", sqe)
}

func TestRingQueueSubmitSingleConsumer(t *testing.T) {
	type opt struct {
		name     string
		jobCount int

		entries uint32
		p       IoUringParams
	}
	ts := []opt{
		{"def-8-256", 8, 256, IoUringParams{}},
		{"def-16-256", 16, 256, IoUringParams{}},
		{"def-32-256", 32, 256, IoUringParams{}},
		{"def-64-256", 64, 256, IoUringParams{}},
		{"def-128-256", 128, 256, IoUringParams{}},
		{"def-128+2-256", 128 + 2, 256, IoUringParams{}}, // passwd 128
		{"def-256-256", 256, 256, IoUringParams{}},

		{"sqpoll-127-256", 127, 256, IoUringParams{Flags: IORING_SETUP_SQPOLL, SqThreadCpu: 4, SqThreadIdle: 10_000}},
		{"sqpoll-128+2-256", 128 + 2, 256, IoUringParams{Flags: IORING_SETUP_SQPOLL, SqThreadCpu: 4, SqThreadIdle: 10_000}},
		{"sqpoll-256-256", 256, 256, IoUringParams{Flags: IORING_SETUP_SQPOLL, SqThreadCpu: 4, SqThreadIdle: 10_000}},

		// we can have other test for queue overflow.
	}
	for _, tc := range ts {

		t.Run(tc.name, func(t *testing.T) {
			ftmp, err := os.CreateTemp(os.TempDir(), "test_iouring_queue_sc_*")
			require.NoError(t, err)
			defer ftmp.Close()
			fdTemp := ftmp.Fd()

			bufPool := sync.Pool{
				New: func() any {
					x := make([]byte, 0, 32)
					return &x
				},
			}

			consumer := func(h *IoUring, ctx context.Context, wg *sync.WaitGroup) {
				var cqe *IoUringCqe
				var err error
				for ctx.Err() == nil {
					err = h.io_uring_wait_cqe(&cqe)
					if err == syscall.EINTR {
						// ignore INTR
						continue
					}
					if err != nil {
						panic(err)
					}
					if cqe.Res < 0 {
						panic(syscall.Errno(-cqe.Res))
					}
					// cqe data check
					if int(cqe.Res) < len("data ") {
						panic(fmt.Sprintf("write less that it should"))
					}
					if (cqe.UserData>>(8<<2))&0xff == 0x00 {
						panic(fmt.Sprintf("cqe userdata should contain canonical address got %+#v", cqe.UserData))
					}

					// put back buf
					bufPool.Put((*[]byte)(unsafe.Pointer(uintptr(cqe.UserData))))
					h.io_uring_cqe_seen(cqe) // necessary
					wg.Done()
				}
			}

			submit := func(t *testing.T, opt *IoUringParams, h *IoUring, expectedSubmitCount int) {
				submitted, err := h.io_uringn_submit()
				assert.NoError(t, err)
				if opt.Flags&IORING_SETUP_SQPOLL == 0 {
					assert.Equal(t, expectedSubmitCount, submitted)
				}
			}

			t.Run("submit_single", func(t *testing.T) {
				var wg sync.WaitGroup

				h := testNewIoUringWithParam(t, 256, &tc.p)
				defer h.Close()

				wg.Add(tc.jobCount)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go consumer(h, ctx, &wg)

				for i := 0; i < tc.jobCount; i++ {
					var sqe *IoUringSqe
					for { // sqe could be nil if SQ is already full so we spin until we got one
						sqe = h.io_uring_get_sqe()
						if sqe != nil {
							break
						}
					}

					bufptr := bufPool.Get().(*[]byte)
					buf := (*bufptr)[:0]
					buf = append(buf, []byte(fmt.Sprintf("data %d\n", i))...)

					PrepWrite(sqe, int(fdTemp), &buf[0], len(buf), 0)
					sqe.UserData = uint64(uintptr(unsafe.Pointer(bufptr)))

					// submit
					submit(t, &tc.p, h, 1)
				}
				wg.Wait()
			})

			t.Run("submit_bulk", func(t *testing.T) {
				var wg sync.WaitGroup

				h := testNewIoUringWithParam(t, 256, &tc.p)
				defer h.Close()

				wg.Add(tc.jobCount)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				go consumer(h, ctx, &wg)

				for i := 0; i < tc.jobCount; i++ {
					sqe := h.io_uring_get_sqe()
					if sqe == nil {
						// spin until we got one
						continue
					}

					bufptr := bufPool.Get().(*[]byte)
					buf := (*bufptr)[:0]
					buf = append(buf, []byte(fmt.Sprintf("data %d\n", i))...)

					PrepWrite(sqe, int(fdTemp), &buf[0], len(buf), 0)
					sqe.UserData = uint64(uintptr(unsafe.Pointer(bufptr)))
				}

				submit(t, &tc.p, h, tc.jobCount)
				wg.Wait()
			})

		})
	}
}
