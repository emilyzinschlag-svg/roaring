package roaring 

import (
	"testing"
)

type adapter[T any] interface {
	add(uint32) (bool, error)
	remove(uint32) (bool, error)
	contains(uint32) (bool, error)
	union(T) (T, error)
	intersect(T) (T, error)
	size() uint64
}

type roaringAdapter struct {
	r *Roaring
}

func (r *Roaring) toAdapter() *roaringAdapter {
	return new (roaringAdapter{r})
}

func (s roaringAdapter) add(item uint32) (bool, error) { 
	return s.r.Add(item)
}

func (s roaringAdapter) remove(item uint32) (bool, error) { 
	return s.r.Remove(item)
}

func (s roaringAdapter) contains(item uint32) (bool, error) { 
	return s.r.Contains(item)
}

func (s roaringAdapter) union(o *Roaring) (*Roaring, error) { 
	return s.r.Union(o)
}

func (s roaringAdapter) intersect(o *Roaring) (*Roaring, error) { 
	return s.r.Intersect(o)
}

func (s roaringAdapter) size() uint64 { 
	return s.r.Size()
}

type containerAdapter struct {
	c *Container
}

func (c *Container) toAdapter() *containerAdapter {
	return new(containerAdapter{c})
}

func (s containerAdapter) convert(item uint32) uint16 {
	if item >= 1 << 16 { panic("item greater than max uint16") }
	return uint16(item)
}

func (s containerAdapter) add(item uint32) (bool, error) { 
	return s.c.add(s.convert(item))
}

func (s containerAdapter) remove(item uint32) (bool, error) { 
	return s.c.remove(s.convert(item))
}

func (s containerAdapter) contains(item uint32) (bool, error) { 
	return s.c.contains(s.convert(item))
}

func (s containerAdapter) union(o *Container) (*Container, error) { 
	return s.c.union(o)
}

func (s containerAdapter) intersect(o *Container) (*Container, error) { 
	return s.c.intersect(o)
}

func (s containerAdapter) size() uint64 { 
	return uint64(s.c.size)
}

func dummy[T any](a adapter[T]) {}

func dummy2() {
	r := MakeRoaring()
	a := roaringAdapter{r}
	dummy(a)
}

func dummy3() {
	c := makeContainer()
	a := containerAdapter{c}
	dummy(a)
}

func RoaringRemoveTester(r *Roaring, shuffledItems []uint32, t *testing.T) {
	removeAllTester(r.toAdapter(), shuffledItems, t)
}

func containerRemoveAllTester(c *Container, shuffledItems []uint16, t *testing.T) {
	castedShuffledItems := make([]uint32, len(shuffledItems))
	for i, item := range shuffledItems {
		castedShuffledItems[i] = uint32(item)
	}
	removeAllTester(c.toAdapter(), castedShuffledItems, t)
}

func removeAllTester[T any](a adapter[T], shuffledItems []uint32, t *testing.T) {
	for _, item := range shuffledItems {
		sizeBefore := a.size()

		gotAdded, err := a.add(item)
		if err != nil {
			t.Fatal(err.Error())
		}
		if gotAdded {
			t.Fatalf("add should return false for item already in: %d", item)
		}
		if sizeBefore != a.size() {
			t.Fatalf("add should not affect size for item already in: %d", item)
		}

		contained, err := a.contains(item)
		if err != nil {
			t.Fatal(err.Error())
		}
		if !contained {
			t.Fatalf("contains should return true for item already in: %d", item)
		}
		if sizeBefore != a.size() {
			t.Fatalf("contains should not affect size. item: %d", item)
		}

		removed, err := a.remove(item)
		if err != nil {
			t.Fatal(err.Error())
		}
		if !removed {
			t.Fatalf("removed should return true for item already in: %d", item)
		}
		if sizeBefore - 1 != a.size() {
			t.Fatalf("want remove to decrement container size: %d", item)
		}

		sizeBefore = a.size()
		contained, err = a.contains(item)
		if err != nil {
			t.Fatal(err.Error())
		}
		if contained {
			t.Fatalf("contains should return false for removed item: %d", item)
		}
		if sizeBefore != a.size() {
			t.Fatalf("contains should not affect size. item: %d", item)
		}
	}
	if a.size() != 0 {
		t.Errorf("empty container should have size 0, got %d", a.size())
	}
}