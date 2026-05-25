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

func TestAddRemoveContainsMany(t *testing.T) {
	tests := []struct {
		name         string
		numbersToAdd int
		multiple     uint16
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
			vec := make([]uint16, tt.numbersToAdd)

			for i := range vec {
				item := uint16(i) * tt.multiple
				vec[i] = item
			}

			shuffled := slices.Clone(vec)
			r := rand.New(rand.NewPCG(tt.pcgInput, tt.pcgInput))
			r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })

			container := makeContainer()

			for i, item := range shuffled {
				_, alreadyIn := uniqueMap[item]
				uniqueMap[item] = struct{}{}

				gotAdded, err := container.add(shuffled[i])
				if err != nil { t.Fatal(err.Error()) }

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

			expected, err := containerFromVector(sortedUnique)
			if err != nil { t.Fatal(err.Error()) }

			shuffledUnique := slices.Clone(sortedUnique)
			r.Shuffle(len(shuffledUnique), func(i, j int) { shuffledUnique[i], shuffledUnique[j] = shuffledUnique[j], shuffledUnique[i] })

			compareContainers(container, expected, t)

			for _, item := range shuffledUnique {
				sizeBefore := container.size 

				gotAdded, err := container.add(item) 
				if err != nil { t.Fatal(err.Error()) }
				if gotAdded { t.Fatalf("add should return false for item already in: %d", item) }
				if sizeBefore != container.size { t.Fatalf("add should not affect size for item already in: %d", item) }

				contained, err := container.contains(item)
				if err != nil { t.Fatal(err.Error()) }
				if !contained { t.Fatalf("contains should return true for item already in: %d", item) }
				if sizeBefore != container.size { t.Fatalf("contains should not affect size. item: %d", item) }

				removed, err := container.remove(item)
				if err != nil { t.Fatal(err.Error()) }
				if !removed { t.Fatalf("removed should return true for item already in: %d", item) }
				if sizeBefore-1 != container.size { t.Fatalf("want remove to decrement container size: %d", item) }
				if container.size <= DEMOTION_THRESHOLD && container.kind != VECTOR { 
					t.Errorf("containers with size below demotion threshold should be vectors") 
				}

				sizeBefore = container.size 
				contained, err = container.contains(item)
				if err != nil { t.Fatal(err.Error()) }
				if contained { t.Fatalf("contains should return false for removed item: %d", item) }
				if sizeBefore != container.size { t.Fatalf("contains should not affect size. item: %d", item) }
			}

			// should now be empty
			if VECTOR != container.kind { t.Errorf("empty container should have kind %d, got %d", VECTOR, container.kind)}
			if 0 != container.size { t.Errorf("empty container should have size 0, got %d", container.size) }
			if 0 != len(container.vector) { t.Errorf("empty container should have vector length 0, got %d", len(container.vector)) }
			if nil != container.bitmap { t.Errorf("empty container should have nil bitmap") }
		})
	}
}

func smallUnionIntersectHelper(vec1 []uint16, vec2 []uint16, expected []uint16, t *testing.T, 
							  f func (*Container, *Container) (*Container, error)) {
	first, err := containerFromVector(vec1)
	if err != nil { t.Fatal(err.Error()) }

	second, err := containerFromVector(vec2)
	if err != nil { t.Fatal(err.Error()) }

	res1, err := f(first, second)
	if err != nil { t.Fatal(err.Error()) }

	res2, err := f(second, first)
	if err != nil { t.Fatal(err.Error()) }

	for _, res := range []*Container{res1, res2} {
		if res.kind != VECTOR {
			t.Errorf("intersection of vectors is not a vector")
		}

		if len(res.vector) != len(expected) {
			t.Errorf("want vector length %d, got %d", len(expected), len(res.vector))
		}

		if res.size != len(expected) {
			t.Errorf("want size %d, got %d", len(expected), res.size)
		}

		if !slices.Equal(res.vector, expected) {
			t.Fatalf("want result %v, got %v", expected, res.vector)
		}
	}
}

