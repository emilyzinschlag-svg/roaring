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
func TestIntersectMany(t *testing.T) {
	tests := []struct {
		name             string
		vec1Numbers      int
		vec1Multiple     int
		vec1Offset 		 int
		vec1Extra 		 []uint32
		vec2Numbers      int
		vec2Multiple     int
		vec2Offset 		 int
		vec2Extra 		 []uint32
	}{
		{
			name:             "both_empty",
			vec1Numbers:      0,
			vec1Multiple:     3,
			vec2Numbers:      0,
			vec2Multiple:     2,
		},
		{
			name:             "one_empty",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     7,
			vec2Numbers:      0,
			vec2Multiple:     2,
		},
		{
			name:             "first",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec2Numbers:      rr.PROMOTION_THRESHOLD,
			vec2Multiple:     2,
		},
		{
			name:             "second",
			vec1Numbers:      rr.PROMOTION_THRESHOLD * 2,
			vec1Multiple:     3,
			vec2Numbers:      rr.PROMOTION_THRESHOLD * 3,
			vec2Multiple:     2,
		},
		{
			name:             "third",
			vec1Numbers:      rr.PROMOTION_THRESHOLD*2 - 3,
			vec1Multiple:     3,
			vec2Numbers:      rr.PROMOTION_THRESHOLD*3 - 2,
			vec2Multiple:     2,
		},
		{
			name:             "fourth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD * 4,
			vec1Multiple:     1,
			vec2Numbers:      rr.PROMOTION_THRESHOLD * 2,
			vec2Multiple:     2,
		},
		{
			name:             "fifth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD - 1,
			vec1Multiple:     3,
			vec2Numbers:      rr.PROMOTION_THRESHOLD,
			vec2Multiple:     2,
		},
		{
			name:             "sixth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD / 2,
			vec1Multiple:     3,
			vec2Numbers:      rr.PROMOTION_THRESHOLD * 2,
			vec2Multiple:     6,
		},
		{
			name:             "seventh",
			vec1Numbers:      rr.PROMOTION_THRESHOLD / 2,
			vec1Multiple:     2,
			vec2Numbers:      rr.PROMOTION_THRESHOLD / 2,
			vec2Multiple:     3,
		},
		{
			name:             "eighth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD / 2,
			vec1Multiple:     1,
			vec2Numbers:      rr.PROMOTION_THRESHOLD / 2,
			vec2Multiple:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr.RoaringLargeUnionIntersectionHelper(
				true,
				tt.vec1Numbers,
				tt.vec1Multiple,
				tt.vec1Offset,
				tt.vec1Extra,
				tt.vec2Numbers,
				tt.vec2Multiple,
				tt.vec2Offset,
				tt.vec2Extra,
				t, 
			)
		})
	}
}

func TestUnionMany(t *testing.T) {
	tests := []struct {
		name             string
		vec1Numbers      int
		vec1Multiple     int
		vec1Offset 		 int
		vec1Extra 		 []uint32
		vec2Numbers      int
		vec2Multiple     int
		vec2Offset 		 int
		vec2Extra		 []uint32
	}{
		{
			name:             "both_empty",
			vec1Numbers:      0,
			vec1Multiple:     3,
			vec2Numbers:      0,
			vec2Multiple:     2,
		},
		{
			name:             "one_empty",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     7,
			vec2Numbers:      0,
			vec2Multiple:     2,
		},
		{
			name:             "first",
			vec1Numbers:      rr.PROMOTION_THRESHOLD / 2,
			vec1Multiple:     2,
			vec1Offset: 	  0,
			vec2Numbers:      rr.PROMOTION_THRESHOLD / 2,
			vec2Multiple:     2,
			vec2Offset: 	  1,
		},
		{
			name:             "second",
			vec1Numbers:      rr.PROMOTION_THRESHOLD - 1,
			vec1Multiple:     3,
			vec2Numbers:      1,
			vec2Multiple:     0,
			vec2Offset: 	  5,
		},
		{
			name:             "third",
			vec1Numbers:      rr.PROMOTION_THRESHOLD - 1,
			vec1Multiple:     3,
			vec2Numbers:      1,
			vec2Multiple:     0,
			vec2Offset: 	  6,
		},
		{
			name:             "fourth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD - 1,
			vec1Multiple:     7,
			vec1Offset:		  3,
			vec2Numbers:      rr.PROMOTION_THRESHOLD - 1,
			vec2Multiple:     11,
			vec2Offset: 	  253,
		},
		{
			name:             "fifth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset:		  4,
			vec2Numbers:      rr.PROMOTION_THRESHOLD,
			vec2Multiple:     3,
			vec2Offset:		  4,
		},
		{
			name:             "sixth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset:		  4,
			vec2Numbers:      rr.PROMOTION_THRESHOLD - 1,
			vec2Multiple:     3,
			vec2Offset:		  4,
		},
		{
			name:             "seventh",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset:		  4,
			vec2Numbers:      14,
			vec2Multiple:     3,
			vec2Offset:		  2,
		},
		{
			name:             "eighth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD * 2,
			vec1Multiple:     3,
			vec1Offset:		  2,
			vec2Numbers:      rr.PROMOTION_THRESHOLD * 2,
			vec2Multiple:     3,
			vec2Offset:		  1,
		},
		{
			name:             "ninth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD * 2,
			vec1Multiple:     2,
			vec2Numbers:      rr.PROMOTION_THRESHOLD * 2,
			vec2Multiple:     3,
		},
		{
			name:             "tenth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD - 1,
			vec1Multiple:     10,
			vec1Offset: 	  7,
			vec2Numbers:      rr.PROMOTION_THRESHOLD - 1,
			vec2Multiple:     10,
			vec2Offset: 	  7,
		},
		{
			name:             "eleventh",
			vec1Numbers:      rr.PROMOTION_THRESHOLD / 2 - 1,
			vec1Multiple:     5,
			vec1Offset: 	  1,
			vec2Numbers:      rr.PROMOTION_THRESHOLD / 2 ,
			vec2Multiple:     5,
			vec2Offset: 	  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr.RoaringLargeUnionIntersectionHelper(
				false,
				tt.vec1Numbers,
				tt.vec1Multiple,
				tt.vec1Offset,
				tt.vec1Extra,
				tt.vec2Numbers,
				tt.vec2Multiple,
				tt.vec2Offset,
				tt.vec2Extra,
				t, 
			)
		})
	}
}
