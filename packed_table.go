package cuckoo

import (
	"fmt"
	"math/rand"
)

// semi sort table
// Using Permutation encoding to save 1 bit per tag
type PackedTable struct {
	kDirBitsPerTag  uint32
	kBitsPerBucket  uint32
	kBytesPerBucket uint32
	kDirBitsMask    uint32

	bitsPerItem uint32
	len         uint32
	numBuckets  uint32
	buckets     []byte
	perm        *PermEncoding
}

// NewPackedTable new a PackedTable
func NewPackedTable() *PackedTable {
	return &PackedTable{}
}

// Init init packed table
func (p *PackedTable) Init(numBucket, tagsPerBucket, bitsPerItem uint32) {
	p.bitsPerItem = bitsPerItem
	p.numBuckets = numBucket

	p.kDirBitsPerTag = bitsPerItem - 4
	p.kBitsPerBucket = (3 + p.kDirBitsPerTag) * 4
	p.kBytesPerBucket = (p.kBitsPerBucket + 7) >> 3
	p.kDirBitsMask = ((1 << p.kDirBitsPerTag) - 1) << 4

	p.len = p.kBytesPerBucket*numBucket + 7
	p.buckets = make([]byte, p.len)
	p.perm = NewPermEncoding()
}

func (p *PackedTable) Insert(i uint32, tag uint32, kickout bool) (oldTag uint32, ok bool) {
	tags := p.readTag(i)

	for j := 0; j < 4; j++ {
		if tags[j] == 0 {
			tags[j] = tag
			p.writeTag(i, tags, true)
			return 0, true
		}
	}

	if !kickout {
		return tag, false
	}

	r := rand.Intn(4)
	oldTag = tags[r]
	tags[r] = tag
	p.writeTag(i, tags, true)
	return oldTag, false
}

func (p *PackedTable) Delete(i uint32, tag uint32) bool {
	tags := p.readTag(i)

	for j := 0; j < 4; j++ {
		if tags[j] == tag {
			tags[j] = 0
			p.writeTag(i, tags, true)
			return true
		}
	}

	return false
}

func (p *PackedTable) Find(i uint32, tag uint32) bool {
	tags := p.readTag(i)

	return tags[0] == tag || tags[1] == tag || tags[2] == tag || tags[3] == tag
}

func (p *PackedTable) SizeInTags() uint32 {
	return p.numBuckets * 4
}

func (p *PackedTable) Info() string {
	return fmt.Sprintf("PackedHashtable with tag size: %v bits \n"+
		"\t\t4 packed bits(3 bits after compression) and %v direct bits\n"+
		"\t\tAssociativity: 4 \n"+
		"\t\tTotal # of rows: %v\n"+
		"\t\tTotal # slots: %v\n",
		p.bitsPerItem, p.kDirBitsPerTag, p.numBuckets, p.SizeInTags())
}

func (p *PackedTable) String() string {
	return "packed_table"
}

func (p *PackedTable) sortPair(a, b *uint32) {
	if (*a & 0x0f) > (*b & 0x0f) {
		*a, *b = *b, *a
	}
}

func (p *PackedTable) sortTags(tags *[4]uint32) {
	p.sortPair(&tags[0], &tags[2])
	p.sortPair(&tags[1], &tags[3])
	p.sortPair(&tags[0], &tags[1])
	p.sortPair(&tags[2], &tags[3])
	p.sortPair(&tags[1], &tags[2])
}

/* read and decode the bucket i, pass the 4 decoded tags to the 2nd arg
 * bucket bits = 12 codeword bits + dir bits of tag1 + dir bits of tag2 ...
 */
