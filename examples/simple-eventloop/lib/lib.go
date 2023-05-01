package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/alphadose/haxmap"
	uring "github.com/ii64/gouring"
)

type Context interface {
	Read()
	Write(buf []byte)
	Close()

	SetContext(ctx context.Context)
	Context() context.Context
}

type EventHandler interface {
	OnAccept(ctx Context, sa syscall.Sockaddr)
	OnRead(ctx Context, buf []byte)
	OnWrite(ctx Context, nb int)
	OnClose(ctx Context)
}

type eventContext struct {
	evloop *Eventloop
	ud     *myUserdata
}

func (e *eventContext) SetContext(ctx context.Context) {
	e.ud.ctx = ctx
}
func (e *eventContext) Context() context.Context {
	return e.ud.ctx
}

func (e eventContext) Read() {
	key, lud := e.evloop.allocUserdata()
	sqe := e.evloop.queueRead(e.ud.fd, key)
	lud.init(sqe.Opcode)
	e.ud.copyTo(lud)
	sqe.UserData = key
}
func (e eventContext) Write(b []byte) {
	key, lud := e.evloop.allocUserdata()
	sqe := e.evloop.queueWrite(e.ud.fd, key, b)
	lud.init(sqe.Opcode)
	e.ud.copyTo(lud)
	sqe.UserData = key
}
func (e eventContext) Close() {
	key, lud := e.evloop.allocUserdata()
	sqe := e.evloop.queueClose(e.ud.fd, key)
	lud.init(sqe.Opcode)
	e.ud.copyTo(lud)
	sqe.UserData = key
}

type myUserdata struct {
	ctx   context.Context
	rsa   syscall.RawSockaddrAny
	rsaSz uintptr
	fd    int
	flag  uint64
	bid   int // buffer id
	op    uring.IoUringOp
}

func (ud *myUserdata) init(op uring.IoUringOp) {
	ud.op = op
	ud.rsaSz = unsafe.Sizeof(ud.rsa)
}

func (ud *myUserdata) copyTo(dst *myUserdata) {
	oldOp := dst.op
	*dst = *ud
	dst.op = oldOp
}

type Eventloop struct {
	ring              *uring.IoUring
	fd                int
	bufSize, bufCount int
	buffers           []byte
	handler           EventHandler
	sema              *uint64
	userdata          *haxmap.Map[uring.UserData, *myUserdata]
	bufGroup          uint16
}

func New(ent uint32, listenFd int, handler EventHandler) *Eventloop {
	ring, err := uring.NewWithParams(ent, &uring.IoUringParams{
		Flags:        uring.IORING_SETUP_SQPOLL,
		SqThreadIdle: 2000,
	})
	if err != nil {
		panic(err)
	}
	bufSize := 0x1000
	bufCount := 2048
	var bufGroup uint16 = 0xffff
	evloop := &Eventloop{
		fd:       listenFd,
		ring:     ring,
		bufSize:  bufSize,
		bufCount: bufCount,
		bufGroup: bufGroup,
		buffers:  make([]byte, bufCount*bufSize),
		userdata: haxmap.New[uring.UserData, *myUserdata](),
		handler:  handler,
		sema:     new(uint64),
	}
	if err := evloop.init(); err != nil {
		panic(err)
	}
	go evloop.runDebugInterface()
	return evloop
}

func (e *Eventloop) runDebugInterface() {
	mux := http.NewServeMux()
	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		dec := json.NewDecoder(r.Body)
		defer r.Body.Close()

		type SubmitRequest struct {
			EventName string `json:"en"`
			Fd        int    `json:"fd"`
			Data      any    `json:"d"`
		}

		var req SubmitRequest
		if err := dec.Decode(&req); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Printf("debug iface error: %s\n", err)
			return
		}

		cx := &eventContext{
			evloop: e,
			ud: &myUserdata{
				ctx: r.Context(),
				fd:  req.Fd,
			},
		}

		switch strings.ToUpper(strings.Trim(req.EventName, " ")) {
		case "WRITE":
			cx.Write([]byte("DEBUG!\n"))
		case "READ":
			cx.Read()
		case "CLOSE":
			cx.Close()
		}

		if _, err := e.ring.Submit(); err != nil {
			goto errInvalid
		}
		w.WriteHeader(http.StatusOK)
		return
	errInvalid:
		w.WriteHeader(http.StatusInternalServerError)
	})
	srv := &http.Server{
		Addr:    ":19110",
		Handler: mux,
	}
	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

