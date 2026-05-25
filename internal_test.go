package roaring

import (
	// "fmt"
	"math/rand/v2"
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
	vec := generateVectorWithOffset(length, multiple, offset)

	res := containerFromVec(vec, t)
	return res
}

func generateVector(length, multiple int) []uint16 {
	return generateVectorWithOffset(length, multiple, 0)
}

func generateVectorWithOffset(length, multiple, offset int) []uint16 {
	res := make([]uint16, length)
	for i := range res {
		res[i] = uint16((i * multiple) + offset)
	}
	return res
}

func containerFromVec(vec []uint16, t *testing.T) *Container {
	res, err := containerFromVector(vec)
	if err != nil {
		t.Fatal(err.Error())
	}

	return res
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
			uniqueMap := make(map[uint16]struct{}) // go idiom for set functionality

			vec := generateVector(tt.numbersToAdd, tt.multiple)

			shuffled := slices.Clone(vec)
			r := rand.New(rand.NewPCG(tt.pcgInput, tt.pcgInput))
			r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

			container := makeContainer()

			for i, item := range shuffled {
				_, alreadyIn := uniqueMap[item]
				uniqueMap[item] = struct{}{}

				gotAdded, err := container.add(shuffled[i])
				if err != nil {
					t.Fatal(err.Error())
				}

				if alreadyIn == gotAdded {
					t.Fatalf("alreadyIn = %t but gotAdded = %t", alreadyIn, gotAdded)
				}
			}

			wantSize := len(uniqueMap)
			if wantSize != container.size {
				t.Errorf("want container size %d, got %d", wantSize, container.size)
			}

			if container.kind != tt.wantKind {
				t.Errorf("want container kind %v, got %v", container.kind, tt.wantKind)
			}

			sortedUnique := make([]uint16, 0, wantSize)
			for k := range uniqueMap {
				sortedUnique = append(sortedUnique, k)
			}
			slices.Sort(sortedUnique)

			expected := containerFromVec(sortedUnique, t)

			shuffledUnique := slices.Clone(sortedUnique)
			r.Shuffle(len(shuffledUnique), func(i, j int) { shuffledUnique[i], shuffledUnique[j] = shuffledUnique[j], shuffledUnique[i] })

			compareContainers(container, expected, t)

			for _, item := range shuffledUnique {
				sizeBefore := container.size

				gotAdded, err := container.add(item)
				if err != nil {
					t.Fatal(err.Error())
				}
				if gotAdded {
					t.Fatalf("add should return false for item already in: %d", item)
				}
				if sizeBefore != container.size {
					t.Fatalf("add should not affect size for item already in: %d", item)
				}

				contained, err := container.contains(item)
				if err != nil {
					t.Fatal(err.Error())
				}
				if !contained {
					t.Fatalf("contains should return true for item already in: %d", item)
				}
				if sizeBefore != container.size {
					t.Fatalf("contains should not affect size. item: %d", item)
				}

				removed, err := container.remove(item)
				if err != nil {
					t.Fatal(err.Error())
				}
				if !removed {
					t.Fatalf("removed should return true for item already in: %d", item)
				}
				if sizeBefore-1 != container.size {
					t.Fatalf("want remove to decrement container size: %d", item)
				}
				if container.size <= DEMOTION_THRESHOLD && container.kind != VECTOR {
					t.Errorf("containers with size below demotion threshold should be vectors")
				}

				sizeBefore = container.size
				contained, err = container.contains(item)
				if err != nil {
					t.Fatal(err.Error())
				}
				if contained {
					t.Fatalf("contains should return false for removed item: %d", item)
				}
				if sizeBefore != container.size {
					t.Fatalf("contains should not affect size. item: %d", item)
				}
			}

			// should now be empty
			if VECTOR != container.kind {
				t.Errorf("empty container should have kind %d, got %d", VECTOR, container.kind)
			}
			if 0 != container.size {
				t.Errorf("empty container should have size 0, got %d", container.size)
			}
			if 0 != len(container.vector) {
				t.Errorf("empty container should have vector length 0, got %d", len(container.vector))
			}
			if nil != container.bitmap {
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

func largeUnionIntersectHelper(tt struct {
	name             string
	vec1Numbers      int
	vec1Multiple     int
	vec2Numbers      int
	vec2Multiple     int
	expectedNumbers  int
	expectedMultiple int
}, t *testing.T, f func(*Container, *Container) (*Container, error)) {

	res1, c1, c2 := generateContainersAndApplyFunc(
		tt.vec1Numbers, tt.vec1Multiple,
		tt.vec2Numbers, tt.vec2Multiple,
		f, t,
	)
	res2 := apply_op(f, c2, c1, t)

	expected := generateContainer(tt.expectedNumbers, tt.expectedMultiple, t)

	compareContainers(res1, expected, t)
	compareContainers(res2, expected, t)
}

func TestIntersectMany(t *testing.T) {
	tests := []struct {
		name             string
		vec1Numbers      int
		vec1Multiple     int
		vec2Numbers      int
		vec2Multiple     int
		expectedNumbers  int
		expectedMultiple int
	}{
		{
			name:             "first",
			vec1Numbers:      PROMOTION_THRESHOLD,
			vec1Multiple:     3,
			vec2Numbers:      PROMOTION_THRESHOLD,
			vec2Multiple:     2,
			expectedNumbers:  1 + PROMOTION_THRESHOLD/3,
			expectedMultiple: 6,
		},
		{
			name:             "second",
			vec1Numbers:      PROMOTION_THRESHOLD * 2,
			vec1Multiple:     3,
			vec2Numbers:      PROMOTION_THRESHOLD * 3,
			vec2Multiple:     2,
			expectedNumbers:  PROMOTION_THRESHOLD,
			expectedMultiple: 6,
		},
		{
			name:             "third",
			vec1Numbers:      PROMOTION_THRESHOLD*2 - 3,
			vec1Multiple:     3,
			vec2Numbers:      PROMOTION_THRESHOLD*3 - 2,
			vec2Multiple:     2,
			expectedNumbers:  PROMOTION_THRESHOLD - 1,
			expectedMultiple: 6,
		},
		{
			name:             "fourth",
			vec1Numbers:      PROMOTION_THRESHOLD * 4,
			vec1Multiple:     1,
			vec2Numbers:      PROMOTION_THRESHOLD * 2,
			vec2Multiple:     2,
			expectedNumbers:  PROMOTION_THRESHOLD * 2,
			expectedMultiple: 2,
		},
		{
			name:             "fifth",
			vec1Numbers:      PROMOTION_THRESHOLD - 1,
			vec1Multiple:     3,
			vec2Numbers:      PROMOTION_THRESHOLD,
			vec2Multiple:     2,
			expectedNumbers:  1 + PROMOTION_THRESHOLD/3,
			expectedMultiple: 6,
		},
		{
			name:             "sixth",
			vec1Numbers:      PROMOTION_THRESHOLD / 2,
			vec1Multiple:     3,
			vec2Numbers:      PROMOTION_THRESHOLD * 2,
			vec2Multiple:     6,
			expectedNumbers:  PROMOTION_THRESHOLD / 4,
			expectedMultiple: 6,
		},
		{
			name:             "seventh",
			vec1Numbers:      PROMOTION_THRESHOLD / 2,
			vec1Multiple:     2,
			vec2Numbers:      PROMOTION_THRESHOLD / 2,
			vec2Multiple:     3,
			expectedNumbers:  1 + PROMOTION_THRESHOLD/6,
			expectedMultiple: 6,
		},
		{
			name:             "eighth",
			vec1Numbers:      PROMOTION_THRESHOLD / 2,
			vec1Multiple:     1,
			vec2Numbers:      PROMOTION_THRESHOLD / 2,
			vec2Multiple:     1,
			expectedNumbers:  PROMOTION_THRESHOLD / 2,
			expectedMultiple: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			largeUnionIntersectHelper(tt, t, (*Container).intersect)
		})
	}
}
