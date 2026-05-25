package roaring

import (
	"math/rand/v2"
	"slices"
	"testing"
)

type adapter[T any, I AllowedInt] interface {
	add(I) (bool, error)
	remove(I) (bool, error)
	contains(I) (bool, error)
	union(T) (T, error)
	intersect(T) (T, error)
	size() uint64
}

type roaringAdapter struct {
	r *Roaring
}

type AllowedInt interface {
	~uint16 | ~uint32
}

func (r *Roaring) toAdapter() *roaringAdapter { return new (roaringAdapter{r}) }
func (s roaringAdapter) add(item uint32) (bool, error) { return s.r.Add(item) }
func (s roaringAdapter) remove(item uint32) (bool, error) { return s.r.Remove(item) }
func (s roaringAdapter) contains(item uint32) (bool, error) { return s.r.Contains(item) }
func (s roaringAdapter) union(o *Roaring) (*Roaring, error) { return s.r.Union(o) }
func (s roaringAdapter) intersect(o *Roaring) (*Roaring, error) { return s.r.Intersect(o) }
func (s roaringAdapter) size() uint64 { return s.r.Size() }

type containerAdapter struct {
	c *Container
}

func (c *Container) toAdapter() *containerAdapter {
	return new(containerAdapter{c})
}
func (s containerAdapter) add(item uint16) (bool, error) { return s.c.add(item) }
func (s containerAdapter) remove(item uint16) (bool, error) { return s.c.remove(item) }
func (s containerAdapter) contains(item uint16) (bool, error) { return s.c.contains(item) }
func (s containerAdapter) union(o *Container) (*Container, error) { return s.c.union(o) }
func (s containerAdapter) intersect(o *Container) (*Container, error) { return s.c.intersect(o) }
func (s containerAdapter) size() uint64 { return uint64(s.c.size) }

func dummy[T any, I AllowedInt](a adapter[T, I]) {}

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

func Shuffle[T any](s []T, pcgInput uint64) {
	r := rand.New(rand.NewPCG(pcgInput, pcgInput))
	r.Shuffle(len(s), func(i, j int) { s[i], s[j] = s[j], s[i] })
}

func generateVector[I AllowedInt](length, multiple int) []I {
	return generateVectorWithOffset[I](length, multiple, 0)
}

func generateVectorWithOffset[I AllowedInt](length, multiple, offset int) []I {
	res := make([]I, length)
	for i := range res {
		res[i] = I((i * multiple) + offset)
	}
	return res
}

func getSortedUnique[I AllowedInt](m map[I]struct{}) []I {
	res := make([]I, 0, len(m))
	for k := range m {
		res = append(res, k)
	}

	slices.Sort(res)
	return res 
}

func RoaringAddRemoveContainsTester(
		name           string,
		numMultiples   int,
		multiple	   int,
		extraItems	   []uint32,
		pcgInput	   uint64,
		r *Roaring, 
		t *testing.T,
	) {
	addRemoveContainsTester(name, numMultiples, multiple, extraItems, pcgInput, r.toAdapter(), t)
}

func addRemoveContainsTester[T any, I AllowedInt](
		name           string,
		numMultiples   int,
		multiple	   int,
		extraItems	   []I,
		pcgInput	   uint64,
		a adapter[T, I], 
		t *testing.T,
	) {
	uniqueMap := make(map[I]struct{}) // go idiom for set functionality

	vec := generateVector[I](numMultiples, multiple)

	shuffled := slices.Clone(vec)
	Shuffle(shuffled, pcgInput)

	for i, item := range shuffled {
		_, alreadyIn := uniqueMap[item]
		uniqueMap[item] = struct{}{}

		contained, err := a.contains(item)
		if err != nil { t.Fatal(err.Error()) }

		if contained != alreadyIn {
			t.Fatalf("want alreadyIn %t, got %t", contained, alreadyIn)
		}

		gotAdded, err := a.add(shuffled[i])
		if err != nil { t.Fatal(err.Error()) }

		if alreadyIn == gotAdded {
			t.Fatalf("alreadyIn = %t but gotAdded = %t", alreadyIn, gotAdded)
		}

		contained, err = a.contains(item)
		if err != nil { t.Fatal(err.Error()) }

		if contained == false {
			t.Fatal("contained should be true for previously added item")
		}
	}

	wantSize := len(uniqueMap)
	if uint64(wantSize) != a.size() {
		t.Errorf("want size %d, got %d", wantSize, a.size())
	}

	sortedUnique := getSortedUnique(uniqueMap)

	shuffledUnique := slices.Clone(sortedUnique)
	Shuffle(shuffledUnique, pcgInput)

	for _, item := range shuffledUnique {
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
