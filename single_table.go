package cuckoo

import (
	"fmt"
	"math/rand"
)

type fingerprint []byte

type singleTable struct {
	numBucket     uint32
	tagsPerBucket uint32
	bitsPerItem   uint32
	tagMask       uint32
	buckets       []fingerprint
}

func (t *singleTable) SizeInTags() uint32 {
	return t.numBucket * t.tagsPerBucket
}

func (t *singleTable) Init(numBucket, tagsPerBucket, bitsPerItem uint32) {
	t.numBucket = numBucket
	t.tagsPerBucket = tagsPerBucket
	t.bitsPerItem = bitsPerItem
	t.tagMask = (1 << bitsPerItem) - 1
	t.buckets = make([]fingerprint, numBucket)
	kBytesPerFingerprint := (bitsPerItem*tagsPerBucket + 7) >> 3
	var i uint32
	for i = 0; i < numBucket; i++ {
		t.buckets[i] = make(fingerprint, kBytesPerFingerprint)
	}
}

func (t *singleTable) Insert(i uint32, tag uint32, kickout bool) (oldTag uint32, ok bool) {
	var j uint32
	for j = 0; j < t.tagsPerBucket; j++ {
		if t.readTag(i, j) == 0 {
			t.writeTag(i, j, tag)
			return 0, true
		}
	}

	if !kickout {
		return tag, false
	}

	var r uint32 = uint32(rand.Intn(int(t.tagsPerBucket)))
	oldTag = t.readTag(i, r)
	t.writeTag(i, r, tag)
	return oldTag, false
}

func (t *singleTable) Delete(i uint32, tag uint32) bool {
	var j uint32
	for j = 0; j < t.tagsPerBucket; j++ {
		if t.readTag(i, j) == tag {
			t.writeTag(i, j, 0)
			return true
		}
	}

	return false
}

func (t *singleTable) Find(i uint32, tag uint32) bool {
	var j uint32
	for j = 0; j < t.tagsPerBucket; j++ {
		if t.readTag(i, j) == tag {
			return true
		}
	}

	return false
}

func (t *singleTable) Info() string {
	return fmt.Sprintf("SingleHashtable with tag size:%v bits \n"+
		"\t\tAssociativity: %v \n"+
		"\t\tTotal # of rows: %v\n"+
		"\t\tTotal # slots: %v\n",
		t.bitsPerItem, t.tagsPerBucket, t.numBucket, t.numBucket*t.tagsPerBucket)
}

func (t *singleTable) String() string {
	return SingleTable
}

func (t *singleTable) writeTag(i, j, tag uint32) {
	fp := t.buckets[i]
	tag = tag & t.tagMask
	/* following code only works for little-endian */
	if t.bitsPerItem == 2 {
		fp[0] |= byte(tag << (2 * j))
	} else if t.bitsPerItem == 4 {
		pos := j >> 1
		if j&1 == 0 {
			fp[pos] &= 0xf0
			fp[pos] |= byte(tag)
		} else {
			fp[pos] &= 0x0f
			fp[pos] |= byte(tag << 4)
		}
	} else if t.bitsPerItem == 8 {
		fp[j] = byte(tag)
	} else if t.bitsPerItem == 12 {
		pos := j + (j >> 1)
		if j&1 == 0 {
			fp[pos] = byte(tag)
			fp[pos+1] &= 0xf0
			fp[pos+1] |= byte(tag >> 8)
		} else {
			fp[pos] &= 0x0f
			fp[pos] |= byte(tag << 4)
			fp[pos+1] = byte(tag >> 4)
		}
	} else if t.bitsPerItem == 16 {
		pos := j << 1
		fp[pos] = byte(tag)
		fp[pos+1] = byte(tag >> 8)
	} else if t.bitsPerItem == 32 {
		pos := j << 2
		fp[pos] = byte(tag)
		fp[pos+1] = byte(tag >> 8)
		fp[pos+2] = byte(tag >> 16)
		fp[pos+3] = byte(tag >> 24)
	}
}

func (t *singleTable) readTag(i, j uint32) uint32 {
	/* following code only works for little-endian */
	fp := t.buckets[i]
	var tag uint32 = 0
	if t.bitsPerItem == 2 {
		tag = uint32(fp[0] >> (j * 2))
	} else if t.bitsPerItem == 4 {
		pos := j >> 1
		tag = uint32(fp[pos] >> ((j & 1) << 2))
	} else if t.bitsPerItem == 8 {
		tag = uint32(fp[j])
	} else if t.bitsPerItem == 12 {
		pos := j + (j >> 1)
		tag = (uint32(fp[pos]) | uint32(fp[pos+1])<<8) >> ((j & 1) << 2)
	} else if t.bitsPerItem == 16 {
		pos := j << 1
		tag = uint32(fp[pos]) | uint32(fp[pos+1])<<8
	} else if t.bitsPerItem == 32 {
		pos := j << 2
		tag = uint32(fp[pos]) | uint32(fp[pos+1])<<8 | uint32(fp[pos+2])<<16 + uint32(fp[pos+3])<<24
	}

	return tag & t.tagMask
}
