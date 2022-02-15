package cuckoo

import (
	"hash"
	"hash/maphash"
)

// Options cuckoo options
type Options struct {
	hf            hash.Hash64
	kicks         int
	numKeys       uint32
	tagsPerBucket uint32
	bitsPerItem   uint32
	table         Table
}

func (o *Options) apply() {
	if o.kicks == 0 {
		o.kicks = 500
	}

	if o.hf == nil {
		hf := &maphash.Hash{}
		hf.SetSeed(hf.Seed())
		o.hf = hf
	}

	if o.table == nil {
		o.table = &singleTable{}
	}

	if o.tagsPerBucket == 0 {
		o.tagsPerBucket = 4
	}

	if o.numKeys == 0 {
		o.numKeys = 10000
	}

	if o.bitsPerItem == 0 {
		o.bitsPerItem = 16
	}
}

type Option func(options *Options)

// WithHash
func WithHash(hf hash.Hash64) Option {
	return func(options *Options) {
		options.hf = hf
	}
}

// WithKickCount
func WithKickCount(kicks int) Option {
	return func(options *Options) {
		options.kicks = kicks
	}
}

func WithTable(t Table) Option {
	return func(options *Options) {
		options.table = t
	}
}

func WithNumKeys(n uint32) Option {
	return func(options *Options) {
		options.numKeys = n
	}
}

func WithTagsPerBucket(n uint32) Option {
	return func(options *Options) {
		options.tagsPerBucket = n
	}
}

// WithBitsPerItem per item has bits count
func WithBitsPerItem(n uint32) Option {
	return func(options *Options) {
		options.bitsPerItem = n
	}
}

type victim struct {
	index uint32
	tag   uint32
	used  bool
}

type Cuckoo struct {
	opt         Options
	count       uint32
	numBucket   uint32
	bitsPerItem uint32
	table       Table
	victim      victim
}

// NewCuckooFilter
func NewCuckooFilter(opts ...Option) *Cuckoo {
	var opt Options
	for _, o := range opts {
		o(&opt)
	}
	opt.apply()
	numBucket := upperPow32(opt.numKeys / opt.tagsPerBucket)
	frac := float64(opt.numKeys) / float64(numBucket*opt.tagsPerBucket)
	if frac > 0.96 {
		numBucket <<= 1
	}
	opt.table.Init(numBucket, opt.tagsPerBucket, opt.bitsPerItem)
	return &Cuckoo{
		opt:         opt,
		table:       opt.table,
		numBucket:   numBucket,
		bitsPerItem: opt.bitsPerItem,
		count:       0,
	}
}

/*
	f = fingerprint(x);
	i1 = hash(x);
	i2 = i1 ⊕ hash(f);
	if bucket[i1] or bucket[i2] has an empty entry then
		add f to that bucket;
		return Done;
	// must relocate existing items;
	i = randomly pick i1 or i2;
	for n = 0; n < MaxNumKicks; n++ do
		randomly select an entry e from bucket[i];
		swap f and the fingerprint stored in entry e;
		i = i ⊕ hash(f);
		if bucket[i] has an empty entry then
			add f to bucket[i];
			return Done;
	// Hashtable is considered full;
	return Failure;
*/
func (c *Cuckoo) Insert(x []byte) bool {
	if c.victim.used {
		return false
	}

	i, tag := c.generateIndexTagHash(x)
	return c.insert(i, tag)
}

/*
	f = fingerprint(x);
	i1 = hash(x);
	i2 = i1 ⊕ hash(f);
	if bucket[i1] or bucket[i2] has f then
		return True;
	return False;
*/
func (c *Cuckoo) Contain(item []byte) bool {
	i1, tag := c.generateIndexTagHash(item)
	i2 := c.altIndex(i1, tag)

	if i1 != c.altIndex(i2, tag) {
		panic("what happened before")
	}

	if c.victim.used &&
		c.victim.tag == tag &&
		(c.victim.index == i1 || c.victim.index == i2) {
		return true
	}

	return c.table.Find(i1, tag) || c.table.Find(i2, tag)
}

func (c *Cuckoo) Delete(item []byte) bool {
	i1, tag := c.generateIndexTagHash(item)
	i2 := c.altIndex(i1, tag)

	if c.victim.used &&
		c.victim.tag == tag &&
		(c.victim.index == i1 || c.victim.index == i2) {
		c.victim.used = false
		return true
	}

	if !c.table.Delete(i1, tag) && !c.table.Delete(i2, tag) {
		return false
	}

	c.count--
	// delete success
	if !c.victim.used {
		return true
	}

	// reinsert victim
	c.insert(c.victim.index, c.victim.tag)
	return true
}

func (c *Cuckoo) insert(i uint32, tag uint32) bool {
	var ok bool
	for cnt := 0; cnt < c.opt.kicks; cnt++ {
		kickout := cnt > 0
		tag, ok = c.table.Insert(i, tag, kickout)
		if ok {
			c.count++
			return true
		}

		i = c.altIndex(i, tag)
	}

	c.victim = victim{
		index: i,
		tag:   tag,
		used:  true,
	}
	return true
}

func (c *Cuckoo) LoadFactor() float64 {
	return 1.0 * float64(c.count) / float64(c.table.SizeInTags())
}

func (c *Cuckoo) BitsPerItem() float64 {
	return 8.0 * float64(c.table.SizeInTags()) / float64(c.count)
}

func (c *Cuckoo) generateIndexTagHash(item []byte) (i, tag uint32) {
	hs := hash64(item, c.opt.hf)
	return c.indexHash(uint32(hs >> 32)), c.tagHash(uint32(hs))
}

func (c *Cuckoo) indexHash(hv uint32) uint32 {
	return hv & (c.numBucket - 1)
}

func (c *Cuckoo) altIndex(i, tag uint32) uint32 {
	// 0x5bd1e995 is the hash constant from MurmurHash2
	return c.indexHash(i ^ (tag * 0x5bd1e995))
}

func (c *Cuckoo) tagHash(hv uint32) uint32 {
	tag := hv & ((1 << c.bitsPerItem) - 1)
	if tag == 0 {
		tag = 1
	}

	return tag
}
