package gouring

import (
	"syscall"

	"github.com/pkg/errors"
)

func New(entries uint, params *IOUringParams) (*Ring, error) {
	r := &Ring{}
	if params != nil {
		r.params = *params
	}
	var err error
	if r.fd, err = setup(r, entries, &r.params); err != nil {
		err = errors.Wrap(err, "setup")
		return nil, err
	}
	return r, nil
}

func (r *Ring) Close() (err error) {
	if err = unsetup(r); err != nil {
		err = errors.Wrap(err, "close")
		return
	}

	if err = syscall.Close(r.fd); err != nil {
		err = errors.Wrap(err, "close")
		return
	}
	// tbd..
	return
}

func (r *Ring) Register(opcode UringRegisterOpcode, arg uintptr, nrArg uint) (ret int, err error) {
	ret, err = register(r, opcode, arg, nrArg)
	if err != nil {
		err = errors.Wrap(err, "register")
		return
	}
	return
}

func (r *Ring) Enter(toSubmit, minComplete uint, flags UringEnterFlag, sig *Sigset_t) (ret int, err error) {
	ret, err = enter(r, toSubmit, minComplete, flags, sig)
	if err != nil {
		err = errors.Wrap(err, "enter")
		return
	}
	return
}

//

func (r *Ring) Params() *IOUringParams {
	return &r.params
}

func (r *Ring) Fd() int {
	return r.fd
}

func (r *Ring) SQ() SQRing {
	return r.sq
}

func (r *Ring) CQ() CQRing {
	return r.cq
}
