package cuckoo

import (
	"fmt"
	"strconv"
	"testing"
)

func TestNewCuckooFilter(t *testing.T) {
	ts := []struct {
		numKeys       uint32
		tagsPerBucket uint32
		bitsPerItem   uint32
		table         Table
	}{
		{
			// default param
			numKeys: 10000,
		},
		{
			numKeys:     10000,
			bitsPerItem: 8,
		},
		{
			// default param
			numKeys:     10000,
			bitsPerItem: 12,
		},
		{
			// default param
			numKeys:     10000,
			bitsPerItem: 13,
			table:       NewPackedTable(),
		},
	}

	for k, te := range ts {
		filter := NewCuckooFilter(
			WithNumKeys(te.numKeys),
			WithBitsPerItem(te.bitsPerItem),
			WithTagsPerBucket(te.tagsPerBucket),
			WithTable(te.table),
		)

		var numInserted int
		for i := 0; i < int(te.numKeys); i++ {
			bs := []byte(strconv.Itoa(i))
			if !filter.Insert(bs) {
				fmt.Printf("break %v\n", i)
				break
			}
			numInserted++
			if !filter.Contain(bs) {
				t.Errorf("find %v fail", i)
			}

			for i := 0; i < numInserted; i++ {
				bs := []byte(strconv.Itoa(i))
				if !filter.Contain(bs) {
					t.Errorf("find %v fail wtf", i)
				}
			}
		}

		var fail int
		// Check if previously inserted items are in the filter, expected
		// true for all items
		for i := 0; i < numInserted; i++ {
			bs := []byte(strconv.Itoa(i))
			if !filter.Contain(bs) {
				fail++
				t.Errorf("find %v fail", i)
			}
		}
		fmt.Printf("total %v inserted %v find fail %v\n", te.numKeys, numInserted, fail)

		var falsePositive, total int
		for i := te.numKeys; i < 2*te.numKeys; i++ {
			bs := []byte(strconv.Itoa(int(i)))
			if filter.Contain(bs) {
				falsePositive++
			}

			total++
		}

		fmt.Printf("false positive rate is %v \n", 100.0*falsePositive/total)

		for i := 0; i < numInserted; i++ {
			bs := []byte(strconv.Itoa(i))
			if !filter.Contain(bs) {
				t.Errorf("find %v fail", i)
			}

			if !filter.Delete(bs) {
				t.Errorf("delete fail %v", i)
			}

			if filter.Contain(bs) {
				t.Errorf("cur %v find %v fail", k, i)
			}
		}
	}
}

func TestCuckoo_Delete(t *testing.T) {
	ts := []struct {
		numKeys       uint32
		tagsPerBucket uint32
		bitsPerItem   uint32
		table         Table
	}{
		{
			// default param
			numKeys: 10000,
		},
		{
			// default param
			numKeys:     10000,
			bitsPerItem: 8,
		},
		{
			// default param
			numKeys:     10000,
			bitsPerItem: 12,
		},
		{
			// default param
			numKeys:     10000,
			bitsPerItem: 13,
			table:       NewPackedTable(),
		},
	}

	for _, te := range ts {
		filter := NewCuckooFilter(
			WithNumKeys(te.numKeys),
			WithBitsPerItem(te.bitsPerItem),
			WithTagsPerBucket(te.tagsPerBucket),
			WithTable(te.table),
		)

		var numInserted int
		for i := 0; i < int(te.numKeys); i++ {
			bs := []byte(strconv.Itoa(i))
			if !filter.Insert(bs) {
				fmt.Printf("break %v\n", i)
				break
			}
			numInserted++
			if !filter.Contain(bs) {
				t.Errorf("find %v fail", i)
			}

			for i := 0; i < numInserted; i++ {
				bs := []byte(strconv.Itoa(i))
				if !filter.Contain(bs) {
					t.Errorf("find %v fail wtf", i)
				}
			}
		}

		var falsePositive, total int
		for i := te.numKeys; i < 2*te.numKeys; i++ {
			bs := []byte(strconv.Itoa(int(i)))
			if filter.Contain(bs) {
				falsePositive++
			}

			total++
		}

		fmt.Printf("false positive rate is %v \n", 100.0*falsePositive/total)

		for i := 0; i < numInserted; i++ {
			bs := []byte(strconv.Itoa(i))
			if !filter.Contain(bs) {
				t.Errorf("find %v fail", i)
			}

			if !filter.Delete(bs) {
				t.Errorf("delete fail %v", i)
			}
		}

		for i := 0; i < numInserted; i++ {
			bs := []byte(strconv.Itoa(i))
			if filter.Contain(bs) {
				t.Errorf("find %v fail", i)
			}
		}
	}
}
