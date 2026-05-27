package roaring

import (
	// "fmt"
	// "math/rand/v2"
	"slices"
	"testing"
)

func TestMakeContainer(t *testing.T) {
	container := makeContainer()
	if container.kind != VECTOR || container.bitmap != nil {
		t.Fatalf("new container is not a vector or has non-nil bitmap")
	}

	if container.size != 0 {
		t.Errorf("new container size is not zero")
	}

	if container.size != len(container.vector) {
		t.Errorf("new container size is not equal to its vector length")
	}
}
func TestAddFew(t *testing.T) {
	tests := []struct {
		name         string
		initialItems []uint16
		addItem      uint16
		wantAdded    bool
		wantKind     ContainerKind
	}{
		{
			name:         "Empty",
			initialItems: []uint16{},
			addItem:      0,
			wantAdded:    true,
			wantKind:     VECTOR,
		},
		{
			name:         "One item with update",
			initialItems: []uint16{5},
			addItem:      0,
			wantAdded:    true,
			wantKind:     VECTOR,
		},
		{
			name:         "One item no update",
			initialItems: []uint16{5},
			addItem:      5,
			wantAdded:    false,
			wantKind:     VECTOR,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := makeContainer()
			c.vector = tt.initialItems
			c.size = len(tt.initialItems)

			gotAdded, err := c.add(tt.addItem)
			if err != nil {
				t.Fatal(err.Error())
			}

			if gotAdded != tt.wantAdded {
				t.Errorf("Add() gotAdded = %t, want %t", gotAdded, tt.wantAdded)
			}

			wantNewSize := len(tt.initialItems)
			if gotAdded {
				wantNewSize++
			}

			if c.size != len(c.vector) {
				t.Errorf("vector container size is not equal to its vector length")
			}

			if c.size != wantNewSize {
				t.Errorf("c.size after add = %d, want %d", c.size, wantNewSize)
			}

			sizeBeforeSecondAdd := c.size

			// second add should have no effect
			gotAdded, err = c.add(tt.addItem)
			if err != nil {
				t.Fatal(err.Error())
			}

			if gotAdded != false {
				t.Errorf("Add() gotAdded = %t, want false since already added", gotAdded)
			}

			if c.size != len(c.vector) {
				t.Errorf("vector container size is not equal to its vector length")
			}

			if c.size != sizeBeforeSecondAdd {
				t.Errorf("c.size after add = %d, want %d", c.size, wantNewSize)
			}
		})
	}
}

func compareContainers(actual *Container, expected *Container, t *testing.T) {
	if expected.kind != actual.kind {
		t.Errorf("want kind %d, got %d", expected.kind, actual.kind)
	}

	switch expected.kind {
	case BITMAP:
		if actual.bitmap == nil {
			t.Fatal("bitmap container has nil bitmap")
		}

		oneBits := bitMapOneBits(actual.bitmap)

		if oneBits != expected.size {
			t.Errorf("want bitmap one bits = %d, got %d", expected.size, oneBits)
		}

		if *actual.bitmap != *expected.bitmap {
			t.Fatalf("actual bitmap and expected do not match")
		}

	case VECTOR:
		if len(actual.vector) != expected.size {
			t.Errorf("want vector length %d, got %d", expected.size, len(actual.vector))
		}

		if actual.size != expected.size {
			t.Errorf("want size %d, got %d", expected.size, actual.size)
		}

		if !slices.Equal(actual.vector, expected.vector) {
			t.Fatalf("actual vector and expected do not match")
		}
	default:
		panic("unrecognized kind")
	}
}

// Returns result, generated1, generated2
// where result if f(generated1, generated2)
func generateContainersAndApplyFunc(
	length1, multiple1, length2, multiple2 int,
	f func(*Container, *Container) (*Container, error),
	t *testing.T) (*Container, *Container, *Container) {
	return generateContainersAndApplyFuncWithOffsets(
		length1, multiple1, 0, length2, multiple2, 0, f, t)
}

func generateContainersAndApplyFuncWithOffsets(length1, multiple1,
	offset1, length2, multiple2, offset2 int,
	f func(*Container, *Container) (*Container, error),
	t *testing.T) (*Container, *Container, *Container) {

	c1 := generateContainerWithOffset(length1, multiple1, offset1, t)
	c2 := generateContainerWithOffset(length2, multiple2, offset2, t)

	res, err := f(c1, c2)
	if err != nil {
		t.Fatal(err.Error())
	}

	return res, c1, c2
}

