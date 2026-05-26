package roaring_test

import (
	// "fmt"
	// "slices"
	"testing"
	rr "roaring"
)

func roaringFromVec(vec []uint32, t *testing.T) *rr.Roaring {
	res := rr.MakeRoaring()
	for _, item := range vec {
		_, err := res.Add(item)
		if err != nil { t.Fatal(err.Error()) }
	}

	return res
}

func apply_op(f func(*rr.Roaring, *rr.Roaring) (*rr.Roaring, error), 
			  r1 *rr.Roaring, r2 *rr.Roaring, t *testing.T) *rr.Roaring {
	res, err := f(r1, r2)
	if err != nil { t.Fatal(err.Error()) }

	return res
}

func TestAddRemoveContainsMany(t *testing.T) {
	tests := []struct {
		name           string
		numMultiples   int
		multiple	   int
		extraItems	   []uint32
		wantContainers int
		pcgInput	   uint64
	}{
		{
			name:           "Empty",
			extraItems: 	[]uint32{},
			wantContainers: 0,
			pcgInput:		1,
		},
		{
			name:           "One",
			extraItems:     []uint32{0},
			wantContainers: 1,
			pcgInput:		2,
		},
		{
			name:           "Two",
			extraItems:		[]uint32{0, 0},
			wantContainers: 1,
			pcgInput:		67,
		},
		{
			name:           "Three",
			extraItems:		[]uint32{0, rr.MAX_CONTAINER_SIZE - 1},
			wantContainers: 1,
			pcgInput:		420,
		},
		{
			name:           "Four",
			extraItems:		[]uint32{0, rr.MAX_CONTAINER_SIZE },
			wantContainers: 2,
			pcgInput:		78,
		},
		{
			name:           "Five",
			extraItems:		[]uint32{
				0, 
				rr.MAX_CONTAINER_SIZE - 1, 
				4 * rr.MAX_CONTAINER_SIZE, 
				5 * rr.MAX_CONTAINER_SIZE,
				6 * rr.MAX_CONTAINER_SIZE - 1,
				9 * rr.MAX_CONTAINER_SIZE - 1,
				9 * rr.MAX_CONTAINER_SIZE,
			},
			wantContainers: 5,
			pcgInput:		19,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roaring := rr.MakeRoaring()
			rr.RoaringAddRemoveContainsTester(tt.numMultiples, tt.multiple, 
											 tt.extraItems, tt.pcgInput, roaring, t)

			if roaring.NumContainers() != 0 {
				t.Errorf("empty roaring should have no containers, got %d", roaring.Size())
			}
		})
	}
}

// func smallUnionIntersectHelper(vec1 []uint16, vec2 []uint16, expectedVec []uint16, t *testing.T,
// 	f func(*Container, *Container) (*Container, error)) {
// 	first := containerFromVec(vec1, t)
// 	second := containerFromVec(vec2, t)
// 	expected := containerFromVec(expectedVec, t)

// 	res1 := apply_op(f, first, second, t)
// 	res2 := apply_op(f, second, first, t)

// 	compareContainers(res1, expected, t)
// 	compareContainers(res2, expected, t)
// }

// func TestIntersectFew(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		vec1     []uint16
// 		vec2     []uint16
// 		expected []uint16
// 	}{
// 		{
// 			name:     "first",
// 			vec1:     []uint16{1, 2},
// 			vec2:     []uint16{2, 3},
// 			expected: []uint16{2},
// 		},
// 		{
// 			name:     "second",
// 			vec1:     []uint16{1, 2},
// 			vec2:     []uint16{3, 4},
// 			expected: []uint16{},
// 		},
// 		{
// 			name:     "third",
// 			vec1:     []uint16{2, 3},
// 			vec2:     []uint16{2, 3},
// 			expected: []uint16{2, 3},
// 		},
// 		{
// 			name:     "fourth",
// 			vec1:     []uint16{1, 2, 4, 6, 7, 9, 12, 15},
// 			vec2:     []uint16{4, 7, 8, 9, 15, 17},
// 			expected: []uint16{4, 7, 9, 15},
// 		},
// 		{
// 			name:     "fifth",
// 			vec1:     []uint16{1, 2, 4, 6, 7, 9, 12, 15},
// 			vec2:     []uint16{},
// 			expected: []uint16{},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			smallUnionIntersectHelper(tt.vec1, tt.vec2, tt.expected, t, (*Container).intersect)
// 		})
// 	}
// }

