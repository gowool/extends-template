package internal

import "unsafe"

func Bytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
