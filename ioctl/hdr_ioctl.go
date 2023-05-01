package ioctl

// Based on `ioctl.h`

const (
	_IOC_NRBITS   = 8
	_IOC_TYPEBITS = 8

	_IOC_SIZEBITS = 14 // OVERRIDE
	_IOC_DIRBITS  = 2  // OVERRIDE

	_IOC_NRMASK   = (1 << _IOC_NRBITS) - 1
	_IOC_TYPEMASK = (1 << _IOC_TYPEBITS) - 1
	_IOC_SIZEMASK = (1 << _IOC_SIZEBITS) - 1
	_IOC_DIRMASK  = (1 << _IOC_DIRBITS) - 1

	_IOC_NRSHIFT   = 0
	_IOC_TYPESHIFT = (_IOC_NRSHIFT + _IOC_NRBITS)
	_IOC_SIZESHIFT = (_IOC_TYPESHIFT + _IOC_TYPEBITS)
	_IOC_DIRSHIFT  = (_IOC_SIZESHIFT + _IOC_SIZEBITS)

	_IOC_NONE  = 0b00 // OVERRIDE
	_IOC_WRITE = 0b01 // OVERRIDE
	_IOC_READ  = 0b10 // OVERRIDE
)

//go:nosplit
func IOC(dir, typ, nr, siz int) int {
	return 0 |
		(dir << _IOC_DIRSHIFT) |
		(typ << _IOC_TYPESHIFT) |
		(nr << _IOC_NRSHIFT) |
		(siz << _IOC_SIZESHIFT)
}

/*
	IO ops
*/

//go:nosplit
func IO(typ, nr int) int { return IOC(_IOC_NONE, typ, nr, 0) }

//go:nosplit
func IOR(typ, nr, siz int) int { return IOC(_IOC_READ, typ, nr, siz) }

//go:nosplit
func IOW(typ, nr, siz int) int { return IOC(_IOC_WRITE, typ, nr, siz) }

//go:nosplit
func IOWR(typ, nr, siz int) int { return IOC(_IOC_WRITE|_IOC_READ, typ, nr, siz) }
