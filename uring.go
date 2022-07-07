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
	err := io_uring_queue_init_params(entries, ring, params)
	if err != nil {
		return nil, err
	}
	return ring, nil
}

func (h *IoUring) Close() {
	h.io_uring_queue_exit()
}

func (h *IoUring) GetSQE() *IoUringSqe {
	return h.io_uring_get_sqe()
}

func (h *IoUring) WaitCQE(cqePtr **IoUringCqe) error {
	return h.io_uring_wait_cqe(cqePtr)
}

func (h *IoUring) Submit() (int, error) {
	return h.io_uringn_submit()
}

func (h *IoUring) SubmitAndWait(waitNr uint32) (int, error) {
	return h.io_uring_submit_and_wait(waitNr)
}
