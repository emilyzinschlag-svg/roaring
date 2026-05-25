package roaring

import (
	"cmp"
	"fmt"
	"math/bits"
	"slices"
)

const (
	PROMOTION_THRESHOLD   = 4096
	DEMOTION_THRESHOLD    = 3072
	WORD_SIZE             = 64
	MAX_CONTAINER_SIZE    = 1 << 16
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
	size    uint64 // max 2^32
}

type Entry struct {
	key       uint16
	container *Container
}

type Container struct {
	kind   ContainerKind
	bitmap *[CONTAINER_BITMAP_SIZE]WORD_TYPE
	vector []uint16
	size   int // max 2^16
}

func makeRoaring() Roaring {
	return Roaring{}
}

func (r *Roaring) Add(item uint32) (bool, error) {
	return r.find(item, true, false)
}

func (r *Roaring) Remove(item uint32) (bool, error) {
	return r.find(item, false, true)
}

func (r *Roaring) Contains(item uint32) (bool, error) {
	return r.find(item, false, false)
}

func (r *Roaring) find(item uint32, addIfMissing bool, removeIfExists bool) (bool, error) {
	if addIfMissing && removeIfExists {
		return false, fmt.Errorf("addIfMissing and removeIfExists cannot both be true")
	}

	upper, lower := uint16(item >> 16), uint16(item & 0xFFFF)

	conversionFunc := func(entry Entry) uint16 { return entry.key }
	containerIdx, alreadyExists := getInsertionIdx(r.entries, upper, conversionFunc)

	if !alreadyExists {
		if addIfMissing {
			container := makeContainer()
			r.entries = slices.Insert(r.entries, containerIdx, Entry{upper, container})
		} else {
			return false, nil
		}
	}

	container := r.entries[containerIdx].container

	res, err := container.find(lower, addIfMissing, removeIfExists)
	if err != nil { return false, err }

	if res && addIfMissing { r.size++ }
	if res && removeIfExists { r.size-- }

	if container.size == 0 {
		r.entries = slices.Delete(r.entries, containerIdx, containerIdx+1)
	}

	return res, nil
}

func (r *Roaring) Size() uint64 {
	return r.size
}

func (r *Roaring) intersect(o *Roaring) (*Roaring, error) {
	res := new(makeRoaring())
	i, j := 0, 0

	for i < len(r.entries) && j < len(o.entries) {
		e1, e2 := r.entries[i], o.entries[j]
		
		if e1.key < e2.key {
			i++
		} else if e1.key > e2.key {
			j++
		} else {
			cont, err := e1.container.intersect(e2.container)
			if err != nil { return nil, err }

			if cont.size != 0 {
				res.entries = append(res.entries, Entry{e1.key, cont})
			}

			i++
			j++
		}
	}

	return res, nil
}

func (r *Roaring) union(o *Roaring) (*Roaring, error) {
	res := new(makeRoaring())
	i, j := 0, 0

	for i < len(r.entries) && j < len(o.entries) {
		e1, e2 := r.entries[i], o.entries[j]

		if e1.key < e2.key {
			res.entries = append(res.entries, *copyEntry(&e1))
			i++
		} else if e1.key > e2.key {
			res.entries = append(res.entries, *copyEntry(&e2))
			j++
		} else {
			cont, err := e1.container.union(e2.container)
			if err != nil { return nil, err }

			res.entries = append(res.entries, Entry{e1.key, cont})
			i++
			j++
		}
	}

	res.entries = append(res.entries, copyEntries(r.entries[i:])...)
	res.entries = append(res.entries, copyEntries(o.entries[j:])...)

	return res, nil
}

func copyEntries(s []Entry) []Entry {
	for i, e := range s {
		s[i] = *copyEntry(&e)
	}
	return s
}

func copyEntry(e *Entry) *Entry {
	return new(Entry{e.key, copyContainer(e.container)})
}

