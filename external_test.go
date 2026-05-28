package roaring_test

import (
	// "slices"
	// "fmt"
	rr "roaring"
	"testing"
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
			name:             "many_same",
			vec1Numbers:      1 << 21,
			vec1Multiple:     2,
			vec2Numbers:      1 << 20,
			vec2Multiple:     2,
		},
		{
			name:             "many_different",
			vec1Numbers:      1 << 20,
			vec1Multiple:     2,
			vec2Numbers:      1 << 19,
			vec2Multiple:     2,
			vec2Offset: 	  1,
		},
		{
			name:             "first",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE,
			vec2Numbers:      rr.PROMOTION_THRESHOLD,
			vec2Multiple:     2,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE,
		},
		{
			name:             "second",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE-1,
			vec2Numbers:      rr.PROMOTION_THRESHOLD,
			vec2Multiple:     2,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE,
		},
		{
			name:             "third",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE-1,
			vec2Numbers:      rr.PROMOTION_THRESHOLD,
			vec2Multiple:     2,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE-1,
		},
		{
			name:             "fourth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE-1,
			vec1Extra: 		  []uint32{
									rr.MAX_CONTAINER_SIZE,
									7 * rr.MAX_CONTAINER_SIZE,
									8 * rr.MAX_CONTAINER_SIZE - 1,
									9 * rr.MAX_CONTAINER_SIZE,
							  },
			vec2Numbers:      rr.PROMOTION_THRESHOLD,
			vec2Multiple:     2,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE-1,
			vec2Extra: 		  []uint32{
									8 * rr.MAX_CONTAINER_SIZE - 1,
							  },
		},
		{
			name:             "fifth",
			vec1Numbers:      rr.MAX_CONTAINER_SIZE,
			vec1Multiple:     4,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE,
			vec1Extra: 		  []uint32{
									9 * rr.MAX_CONTAINER_SIZE, 10 * rr.MAX_CONTAINER_SIZE - 1,
							  },
			vec2Numbers:      rr.MAX_CONTAINER_SIZE * 3 - 1,
			vec2Multiple:     2,
			vec2Offset:		  0,
			vec2Extra: 		  []uint32{
									9 * rr.MAX_CONTAINER_SIZE + 1,
							  },
		},
		{
			name:             "sixth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     4,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE - 4,
			vec1Extra: 		  []uint32{
									rr.MAX_CONTAINER_SIZE + 2,
							  },
			vec2Numbers:      rr.MAX_CONTAINER_SIZE * 4,
			vec2Multiple:     2,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE,
			vec2Extra: 		  []uint32{
								rr.MAX_CONTAINER_SIZE - 4,
								1,
							  },
		},
		{
			name:             "seventh",
			vec1Numbers:      0,
			vec1Multiple:     1,
			vec1Offset: 	  0,
			vec1Extra: 		  []uint32{
									0, rr.MAX_CONTAINER_SIZE - 1, rr.MAX_CONTAINER_SIZE * 2,
							  },
			vec2Numbers:      rr.MAX_CONTAINER_SIZE * 3,
			vec2Multiple:     1,
			vec2Offset:		  0,
			vec2Extra: 		  []uint32{
								rr.MAX_CONTAINER_SIZE * 3,
							  },
		},
		{
			name:             "eighth",
			vec1Numbers:      0,
			vec1Multiple:     1,
			vec1Offset: 	  0,
			vec1Extra: 		  []uint32{
									rr.MAX_CONTAINER_SIZE * 1024, 
									rr.MAX_CONTAINER_SIZE * 1025 - 1, 
									rr.MAX_CONTAINER_SIZE * 1026,
									rr.MAX_CONTAINER_SIZE * 5,
							  },
			vec2Numbers:      rr.MAX_CONTAINER_SIZE * 3 + 2,
			vec2Multiple:     1,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE * 1023,
			vec2Extra: 		  []uint32{
								rr.MAX_CONTAINER_SIZE * 1,
								0, 
								23,
								7,
								rr.MAX_CONTAINER_SIZE * 5,
								rr.MAX_CONTAINER_SIZE * 5 - 1,
							  },
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
			name:             "many_same",
			vec1Numbers:      1 << 21,
			vec1Multiple:     2,
			vec2Numbers:      1 << 20,
			vec2Multiple:     2,
		},
		{
			name:             "many_different",
			vec1Numbers:      1 << 20,
			vec1Multiple:     2,
			vec2Numbers:      1 << 19,
			vec2Multiple:     2,
			vec2Offset: 	  1,
		},
		{
			name:             "first",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE,
			vec2Numbers:      rr.PROMOTION_THRESHOLD,
			vec2Multiple:     2,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE,
		},
		{
			name:             "second",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE-1,
			vec2Numbers:      1,
			vec2Multiple:     1,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE,
		},
		{
			name:             "third",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE-1,
			vec2Numbers:      rr.PROMOTION_THRESHOLD,
			vec2Multiple:     2,
			vec2Offset:		  0,
		},
		{
			name:             "fourth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     rr.MAX_CONTAINER_SIZE,
			vec1Offset: 	  0,
			vec1Extra: 		  []uint32{
									9 * rr.MAX_CONTAINER_SIZE,
									rr.MAX_CONTAINER_SIZE,
									7 * rr.MAX_CONTAINER_SIZE,
									8 * rr.MAX_CONTAINER_SIZE - 1,
							  },
			vec2Numbers:      rr.PROMOTION_THRESHOLD,
			vec2Multiple:     rr.MAX_CONTAINER_SIZE,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE,
			vec2Extra: 		  []uint32{
									8 * rr.MAX_CONTAINER_SIZE - 1,
							  },
		},
		{
			name:             "fifth",
			vec1Numbers:      rr.MAX_CONTAINER_SIZE,
			vec1Multiple:     4,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE,
			vec1Extra: 		  []uint32{
									9 * rr.MAX_CONTAINER_SIZE, 10 * rr.MAX_CONTAINER_SIZE - 1,
							  },
			vec2Numbers:      rr.MAX_CONTAINER_SIZE * 3 - 1,
			vec2Multiple:     2,
			vec2Offset:		  0,
			vec2Extra: 		  []uint32{
									9 * rr.MAX_CONTAINER_SIZE + 1,
							  },
		},
		{
			name:             "sixth",
			vec1Numbers:      rr.PROMOTION_THRESHOLD,
			vec1Multiple:     4,
			vec1Offset: 	  rr.MAX_CONTAINER_SIZE - 4,
			vec1Extra: 		  []uint32{
									rr.MAX_CONTAINER_SIZE + 2,
							  },
			vec2Numbers:      rr.MAX_CONTAINER_SIZE * 4,
			vec2Multiple:     2,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE,
			vec2Extra: 		  []uint32{
								1,
							  },
		},
		{
			name:             "seventh",
			vec1Numbers:      0,
			vec1Multiple:     1,
			vec1Offset: 	  0,
			vec1Extra: 		  []uint32{
									0, rr.MAX_CONTAINER_SIZE - 1, rr.MAX_CONTAINER_SIZE * 2,
							  },
			vec2Numbers:      rr.MAX_CONTAINER_SIZE * 3,
			vec2Multiple:     1,
			vec2Offset:		  0,
			vec2Extra: 		  []uint32{
								rr.MAX_CONTAINER_SIZE * 3,
							  },
		},
		{
			name:             "eighth",
			vec1Numbers:      0,
			vec1Multiple:     1,
			vec1Offset: 	  0,
			vec1Extra: 		  []uint32{
									rr.MAX_CONTAINER_SIZE * 1024, 
									rr.MAX_CONTAINER_SIZE * 1025 - 1, 
									rr.MAX_CONTAINER_SIZE * 1026,
									rr.MAX_CONTAINER_SIZE * 5,
							  },
			vec2Numbers:      rr.MAX_CONTAINER_SIZE * 3 + 2,
			vec2Multiple:     1,
			vec2Offset:		  rr.MAX_CONTAINER_SIZE * 1023,
			vec2Extra: 		  []uint32{
								rr.MAX_CONTAINER_SIZE * 1,
								0, 
								23,
								7,
								rr.MAX_CONTAINER_SIZE * 5,
								rr.MAX_CONTAINER_SIZE * 5 - 1,
							  },
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