var udPool = sync.Pool{
	New: func() any {
		return new(myUserdata)
	},
}

func (e *Eventloop) allocUserdata() (key uring.UserData, val *myUserdata) {
	// val = udPool.Get().(*myUserdata)
	val = new(myUserdata)
	key.SetUnsafe(unsafe.Pointer(val))
	e.userdata.Set(key, val)
	return
}
func (e *Eventloop) freeUserdata(key uring.UserData) {
	// v, exist := e.userdata.Get(key)
	// if exist {
	// 	udPool.Put(v)
	e.userdata.Del(key)
	// }
}

func (e *Eventloop) getBuf(bid int) []byte {
	start := e.bufSize * bid
	end := start + e.bufSize
	return e.buffers[start:end]
}

func (e *Eventloop) init() error {
	// queue accept mshot
	sqe := e.ring.GetSqe()
	key, ud := e.allocUserdata()
	uring.PrepAcceptMultishot(sqe, e.fd, &ud.rsa, &ud.rsaSz, 0)
	ud.init(sqe.Opcode)
	sqe.UserData = key

	// queue init provide buffers
	sqe = e.ring.GetSqe()
	uring.PrepProvideBuffers(sqe, unsafe.Pointer(&e.buffers[0]), e.bufSize, e.bufCount, e.bufGroup, 0)

	// wait for init provide buffers
	submitted, err := e.ring.SubmitAndWait(1)
	if err != nil {
		return err
	}
	if submitted != 2 {
		panic("MUST submit 2 sqes")
	}

	var cqe *uring.IoUringCqe
	if err = e.ring.WaitCqe(&cqe); err != nil {
		return err
	}
	if cqe.Res < 0 {
		err = syscall.Errno(-cqe.Res)
		return err
	}
	e.ring.SeenCqe(cqe)
	return nil
}

func (e *Eventloop) queueProvideBuffer(bid int, ud uring.UserData) *uring.IoUringSqe {
	fmt.Printf("[DEBUG] QU:PRV_BUF\n")
	sqe := e.ring.GetSqe()
	uring.PrepProvideBuffers(sqe, unsafe.Pointer(&e.getBuf(bid)[0]), e.bufSize, 1, e.bufGroup, bid)
	sqe.UserData = ud
	sqe.Flags |= uring.IOSQE_IO_LINK
	return sqe
}
func (e *Eventloop) queueRead(fd int, ud uring.UserData) *uring.IoUringSqe {
	fmt.Printf("[DEBUG] QU:READ fd=%d\n", fd)
	sqe := e.ring.GetSqe()
	uring.PrepRead(sqe, fd, nil, e.bufSize, 0)
	sqe.Flags |= uring.IOSQE_BUFFER_SELECT
	sqe.Flags |= uring.IOSQE_IO_LINK
	sqe.SetBufGroup(e.bufGroup)
	sqe.UserData = ud
	return sqe
}
func (e *Eventloop) queueWrite(fd int, ud uring.UserData, buf []byte) *uring.IoUringSqe {
	fmt.Printf("[DEBUG] QU:WRITE fd=%d\n", fd)
	sqe := e.ring.GetSqe()
	uring.PrepWrite(sqe, fd, &buf[0], len(buf), 0)
	sqe.Flags |= uring.IOSQE_IO_LINK
	sqe.UserData = ud
	return sqe
}
func (e *Eventloop) queueClose(fd int, ud uring.UserData) *uring.IoUringSqe {
	fmt.Printf("[DEBUG] QU:CLOSE fd=%d\n", fd)
	sqe := e.ring.GetSqe()
	uring.PrepClose(sqe, fd)
	sqe.Flags |= uring.IOSQE_IO_LINK
	sqe.UserData = ud
	return sqe
}

func (e *Eventloop) queueGraceShutdownNop() *uring.IoUringSqe {
	fmt.Printf("[DEBUG] QU:GRACE_SHUTDOWN\n")
	sqe := e.ring.GetSqe()
	uring.PrepNop(sqe)
	sqe.Flags |= uring.IOSQE_IO_LINK

	key, ud := e.allocUserdata()
	ud.init(sqe.Opcode)
	ud.fd = e.fd
	ud.flag = 0xDEAD_DEAD_DEAD_DEAD

	sqe.UserData = key
	return sqe
}

