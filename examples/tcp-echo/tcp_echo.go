package main

// Based from https://github.com/frevib/io_uring-echo-server/blob/a42497e4a7b1452329f6b2aa2cbcc25c2e422391/io_uring_echo_server.c

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"syscall"
	"unsafe"

	uring "github.com/ii64/gouring"
	"golang.org/x/sys/unix"
)

const (
	OP_ACCEPT = 1 << 0 // uring.IORING_OP_ACCEPT
	OP_READ   = 1 << 1 // uring.IORING_OP_READ
	OP_WRITE  = 1 << 2 // uring.IORING_OP_WRITE
	OP_PRBUF  = 1 << 3 // uring.IORING_OP_PROVIDE_BUFFERS
)

type MyUserdata struct {
	Fd    uint32
	Flags uint16
	BufID uint16
}

func UnpackUD(p uring.UserData) MyUserdata {
	return *(*MyUserdata)(unsafe.Pointer(&p))
}
func (ud MyUserdata) PackUD() uring.UserData {
	return *(*uring.UserData)(unsafe.Pointer(&ud))
}

func _SizeChecker() {
	var x [1]struct{}
	_ = x[unsafe.Sizeof(MyUserdata{})-unsafe.Sizeof(uring.UserData(0))]
}

func runServer() (err error) {
	var ring *uring.IoUring
	ring, err = uring.New(64, 0)
	if err != nil {
		return
	}
	defer ring.Close()

	probe := ring.GetProbeRing()
	fmt.Printf("probe: %+#v\n", probe)

	var ln net.Listener
	ln, err = net.Listen("tcp", "0.0.0.0:11337")
	if err != nil {
		return err
	}
	defer ln.Close()

	var file *os.File
	if file, err = ln.(*net.TCPListener).File(); err != nil {
		return
	}
	defer file.Close()
	fd := int(file.Fd())
	if err = unix.SetNonblock(fd, false); err != nil {
		return
	}

	var (
		rsa   syscall.RawSockaddrAny
		rsaSz uintptr
	)

	rsaSz = unsafe.Sizeof(rsa)

	const BUF_GID = 0xdead
	const BUF_SIZE = 0x1000
	const BUF_COUNT = 2048
	UD_ACCEPT_MSHOT := MyUserdata{
		Fd:    uint32(fd),
		Flags: OP_ACCEPT,
		BufID: ^uint16(0),
	}.PackUD()
	var sqe *uring.IoUringSqe
	var bufs [BUF_COUNT][BUF_SIZE]byte
	var submitted int

	// Q accept multishot
	sqe = ring.GetSqe()
	uring.PrepAcceptMultishot(sqe, fd, &rsa, &rsaSz, 0)
	sqe.UserData = UD_ACCEPT_MSHOT

	// Q init provide buffers
	sqe = ring.GetSqe()
	uring.PrepProvideBuffers(sqe, unsafe.Pointer(&bufs[0][0]), BUF_SIZE, BUF_COUNT, BUF_GID, 0)

	queueRead := func(fd int, buf *byte, nb int) *uring.IoUringSqe {
		sqe := ring.GetSqe()
		uring.PrepRead(sqe, fd, buf, nb, 0)
		return sqe
	}
	queueWrite := func(fd int, buf *byte, nb int) *uring.IoUringSqe {
		sqe := ring.GetSqe()
		uring.PrepWrite(sqe, fd, buf, nb, 0)
		return sqe
	}
	queueProvideBuf := func(index uint16) *uring.IoUringSqe {
		sqe := ring.GetSqe()
		uring.PrepProvideBuffers(sqe, unsafe.Pointer(&bufs[index]), BUF_SIZE, 1, BUF_GID, int(index))
		return sqe
	}
	_ = queueRead
	_ = queueWrite
	_ = queueProvideBuf

	// wait 1 for provide buf
	if submitted, err = ring.SubmitAndWait(1); err != nil {
		return
	}
	fmt.Printf("Submitted: %d\n", submitted)

	var cqe *uring.IoUringCqe
	err = ring.WaitCqe(&cqe) // init provide buffer result
	if err != nil {
		panic(err)
	} else if cqe.Res < 0 {
		panic(syscall.Errno(-cqe.Res))
	}
	ring.SeenCqe(cqe)

	for {
		err = ring.WaitCqe(&cqe)
		if err == syscall.EINTR {
			runtime.Gosched()
			continue
		} else if err != nil {
			return
		}
		ud := UnpackUD(cqe.UserData)
		fmt.Printf("cqe=%+#v ud=%+#v\n", cqe, ud)

		switch {
		case cqe.Res == -int32(syscall.ENOBUFS):
			panic("RAN OUT OF BUFFER!")

		case ud.Flags == OP_PRBUF:
			if cqe.Res < 0 {
				panic(syscall.Errno(-cqe.Res))
			}

		case ud.Flags == OP_ACCEPT:
			var sa syscall.Sockaddr
			sa, err = anyToSockaddr(&rsa)
			if err != nil {
				panic(err)
			}
			fd := cqe.Res
			fmt.Printf("CQE=%+#v rsaSz=%d sa=%+#v\n", cqe, rsaSz, sa)

			if fd < 0 {
				goto skip_no_submit
			}

			// Read from client socket
			sqe = queueRead(int(fd), nil, BUF_COUNT)
			sqe.Flags |= 0 |
				uring.IOSQE_BUFFER_SELECT
			sqe.SetBufGroup(BUF_GID)
			sqe.UserData = MyUserdata{
				Fd:    uint32(fd),
				Flags: OP_READ,
				BufID: ^uint16(0),
			}.PackUD()

		case ud.Flags == OP_READ:
			if !(cqe.Flags&uring.IORING_CQE_F_BUFFER != 0) {
				panic("MUST PROVIDE BUFFER")
			}
			nb := cqe.Res
			bid := uint16(cqe.Flags >> 16)
			if cqe.Res <= 0 {
				// read failed, re-add the buffer
				sqe = queueProvideBuf(bid)
				sqe.UserData = MyUserdata{
					Fd:    ud.Fd,
					Flags: OP_PRBUF,
					BufID: ^uint16(0),
				}.PackUD()
				// connection closed or error
				syscall.Close(int(ud.Fd))
			} else {
				// bytes have been read into bufs, now add write to socket sqe
				sqe = queueWrite(int(ud.Fd), &bufs[bid][0], int(nb))
				sqe.UserData = MyUserdata{
					Fd:    ud.Fd,
					Flags: OP_WRITE,
					BufID: bid,
				}.PackUD()
			}
		case ud.Flags == OP_WRITE:
			// write has been completed, first re-add the buffer
			sqe = queueProvideBuf(ud.BufID)
			sqe.UserData = MyUserdata{
				Fd:    ud.Fd,
				Flags: OP_PRBUF,
				BufID: ^uint16(0),
			}.PackUD()

			// Read from client socket
			sqe = queueRead(int(ud.Fd), nil, BUF_COUNT)
			sqe.Flags |= 0 |
				uring.IOSQE_BUFFER_SELECT
			sqe.SetBufGroup(BUF_GID)
			sqe.UserData = MyUserdata{
				Fd:    ud.Fd,
				Flags: OP_READ,
				BufID: ^uint16(0),
			}.PackUD()

		}

		// skip:
		if submitted, err = ring.Submit(); err != nil {
			panic(err)
		} else {
			println("submitted", submitted)
		}
	skip_no_submit:
		ring.SeenCqe(cqe)
	}
}

func main() {

	if err := runServer(); err != nil {
		panic(err)
	}

}

//go:linkname anyToSockaddr syscall.anyToSockaddr
func anyToSockaddr(rsa *syscall.RawSockaddrAny) (syscall.Sockaddr, error)

//go:linkname sockaddrToTCP net.sockaddrToTCP
func sockaddrToTCP(sa syscall.Sockaddr) net.Addr
