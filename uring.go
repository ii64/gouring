package gouring

func New(entries uint32, flags uint32) (*IoUring, error) {
	ring := &IoUring{}
	err := io_uring_queue_init(entries, ring, flags)
	if err != nil {
		return nil, err
	}
	return ring, nil
}

func NewWithParamms(entries uint32, params *IoUringParams) (*IoUring, error) {
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
