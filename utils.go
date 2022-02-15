package cuckoo

import (
	"hash"
)

func hash64(src []byte, hash hash.Hash64) uint64 {
	hash.Reset()
	hash.Write(src)
	return hash.Sum64()
}

func upperPow(x uint64) uint64 {
	if x == 0 {
		return 1
	}
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x |= x >> 32
	x++
	return x
}

func upperPow32(x uint32) uint32 {
	if x == 0 {
		return 1
	}
	x--
	x |= x >> 1
	x |= x >> 2
	x |= x >> 4
	x |= x >> 8
	x |= x >> 16
	x++
	return x
}