func (p *PackedTable) readTag(i uint32) [4]uint32 {
	var tags [4]uint32
	var codeword uint16
	if p.bitsPerItem == 5 {
		// 1 dirbits per tag, 16 bits per bucket
		pos := i * 2
		tag := uint16(p.buckets[pos]) | uint16(p.buckets[pos+1])<<8
		codeword = tag & 0x0fff
		tags[0] = uint32(tag>>8) & p.kDirBitsMask
		tags[1] = uint32(tag>>9) & p.kDirBitsMask
		tags[2] = uint32(tag>>10) & p.kDirBitsMask
		tags[3] = uint32(tag>>11) & p.kDirBitsMask
	} else if p.bitsPerItem == 6 {
		// 2 dirbits per tag, 20 bits per bucket
		pos := (20 * i) >> 3
		tag := uint32(p.buckets[pos]) | uint32(p.buckets[pos+1])<<8 |
			uint32(p.buckets[pos+2])<<16 | uint32(p.buckets[pos+3])<<24
		codeword = (uint16(tag) >> ((i & 1) << 2)) & 0x0fff
		tags[0] = (tag >> (8 + ((i & 1) << 2))) & p.kDirBitsMask
		tags[1] = (tag >> (10 + ((i & 1) << 2))) & p.kDirBitsMask
		tags[2] = (tag >> (12 + ((i & 1) << 2))) & p.kDirBitsMask
		tags[3] = (tag >> (14 + ((i & 1) << 2))) & p.kDirBitsMask
	} else if p.bitsPerItem == 7 {
		// 3 dirbits per tag, 24 bits per bucket
		pos := (i << 1) + i
		tag := uint32(p.buckets[pos]) | uint32(p.buckets[pos+1])<<8 |
			uint32(p.buckets[pos+2])<<16 | uint32(p.buckets[pos+3])<<24
		codeword = uint16(tag) & 0x0fff
		tags[0] = (tag >> 8) & p.kDirBitsMask
		tags[1] = (tag >> 11) & p.kDirBitsMask
		tags[2] = (tag >> 14) & p.kDirBitsMask
		tags[3] = (tag >> 17) & p.kDirBitsMask
	} else if p.bitsPerItem == 8 {
		// 4 dirbits per tag, 28 bits per bucket
		pos := (28 * i) >> 3
		tag := uint32(p.buckets[pos]) | uint32(p.buckets[pos+1])<<8 |
			uint32(p.buckets[pos+2])<<16 | uint32(p.buckets[pos+3])<<24
		codeword = (uint16(tag) >> ((i & 1) << 2)) & 0x0fff
		tags[0] = (tag >> (8 + ((i & 1) << 2))) & p.kDirBitsMask
		tags[1] = (tag >> (12 + ((i & 1) << 2))) & p.kDirBitsMask
		tags[2] = (tag >> (16 + ((i & 1) << 2))) & p.kDirBitsMask
		tags[3] = (tag >> (20 + ((i & 1) << 2))) & p.kDirBitsMask
	} else if p.bitsPerItem == 9 {
		// 5 dirbits per tag, 32 bits per bucket
		pos := i * 4
		tag := uint32(p.buckets[pos]) | uint32(p.buckets[pos+1])<<8 |
			uint32(p.buckets[pos+2])<<16 | uint32(p.buckets[pos+3])<<24
		codeword = uint16(tag) & 0x0fff
		tags[0] = (tag >> 8) & p.kDirBitsMask
		tags[1] = (tag >> 13) & p.kDirBitsMask
		tags[2] = (tag >> 18) & p.kDirBitsMask
		tags[3] = (tag >> 23) & p.kDirBitsMask
	} else if p.bitsPerItem == 13 {
		// 9 dirbits per tag,  48 bits per bucket
		pos := i * 6
		tag := uint64(p.buckets[pos]) | uint64(p.buckets[pos+1])<<8 |
			uint64(p.buckets[pos+2])<<16 | uint64(p.buckets[pos+3])<<24 |
			uint64(p.buckets[pos+4])<<32 | uint64(p.buckets[pos+5])<<40 |
			uint64(p.buckets[pos+6])<<48 | uint64(p.buckets[pos+7])<<56
		codeword = uint16(tag) & 0x0fff

		tags[0] = uint32(tag>>8) & p.kDirBitsMask
		tags[1] = uint32(tag>>17) & p.kDirBitsMask
		tags[2] = uint32(tag>>26) & p.kDirBitsMask
		tags[3] = uint32(tag>>35) & p.kDirBitsMask
	} else if p.bitsPerItem == 17 {
		// 13 dirbits per tag, 64 bits per bucket
		pos := i << 3
		tag := uint64(p.buckets[pos]) | uint64(p.buckets[pos+1])<<8 |
			uint64(p.buckets[pos+2])<<16 | uint64(p.buckets[pos+3])<<24 |
			uint64(p.buckets[pos+4])<<32 | uint64(p.buckets[pos+5])<<40 |
			uint64(p.buckets[pos+6])<<48 | uint64(p.buckets[pos+7])<<56

		codeword = uint16(tag) & 0x0fff
		tags[0] = uint32(tag>>8) & p.kDirBitsMask
		tags[1] = uint32(tag>>21) & p.kDirBitsMask
		tags[2] = uint32(tag>>34) & p.kDirBitsMask
		tags[3] = uint32(tag>>47) & p.kDirBitsMask
	}

	/* codeword is the lowest 12 bits in the bucket */
	lowBits := p.perm.Decode(codeword)

	tags[0] |= uint32(lowBits[0])
	tags[1] |= uint32(lowBits[1])
	tags[2] |= uint32(lowBits[2])
	tags[3] |= uint32(lowBits[3])

	return tags
}