// func TestUnionFew(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		vec1     []uint16
// 		vec2     []uint16
// 		expected []uint16
// 	}{
// 		{
// 			name:     "first",
// 			vec1:     []uint16{1, 2},
// 			vec2:     []uint16{2, 3},
// 			expected: []uint16{1, 2, 3},
// 		},
// 		{
// 			name:     "second",
// 			vec1:     []uint16{1, 2},
// 			vec2:     []uint16{3, 4},
// 			expected: []uint16{1, 2, 3, 4},
// 		},
// 		{
// 			name:     "third",
// 			vec1:     []uint16{2, 3},
// 			vec2:     []uint16{2, 3},
// 			expected: []uint16{2, 3},
// 		},
// 		{
// 			name:     "fourth",
// 			vec1:     []uint16{1, 2, 4, 6, 7, 9, 12, 15},
// 			vec2:     []uint16{4, 7, 8, 9, 15, 17},
// 			expected: []uint16{1, 2, 4, 6, 7, 8, 9, 12, 15, 17},
// 		},
// 		{
// 			name:     "fifth",
// 			vec1:     []uint16{1, 2, 4, 6, 7, 9, 12, 15},
// 			vec2:     []uint16{},
// 			expected: []uint16{1, 2, 4, 6, 7, 9, 12, 15},
// 		},
// 		{
// 			name:     "sixth",
// 			vec1:     []uint16{},
// 			vec2:     []uint16{},
// 			expected: []uint16{},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			smallUnionIntersectHelper(tt.vec1, tt.vec2, tt.expected, t, (*Container).union)
// 		})
// 	}
// }

// func largeUnionIntersectHelper(tt struct {
// 	name             string
// 	vec1Numbers      int
// 	vec1Multiple     int
// 	vec1Offset 		 int
// 	vec2Numbers      int
// 	vec2Multiple     int
// 	vec2Offset 		 int
// }, t *testing.T, f func(*Container, *Container) (*Container, error),
// 	generateExpectedVec func([]uint16, []uint16) []uint16,
// ) {
// 	v1 := generateVectorWithOffset(tt.vec1Numbers, tt.vec1Multiple, tt.vec1Offset)
// 	v2 := generateVectorWithOffset(tt.vec2Numbers, tt.vec2Multiple, tt.vec2Offset)

// 	c1, c2 := containerFromVec(v1, t), containerFromVec(v2, t)
// 	res1 := apply_op(f, c1, c2, t)
// 	res2 := apply_op(f, c2, c1, t)

// 	expectedVec := generateExpectedVec(v1, v2)
// 	expected := containerFromVec(expectedVec, t)

// 	compareContainers(res1, expected, t)
// 	compareContainers(res2, expected, t)

// 	// ensure no mutation (i.e. expected is completely distinct)
// 	added := false
// 	var err error
// 	for i := uint16(1); !added; i++ {
// 		added, err = expected.add(i)
// 		if err != nil { t.Fatal(err.Error()) }
// 		if i == 0 { return } // overflowed
// 	}

// 	for _, res := range []*Container{res1, res2} {
// 		if expected.size != res.size + 1 {
// 			t.Error("expected size should be one greater than res size")
// 		}

// 		checkConcreteSize(res, t)
// 	}
// }

// func TestIntersectMany(t *testing.T) {
// 	tests := []struct {
// 		name             string
// 		vec1Numbers      int
// 		vec1Multiple     int
// 		vec1Offset 		 int
// 		vec1Extra 		 []uint32
// 		vec2Numbers      int
// 		vec2Multiple     int
// 		vec2Offset 		 int
// 		vec2Extra 		 []uint32
// 	}{
		
// 	}

// 	vectorwiseIntersect := func(vec1 []uint32, vec2 []uint32) []uint32 {
// 		elems := make(map[uint16]struct{})
// 		for _, item := range vec1 {
// 			elems[item] = struct{}{}
// 		}
// 		var res []uint16
// 		for _, item := range vec2 {
// 			_, ok := elems[item]
// 			if ok {
// 				res = append(res, item)
// 			}
// 		}
// 		slices.Sort(res)
// 		return res
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			largeUnionIntersectHelper(tt, t, (*Container).intersect, vectorwiseIntersect)
// 		})
// 	}
// }