func generateContainer(length, multiple int, t *testing.T) *Container {
	return generateContainerWithOffset(length, multiple, 0, t)
}

func generateContainerWithOffset(length, multiple, offset int, t *testing.T) *Container {
	vec := generateVectorWithOffset[uint16](length, multiple, offset)

	res := containerFromVec(vec, t)
	return res
}

func checkConcreteSize(c *Container, t *testing.T) {
	switch c.kind {
	case VECTOR:
		if len(c.vector) != c.size {
			t.Errorf("want container vector size %d, got %d", 
				c.size, len(c.vector))
		}

		if c.size > 0 && c.vector == nil {
			t.Fatal("nonempty container with vector kind must have non-nil vector")
		}
	case BITMAP:
		if c.bitmap == nil {
			t.Fatal("container with bitmap kind must have non-nil bitmap")
		}

		oneBits := bitMapOneBits(c.bitmap)
		if oneBits != c.size {
			t.Errorf("want container one bits %d, got %d", 
				c.size, oneBits)
		}
	default:
		panic("unknown kind")
	}
}

func apply_op(f func(*Container, *Container) (*Container, error), 
			  c1 *Container, c2 *Container, t *testing.T) *Container {
	res, err := f(c1, c2)
	if err != nil { t.Fatal(err.Error()) }

	return res
}

func TestAddRemoveContainsMany(t *testing.T) {
	tests := []struct {
		name         string
		numbersToAdd int
		multiple     int
		wantKind     ContainerKind
		pcgInput     uint64
	}{
		{
			name:         "Thousand",
			numbersToAdd: 1000,
			multiple:     3,
			wantKind:     VECTOR,
			pcgInput:     31,
		},
		{
			name:         "Before Switch Threshold",
			numbersToAdd: PROMOTION_THRESHOLD - 1,
			multiple:     1,
			wantKind:     VECTOR,
			pcgInput:     37,
		},
		{
			name:         "Switch Threshold",
			numbersToAdd: PROMOTION_THRESHOLD,
			multiple:     1,
			wantKind:     BITMAP,
			pcgInput:     50,
		},
		{
			name:         "Five Thousand",
			numbersToAdd: 5000,
			multiple:     3,
			wantKind:     BITMAP,
			pcgInput:     32,
		},
		{
			name:         "Max Unique",
			numbersToAdd: MAX_CONTAINER_SIZE,
			multiple:     1,
			wantKind:     BITMAP,
			pcgInput:     19,
		},
		{
			name:         "Repeats 1",
			numbersToAdd: MAX_CONTAINER_SIZE * 2,
			multiple:     1,
			wantKind:     BITMAP,
			pcgInput:     18,
		},
		{
			name:         "Repeats 2",
			numbersToAdd: MAX_CONTAINER_SIZE * 2,
			multiple:     2,
			wantKind:     BITMAP,
			pcgInput:     15,
		},
		{
			name:         "Repeats 3",
			numbersToAdd: MAX_CONTAINER_SIZE * 2,
			multiple:     2,
			wantKind:     BITMAP,
			pcgInput:     16,
		},
		{
			name:         "Repeats But Still Vector",
			numbersToAdd: 1 << 17,
			multiple:     1 << 8,
			wantKind:     VECTOR,
			pcgInput:     17,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cont := makeContainer()
			addRemoveContainsTester(tt.numbersToAdd, tt.multiple, []uint16{},
									tt.pcgInput, cont.toAdapter(), t)

			// container := makeContainer()

			// wantSize := len(uniqueMap)
			// if wantSize != container.size {
			// 	t.Errorf("want container size %d, got %d", wantSize, container.size)
			// }

			// if container.kind != tt.wantKind {
			// 	t.Errorf("want container kind %v, got %v", container.kind, tt.wantKind)
			// }

			// compareContainers(container, expected, t)

			// should now be empty
			if cont.kind != VECTOR {
				t.Errorf("empty container should have kind %d, got %d", VECTOR, cont.kind)
			}
			if len(cont.vector) != 0 {
				t.Errorf("empty container should have vector length 0, got %d", len(cont.vector))
			}
			if cont.bitmap != nil {
				t.Errorf("empty container should have nil bitmap")
			}
		})
	}
}