func (e *Eventloop) Run() {
	var cqe *uring.IoUringCqe
	var err error
	var i = 0

	var pokeCounter uint64 = 0
	qEventPoke := func() {
		sqe := e.ring.GetSqe()
		uring.PrepNop(sqe)
		sqe.UserData = 0xBEEF_BEEF_BEEF_BEEF
		atomic.AddUint64(&pokeCounter, 1)
		if _, err := e.ring.Submit(); err != nil {
			panic(err)
		}
	}
	if false {
		go func() {
			for {
				fmt.Printf("[MON] POKE COUNT: %d\n", pokeCounter)
				time.Sleep(time.Second * 2)
			}
		}()
		qEventPoke()
	}

	for {
		if atomic.CompareAndSwapUint64(e.sema, 1, 0) {
			break
		}
		if err = e.ring.WaitCqe(&cqe); err == syscall.EINTR || err == syscall.EAGAIN || err == syscall.ETIME {
			fmt.Printf("[DEBUG][%d] CQE WAIT err=%+#v\n", i, err)
			runtime.Gosched() // relax, do other thing while waiting for IO
			continue
		} else if err != nil {
			panic(err)
		}
		if cqe.UserData == 0xBEEF_BEEF_BEEF_BEEF {
			e.ring.SeenCqe(cqe)
			qEventPoke()
			continue
		}

		ctx := &eventContext{
			evloop: e,
		}
		ud, ok := e.userdata.Get(cqe.UserData)
		if !ok {
			goto skip_no_submit
		}
		ctx.ud = ud

		switch ud.op {
		case uring.IORING_OP_NOP:
			switch {
			case e.fd == ud.fd && ud.flag == 0xDEAD_DEAD_DEAD_DEAD:
				break
			}

		case uring.IORING_OP_ACCEPT:
			fmt.Printf("[DEBUG][%d] EV:ACCEPT\n", i)
			var sa syscall.Sockaddr
			sa, err = anyToSockaddr(&ud.rsa)
			if err != nil {
				panic(err)
			}
			fd := cqe.Res
			if fd < 0 {
				goto skip_no_submit
			}
			ud.fd = int(fd)
			e.handler.OnAccept(ctx, sa)

		case uring.IORING_OP_READ:
			fmt.Printf("[DEBUG][%d] EV:READ\n", i)
			if !(cqe.Flags&uring.IORING_CQE_F_BUFFER != 0) {
				panic("MUST PROVIDE BUFFER")
			}
			nb := cqe.Res
			bid := uint16(cqe.Flags >> 16)
			if cqe.Res <= 0 {
				e.queueClose(ud.fd, cqe.UserData)
			} else {
				e.handler.OnRead(ctx, e.getBuf(int(bid))[:nb])
			}
			e.queueProvideBuffer(int(bid), 0)

		case uring.IORING_OP_WRITE:
			fmt.Printf("[DEBUG][%d] EV:WRITE\n", i)
			e.handler.OnWrite(ctx, int(cqe.Res))

		case uring.IORING_OP_CLOSE:
			fmt.Printf("[DEBUG][%d] EV:CLOSE\n", i)
			e.handler.OnClose(ctx)

		}

		if ud.op != uring.IORING_OP_ACCEPT { // don't remove mshot UD
			e.freeUserdata(cqe.UserData)
		}

		e.reportSubmit(-1)
		if submitted, err := e.ring.Submit(); err != nil {
			panic(err)
		} else {
			_ = submitted
			e.reportSubmit(submitted)
		}
	skip_no_submit:
		e.ring.SeenCqe(cqe)
		i++
	}
}

func (e *Eventloop) reportSubmit(submitted int) {
	sq := &e.ring.Sq
	p := (*[2]unsafe.Pointer)(unsafe.Pointer(sq))
	khead := p[0]
	ktail := p[1]
	fmt.Printf("submitted: %d ht=(%d,%d) kht=(%p,%p)\n", submitted,
		sq.SqeHead, sq.SqeTail,
		khead, ktail)
}

func (e *Eventloop) Stop() error {
	// mark spin to stop
	atomic.StoreUint64(e.sema, 1)
	// if the spin waiting for an event
	// submit NOP with flag
	e.queueGraceShutdownNop()
	if _, err := e.ring.Submit(); err != nil {
		return err
	}
	return nil
}

func (e *Eventloop) Close() {
	e.ring.Close()
}

//go:linkname anyToSockaddr syscall.anyToSockaddr
func anyToSockaddr(rsa *syscall.RawSockaddrAny) (syscall.Sockaddr, error)

//go:linkname sockaddrToTCP net.sockaddrToTCP
func sockaddrToTCP(sa syscall.Sockaddr) net.Addr
