package roaring

import (
	"fmt"	
	"cmp"
	"math/bits"
	"slices"
)

const (
	ROARING_THRESHOLD = 4096
	WORD_SIZE = 64
	MAX_CONTAINER_SIZE = 1 << 16
	CONTAINER_BITMAP_SIZE = MAX_CONTAINER_SIZE / WORD_SIZE
)
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
	kind ContainerKind
	bitmap *[CONTAINER_BITMAP_SIZE]WORD_TYPE
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
		r.entries = slices.Insert(r.entries, insertionIdx, Entry{upper, container})
	}

	var container *Container = r.entries[insertionIdx].container
	added, err := container.add(lower)
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

func containerFromBitMap(bitmap *[CONTAINER_BITMAP_SIZE]WORD_TYPE) *Container {
	size := bitMapOneBits(bitmap)
	if size < ROARING_THRESHOLD {
		res := makeContainer() 
		for wordIdx, word := range bitmap {
			for i := 0; i < WORD_SIZE; i++ {
				if word & 1 == 1 {
					res.add(uint16(wordIdx * WORD_SIZE + i))
				}
				word >>= 1
			}
		}
		return res
	} else {
		return new(Container{BITMAP, bitmap, nil, size})
	}
}

func bitMapOneBits(bitmap *[CONTAINER_BITMAP_SIZE]WORD_TYPE) int {
	oneBits := 0
	for _, word := range bitmap {
		oneBits += bits.OnesCount64(uint64(word))
	}
	return oneBits
}

func (c *Container) add(item uint16) (bool, error) {
	return c.find(item, true, false)
}

func (c *Container) remove(item uint16) (bool, error) {
	return c.find(item, false, true)
}

func (c *Container) contains(item uint16) (bool, error) {
	return c.find(item, false, false)
}

func (c *Container) find(item uint16, addIfMissing bool, removeIfExists bool) (bool, error) {
	if addIfMissing && removeIfExists {
		return false, fmt.Errorf("addIfMissing and removeIfExists cannot both be true")
	}

	switch c.kind {
	case BITMAP:
		return c.findInBitMap(item, addIfMissing, removeIfExists)
	case VECTOR:
		res1, res2 := c.findInVector(item, addIfMissing, removeIfExists)
		if c.size >= ROARING_THRESHOLD {
			c.changeToBitMap()
		}
		return res1, res2
	default:
		panic("unknown container kind")
	}
}

func (c *Container) findInBitMap(item uint16, addIfMissing bool, removeIfExists bool) (bool, error) {
	if c.kind != BITMAP { 
		return false, fmt.Errorf("container is not a bitmap")
	}
	wordIdx, bit := item / WORD_SIZE, item % WORD_SIZE
	
	itemExists := (c.bitmap[wordIdx] >> bit) & 1 == 1

	if itemExists {
		if removeIfExists {
			c.bitmap[wordIdx] &= ^(WORD_TYPE(1) << bit)
			c.size--
			return true, nil
		}
		return !addIfMissing, nil
	} else {
		if addIfMissing {
			c.bitmap[wordIdx] |= WORD_TYPE(1) << bit 
			c.size++
			return true, nil
		}
		return !removeIfExists, nil
	}	
}

func (c *Container) findInVector(item uint16, addIfMissing bool, removeIfExists bool) (bool, error) {
	if c.kind != VECTOR { 
		return false, fmt.Errorf("container is not a vector")
	}

	itemIdx, alreadyExists := getInsertionIdx(c.vector, item, func(num uint16) uint16 { return num } )
	if alreadyExists {

		if removeIfExists {
			c.vector = slices.Delete(c.vector, itemIdx, itemIdx+1)
			c.size--
			return true, nil
		}

		return !addIfMissing, nil

	} else {

		if addIfMissing {
			c.vector = slices.Insert(c.vector, itemIdx, item)
			c.size++
			return true, nil
		}

		return !removeIfExists, nil
	}
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

	c.bitmap = new([CONTAINER_BITMAP_SIZE]WORD_TYPE)
	c.size = 0

	for _, item := range c.vector {
		c.add(item)
	}
	c.vector = nil
	return nil
}

func (c *Container) intersect(o *Container) (*Container, error) {
	switch o.kind {
	case VECTOR:
		return c.intersectVector(o)
	case BITMAP:
		return c.intersectBitMap(o)
	default:
		panic("unrecognized container kind")
	}
}

func (c *Container) intersectVector(o *Container) (*Container, error) {
	switch c.kind {
	case VECTOR:
		res := makeContainer()
		i, j := 0, 0
		for i < c.size && j < o.size {
			if c.vector[i] < o.vector[j] {
				i++
			} else if o.vector[j] < c.vector[i] {
				j++
			} else {
				ok, err := res.add(c.vector[i])
				if err != nil { return nil, err }
				if !ok { return nil, fmt.Errorf("result already contained %d", c.vector[i]) }
				i++
				j++
			}
		}
		return res, nil
	case BITMAP:
		return o.intersectBitMap(c)
	default:
		panic("unrecognized container kind")
	}
}

func (c *Container) intersectBitMap(o *Container) (*Container, error) {
	switch c.kind {
	case VECTOR:
		res := makeContainer()

		for i := range c.vector {
			item := c.vector[i]
			wordIdx, bit := item / WORD_SIZE, item % WORD_SIZE
			
			if (c.bitmap[wordIdx] >> bit) & 1 == 1 {
				ok, err := res.add(c.vector[i])
				if err != nil { return nil, err }
				if !ok { return nil, fmt.Errorf("result already contained %d", c.vector[i]) }
			}
		}

		return res, nil
	case BITMAP:
		var bitmap [CONTAINER_BITMAP_SIZE]WORD_TYPE
		for i := range c.bitmap {
			bitmap[i] = c.bitmap[i] & o.bitmap[i]
		}
		return containerFromBitMap(&bitmap), nil
		
	default:
		panic("unrecognized container kind")
	}
}

func (c *Container) union(o *Container) {

}
