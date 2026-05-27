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
	union(T) (adapter[T, I], error)
	intersect(T) (adapter[T, I], error)
	size() uint64
	getMember() T
	concreteSize() uint64
}

type roaringAdapter struct {
	r *Roaring
}

type AllowedInt interface {
	~uint16 | ~uint32
}

func (r *Roaring) toAdapter() *roaringAdapter { return new(roaringAdapter{r}) }
func (s roaringAdapter) add(item uint32) (bool, error) { return s.r.Add(item) }
func (s roaringAdapter) remove(item uint32) (bool, error) { return s.r.Remove(item) }
func (s roaringAdapter) contains(item uint32) (bool, error) { return s.r.Contains(item) }
func (s roaringAdapter) union(o *Roaring) (adapter[*Roaring, uint32], error) { 
	roar, err := s.r.Union(o)
	if err != nil { return roaringAdapter{}, err }
	return *roar.toAdapter(), err
}
func (s roaringAdapter) intersect(o *Roaring) (adapter[*Roaring, uint32], error) { 
	roar, err := s.r.Intersect(o)
	if err != nil { return roaringAdapter{}, err }
	return *roar.toAdapter(), err
}
func (s roaringAdapter) size() uint64 { return s.r.Size() }
func (s roaringAdapter) getMember() *Roaring { return s.r }
func (s roaringAdapter) concreteSize() uint64 { return s.r.concreteSize() }

type containerAdapter struct {
	c *Container
}

func (c *Container) toAdapter() *containerAdapter {
	return new(containerAdapter{c})
}
func (s containerAdapter) add(item uint16) (bool, error) { return s.c.add(item) }
func (s containerAdapter) remove(item uint16) (bool, error) { return s.c.remove(item) }
func (s containerAdapter) contains(item uint16) (bool, error) { return s.c.contains(item) }
func (s containerAdapter) union(o *Container) (adapter[*Container, uint16], error) { 
	cont, err := s.c.union(o)
	if err != nil { return containerAdapter{}, err }
	return *cont.toAdapter(), err
}
func (s containerAdapter) intersect(o *Container) (adapter[*Container, uint16], error) { 
	cont, err := s.c.intersect(o)
	if err != nil { return containerAdapter{}, err }
	return *cont.toAdapter(), err
}
func (s containerAdapter) size() uint64 { return uint64(s.c.size) }
func (s containerAdapter) getMember() *Container { return s.c }
func (s containerAdapter) concreteSize() uint64 { return s.c.concreteSize() }


func (r *Roaring) concreteSize() uint64 {
	res := uint64(0)
	for _, entry := range r.entries {
		res += entry.container.concreteSize()
	}
	return res
}

func (c *Container) concreteSize() uint64 {
	switch c.kind {
	case VECTOR:
		return uint64(len(c.vector))
	case BITMAP:
		return uint64(bitMapOneBits(c.bitmap))
	default:
		panic("unknown kind")
	}
}

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
		numMultiples   int,
		multiple	   int,
		extraItems	   []uint32,
		pcgInput	   uint64,
		r *Roaring, 
		t *testing.T,
	) {
	addRemoveContainsTester(numMultiples, multiple, extraItems, pcgInput, r.toAdapter(), t)
}

