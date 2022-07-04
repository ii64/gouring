package gouring

import (
	"syscall"
	"unsafe"
)

func PrepRW(op IoUringOp, sqe *IoUringSqe, fd int,
	addr unsafe.Pointer, len int, offset uint64) {
	sqe.Opcode = op
	sqe.Flags = 0
	sqe.IoPrio = 0
	sqe.Fd = int32(fd)
	sqe.SetOffset(offset)
	sqe.SetAddr(uint64(uintptr(addr)))
	sqe.Len = uint32(len)
	sqe.SetOpFlags(0)
	sqe.SetBufIndex(0)
	sqe.Personality = 0
	sqe.SetFileIndex(0)
	sqe.Addr3 = 0
	sqe.__pad2[0] = 0
}

func PrepNop(sqe *IoUringSqe) {
	PrepRW(IORING_OP_NOP, sqe, -1, nil, 0, 0)
}

func PrepTimeout(sqe *IoUringSqe, ts *syscall.Timespec, count uint32, flags uint32) {
	PrepRW(IORING_OP_TIMEOUT, sqe, -1, unsafe.Pointer(ts), 1, uint64(count))
	sqe.SetTimeoutFlags(flags)
}

func PrepReadv(sqe *IoUringSqe, fd int,
	iov *syscall.Iovec, nrVecs int,
	offset uint64) {
	PrepRW(IORING_OP_READV, sqe, fd, unsafe.Pointer(iov), nrVecs, offset)
}
func PrepReadv2(sqe *IoUringSqe, fd int,
	iov *syscall.Iovec, nrVecs int,
	offset uint64, flags uint32) {
	PrepReadv(sqe, fd, iov, nrVecs, offset)
	sqe.SetRwFlags(flags)
}

func PrepWritev(sqe *IoUringSqe, fd int,
	iov *syscall.Iovec, nrVecs int,
	offset uint64) {
	PrepRW(IORING_OP_WRITEV, sqe, fd, unsafe.Pointer(iov), nrVecs, offset)
}
func PrepWritev2(sqe *IoUringSqe, fd int,
	iov *syscall.Iovec, nrVecs int,
	offset uint64, flags uint32) {
	PrepWritev(sqe, fd, iov, nrVecs, offset)
	sqe.SetRwFlags(flags)
}

func PrepAccept(sqe *IoUringSqe, fd int, rsa *syscall.RawSockaddrAny, rsaSz *uintptr, flags int) {
	*rsaSz = syscall.SizeofSockaddrAny
	PrepRW(IORING_OP_ACCEPT, sqe, fd, unsafe.Pointer(rsa), 0, uint64(uintptr(unsafe.Pointer(rsaSz))))
	sqe.SetAcceptFlags(uint32(flags))
}