func smallUnionIntersectHelper(vec1 []uint16, vec2 []uint16, expectedVec []uint16, t *testing.T,
	f func(*Container, *Container) (*Container, error)) {
	first := containerFromVec(vec1, t)
	second := containerFromVec(vec2, t)
	expected := containerFromVec(expectedVec, t)

	res1 := apply_op(f, first, second, t)
	res2 := apply_op(f, second, first, t)

	compareContainers(res1, expected, t)
	compareContainers(res2, expected, t)
}

func TestIntersectFew(t *testing.T) {
	tests := []struct {
		name     string
		vec1     []uint16
		vec2     []uint16
		expected []uint16
	}{
		{
			name:     "first",
			vec1:     []uint16{1, 2},
			vec2:     []uint16{2, 3},
			expected: []uint16{2},
		},
		{
			name:     "second",
			vec1:     []uint16{1, 2},
			vec2:     []uint16{3, 4},
			expected: []uint16{},
		},
		{
			name:     "third",
			vec1:     []uint16{2, 3},
			vec2:     []uint16{2, 3},
			expected: []uint16{2, 3},
		},
		{
			name:     "fourth",
			vec1:     []uint16{1, 2, 4, 6, 7, 9, 12, 15},
			vec2:     []uint16{4, 7, 8, 9, 15, 17},
			expected: []uint16{4, 7, 9, 15},
		},
		{
			name:     "fifth",
			vec1:     []uint16{1, 2, 4, 6, 7, 9, 12, 15},
			vec2:     []uint16{},
			expected: []uint16{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smallUnionIntersectHelper(tt.vec1, tt.vec2, tt.expected, t, (*Container).intersect)
		})
	}
}

func TestUnionFew(t *testing.T) {
	tests := []struct {
		name     string
		vec1     []uint16
		vec2     []uint16
		expected []uint16
	}{
		{
			name:     "first",
			vec1:     []uint16{1, 2},
			vec2:     []uint16{2, 3},
			expected: []uint16{1, 2, 3},
		},
		{
			name:     "second",
			vec1:     []uint16{1, 2},
			vec2:     []uint16{3, 4},
			expected: []uint16{1, 2, 3, 4},
		},
		{
			name:     "third",
			vec1:     []uint16{2, 3},
			vec2:     []uint16{2, 3},
			expected: []uint16{2, 3},
		},
		{
			name:     "fourth",
			vec1:     []uint16{1, 2, 4, 6, 7, 9, 12, 15},
			vec2:     []uint16{4, 7, 8, 9, 15, 17},
			expected: []uint16{1, 2, 4, 6, 7, 8, 9, 12, 15, 17},
		},
		{
			name:     "fifth",
			vec1:     []uint16{1, 2, 4, 6, 7, 9, 12, 15},
			vec2:     []uint16{},
			expected: []uint16{1, 2, 4, 6, 7, 9, 12, 15},
		},
		{
			name:     "sixth",
			vec1:     []uint16{},
			vec2:     []uint16{},
			expected: []uint16{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			smallUnionIntersectHelper(tt.vec1, tt.vec2, tt.expected, t, (*Container).union)
		})
	}
}

