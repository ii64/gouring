package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"syscall"
	"time"

	"github.com/ii64/gouring"
)

// usage: go tool pprof -http=:9001 pprof.cpu

var fs = flag.NewFlagSet("perf", flag.ExitOnError)

var (
	entries      uint
	sqPoll       bool
	sqThreadCpu  uint
	sqThreadIdle uint

	N    uint
	noti uint
)

func init() {
	fs.UintVar(&entries, "entries", 256, "Entries")
	fs.BoolVar(&sqPoll, "sqpoll", false, "Enable SQPOLL")
	fs.UintVar(&sqThreadCpu, "sqthreadcpu", 4, "SQ Thread CPU")
	fs.UintVar(&sqThreadIdle, "sqthreadidle", 10_000, "SQ Thread idle")

	fs.UintVar(&N, "n", 10_000, "N times")
	fs.UintVar(&noti, "noti", 10_000, "Notify per attempt N")
}

func main() {
	err := fs.Parse(os.Args[1:])
	if err != nil {
		panic(err)
	}

	// check entries size
	if entries > uint(^uint32(0)) {
		panic("entries overflow.")
	}

	params := &gouring.IoUringParams{}
	if sqPoll {
		params.Flags |= gouring.IORING_SETUP_SQPOLL
		params.SqThreadCpu = uint32(sqThreadCpu)
		params.SqThreadIdle = uint32(sqThreadIdle)
	}

	h, err := gouring.NewWithParams(uint32(entries), params)
	if err != nil {
		panic(err)
	}
	defer h.Close()

	f, err := os.Create("pprof.cpu")
	if err != nil {
		panic(err)
	}

	fmt.Println("performing...")

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	var i, j uint
	var sqe *gouring.IoUringSqe
	var cqe *gouring.IoUringCqe
	var submitted int

	startTime := time.Now()
	for i = 0; i < N; i++ {
		if i%noti == 0 { // notify
			fmt.Printf("n:%d e:%s\n", j, time.Now().Sub(startTime))
		}

		for j = 0; j < entries; j++ {
			sqe = h.GetSQE()
			if sqe == nil { // spinning wait...
				runtime.Gosched()
				continue
			}
			gouring.PrepNop(sqe)
			sqe.UserData = uint64(i + j)
		}
		submitted, err = h.Submit()
		if err != nil {
			panic(err)
		}

		if i%noti == 0 { // notify
			fmt.Printf(" >> submitted %d\n", submitted)
		}

		for j = 0; j < entries; j++ {
			err = h.WaitCQE(&cqe)
			if err != nil {
				panic(err)
			}
			if cqe == nil {
				panic("cqe is nil!")
			}
			if cqe.Res < 0 {
				panic(syscall.Errno(-cqe.Res))
			}
			_ = cqe
		}
	}
	_ = submitted
	_ = err
}