func copyContainer(c *Container) *Container {
	res := new(Container{c.kind, c.bitmap, c.vector, c.size})
	switch c.kind {
	case BITMAP:
		bitmap := new([CONTAINER_BITMAP_SIZE]WORD_TYPE)
		*bitmap = *c.bitmap 
		res.bitmap = bitmap
	case VECTOR:
		vector := slices.Clone(c.vector)
		res.vector = vector
	default:
		panic("unrecognized kind")
	}

	return res
}

func makeContainer() *Container {
	var vec []uint16
	return new(Container{VECTOR, nil, vec, 0})
}

func containerFromBitMap(bitmap *[CONTAINER_BITMAP_SIZE]WORD_TYPE) *Container {
	size := bitMapOneBits(bitmap)
	if size < PROMOTION_THRESHOLD {
		res := makeContainer()
		res.vector = bitMapToVector(bitmap)

		res.size = len(res.vector)
		return res
	} else {
		return new(Container{BITMAP, bitmap, nil, size})
	}
}

func containerFromVector(vector []uint16) (*Container, error) {
	if len(vector) < PROMOTION_THRESHOLD {
		return new(Container{VECTOR, nil, vector, len(vector)}), nil
	} else {
		bitmap := new([CONTAINER_BITMAP_SIZE]WORD_TYPE)
		prev := -1
		for _, item := range vector {
			if int(item) <= prev {
				err := fmt.Errorf("vector is not strictly increasing")
				return nil, err
			}
			prev = int(item)

			wordIdx, bit := item / WORD_SIZE, item % WORD_SIZE
			bitmap[wordIdx] |= WORD_TYPE(1) << bit
		}

		return new(Container{BITMAP, bitmap, nil, len(vector)}), nil
	}
}

func bitMapOneBits(bitmap *[CONTAINER_BITMAP_SIZE]WORD_TYPE) int {
	oneBits := 0
	for _, word := range bitmap {
		oneBits += bits.OnesCount64(uint64(word))
	}
	return oneBits
}

func bitMapToVector(bitmap *[CONTAINER_BITMAP_SIZE]WORD_TYPE) []uint16 {
	var res []uint16
	for wordIdx, word := range bitmap {
		if word == 0 {
			continue
		}
		for i := 0; i < WORD_SIZE; i++ {
			if word & 1 == 1 {
				res = append(res, uint16(wordIdx*WORD_SIZE + i))
			}
			word >>= 1
		}
	}
	return res
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
		res, err := c.findInBitMap(item, addIfMissing, removeIfExists)
		if err != nil { return false, err }

		if c.size <= DEMOTION_THRESHOLD {
			err = c.changeToVector()
			if err != nil { return false, err }
		}

		return res, err

	case VECTOR:
		res, err := c.findInVector(item, addIfMissing, removeIfExists)
		if err != nil { return res, err }

		if c.size >= PROMOTION_THRESHOLD {
			err = c.changeToBitMap()
			if err != nil { return false, err }
		}

		return res, err
	default:
		panic("unknown container kind")
	}
}

func (c *Container) findInBitMap(item uint16, addIfMissing bool, removeIfExists bool) (bool, error) {
	if c.kind != BITMAP {
		return false, fmt.Errorf("container is not a bitmap")
	}
	wordIdx, bit := item/WORD_SIZE, item%WORD_SIZE

	itemExists := (c.bitmap[wordIdx]>>bit)&1 == 1

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
		return false, nil
	}
}

func (c *Container) findInVector(item uint16, addIfMissing bool, removeIfExists bool) (bool, error) {
	if c.kind != VECTOR {
		return false, fmt.Errorf("container is not a vector")
	}

	itemIdx, alreadyExists := getInsertionIdx(c.vector, item, func(num uint16) uint16 { return num })
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

		return false, nil
	}
}

func getInsertionIdx[T any, R cmp.Ordered](s []T, target R, conversionFunc func(T) R) (int, bool) {
	lo, hi := 0, len(s)-1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		atMid := conversionFunc(s[mid])
		if atMid == target {
			return mid, true
		} else if target < atMid {
			hi = mid - 1
		} else {
			lo = mid + 1
		}
	}
	return lo, false
}

