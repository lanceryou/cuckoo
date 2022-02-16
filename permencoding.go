package cuckoo

const (
	N_ENTS = 3876
)

type PermEncoding struct {
	/* unpack one 2-byte number to four 4-bit numbers */
	// inline void unpack(const uint16_t in, const uint8_t out[4]) const {
	//     (*(uint16_t *)out)      = in & 0x0f0f;
	//     (*(uint16_t *)(out +2)) = (in >> 4) & 0x0f0f;
	// }
	decTable [N_ENTS]uint16
	encTable [1 << 16]uint16
}

func NewPermEncoding() *PermEncoding {
	p := &PermEncoding{}
	var idx uint16
	p.genTables(0, 0, [4]uint8{}, &idx)
	return p
}

func (p *PermEncoding) Encode(lowBits [4]uint8) uint16 {
	return p.encTable[p.pack(lowBits)]
}

func (p *PermEncoding) Decode(codeword uint16) [4]uint8 {
	return p.unpack(p.decTable[codeword])
}

func (p *PermEncoding) DecItem(codeword uint16) uint16 {
	return p.decTable[codeword]
}

func (p *PermEncoding) unpack(in uint16) [4]uint8 {
	var out [4]uint8
	out[0] = uint8(in & 0x000f)
	out[2] = uint8((in >> 4) & 0x000f)
	out[1] = uint8((in >> 8) & 0x000f)
	out[3] = uint8((in >> 12) & 0x000f)
	return out
}

func (p *PermEncoding) pack(in [4]uint8) uint16 {
	in1 := (uint16(in[0]) | uint16(in[1])<<8) & 0x0f0f
	in2 := (uint16(in[2]) | uint16(in[3])<<8) << 4

	return in1 | in2
}

func (p *PermEncoding) genTables(base, k int, dst [4]uint8, idx *uint16) {
	for i := base; i < 16; i++ {
		/* for fast comparison in binary_search in little-endian machine */
		dst[k] = uint8(i)
		if k+1 < 4 {
			p.genTables(i, k+1, dst, idx)
		} else {
			p.decTable[*idx] = p.pack(dst)
			p.encTable[p.pack(dst)] = *idx
			*idx++
		}
	}
}