// func TestUnionMany(t *testing.T) {
// 	tests := []struct {
// 		name             string
// 		vec1Numbers      int
// 		vec1Multiple     int
// 		vec1Offset 		 int
// 		vec2Numbers      int
// 		vec2Multiple     int
// 		vec2Offset 		 int
// 	}{
// 		{
// 			name:             "both_empty",
// 			vec1Numbers:      0,
// 			vec1Multiple:     3,
// 			vec2Numbers:      0,
// 			vec2Multiple:     2,
// 		},
// 		{
// 			name:             "one_empty",
// 			vec1Numbers:      PROMOTION_THRESHOLD,
// 			vec1Multiple:     7,
// 			vec2Numbers:      0,
// 			vec2Multiple:     2,
// 		},
// 		{
// 			name:             "first",
// 			vec1Numbers:      PROMOTION_THRESHOLD / 2,
// 			vec1Multiple:     2,
// 			vec1Offset: 	  0,
// 			vec2Numbers:      PROMOTION_THRESHOLD / 2,
// 			vec2Multiple:     2,
// 			vec2Offset: 	  1,
// 		},
// 		{
// 			name:             "second",
// 			vec1Numbers:      PROMOTION_THRESHOLD - 1,
// 			vec1Multiple:     3,
// 			vec2Numbers:      1,
// 			vec2Multiple:     0,
// 			vec2Offset: 	  5,
// 		},
// 		{
// 			name:             "third",
// 			vec1Numbers:      PROMOTION_THRESHOLD - 1,
// 			vec1Multiple:     3,
// 			vec2Numbers:      1,
// 			vec2Multiple:     0,
// 			vec2Offset: 	  6,
// 		},
// 		{
// 			name:             "fourth",
// 			vec1Numbers:      PROMOTION_THRESHOLD - 1,
// 			vec1Multiple:     7,
// 			vec1Offset:		  3,
// 			vec2Numbers:      PROMOTION_THRESHOLD - 1,
// 			vec2Multiple:     11,
// 			vec2Offset: 	  253,
// 		},
// 		{
// 			name:             "fifth",
// 			vec1Numbers:      PROMOTION_THRESHOLD,
// 			vec1Multiple:     3,
// 			vec1Offset:		  4,
// 			vec2Numbers:      PROMOTION_THRESHOLD,
// 			vec2Multiple:     3,
// 			vec2Offset:		  4,
// 		},
// 		{
// 			name:             "sixth",
// 			vec1Numbers:      PROMOTION_THRESHOLD,
// 			vec1Multiple:     3,
// 			vec1Offset:		  4,
// 			vec2Numbers:      PROMOTION_THRESHOLD - 1,
// 			vec2Multiple:     3,
// 			vec2Offset:		  4,
// 		},
// 		{
// 			name:             "seventh",
// 			vec1Numbers:      PROMOTION_THRESHOLD,
// 			vec1Multiple:     3,
// 			vec1Offset:		  4,
// 			vec2Numbers:      14,
// 			vec2Multiple:     3,
// 			vec2Offset:		  2,
// 		},
// 		{
// 			name:             "eighth",
// 			vec1Numbers:      PROMOTION_THRESHOLD * 2,
// 			vec1Multiple:     3,
// 			vec1Offset:		  2,
// 			vec2Numbers:      PROMOTION_THRESHOLD * 2,
// 			vec2Multiple:     3,
// 			vec2Offset:		  1,
// 		},
// 		{
// 			name:             "ninth",
// 			vec1Numbers:      PROMOTION_THRESHOLD * 2,
// 			vec1Multiple:     2,
// 			vec2Numbers:      PROMOTION_THRESHOLD * 2,
// 			vec2Multiple:     3,
// 		},
// 		{
// 			name:             "tenth",
// 			vec1Numbers:      PROMOTION_THRESHOLD - 1,
// 			vec1Multiple:     10,
// 			vec1Offset: 	  7,
// 			vec2Numbers:      PROMOTION_THRESHOLD - 1,
// 			vec2Multiple:     10,
// 			vec2Offset: 	  7,
// 		},
// 		{
// 			name:             "eleventh",
// 			vec1Numbers:      PROMOTION_THRESHOLD / 2 - 1,
// 			vec1Multiple:     5,
// 			vec1Offset: 	  1,
// 			vec2Numbers:      PROMOTION_THRESHOLD / 2 ,
// 			vec2Multiple:     5,
// 			vec2Offset: 	  3,
// 		},
// 	}

// 	vectorwiseUnion := func(vec1 []uint16, vec2 []uint16) []uint16 {
// 		elems := make(map[uint16]struct{})
// 		for _, item := range vec1 {
// 			elems[item] = struct{}{}
// 		}
// 		for _, item := range vec2 {
// 			elems[item] = struct{}{}
// 		}
// 		return getSortedUnique(elems)
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			largeUnionIntersectHelper(tt, t, (*Container).union, vectorwiseUnion)
// 		})
// 	}
// }
