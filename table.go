package cuckoo

const (
	SingleTable = "single-table"
)

var (
	_ Table = &singleTable{}
	_ Table = &PackedTable{}
)

type Table interface {
	Init(numBucket, tagsPerBucket, bitsPerItem uint32)
	Insert(i uint32, tag uint32, kickout bool) (oldTag uint32, ok bool)
	Delete(i uint32, tag uint32) bool
	Find(i1 uint32, tag uint32) bool
	SizeInTags() uint32
	Info() string
	String() string
}