func TestIntersectFew(t *testing.T) {
	tests := []struct {
		name string
		vec1 []uint16 
		vec2 []uint16
		expected []uint16
	}{
		{
			name: "first",
			vec1: []uint16{1, 2},
			vec2: []uint16{2, 3},
			expected: []uint16{2},
		},
		{
			name: "second",
			vec1: []uint16{1, 2},
			vec2: []uint16{3, 4},
			expected: []uint16{},
		},
		{
			name: "third",
			vec1: []uint16{2, 3},
			vec2: []uint16{2, 3},
			expected: []uint16{2, 3},
		},
		{
			name: "fourth",
			vec1: []uint16{1, 2, 4, 6, 7, 9, 12, 15},
			vec2: []uint16{4, 7, 8, 9, 15, 17},
			expected: []uint16{4, 7, 9, 15},
		},
		{
			name: "fifth",
			vec1: []uint16{1, 2, 4, 6, 7, 9, 12, 15},
			vec2: []uint16{},
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
		name string
		vec1 []uint16 
		vec2 []uint16
		expected []uint16
	}{
		{
			name: "first",
			vec1: []uint16{1, 2},
			vec2: []uint16{2, 3},
			expected: []uint16{1, 2, 3},
		},
		{
			name: "second",
			vec1: []uint16{1, 2},
			vec2: []uint16{3, 4},
			expected: []uint16{1, 2, 3, 4},
		},
		{
			name: "third",
			vec1: []uint16{2, 3},
			vec2: []uint16{2, 3},
			expected: []uint16{2, 3},
		},
		{
			name: "fourth",
			vec1: []uint16{1, 2, 4, 6, 7, 9, 12, 15},
			vec2: []uint16{4, 7, 8, 9, 15, 17},
			expected: []uint16{1, 2, 4, 6, 7, 8, 9, 12, 15, 17},
		},
		{
			name: "fifth",
			vec1: []uint16{1, 2, 4, 6, 7, 9, 12, 15},
			vec2: []uint16{},
			expected: []uint16{1, 2, 4, 6, 7, 9, 12, 15},
		},
		{
			name: "sixth",
			vec1: []uint16{},
			vec2: []uint16{},
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
		name string
		vec1Numbers int 
		vec1Multiple int
		vec2Numbers int
		vec2Multiple int
		expectedNumbers int 
		expectedMultiple int
	}, t *testing.T, f func(*Container, *Container) (*Container, error)) {
	
	vec1 := make([]uint16, tt.vec1Numbers)
	vec2 := make([]uint16, tt.vec2Numbers)

	for i := 0; i < tt.vec1Numbers; i++ {
		vec1[i] = uint16(i * tt.vec1Multiple)
	}

	for i := 0; i < tt.vec2Numbers; i++ {
		vec2[i] = uint16(i * tt.vec2Multiple)
	}

	c1, err := containerFromVector(vec1)
	if err != nil { t.Fatal(err.Error()) }

	c2, err := containerFromVector(vec2)
	if err != nil { t.Fatal(err.Error()) }

	res1, err := f(c1, c2)
	if err != nil { t.Fatal(err.Error()) }

	res2, err := f(c2, c1)
	if err != nil { t.Fatal(err.Error()) }

	expectedVec := make([]uint16, tt.expectedNumbers)

	for i := 0; i < tt.expectedNumbers; i++ {
		expectedVec[i] = uint16(i * tt.expectedMultiple)
	}

	expected, err := containerFromVector(expectedVec)
	if err != nil { t.Fatal(err.Error()) }

	for _, res := range []*Container{res1, res2} {
		if res.kind != expected.kind {
			t.Errorf("want kind %d, got %d", expected.kind, res.kind)
		}

		compareContainers(res, expected, t)
	}
}

func TestIntersectMany(t *testing.T) {
	tests := []struct {
		name string
		vec1Numbers int 
		vec1Multiple int
		vec2Numbers int
		vec2Multiple int
		expectedNumbers int 
		expectedMultiple int
	}{
		{
			name: "first",
			vec1Numbers: PROMOTION_THRESHOLD,
			vec1Multiple: 3,
			vec2Numbers: PROMOTION_THRESHOLD,
			vec2Multiple: 2,
			expectedNumbers: 1 + PROMOTION_THRESHOLD / 3,
			expectedMultiple: 6,
		},
		{
			name: "second",
			vec1Numbers: PROMOTION_THRESHOLD * 2,
			vec1Multiple: 3,
			vec2Numbers: PROMOTION_THRESHOLD * 3,
			vec2Multiple: 2,
			expectedNumbers: PROMOTION_THRESHOLD,
			expectedMultiple: 6,
		},
		{
			name: "third",
			vec1Numbers: PROMOTION_THRESHOLD * 2 - 3,
			vec1Multiple: 3,
			vec2Numbers: PROMOTION_THRESHOLD * 3 - 2,
			vec2Multiple: 2,
			expectedNumbers: PROMOTION_THRESHOLD - 1,
			expectedMultiple: 6,
		},
		{
			name: "fourth",
			vec1Numbers: PROMOTION_THRESHOLD * 4,
			vec1Multiple: 1,
			vec2Numbers: PROMOTION_THRESHOLD * 2,
			vec2Multiple: 2,
			expectedNumbers: PROMOTION_THRESHOLD * 2,
			expectedMultiple: 2,
		},
		{
			name: "fifth",
			vec1Numbers: PROMOTION_THRESHOLD - 1,
			vec1Multiple: 3,
			vec2Numbers: PROMOTION_THRESHOLD,
			vec2Multiple: 2,
			expectedNumbers: 1 + PROMOTION_THRESHOLD / 3,
			expectedMultiple: 6,
		},
		{
			name: "sixth",
			vec1Numbers: PROMOTION_THRESHOLD / 2,
			vec1Multiple: 3,
			vec2Numbers: PROMOTION_THRESHOLD * 2,
			vec2Multiple: 6,
			expectedNumbers: PROMOTION_THRESHOLD / 4,
			expectedMultiple: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			largeUnionIntersectHelper(tt, t, (*Container).intersect)
		})
	}
}