func (c *Container) changeToBitMap() error {
	if c.size < PROMOTION_THRESHOLD {
		return fmt.Errorf("size not great enough to warrant change to bitmap: %d", c.size)
	}

	c.kind = BITMAP
	size := c.size


	c.bitmap = new([CONTAINER_BITMAP_SIZE]WORD_TYPE)

	for _, item := range c.vector {
		c.add(item)
	}

	c.size = size
	c.vector = nil
	return nil
}

func (c *Container) changeToVector() error {
	if c.size > DEMOTION_THRESHOLD {
		return fmt.Errorf("size to great to warrant change to vector: %d", c.size)
	}

	c.kind = VECTOR
	bitmap := c.bitmap
	c.bitmap = nil 

	c.vector = bitMapToVector(bitmap)
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
		var resVec []uint16
		i, j := 0, 0

		for i < c.size && j < o.size {
			if c.vector[i] < o.vector[j] {
				i++
			} else if c.vector[i] > o.vector[j] {
				j++
			} else {
				resVec = append(resVec, c.vector[i])
				i++
				j++
			}
		}

		return containerFromVector(resVec)
	case BITMAP:
		return o.intersectBitMap(c)
	default:
		panic("unrecognized container kind")
	}
}

func (c *Container) intersectBitMap(o *Container) (*Container, error) {
	switch c.kind {
	case VECTOR:
		resVec := make([]uint16, 0, min(c.size, o.size))

		for _, item := range c.vector {
			wordIdx, bit := item / WORD_SIZE, item % WORD_SIZE

			if (o.bitmap[wordIdx] >> bit) & 1 == 1 {
				resVec = append(resVec, item)
			}
		}

		return containerFromVector(resVec)
	case BITMAP:
		bitmap := new([CONTAINER_BITMAP_SIZE]WORD_TYPE)
		for i := range c.bitmap {
			bitmap[i] = c.bitmap[i] & o.bitmap[i]
		}

		return containerFromBitMap(bitmap), nil
	default:
		panic("unrecognized container kind")
	}
}

func (c *Container) union(o *Container) (*Container, error) {
	switch o.kind {
	case VECTOR:
		return c.unionVector(o)
	case BITMAP:
		return c.unionBitMap(o), nil
	default:
		panic("unrecognized container kind")
	}
}

func (c *Container) unionVector(o *Container) (*Container, error) {
	switch c.kind {
	case VECTOR:
		resVec := make([]uint16, 0, c.size + o.size) 
		i, j := 0, 0
		for i < c.size && j < o.size {
			if c.vector[i] < o.vector[j] {
				resVec = append(resVec, c.vector[i])
				i++
			} else if c.vector[i] > o.vector[j] {
				resVec = append(resVec, o.vector[j])
				j++
			} else {
				resVec = append(resVec, c.vector[i])
				i++
				j++
			}
		}

		resVec = append(resVec, c.vector[i:]...)
		resVec = append(resVec, o.vector[j:]...)

		return containerFromVector(resVec)
	case BITMAP:
		return o.unionBitMap(c), nil
	default:
		panic("unrecognized container kind")
	}
}

func (c *Container) unionBitMap(o *Container) *Container {
	switch c.kind {
	case VECTOR:
		bitmap := new([CONTAINER_BITMAP_SIZE]WORD_TYPE)
		*bitmap = *o.bitmap

		for i := range c.vector {
			item := c.vector[i]
			wordIdx, bit := item / WORD_SIZE, item % WORD_SIZE
			bitmap[wordIdx] |= 1 << bit
		}

		return containerFromBitMap(bitmap)
	case BITMAP:
		bitmap := new([CONTAINER_BITMAP_SIZE]WORD_TYPE)
		for i := range c.bitmap {
			bitmap[i] = c.bitmap[i] | o.bitmap[i]
		}

		return containerFromBitMap(bitmap)
	default:
		panic("unrecognized container kind")
	}
}
