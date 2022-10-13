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
	sqe.IoUringSqe_Union1 = IoUringSqe_Union1(offset)                    // union1
	sqe.IoUringSqe_Union2 = *(*IoUringSqe_Union2)(unsafe.Pointer(&addr)) // union2
	sqe.Len = uint32(len)
	sqe.IoUringSqe_Union3 = 0 // sqe.SetOpFlags(0) // union3
	sqe.IoUringSqe_Union4 = 0 // sqe.SetBufIndex(0) // union4
	sqe.Personality = 0
	sqe.IoUringSqe_Union5 = 0 // sqe.SetFileIndex(0) // union5
	sqe.IoUringSqe_Union6 = IoUringSqe_Union6{}
}

func PrepNop(sqe *IoUringSqe) {
	PrepRW(IORING_OP_NOP, sqe, -1, nil, 0, 0)
}

func PrepTimeout(sqe *IoUringSqe, ts *syscall.Timespec, count uint32, flags uint32) {
	PrepRW(IORING_OP_TIMEOUT, sqe, -1, unsafe.Pointer(ts), 1, uint64(count))
	sqe.SetTimeoutFlags(flags)
}

func PrepTimeoutRemove(sqe *IoUringSqe, userDaata uint64, flags uint32) {
	PrepRW(IORING_OP_TIMEOUT_REMOVE, sqe, -1, nil, 0, 0)
	sqe.SetAddr_Value(userDaata)
	sqe.SetTimeoutFlags(flags)
}

func PrepTimeoutUpdate(sqe *IoUringSqe, ts *syscall.Timespec, userData uint64, flags uint32) {
	PrepRW(IORING_OP_TIMEOUT_REMOVE, sqe, -1, nil, 0, 0)
	sqe.SetAddr_Value(userData)
	sqe.SetTimeoutFlags(flags | IORING_TIMEOUT_UPDATE)
}

/*
	"Syscall" OP
*/

func PrepSplice(sqe *IoUringSqe, fdIn int, offIn uint64, fdOut int, offOut uint64, nb int, spliceFlags uint32) {
	PrepRW(IORING_OP_SPLICE, sqe, fdOut, nil, nb, offOut)
	sqe.SetSpliceOffsetIn(offIn)
	sqe.SetSpliceFdIn(int32(fdIn))
	sqe.SetSpliceFlags(spliceFlags)
}

func PrepTee(sqe *IoUringSqe, fdIn int, fdOut int, nb int, spliceFlags uint32) {
	PrepRW(IORING_OP_TEE, sqe, fdOut, nil, nb, 0)
	sqe.SetSpliceOffsetIn(0)
	sqe.SetSpliceFdIn(int32(fdIn))
	sqe.SetSpliceFlags(spliceFlags)
}

