package roaring

import (
	"fmt"	
	"cmp"
)

const ROARING_THRESHOLD = 4096

const WORD_SIZE = 64
type WORD_TYPE uint64

type ContainerKind uint8
const (
	BITMAP ContainerKind = iota
	VECTOR
)

type Roaring struct {
	entries []Entry
	size uint64 // max 2^32
}

type Entry struct {
	key uint16
	container *Container
}

type Container struct {
	// 2^16 = 2^6 * 2^10 where 2^6 = 64 (bits)
	kind ContainerKind
	bitmap *[1 << 10]WORD_TYPE
	vector []uint16
	size int // max 2^16
}

func makeRoaring() Roaring {
	return Roaring{}
}

func (r *Roaring) Add(item uint32) (bool, error) {
	upper, lower := uint16(item >> 16), uint16(item & 0xFFFF)

	conversionFunc := func(entry Entry) uint16 { return entry.key } 
	insertionIdx, alreadyExists := getInsertionIdx(r.entries, upper, conversionFunc)
	if !alreadyExists {
		container := makeContainer()
		entries, err := addToSlice(r.entries, insertionIdx, Entry{upper, container})
		if err != nil { return false, err }

		r.entries = entries
	}

	var container *Container = r.entries[insertionIdx].container
	added, err := container.Add(lower)
	if err == nil && added == true {
		r.size++
	}
	return added, err
}

func (r * Roaring) Size() uint64 {
	return r.size
}

func makeContainer() *Container {
	var vec []uint16
	return new(Container{VECTOR, nil, vec, 0})
}

func (c *Container) Add(item uint16) (bool, error) {
	switch c.kind {
	case BITMAP:
		return c.addToBitMap(item)
	case VECTOR:
		res1, res2 := c.addToVector(item)
		if c.size >= ROARING_THRESHOLD {
			c.changeToBitMap()
		}
		return res1, res2
	default:
		panic("unknown container kind")
	}
}

func (c *Container) addToBitMap(item uint16) (bool, error) {
	if c.kind != BITMAP { 
		return false, fmt.Errorf("container is not a bitmap")
	}
	wordIdx, bit := item / WORD_SIZE, item % WORD_SIZE
	
	if (c.bitmap[wordIdx] >> bit) & 1 == 1 {
		return false, nil
	}

	c.bitmap[wordIdx] |= 1 << bit 
	c.size++
	return true, nil
}

func (c *Container) addToVector(item uint16) (bool, error) {
	if c.kind != VECTOR { 
		return false, fmt.Errorf("container is not a vector")
	}

	// lo = hi + 1 and vec[hi] < num < vec[lo]
	// so insert at lo idx
	insertionIdx, alreadyExists := getInsertionIdx(c.vector, item, func(num uint16) uint16 { return num } )
	if alreadyExists { return false, nil }

	vec, err := addToSlice(c.vector, insertionIdx, item)
	if err != nil { return false, err }
	
	c.vector = vec
	c.size++
	return true, nil
}

func getInsertionIdx[T any, R cmp.Ordered](s []T, target R, conversionFunc func(T) R) (int, bool) {
	lo, hi := 0, len(s) - 1
	for lo <= hi {
		mid := lo + (hi - lo) / 2
		atMid := conversionFunc(s[mid])
		if atMid == target {
			return mid, true
		} else if (target < atMid) {
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}
	return lo, false
}

func (c *Container) changeToBitMap() error {
	if c.size < ROARING_THRESHOLD {
		return fmt.Errorf("size not great enough to warrant change to bitmap: %d", c.size)
	}

	c.kind = BITMAP

	c.bitmap = new([1 << 10]WORD_TYPE)
	c.size = 0

	for _, item := range c.vector {
		c.addToBitMap(item)
	}
	c.vector = nil
	return nil
}

func addToSlice[T any](s []T, idx int, item T) ([]T, error) {
	if idx < 0 || idx > len(s) {
		return nil, fmt.Errorf("index out of bounds: %d", idx)
	}

	var dummy T
	s = append(s, dummy)
	copy(s[idx+1:],s[idx:])
	s[idx] = item
	return s, nil
}
