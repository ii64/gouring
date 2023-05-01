package nvme

import (
	"unsafe"

	"github.com/ii64/gouring/ioctl"
)

// Based on `nvme.h` w.r.t. `linux/nvme_ioctl.h`

const (
	SizeofNvmeAdminCmd      = unsafe.Sizeof(NvmeAdminCmd{})
	SizeofNvmeUserIo        = unsafe.Sizeof(NvmeUserIo{})
	SizeofNvmePassthruCmd   = unsafe.Sizeof(NvmePassthruCmd{})
	SizeofNvmePassthruCmd64 = unsafe.Sizeof(NvmePassthruCmd64{})
	SizeofNvmeUringCmd      = unsafe.Sizeof(NvmeUringCmd{})
	SizeofNvmeIdNs          = unsafe.Sizeof(NvmeIdNs{})
	SizeofNvmeLbaf          = unsafe.Sizeof(NvmeLbaf{})

	NVME_DEFAULT_IOCTL_TIMEOUT = 0
	NVME_IDENTIFY_DATA_SIZE    = 0x1000
	NVME_IDENTIFY_CSI_SHIFT    = 24
	NVME_IDENTIFY_CNS_NS       = 0
	NVME_CSI_NVM               = 0
)

func _SizeChecker() {
	var x [1]struct{}
	_ = x[SizeofNvmeAdminCmd-72]
	_ = x[SizeofNvmeUserIo-48]
	_ = x[SizeofNvmePassthruCmd-72]
	_ = x[SizeofNvmePassthruCmd64-80]
	_ = x[SizeofNvmeUringCmd-72]
	_ = x[SizeofNvmeLbaf-4]
	_ = x[SizeofNvmeIdNs-0x1000]
}

func NVME_IOCTL_ID() int           { return ioctl.IO('N', 0x40) }
func NVME_IOCTL_ADMIN_CMD() int    { return ioctl.IOWR('N', 0x41, int(SizeofNvmeAdminCmd)) }
func NVME_IOCTL_SUBMIT_IO() int    { return ioctl.IOW('N', 0x42, int(SizeofNvmeUserIo)) }
func NVME_IOCTL_IO_CMD() int       { return ioctl.IOR('N', 0x43, int(SizeofNvmePassthruCmd)) }
func NVME_IOCTL_RESET() int        { return ioctl.IO('N', 0x44) }
func NVME_IOCTL_SUBSYS_RESET() int { return ioctl.IO('N', 0x45) }
func NVME_IOCTL_RESCAN() int       { return ioctl.IO('N', 0x46) }
func NVME_IOCTL_ADMIN64_CMD() int  { return ioctl.IOWR('N', 0x47, int(SizeofNvmePassthruCmd64)) }
func NVME_IOCTL_IO64_CMD() int     { return ioctl.IOWR('N', 0x48, int(SizeofNvmePassthruCmd64)) }
func NVME_IOCTL_IO64_CMD_VEC() int { return ioctl.IOWR('N', 0x49, int(SizeofNvmePassthruCmd64)) }

func NVME_URING_CMD_IO() int        { return ioctl.IOWR('N', 0x80, int(SizeofNvmeUringCmd)) }
func NVME_URING_CMD_IO_VEC() int    { return ioctl.IOWR('N', 0x81, int(SizeofNvmeUringCmd)) }
func NVME_URING_CMD_ADMIN() int     { return ioctl.IOWR('N', 0x82, int(SizeofNvmeUringCmd)) }
func NVME_URING_CMD_ADMIN_VEC() int { return ioctl.IOWR('N', 0x83, int(SizeofNvmeUringCmd)) }

// nvme_admin_opcode
const (
	NVME_ADMIN_IDENTIFY = 0x06
)

// nvme_io_opcode
const (
	NVME_CMD_WRITE = 0x01
	NVME_CMD_READ  = 0x02
)

type NvmeAdminCmd = NvmePassthruCmd

type NvmeUserIo struct {
	Opcode   uint8
	Flags    uint8
	Control  uint16
	Nblocks  uint16
	Rsvd     uint16
	Metadata uint64
	Addr     uint64
	Slba     uint64
	Dsmgmt   uint32
	Reftag   uint32
	Apptag   uint16
	Appmask  uint16
	_pad     [4]byte
}

type NvmePassthruCmd struct {
	Opcode uint8
	Flags  uint8
	Rsvd1  uint16
	Nsid   uint32
	Cdw2,
	Cdw3 uint32
	Metadata    uint64
	Addr        uint64
	MetadataLen uint32
	DataLen     uint32
	Cdw10,
	Cdw11,
	Cdw12,
	Cdw13,
	Cdw14,
	Cdw15 uint32
	TimeoutMs uint32
	Result    uint32
}

type NvmePassthruCmd64_Union1 uint32

func (u *NvmePassthruCmd64_Union1) SetDataLen(v uint32)  { *u = NvmePassthruCmd64_Union1(v) }
func (u *NvmePassthruCmd64_Union1) SetVecCount(v uint32) { *u = NvmePassthruCmd64_Union1(v) }

type NvmePassthruCmd64 struct {
	Opcode uint8
	Flags  uint8
	Rsvd1  uint16
	Nsid   uint32
	Cdw2,
	Cdw3 uint32
	Metadata    uint64
	Addr        uint64
	MetadataLen uint32
	// union {
	// 	__u32	data_len; /* for non-vectored io */
	// 	__u32	vec_cnt; /* for vectored io */
	// };
	NvmePassthruCmd64_Union1
	Cdw10,
	Cdw11,
	Cdw12,
	Cdw13,
	Cdw14,
	Cdw15 uint32
	TimeoutMs uint32
	Rsvd2     uint32
	Result    uint64
}

type NvmeUringCmd struct {
	Opcode      uint8
	Flags       uint8
	Rsvd1       uint16
	Nsid        uint32
	Cdw2, Cdw3  uint32
	Metadata    uint64
	Addr        uint64
	MetadataLen uint32
	DataLen     uint32

	Cdw10, Cdw11, Cdw12, Cdw13, Cdw14, Cdw15 uint32

	TimeoutMs uint32
	Rsvd2     uint32
}

type NvmeLbaf struct {
	Ms uint16 // bo: Little
	Ds uint8
	Rp uint8
}

type NvmeIdNs struct {
	Nsze,
	Ncap,
	Nuse uint64 // bo: Little
	Nsfeat,
	Nlbaf,
	Flbas,
	Mc,
	Dpc,
	Dps,
	Nmic,
	Rescap,
	Fpi,
	Dlfeat uint8
	Nawun,
	Nawupf,
	Nacwu,
	Nabsn,
	Nabo,
	Nabspf,
	Noiob uint16 // bo: Little
	Nvmcap [16]byte
	Npwg,
	Npwa,
	Npdg,
	Npda,
	Nows uint16 // bo: Little
	Msrl     uint16 // bo: Little
	Mcl      uint32 // bo: Little
	Msrc     uint8
	Resvd81  [11]byte
	Anagrpid uint32 // bo: Little
	Rsvd96   [3]byte
	Nsattr   uint8
	Nvmsetid uint16 // bo: Little
	Endgid   uint16 // bo: Little
	Nguid    [16]byte
	Eui64    [8]byte
	Lbaf     [16]NvmeLbaf
	Rsvd192  [192]byte
	Vs       [3712]byte
}
