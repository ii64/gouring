package main

import (
	"fmt"
	"syscall"
	"unsafe"

	uring "github.com/ii64/gouring"
	nvme "github.com/ii64/gouring/nvme"
	"golang.org/x/sys/unix"
)

// NOTICE NOTICE NOTICE NOTICE NOTICE
//
//   This example is performing **READ** access to NVMe via low-level control device.
//
// NOTICE NOTICE NOTICE NOTICE NOTICE

var (
	// hardcoded device path
	// devicePath = "/dev/nvme0n1"
	devicePath = "/dev/ng0n1"

	nsid     uint32
	lbaSize  uint32
	lbaShift int
	BS       uint64 = 8192
)

func DoNvmeGetInfo(devPath string) error {
	fd, err := unix.Open(devPath, unix.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer func() {
		if err := unix.Close(fd); err != nil {
			panic(err)
		}
	}()

	var (
		ns  nvme.NvmeIdNs
		cmd nvme.NvmePassthruCmd
	)

	nsidRet, err := sys_ioctl(fd, uintptr(nvme.NVME_IOCTL_ID()), 0)
	if err != nil {
		return err
	}
	nsid = uint32(nsidRet)

	cmd = nvme.NvmePassthruCmd{
		Opcode:    nvme.NVME_ADMIN_IDENTIFY,
		Nsid:      nsid,
		Addr:      uint64(uintptr(unsafe.Pointer(&ns))),
		DataLen:   nvme.NVME_IDENTIFY_DATA_SIZE,
		Cdw10:     nvme.NVME_IDENTIFY_CNS_NS,
		Cdw11:     nvme.NVME_CSI_NVM << nvme.NVME_IDENTIFY_CSI_SHIFT,
		TimeoutMs: nvme.NVME_DEFAULT_IOCTL_TIMEOUT,
	}
	_, err = sys_ioctl(fd, uintptr(nvme.NVME_IOCTL_ADMIN_CMD()), uintptr(unsafe.Pointer(&cmd)))
	if err != nil {
		return err
	}

	lbaSize = 1 << ns.Lbaf[(ns.Flbas&0x0F)].Ds
	lbaShift = ilog2(uint32(lbaSize))

	return nil
}

func DoIoUring(devPath string) error {
	ring, err := uring.New(64,
		uring.IORING_SETUP_IOPOLL|
			uring.IORING_SETUP_SQE128|uring.IORING_SETUP_CQE32)
	if err != nil {
		return err
	}
	defer ring.Close()

	fd, err := unix.Open(devicePath, unix.O_RDONLY, 0) // 0 as it O_RDONLY
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	var bufs [10][0x1000]byte
	var sqe *uring.IoUringSqe
	sqe = ring.GetSqe()

	buf := bufs[1]
	bufSz := len(buf)
	uring.PrepRead(sqe, fd, &buf[0], bufSz, 0)

	sqe.SetCmdOp(uint32(nvme.NVME_URING_CMD_IO()))
	sqe.Opcode = uring.IORING_OP_URING_CMD

	var off uint64 = 0
	var i uint32 = 1
	sqe.UserData.SetUint64(uint64(off<<32) | uint64(i)) // temp

	var slba uint64 = off >> lbaShift
	var nlb uint64 = BS>>lbaShift - 1
	// zero and init
	cmd := nvme.NvmeUringCmd{
		Opcode: nvme.NVME_CMD_READ,

		// cdw10 and cdw11 represent starting lba
		Cdw10: uint32(slba & 0xffff_ffff),
		Cdw11: uint32(slba >> 32),
		// represent number of lba's for read/write
		Cdw12: uint32(nlb),

		Nsid: nsid,

		Addr:    uint64(uintptr(unsafe.Pointer(&buf[0]))),
		DataLen: uint32(bufSz),
	}
	cmdPtr := (*nvme.NvmeUringCmd)(sqe.GetCmd())
	*cmdPtr = cmd // copy

	fmt.Printf("CMD %+#v\n", cmdPtr)

	submitted, err := ring.SubmitAndWait(1)
	if err != nil {
		return err
	}
	fmt.Println("submitted", submitted)

	var cqe *uring.IoUringCqe
	// for i := 0; i < 2; i++ {
	if err := ring.WaitCqe(&cqe); err != nil {
		return err
	}
	fmt.Printf("CQE:\t%+#v\n", cqe)
	cqeExtra := (*[2]uint64)(cqe.GetBigCqe())
	fmt.Printf("CQE Extra:\t%+#v\n", cqeExtra)
	fmt.Printf("Buffer: %+#v\n", buf)
	fmt.Printf("=========\n")
	ring.SeenCqe(cqe)
	// }
	return nil
}

func main() {
	err := DoNvmeGetInfo(devicePath)
	if err != nil {
		panic(err)
	}
	fmt.Printf("lbaSize: %d lbaShift: %d\n", lbaSize, lbaShift)

	if err := DoIoUring(devicePath); err != nil {
		panic(err)
	}

}

func sys_ioctl(fd int, a1, a2 uintptr) (int, error) {
	r1, _, err := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(fd), a1, a2)
	if err != 0 {
		return 0, err
	}
	return int(r1), nil
}

func ilog2(i uint32) int {
	log := -1
	for i > 0 {
		i >>= 1
		log++
	}
	return log
}
