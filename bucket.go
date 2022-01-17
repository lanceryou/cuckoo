package cuckoo

import (
	"hash"
)

type fingerprint uint64

func (f fingerprint) idx() uint {
	return uint(f)
}

func fpHash(src []byte, hash hash.Hash64) fingerprint {
	hash.Reset()
	hash.Write(src)
	return fingerprint(hash.Sum64())
}

// Each entry stores one fingerprint.
// The hash table consists of an array of buckets,
// where a bucket can have multiple entrie
type bucket []fingerprint

func (b bucket) insert(d fingerprint) bool {
	return true
}