func addRemoveContainsTester[T any, I AllowedInt](
		numMultiples   int,
		multiple	   int,
		extraItems	   []I,
		pcgInput	   uint64,
		a adapter[T, I], 
		t *testing.T,
	) {
	uniqueMap := make(map[I]struct{}) // go idiom for set functionality

	vec := generateVector[I](numMultiples, multiple)
	vec = append(vec, extraItems...)

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

func vectorwiseIntersect[I AllowedInt](vec1 []I, vec2 []I) []I {
	elems := make(map[I]struct{})
	for _, item := range vec1 {
		elems[item] = struct{}{}
	}
	var res []I
	for _, item := range vec2 {
		_, ok := elems[item]
		if ok {
			res = append(res, item)
		}
	}
	slices.Sort(res)
	return res
}

func vectorwiseUnion[I AllowedInt](vec1 []I, vec2 []I) []I {
	elems := make(map[I]struct{})
	for _, item := range vec1 {
		elems[item] = struct{}{}
	}
	for _, item := range vec2 {
		elems[item] = struct{}{}
	}
	return getSortedUnique(elems)
}

func applyOp[T any, I AllowedInt](f func(adapter[T, I], T) (adapter[T, I], error), 
			  a1 adapter[T, I], a2 adapter[T, I], t *testing.T) adapter[T, I] {
	res, err := f(a1, a2.getMember())
	if err != nil { t.Fatal(err.Error()) }

	return res
}

func validateAdapter[T any, I AllowedInt](actual adapter[T, I], expected []I, t *testing.T) {
	if actual.size() != uint64(len(expected)) {
		t.Fatalf("want size %d, got %d", len(expected), actual.size())
	}

	concrete := actual.concreteSize()
	if concrete != uint64(len(expected)) {
		t.Fatalf("want concrete size %d, got %d", len(expected), concrete)
	}

	for _, item := range expected {
		contained, err :=  actual.contains(item) 
		if err != nil { t.Fatal(err.Error()) }
		if !contained {
			t.Fatalf("item %d not contained in actual", item)
		}
	}
}

func containerFromVec(vec []uint16, t *testing.T) *Container {
	res, err := containerFromVector(vec)
	if err != nil {
		t.Fatal(err.Error())
	}

	return res
}

func containerAdapterFromVec(vec []uint16, t *testing.T) adapter[*Container, uint16] {
	c := containerFromVec(vec, t)
	return c.toAdapter()
}

func roaringFromVec(vec []uint32, t *testing.T) *Roaring {
	roar := MakeRoaring()
	for _, item := range vec {
		_, err := roar.Add(item)
		if err != nil { t.Fatal(err.Error()) }
	}
	return roar
}

func roaringAdapterFromVec(vec []uint32, t *testing.T) adapter[*Roaring, uint32] {
	r := roaringFromVec(vec, t)
	return r.toAdapter()
}

func containerLargeUnionIntersectionHelper(
	isIntersect 	 bool,
	vec1Numbers      int,
	vec1Multiple     int,
	vec1Offset 		 int,
	vec1Extra 		 []uint16,
	vec2Numbers      int,
	vec2Multiple     int,
	vec2Offset 		 int,
	vec2Extra 		 []uint16,
	t *testing.T, 
) {
	var adapterFunc func(adapter[*Container, uint16], *Container) (adapter[*Container, uint16], error)
	var generateExpectedVec func([]uint16, []uint16) []uint16

	if isIntersect {
		adapterFunc = adapter[*Container, uint16].intersect
		generateExpectedVec = vectorwiseIntersect
	} else {
		adapterFunc = adapter[*Container, uint16].union
		generateExpectedVec = vectorwiseUnion
	}
	largeUnionIntersectHelper(
		vec1Numbers,
		vec1Multiple,
		vec1Offset,
		vec1Extra,
		vec2Numbers,
		vec2Multiple,
		vec2Offset,
		vec2Extra,
		t,
		adapterFunc,
		generateExpectedVec,
		containerAdapterFromVec,
	)
}

func RoaringLargeUnionIntersectionHelper(
	isIntersect 	 bool,
	vec1Numbers      int,
	vec1Multiple     int,
	vec1Offset 		 int,
	vec1Extra 		 []uint32,
	vec2Numbers      int,
	vec2Multiple     int,
	vec2Offset 		 int,
	vec2Extra 		 []uint32,
	t *testing.T, 
) {
	var adapterFunc func(adapter[*Roaring, uint32], *Roaring) (adapter[*Roaring, uint32], error)
	var generateExpectedVec func([]uint32, []uint32) []uint32

	if isIntersect {
		adapterFunc = adapter[*Roaring, uint32].intersect
		generateExpectedVec = vectorwiseIntersect
	} else {
		adapterFunc = adapter[*Roaring, uint32].union
		generateExpectedVec = vectorwiseUnion
	}

	largeUnionIntersectHelper(
		vec1Numbers,
		vec1Multiple,
		vec1Offset,
		vec1Extra,
		vec2Numbers,
		vec2Multiple,
		vec2Offset,
		vec2Extra,
		t,
		adapterFunc,
		generateExpectedVec,
		roaringAdapterFromVec,
	)
}

func largeUnionIntersectHelper[T any, I AllowedInt](
	vec1Numbers      int,
	vec1Multiple     int,
	vec1Offset 		 int,
	vec1Extra		 []I,
	vec2Numbers      int,
	vec2Multiple     int,
	vec2Offset 		 int,
	vec2Extra 		 []I,
	t *testing.T, 
	f func(adapter[T, I], T) (adapter[T, I], error),
	generateExpectedVec func([]I, []I) []I,
	adapterFromVec func([]I, *testing.T) adapter[T, I],
) {
	v1 := generateVectorWithOffset[I](vec1Numbers, vec1Multiple, vec1Offset)
	v1 = append(v1, vec1Extra...)
	slices.Sort(v1)

	v2 := generateVectorWithOffset[I](vec2Numbers, vec2Multiple, vec2Offset)
	v2 = append(v2, vec2Extra...)
	slices.Sort(v2)

	a1, a2 := adapterFromVec(v1, t), adapterFromVec(v2, t)

	res1 := applyOp(f, a1, a2, t)
	res2 := applyOp(f, a2, a1, t)

	expectedVec := generateExpectedVec(v1, v2)
	expected := adapterFromVec(expectedVec, t)

	validateAdapter(res1, expectedVec, t)
	validateAdapter(res2, expectedVec, t)

	// ensure no mutation (i.e. expected is completely distinct)
	added := false
	var err error
	for i := I(1); !added; i++ {
		added, err = expected.add(i)
		if err != nil { t.Fatal(err.Error()) }
		if i == 0 { return } // overflowed
	}

	validateAdapter(res1, expectedVec, t)
	validateAdapter(res2, expectedVec, t)
}
