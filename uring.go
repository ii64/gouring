package gouring

func New(entries uint32, flags uint32) (*IoUring, error) {
	ring := &IoUring{}
	p := new(IoUringParams)
	p.Flags = flags
	err := io_uring_queue_init_params(entries, ring, p)
	if err != nil {
		return nil, err
	}
	return ring, nil
}

func NewWithParams(entries uint32, params *IoUringParams) (*IoUring, error) {
	ring := &IoUring{}
	if params == nil {
		params = new(IoUringParams)
	}
	err := io_uring_queue_init_params(entries, ring, params)
	if err != nil {
		return nil, err
	}
	return ring, nil
}

func (h *IoUring) Close() {
	h.io_uring_queue_exit()
}

func (h *IoUring) GetSqe() *IoUringSqe {
	return h.io_uring_get_sqe()
}

func (h *IoUring) WaitCqe(cqePtr **IoUringCqe) error {
	return h.io_uring_wait_cqe(cqePtr)
}

func (h *IoUring) SeenCqe(cqe *IoUringCqe) {
	h.io_uring_cqe_seen(cqe)
}

func (h *IoUring) Submit() (int, error) {
	return h.io_uring_submit()
}

func (h *IoUring) SubmitAndWait(waitNr uint32) (int, error) {
	return h.io_uring_submit_and_wait(waitNr)
}
