package gouring

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testNewIoUring(t *testing.T, entries uint32, flags uint32) *IoUring {
	h, err := New(entries, flags)
	assert.NoError(t, err)
	assert.NotNil(t, h)
	return h
}
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
	ts := []int{
		8,
		32,
		64,
		128 + 2,
		256,
		// we can have other test for queue overflow.
	}
	for i := range ts {
		jobCount := ts[i]

		t.Run(fmt.Sprintf("jobsz-%d", jobCount), func(t *testing.T) {
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

			consumer := func(h *IoUring, wg *sync.WaitGroup) {
				var cqe *IoUringCqe
				var err error
				for {
					err = h.io_uring_wait_cqe(&cqe)
					if err == syscall.EINTR {
						// ignore interrupt
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

			submit := func(t *testing.T, h *IoUring, expectedSubmitCount int) {
				submitted, err := h.io_uringn_submit()
				assert.NoError(t, err)
				assert.Equal(t, expectedSubmitCount, submitted)
			}

			t.Run("submit_single", func(t *testing.T) {
				var wg sync.WaitGroup

				h := testNewIoUring(t, 256, 0)
				defer h.Close()

				wg.Add(jobCount)
				go consumer(h, &wg)

				for i := 0; i < jobCount; i++ {
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

					// submit
					submit(t, h, 1)
				}
				wg.Wait()
			})

			t.Run("submit_bulk", func(t *testing.T) {
				var wg sync.WaitGroup

				h := testNewIoUring(t, 256, 0)
				defer h.Close()

				wg.Add(jobCount)
				go consumer(h, &wg)

				for i := 0; i < jobCount; i++ {
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

				submit(t, h, jobCount)
				wg.Wait()
			})

		})
	}
}
