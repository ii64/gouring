package gouring

type UringSQEFlag = uint8

const (
	IOSQE_FIXED_FILE_BIT = iota
	IOSQE_IO_DRAIN_BIT
	IOSQE_IO_LINK_BIT
	IOSQE_TO_HARDLINK_BIT
	IOSQE_ASYNC_BIT
	IOSQE_BUFFER_SELECT_BIT
	IOSQE_CQE_SKIP_BIT
)

const (
	IOSQE_FIXED_FILE       UringSQEFlag = 1 << IOSQE_FIXED_FILE_BIT
	IOSQE_IO_DRAIN         UringSQEFlag = 1 << IOSQE_IO_DRAIN_BIT
	IOSQE_IO_LINK          UringSQEFlag = 1 << IOSQE_IO_LINK_BIT
	IOSQE_TO_HARDLINK      UringSQEFlag = 1 << IOSQE_TO_HARDLINK_BIT
	IOSQE_ASYNC            UringSQEFlag = 1 << IOSQE_ASYNC_BIT
	IOSQE_BUFFER_SELECT    UringSQEFlag = 1 << IOSQE_BUFFER_SELECT_BIT
	IOSQE_CQE_SKIP_SUCCESS UringSQEFlag = 1 << IOSQE_CQE_SKIP_BIT
)

//

// io_uring_setup() flags
type UringSetupFlag = uint32

const (
	IORING_SETUP_IOPOLL UringSetupFlag = 1 << iota
	IORING_SETUP_SQPOLL
	IORING_SETUP_SQ_AFF
	IORING_SETUP_SQSIZE
	IORING_SETUP_CLAMP
	IORING_SETUP_ATTACH_WQ
	IORING_SETUP_R_DISABLED
)

//

// uring  op code
type UringOpcode = uint8

const (
	IORING_OP_NOP UringOpcode = iota
	IORING_OP_READV
	IORING_OP_WRITEV
	IORING_OP_FSYNC
	IORING_OP_READ_FIXED
	IORING_OP_WRITE_FIXED
	IORING_OP_POLL_ADD
	IORING_OP_POLL_REMOVE
	IORING_OP_SYNC_FILE_RANGE
	IORING_OP_SENDMSG
	IORING_OP_RECVMSG
	IORING_OP_TIMEOUT
	IORING_OP_TIMEOUT_REMOVE
	IORING_OP_ACCEPT
	IORING_OP_ASYNC_CANCEL
	IORING_OP_LINK_TIMEOUT
	IORING_OP_CONNECT
	IORING_OP_FALLOCATE
	IORING_OP_OPENAT
	IORING_OP_CLOSE
	IORING_OP_FILES_UPDATE
	IORING_OP_STATX
	IORING_OP_READ
	IORING_OP_WRITE
	IORING_OP_FADVISE
	IORING_OP_MADVISE
	IORING_OP_SEND
	IORING_OP_RECV
	IORING_OP_OPENAT2
	IORING_OP_EPOLL_CTL
	IORING_OP_SPLICE
	IORING_OP_PROVIDE_BUFFERS
	IORING_OP_REMOVE_BUFFERS
	IORING_OP_TEE
	IORING_OP_SHUTDOWN
	IORING_OP_RENAMEAT
	IORING_OP_UNLINKAT
	IORING_OP_MKDIRAT
	IORING_OP_SYMLINKAT
	IORING_OP_LINKAT

	/* this goes last, obviously */
	IORING_OP_LAST
)

// sqe->fsync_flags
const IORING_FSYNC_DATASYNC uint32 = 1 << 0

// sqe->timeout_flags
const IORING_TIMEOUT_ABS uint32 = 1 << 0

// sqe->splice_flags
// extends splice(2) flags
const SPLICE_F_FD_IN_FIXED uint32 = 1 << 31

//

// cqe->flags
type UringCQEFlag = uint32

const IORING_CQE_F_BUFFER UringCQEFlag = 1 << 8
const IORING_CQE_BUFFER_SHIFT UringCQEFlag = 16

//

// Magic offsets for the application to mmap the data it needs
type UringOffset = int64

const (
	IORING_OFF_SQ_RING UringOffset = 0
	IORING_OFF_CQ_RING UringOffset = 0x8000000
	IORING_OFF_SQES    UringOffset = 0x10000000
)

//

// sq_ring->flags
type UringSQ = uint32

const (
	IORING_SQ_NEED_WAKEUP UringSQ = 1 << iota // needs io_uring_enter wakeup
	IORING_SQ_CQ_OVERFLOW                     // CQ Ring is overflow
)

//

// cq_ring->flags
type UringCQ = uint32

const IORING_CQ_EVENTFD_DISABLED = 1 << 0

//

// io_uring_enter(2) flag
type UringEnterFlag = uint32

const (
	IORING_ENTER_GETEVENTS UringEnterFlag = 1 << iota
	IORING_ENTER_SQ_WAKEUP
	IORING_ENTER_SQ_WAIT
	IORING_ENTER_EXT_ARG
)

//

// io_uring_params->features flags
type UringParamFeatureFlag = uint32

const (
	IORING_FEAT_SINGLE_MMAP UringParamFeatureFlag = 1 << iota
	IORING_FEAT_NODROP
	IORING_FEAT_SUBMIT_STABLE
	IORING_FEAT_RW_CUR_POS
	IORING_FEAT_CUR_PERSONALITY
	IORING_FEAT_FAST_POLL
	IORING_FEAT_POLL_32BITS
	IORING_FEAT_SQPOLL_NONFIXED
	IORING_FEAT_EXT_ARG
	IORING_FEAT_NATIVE_WORKERS
	IORING_FEAT_RSRC_TAGS
	IORING_FEAT_CQE_SKIP
)

//

type UringRegisterOpcode = uint

const (
	IORING_REGISTER_BUFFERS UringRegisterOpcode = iota
	IORING_UREGISTER_BUFFERS

	IORING_REGISTER_FILES
	IORING_UNREGISTER_FILES

	IORING_REGISTER_EVENTFD
	IORING_UNREGISTER_EVENTFD

	IORING_REGISTER_FILES_UPDATE
	IORING_REGISTER_EVENTFD_ASYNC
	IORING_REGISTER_PROBE

	IORING_REGISTER_PERSONALITY
	IORING_UNREGISTER_PERSONALITY

	IORING_REGISTER_RESTRICTIONS
	IORING_REGISTER_ENABLE_RINGS

	/* extended with tagging */
	IORING_REGISTER_FILES2
	IORING_REGISTER_FILES_UPDATE2
	IORING_REGISTER_BUFFERS2
	IORING_REGISTER_BUFFERS_UPDATE

	/* set/clear io-wq thread affinities */
	IORING_REGISTER_IOWQ_AFF
	IORING_UNREGISTER_IOWQ_AFF

	/* set/get max number of io-wq affinities */
	IORING_REGISTER_IOWQ_MAX_WORKERS

	// BPF soon

	//
	/* this goes last */
	IORING_REGISTER_LAST
)

//

const IO_URING_OP_SUPPORTED = 1 << 0

//

// io_uring_restriction->opcode values
type UringRestrictionOpcode = uint32

const (
	IORING_RESTRICTION_REGISTER_OP UringRestrictionOpcode = iota
	IORING_RESTRICTION_SQE_OP
	IORING_RESTRICTION_SQE_FLAGS_ALLOWED
	IORINGN_RESTRICTION_SQE_FLAGS_REQUIRED

	IORING_RESTRICTION_LAST
)