/* Tag = 4 low bits + x high bits
 * L L L L H H H H ...
 */
func (p *PackedTable) writeTag(i uint32, tags [4]uint32, sort bool) {
	if sort {
		p.sortTags(&tags)
	}

	var lowBits [4]uint8
	lowBits[0] = uint8(tags[0] & 0x0f)
	lowBits[1] = uint8(tags[1] & 0x0f)
	lowBits[2] = uint8(tags[2] & 0x0f)
	lowBits[3] = uint8(tags[3] & 0x0f)

	var highBits [4]uint32
	highBits[0] = tags[0] & 0xfffffff0
	highBits[1] = tags[1] & 0xfffffff0
	highBits[2] = tags[2] & 0xfffffff0
	highBits[3] = tags[3] & 0xfffffff0

	codeword := p.perm.Encode(lowBits)
	pos := (p.kBitsPerBucket * i) >> 3
	if p.kBitsPerBucket == 16 {
		// 1 dirbits per tag
		tag := codeword | uint16(highBits[0]<<8) | uint16(highBits[1]<<9) |
			uint16(highBits[2]<<10) | uint16(highBits[3]<<11)
		p.buckets[pos] = byte(tag)
		p.buckets[pos+1] = byte(tag >> 8)
	} else if p.kBitsPerBucket == 20 {
		// 2 dirbits per tag
		tag := uint32(p.buckets[pos]) | uint32(p.buckets[pos+1])<<8 |
			uint32(p.buckets[pos+2])<<16 | uint32(p.buckets[pos+3])<<24
		if (i & 0x0001) == 0 {
			tag &= 0xfff00000
			tag |= uint32(codeword) | (highBits[0] << 8) |
				(highBits[1] << 10) | (highBits[2] << 12) |
				(highBits[3] << 14)
		} else {
			tag &= 0xff00000f
			tag |= uint32(codeword)<<4 | (highBits[0] << 12) |
				(highBits[1] << 14) | (highBits[2] << 16) |
				(highBits[3] << 18)
		}

		p.buckets[pos] = byte(tag)
		p.buckets[pos+1] = byte(tag >> 8)
		p.buckets[pos+2] = byte(tag >> 16)
		p.buckets[pos+3] = byte(tag >> 24)
	} else if p.kBitsPerBucket == 24 {
		// 3 dirbits per tag
		tag := uint32(p.buckets[pos]) | uint32(p.buckets[pos+1])<<8 |
			uint32(p.buckets[pos+2])<<16 | uint32(p.buckets[pos+3])<<24
		tag &= 0xff000000
		tag |= uint32(codeword) | (highBits[0] << 8) |
			(highBits[1] << 11) | (highBits[2] << 14) |
			(highBits[3] << 17)

		p.buckets[pos] = byte(tag)
		p.buckets[pos+1] = byte(tag >> 8)
		p.buckets[pos+2] = byte(tag >> 16)
		p.buckets[pos+3] = byte(tag >> 24)
	} else if p.kBitsPerBucket == 28 {
		// 4 dirbits per tag
		tag := uint32(p.buckets[pos]) | uint32(p.buckets[pos+1])<<8 |
			uint32(p.buckets[pos+2])<<16 | uint32(p.buckets[pos+3])<<24
		if (i & 0x0001) == 0 {
			tag &= 0xf0000000
			tag |= uint32(codeword) | (highBits[0] << 8) |
				(highBits[1] << 12) | (highBits[2] << 16) |
				(highBits[3] << 20)
		} else {
			tag &= 0x0000000f
			tag |= uint32(codeword)<<4 | (highBits[0] << 12) |
				(highBits[1] << 16) | (highBits[2] << 20) |
				(highBits[3] << 24)
		}

		p.buckets[pos] = byte(tag)
		p.buckets[pos+1] = byte(tag >> 8)
		p.buckets[pos+2] = byte(tag >> 16)
		p.buckets[pos+3] = byte(tag >> 24)
	} else if p.kBitsPerBucket == 32 {
		// 5 dirbits per tag
		tag := uint32(codeword) | (highBits[0] << 8) | (highBits[1] << 13) |
			(highBits[2] << 18) | (highBits[3] << 23)

		p.buckets[pos] = byte(tag)
		p.buckets[pos+1] = byte(tag >> 8)
		p.buckets[pos+2] = byte(tag >> 16)
		p.buckets[pos+3] = byte(tag >> 24)
	} else if p.kBitsPerBucket == 48 {
		tag := uint64(p.buckets[pos]) | uint64(p.buckets[pos+1])<<8 |
			uint64(p.buckets[pos+2])<<16 | uint64(p.buckets[pos+3])<<24 |
			uint64(p.buckets[pos+4])<<32 | uint64(p.buckets[pos+5])<<40 |
			uint64(p.buckets[pos+6])<<48 | uint64(p.buckets[pos+7])<<56
		// 9 dirbits per tag
		tag &= 0xffff000000000000
		tag |= uint64(codeword) | uint64(highBits[0])<<8 |
			uint64(highBits[1])<<17 | uint64(highBits[2])<<26 | uint64(highBits[3])<<35

		p.buckets[pos] = byte(tag)
		p.buckets[pos+1] = byte(tag >> 8)
		p.buckets[pos+2] = byte(tag >> 16)
		p.buckets[pos+3] = byte(tag >> 24)
		p.buckets[pos+4] = byte(tag >> 32)
		p.buckets[pos+5] = byte(tag >> 40)
		p.buckets[pos+6] = byte(tag >> 48)
		p.buckets[pos+7] = byte(tag >> 56)

	} else if p.kBitsPerBucket == 64 {
		// 13 dirbits per tag
		tag := uint64(codeword) | uint64(highBits[0])<<8 |
			uint64(highBits[1])<<21 | uint64(highBits[2])<<34 |
			uint64(highBits[3])<<47

		p.buckets[pos] = byte(tag)
		p.buckets[pos+1] = byte(tag >> 8)
		p.buckets[pos+2] = byte(tag >> 16)
		p.buckets[pos+3] = byte(tag >> 24)
		p.buckets[pos+4] = byte(tag >> 32)
		p.buckets[pos+5] = byte(tag >> 40)
		p.buckets[pos+6] = byte(tag >> 48)
		p.buckets[pos+7] = byte(tag >> 56)
	}
}