func PrepRead(sqe *IoUringSqe, fd int, buf *byte, nb int, offset uint64) {
	PrepRW(IORING_OP_READ, sqe, fd, unsafe.Pointer(buf), nb, offset)
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
func PrepReadFixed(sqe *IoUringSqe, fd int,
	buf *byte, nb int,
	offset uint64, bufIndex uint16) {
	PrepRW(IORING_OP_READ_FIXED, sqe, fd, unsafe.Pointer(buf), nb, offset)
	sqe.SetBufIndex(bufIndex)
}

func PrepWrite(sqe *IoUringSqe, fd int, buf *byte, nb int, offset uint64) {
	PrepRW(IORING_OP_WRITE, sqe, fd, unsafe.Pointer(buf), nb, offset)
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
func PrepWriteFixed(sqe *IoUringSqe, fd int,
	buf *byte, nb int,
	offset uint64, bufIndex uint16) {
	PrepRW(IORING_OP_WRITE_FIXED, sqe, fd, unsafe.Pointer(buf), nb, offset)
	sqe.SetBufIndex(bufIndex)
}

func PrepAccept(sqe *IoUringSqe, fd int, rsa *syscall.RawSockaddrAny, rsaSz *uintptr, flags uint32) {
	PrepRW(IORING_OP_ACCEPT, sqe, fd, unsafe.Pointer(rsa), 0, uint64(uintptr(unsafe.Pointer(rsaSz))))
	sqe.SetAcceptFlags(uint32(flags))
}
func PrepAcceptDirect(sqe *IoUringSqe, fd int, rsa *syscall.RawSockaddrAny, rsaSz *uintptr, flags uint32, fileIndex int) {
	PrepAccept(sqe, fd, rsa, rsaSz, flags)
	__io_uring_set_target_fixed_file(sqe, uint32(fileIndex))
}

func PrepConnect(sqe *IoUringSqe, fd int, rsa *syscall.RawSockaddrAny, rsaSz uintptr) {
	PrepRW(IORING_OP_CONNECT, sqe, fd, unsafe.Pointer(rsa), 0, uint64(rsaSz))
}

func PrepRecvmsg(sqe *IoUringSqe, fd int, msg *syscall.Msghdr, flags uint) {
	PrepRW(IORING_OP_RECVMSG, sqe, fd, unsafe.Pointer(msg), 1, 0)
	sqe.SetMsgFlags(uint32(flags))
}

func PrepSendmsg(sqe *IoUringSqe, fd int, msg *syscall.Msghdr, flags uint32) {
	PrepRW(IORING_OP_SENDMSG, sqe, fd, unsafe.Pointer(msg), 1, 0)
	sqe.SetMsgFlags(flags)
}
func PrepSendmsgZc(sqe *IoUringSqe, fd int, msg *syscall.Msghdr, flags uint32) {
	PrepSendmsg(sqe, fd, msg, flags)
	sqe.Opcode |= IORING_OP_SENDMSG_ZC
}

func PrepClose(sqe *IoUringSqe, fd int) {
	PrepRW(IORING_OP_CLOSE, sqe, fd, nil, 0, 0)
}
func PrepCloseDirect(sqe *IoUringSqe, fileIndex uint32) {
	PrepClose(sqe, 0)
	__io_uring_set_target_fixed_file(sqe, fileIndex)
}

func PrepFilesUpdate(sqe *IoUringSqe, fds []int32, offset int) {
	PrepRW(IORING_OP_FILES_UPDATE, sqe, -1, unsafe.Pointer(&fds[0]), len(fds), uint64(offset))
}

//	func PrepFallocate(sqe *IoUringSqe, fd int, mode int, offset uint64, length uint64) {
//		PrepRW(IORING_OP_FALLOCATE, sqe, fd, )
//	}

func PrepOpenat(sqe *IoUringSqe, dfd int, path *byte, flags uint32, mode int) {
	PrepRW(IORING_OP_OPENAT, sqe, dfd, unsafe.Pointer(path), mode, 0)
	sqe.SetOpenFlags(flags)
}
func PrepOpenatDirect(sqe *IoUringSqe, dfd int, path *byte, flags uint32, mode int, fileIndex uint32) {
	PrepOpenat(sqe, dfd, path, flags, mode)
	__io_uring_set_target_fixed_file(sqe, fileIndex)
}

func PrepFadvise(sqe *IoUringSqe, fd int, offset uint64, length int, advice uint32) {
	PrepRW(IORING_OP_FADVISE, sqe, fd, nil, length, offset)
	sqe.SetFadviseAdvice(advice)
}
func PrepMadvise(sqe *IoUringSqe, addr unsafe.Pointer, length int, advice uint32) {
	PrepRW(IORING_OP_MADVISE, sqe, -1, addr, length, 0)
	sqe.SetFadviseAdvice(advice)
}

func PrepSend(sqe *IoUringSqe, sockfd int, buf *byte, length int, flags uint32) {
	PrepRW(IORING_OP_SEND, sqe, sockfd, unsafe.Pointer(buf), length, 0)
	sqe.SetMsgFlags(flags)
}
func PrepSendZc(sqe *IoUringSqe, sockfd int, buf *byte, length int, flags uint32, zcFlags uint16) {
	PrepRW(IORING_OP_SEND_ZC, sqe, sockfd, unsafe.Pointer(buf), length, 0)
	sqe.SetMsgFlags(flags)
	sqe.IoPrio = uint16(zcFlags)
}
func PrepSendZcFixed(sqe *IoUringSqe, sockfd int, buf *byte, length int, flags uint32, zcFlags uint16, bufIndex uint16) {
	PrepSendZc(sqe, sockfd, buf, length, flags, zcFlags)
	sqe.IoPrio |= IORING_RECVSEND_FIXED_BUF
	sqe.SetBufIndex(bufIndex)
}

// statx
//send
//recv
//openat2
//openat2Direct
//epollCtl
//provide_buffers
//remove_buffers
//shutdown
//unlinkat
//unlink
//renameat
//rename
//sync_file_range
//mkdirat
//mkdir
//symlinkat
//symlink
//linkat
//link
//msg_ring
//getxattr
//setxattr
//fgetxattr
//fsetxattr

func PrepSocket(sqe *IoUringSqe, domain int, _type int, protocol int, flags uint32) {
	PrepRW(IORING_OP_SOCKET, sqe, domain, nil, protocol, uint64(_type))
	sqe.SetRwFlags(flags)
}
func PrepSocketDirect(sqe *IoUringSqe, domain int, _type int, protocol int, fileIndex uint32, flags uint32) {
	PrepRW(IORING_OP_SOCKET, sqe, domain, nil, protocol, uint64(_type))
	sqe.SetRwFlags(flags)
	__io_uring_set_target_fixed_file(sqe, fileIndex)
}
func PrepSocketDirectAlloc(sqe *IoUringSqe, domain int, _type int, protocol int, flags uint32) {
	PrepRW(IORING_OP_SOCKET, sqe, domain, nil, protocol, uint64(_type))
	sqe.SetRwFlags(flags)
	__io_uring_set_target_fixed_file(sqe, IORING_FILE_INDEX_ALLOC-1)
}

/*
	Poll
*/

func PrepPollAdd(sqe *IoUringSqe, fd int, pollMask uint32) {
	PrepRW(IORING_OP_POLL_ADD, sqe, fd, nil, 0, 0)
	sqe.SetPoll32Events(pollMask) // TODO: check endiannes
}
func PrepPollMultishot(sqe *IoUringSqe, fd int, pollMask uint32) {
	PrepPollAdd(sqe, fd, pollMask)
	sqe.Len = IORING_POLL_ADD_MULTI
}
func PrepPollRemove(sqe *IoUringSqe, userdata UserData) {
	PrepRW(IORING_OP_POLL_REMOVE, sqe, -1, nil, 0, 0)
	sqe.SetAddr(userdata.GetUnsafe())
}
func PrepPollUpdate(sqe *IoUringSqe, oldUserdata UserData, newUserdata UserData, pollMask uint32, flags int) {
	PrepRW(IORING_OP_POLL_REMOVE, sqe, -1, nil, flags, newUserdata.GetUint64())
	sqe.SetAddr(oldUserdata.GetUnsafe())
	sqe.SetPoll32Events(pollMask) // TODO: check endiannes
}

func PrepFsync(sqe *IoUringSqe, fd int, fsyncFlags uint32) {
	PrepRW(IORING_OP_FSYNC, sqe, fd, nil, 0, 0)
	sqe.SetFsyncFlags(fsyncFlags)
}

func PrepCancel64(sqe *IoUringSqe, ud UserData, flags uint32) {
	PrepRW(IORING_OP_ASYNC_CANCEL, sqe, -1, nil, 0, 0)
	sqe.SetAddr(ud.GetUnsafe())
	sqe.SetCancelFlags(flags)
}
func PrepCancel(sqe *IoUringSqe, ud UserData, flags uint32) {
	PrepCancel64(sqe, UserData(ud.GetUintptr()), flags)
}
func PrepCancelFd(sqe *IoUringSqe, fd int, flags uint32) {
	PrepRW(IORING_OP_ASYNC_CANCEL, sqe, fd, nil, 0, 0)
	sqe.SetCancelFlags(flags | IORING_ASYNC_CANCEL_FD)
}

func PrepLinkTimeout(sqe *IoUringSqe, ts *syscall.Timespec, flags uint32) {
	PrepRW(IORING_OP_LINK_TIMEOUT, sqe, -1, unsafe.Pointer(ts), 1, 0)
	sqe.SetTimeoutFlags(flags)
}

/*
	Multishot
*/

func PrepMultishotAccept(sqe *IoUringSqe, fd int, rsa *syscall.RawSockaddrAny, rsaSz *uintptr, flags uint32) {
	PrepAccept(sqe, fd, rsa, rsaSz, flags)
	sqe.IoPrio |= IORING_ACCEPT_MULTISHOT
}

func PrepMultishotAcceptDirect(sqe *IoUringSqe, fd int, rsa *syscall.RawSockaddrAny, rsaSz *uintptr, flags uint32) {
	PrepMultishotAccept(sqe, fd, rsa, rsaSz, flags)
	__io_uring_set_target_fixed_file(sqe, IORING_FILE_INDEX_ALLOC-1)
}

//go:nosplit
func __io_uring_set_target_fixed_file(sqe *IoUringSqe, fileIndex uint32) {
	sqe.SetFileIndex(fileIndex)
}
