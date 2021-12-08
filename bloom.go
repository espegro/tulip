package main

import (
	"github.com/twmb/murmur3"
	"math/rand"
	"time"
)

//
// Bloom filter functions
//

// Returns a new BloomFilter object,
func NewBloom(size uint, hashnum uint, decay uint, max uint) *BloomFilter {
	b := new(BloomFilter)
	b.Elements = make([]uint, 1)
	b.Rounds = hashnum
	b.Size = size
	b.Decay = decay
	b.Max = max
	b.Bitset = make([]uint, size)
	b.Elements[0] = 0
	return b
}

// Add item to filter
func (b *BloomFilter) Add(item []byte) {
	rand.Seed(time.Now().UnixNano())
	maplock.Lock()
	for r := uint(0); r < b.Decay; r++ {
		p := rand.Intn(int(b.Size))
		v := int(b.Bitset[p])
		v = v - 1
		if v < 0 {
			v = 0
		}
		b.Bitset[p] = uint(v)
	}
	for i := uint(0); i < b.Rounds; i++ {
		h64 := murmur3.SeedNew64(uint64(i))
		h64.Write(item)
		b.Bitset[h64.Sum64()%uint64(b.Size)] = b.Max
	}
	b.Elements[0] = b.Elements[0] + 1
	maplock.Unlock()
}

// Test if item is in filter
func (b *BloomFilter) Test(item []byte) bool {
	maplock.Lock()
	defer maplock.Unlock()
	for i := uint(0); i < b.Rounds; i++ {
		h64 := murmur3.SeedNew64(uint64(i))
		h64.Write(item)
		if b.Bitset[h64.Sum64()%uint64(b.Size)] > 0 {
			return true
		} else {
			return false
		}
	}
	return false
}

// Reset filter
func (b *BloomFilter) Reset() {
	maplock.Lock()
	defer maplock.Unlock()
	for i := uint(0); i < b.Size; i++ {
		b.Bitset[i] = 0
	}
	b.Elements[0] = 0
}

// Add if not set
func (b *BloomFilter) AddIfNotSet(item []byte) bool {
	if !b.Test(item) {
		b.Add(item)
		return false
	}
	return true
}
