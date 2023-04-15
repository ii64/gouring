package gouring

import (
	"syscall"
	"unsafe"
)

// func io_uring_queue_init(entries uint32, ring *IoUring, flags uint32) error {
// 	p := new(IoUringParams)
// 	p.Flags = flags
// 	return io_uring_queue_init_params(entries, ring, p)
// }

func io_uring_queue_init_params(entries uint32, ring *IoUring, p *IoUringParams) error {
	fd, err := io_uring_setup(uintptr(entries), p)
	if err != nil {
		return err
	}
	err = io_uring_queue_mmap(fd, p, ring)
	if err != nil {
		return err
	}

	// Directly map SQ slots to SQEs
	sqArray := ring.Sq.Array
	sqEntries := *ring.Sq._KRingEntries()
	var index uint32
	for index = 0; index < sqEntries; index++ {
		*uint32Array_Index(sqArray, uintptr(index)) = index
	}
	ring.Features = p.Features
	return nil
}

func (ring *IoUring) io_uring_queue_exit() {
	sq := &ring.Sq
	cq := &ring.Cq
	sqeSize := SizeofIoUringSqe
	if ring.Flags&IORING_SETUP_SQE128 != 0 {
		sqeSize += Align128IoUringSqe
	}
	munmap(unsafe.Pointer(sq.Sqes), sqeSize*uintptr(*sq._KRingEntries()))
	io_uring_unmap_rings(sq, cq)
	/*
	 * Not strictly required, but frees up the slot we used now rather
	 * than at process exit time.
	 */
	if ring.IntFlags&INT_FLAG_REG_RING != 0 {
		ring.io_uring_unregister_ring_fd()
	}
	syscall.Close(int(ring.RingFd))
}

func io_uring_queue_mmap(fd int, p *IoUringParams, ring *IoUring) error {
	err := io_uring_mmap(fd, p, &ring.Sq, &ring.Cq)
	if err != nil {
		return err
	}
	ring.Flags = p.Flags
	ring.RingFd, ring.EnterRingFd = fd, fd
	ring.IntFlags = 0
	return nil
}

func io_uring_mmap(fd int, p *IoUringParams, sq *IoUringSq, cq *IoUringCq) (err error) {
	size := SizeofIoUringCqe
	if p.Flags&IORING_SETUP_CQE32 != 0 {
		size += SizeofIoUringCqe
	}

	sq.RingSz = p.SqOff.Array + p.SqEntries*uint32(SizeofUnsigned)
	cq.RingSz = p.CqOff.Cqes + p.CqEntries*uint32(size)

	if p.Features&IORING_FEAT_SINGLE_MMAP != 0 {
		if cq.RingSz > sq.RingSz {
			sq.RingSz = cq.RingSz
		}
		cq.RingSz = sq.RingSz
	}
	// alloc sq ring
	sq.RingPtr, err = mmap(nil, uintptr(sq.RingSz),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_POPULATE,
		fd, IORING_OFF_SQ_RING)
	if err != nil {
		return
	}

	if p.Features&IORING_FEAT_SINGLE_MMAP != 0 {
		cq.RingPtr = sq.RingPtr
	} else {
		// alloc cq ring
		cq.RingPtr, err = mmap(nil, uintptr(cq.RingSz),
			syscall.PROT_READ|syscall.PROT_WRITE,
			syscall.MAP_SHARED|syscall.MAP_POPULATE,
			fd, IORING_OFF_CQ_RING)
		if err != nil {
			// goto errLabel
			io_uring_unmap_rings(sq, cq)
			return
		}
	}

	//sq
	sq.khead = unsafe.Add(sq.RingPtr, p.SqOff.Head)
	sq.ktail = unsafe.Add(sq.RingPtr, p.SqOff.Tail)
	sq.kringMask = unsafe.Add(sq.RingPtr, p.SqOff.RingMask)
	sq.kringEntries = unsafe.Add(sq.RingPtr, p.SqOff.RingEntries)
	sq.kflags = unsafe.Add(sq.RingPtr, p.SqOff.Flags)
	sq.kdropped = unsafe.Add(sq.RingPtr, p.SqOff.Dropped)
	sq.Array = (uint32Array)(unsafe.Add(sq.RingPtr, p.SqOff.Array))

	size = SizeofIoUringSqe
	if p.Flags&IORING_SETUP_SQE128 != 0 {
		size += Align128IoUringSqe
	}
	var sqeAddr unsafe.Pointer
	sqeAddr, err = mmap(nil, size*uintptr(p.SqEntries),
		syscall.PROT_READ|syscall.PROT_WRITE,
		syscall.MAP_SHARED|syscall.MAP_POPULATE,
		fd, IORING_OFF_SQES)
	if err != nil {
		//errLabel:
		io_uring_unmap_rings(sq, cq)
		return
	}
	sq.Sqes = (ioUringSqeArray)(sqeAddr)

	//cq
	cq.khead = unsafe.Add(cq.RingPtr, p.CqOff.Head)
	cq.ktail = unsafe.Add(cq.RingPtr, p.CqOff.Tail)
	cq.kringMask = unsafe.Add(cq.RingPtr, p.CqOff.RingMask)
	cq.kringEntries = unsafe.Add(cq.RingPtr, p.CqOff.RingEntries)
	cq.koverflow = unsafe.Add(cq.RingPtr, p.CqOff.Overflow)
	cq.Cqes = (ioUringCqeArray)(unsafe.Add(cq.RingPtr, p.CqOff.Cqes))

	if p.CqOff.Flags != 0 {
		cq.kflags = (unsafe.Pointer(uintptr(cq.RingPtr) + uintptr(p.CqOff.Flags)))
	}

	sq.RingMask = *sq._KRingMask()
	sq.RingEntries = *sq._KRingEntries()
	cq.RingMask = *cq._KRingMask()
	cq.RingEntries = *cq._KRingEntries()
	return nil
}

func io_uring_unmap_rings(sq *IoUringSq, cq *IoUringCq) error {
	munmap(sq.RingPtr, uintptr(sq.RingSz))
	if cq.RingPtr != nil && cq.RingPtr != sq.RingPtr {
		munmap(cq.RingPtr, uintptr(cq.RingSz))
	}
	return nil
}

func io_uring_get_probe_ring(ring *IoUring) (probe *IoUringProbe) {
	// len := SizeofIoUringProbe + 256*SizeofIouringProbeOp
	probe = new(IoUringProbe)
	r := ring.io_uring_register_probe(probe, 256)
	if r >= 0 {
		return
	}
	return nil
}

func (ring *IoUring) io_uring_get_probe_ring() (probe *IoUringProbe) {
	return io_uring_get_probe_ring(ring)
}
