# cuckoo filter

## Overview

Cuckoo filter is a Bloom filter replacement for approximated set-membership queries. While Bloom filters are well-known space-efficient data structures to serve queries like "if item x is in a set?", they do not support deletion. Their variances to enable deletion (like counting Bloom filters) usually require much more space.

Cuckoo ﬁlters provide the ﬂexibility to add and remove items dynamically. A cuckoo filter is based on cuckoo hashing (and therefore named as cuckoo filter). It is essentially a cuckoo hash table storing each key's fingerprint. Cuckoo hash tables can be highly compact, thus a cuckoo filter could use less space than conventional Bloom ﬁlters, for applications that require low false positive rates (< 3%).

For details about the algorithm and citations please use:

["Cuckoo Filter: Practically Better Than Bloom"](http://www.cs.cmu.edu/~binfan/papers/conext14_cuckoofilter.pdf) in proceedings of ACM CoNEXT 2014 by Bin Fan, Dave Andersen and Michael Kaminsky

## API
A cuckoo filter supports following operations:
+ Insert([]byte)  insert an item to the filter
+ Contain([]byte) return if item is already in the filter. Note that this method may return false positive results like Bloom filters
+ Delete([]byte) delete the given item from the filter. Note that to use this method, it must be ensured that this item is in the filter (e.g., based on records on external storage); otherwise, a false item may be deleted.

## Example usage:
```go
// default option
filter := NewCuckooFilter()
// insert item "12" into filter
filter.Insert([]byte("12"))

assert(filter.Contain("12") == true)
```