func TestIntersectMany(t *testing.T) {
	tests := []struct {
		name             string
		vec1Numbers      int
		vec1Multiple     int
		vec1Offset 		 int
		vec1Extra 		 []uint16
		vec2Numbers      int
		vec2Multiple     int
		vec2Offset 		 int
		vec2Extra 		 []uint16
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
			vec1Numbers:      PROMOTION_THRESHOLD,
			vec1Multiple:     7,
			vec2Numbers:      0,
			vec2Multiple:     2,
		},
		{
			name:             "first",
			vec1Numbers:      PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec2Numbers:      PROMOTION_THRESHOLD,
			vec2Multiple:     2,
		},
		{
			name:             "second",
			vec1Numbers:      PROMOTION_THRESHOLD * 2,
			vec1Multiple:     3,
			vec2Numbers:      PROMOTION_THRESHOLD * 3,
			vec2Multiple:     2,
		},
		{
			name:             "third",
			vec1Numbers:      PROMOTION_THRESHOLD*2 - 3,
			vec1Multiple:     3,
			vec2Numbers:      PROMOTION_THRESHOLD*3 - 2,
			vec2Multiple:     2,
		},
		{
			name:             "fourth",
			vec1Numbers:      PROMOTION_THRESHOLD * 4,
			vec1Multiple:     1,
			vec2Numbers:      PROMOTION_THRESHOLD * 2,
			vec2Multiple:     2,
		},
		{
			name:             "fifth",
			vec1Numbers:      PROMOTION_THRESHOLD - 1,
			vec1Multiple:     3,
			vec2Numbers:      PROMOTION_THRESHOLD,
			vec2Multiple:     2,
		},
		{
			name:             "sixth",
			vec1Numbers:      PROMOTION_THRESHOLD / 2,
			vec1Multiple:     3,
			vec2Numbers:      PROMOTION_THRESHOLD * 2,
			vec2Multiple:     6,
		},
		{
			name:             "seventh",
			vec1Numbers:      PROMOTION_THRESHOLD / 2,
			vec1Multiple:     2,
			vec2Numbers:      PROMOTION_THRESHOLD / 2,
			vec2Multiple:     3,
		},
		{
			name:             "eighth",
			vec1Numbers:      PROMOTION_THRESHOLD / 2,
			vec1Multiple:     1,
			vec2Numbers:      PROMOTION_THRESHOLD / 2,
			vec2Multiple:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerLargeUnionIntersectionHelper(
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
		vec1Extra 		 []uint16
		vec2Numbers      int
		vec2Multiple     int
		vec2Offset 		 int
		vec2Extra		 []uint16
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
			vec1Numbers:      PROMOTION_THRESHOLD,
			vec1Multiple:     7,
			vec2Numbers:      0,
			vec2Multiple:     2,
		},
		{
			name:             "first",
			vec1Numbers:      PROMOTION_THRESHOLD / 2,
			vec1Multiple:     2,
			vec1Offset: 	  0,
			vec2Numbers:      PROMOTION_THRESHOLD / 2,
			vec2Multiple:     2,
			vec2Offset: 	  1,
		},
		{
			name:             "second",
			vec1Numbers:      PROMOTION_THRESHOLD - 1,
			vec1Multiple:     3,
			vec2Numbers:      1,
			vec2Multiple:     0,
			vec2Offset: 	  5,
		},
		{
			name:             "third",
			vec1Numbers:      PROMOTION_THRESHOLD - 1,
			vec1Multiple:     3,
			vec2Numbers:      1,
			vec2Multiple:     0,
			vec2Offset: 	  6,
		},
		{
			name:             "fourth",
			vec1Numbers:      PROMOTION_THRESHOLD - 1,
			vec1Multiple:     7,
			vec1Offset:		  3,
			vec2Numbers:      PROMOTION_THRESHOLD - 1,
			vec2Multiple:     11,
			vec2Offset: 	  253,
		},
		{
			name:             "fifth",
			vec1Numbers:      PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset:		  4,
			vec2Numbers:      PROMOTION_THRESHOLD,
			vec2Multiple:     3,
			vec2Offset:		  4,
		},
		{
			name:             "sixth",
			vec1Numbers:      PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset:		  4,
			vec2Numbers:      PROMOTION_THRESHOLD - 1,
			vec2Multiple:     3,
			vec2Offset:		  4,
		},
		{
			name:             "seventh",
			vec1Numbers:      PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec1Offset:		  4,
			vec2Numbers:      14,
			vec2Multiple:     3,
			vec2Offset:		  2,
		},
		{
			name:             "eighth",
			vec1Numbers:      PROMOTION_THRESHOLD * 2,
			vec1Multiple:     3,
			vec1Offset:		  2,
			vec2Numbers:      PROMOTION_THRESHOLD * 2,
			vec2Multiple:     3,
			vec2Offset:		  1,
		},
		{
			name:             "ninth",
			vec1Numbers:      PROMOTION_THRESHOLD * 2,
			vec1Multiple:     2,
			vec2Numbers:      PROMOTION_THRESHOLD * 2,
			vec2Multiple:     3,
		},
		{
			name:             "tenth",
			vec1Numbers:      PROMOTION_THRESHOLD - 1,
			vec1Multiple:     10,
			vec1Offset: 	  7,
			vec2Numbers:      PROMOTION_THRESHOLD - 1,
			vec2Multiple:     10,
			vec2Offset: 	  7,
		},
		{
			name:             "eleventh",
			vec1Numbers:      PROMOTION_THRESHOLD / 2 - 1,
			vec1Multiple:     5,
			vec1Offset: 	  1,
			vec2Numbers:      PROMOTION_THRESHOLD / 2 ,
			vec2Multiple:     5,
			vec2Offset: 	  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			containerLargeUnionIntersectionHelper(
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
