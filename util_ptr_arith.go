package gouring

import "unsafe"

type uint32Array *uint32

func uint32Array_Index(u uint32Array, i uintptr) *uint32 {
	return (*uint32)(unsafe.Pointer(uintptr(unsafe.Pointer(u)) + SizeofUint32*i))
}

type ioUringSqeArray *IoUringSqe

func ioUringSqeArray_Index(u ioUringSqeArray, i uintptr) *IoUringSqe {
	return (*IoUringSqe)(unsafe.Pointer(uintptr(unsafe.Pointer(u)) + SizeofIoUringSqe*i))
}

type ioUringCqeArray *IoUringCqe

func ioUringCqeArray_Index(u ioUringCqeArray, i uintptr) *IoUringCqe {
	return (*IoUringCqe)(unsafe.Pointer(uintptr(unsafe.Pointer(u)) + SizeofIoUringCqe*i))
}
