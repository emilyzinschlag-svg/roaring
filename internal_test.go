package roaring

import (
	"math/bits"
	"math/rand/v2"
	"testing"
	"slices"
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
func TestVectorAddFew(t *testing.T) {
	tests := []struct {
		name string
		initialItems []uint16 
		addItem uint16 
		wantAdded bool 
		wantKind ContainerKind
	} {
		{
			name: "Empty",
			initialItems: []uint16{},
			addItem: 0,
			wantAdded: true,
			wantKind: VECTOR,
		},
		{
			name: "One item with update",
			initialItems: []uint16{5},
			addItem: 0,
			wantAdded: true,
			wantKind: VECTOR,
		},
		{
			name: "One item no update",
			initialItems: []uint16{5},
			addItem: 5,
			wantAdded: false,
			wantKind: VECTOR,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := makeContainer()
			c.vector = tt.initialItems
			c.size = len(tt.initialItems)

			gotAdded, err := c.Add(tt.addItem)
			if err != nil { t.Fatal(err.Error()) }

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
			gotAdded, err = c.Add(tt.addItem)
			if err != nil { t.Fatal(err.Error()) }

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

func TestAddMany(t *testing.T) {
	tests := []struct {
		name string 
		numbersToAdd int 
		multiple uint16
		wantKind ContainerKind
		pcgInput uint64
	}{
		{
			name: "Thousand",
			numbersToAdd: 1000,
			multiple: 3,
			wantKind: VECTOR,
			pcgInput: 31,
		},
		{
			name: "Before Switch Threshold",
			numbersToAdd: ROARING_THRESHOLD - 1,
			multiple: 1,
			wantKind: VECTOR,
			pcgInput: 37,
		},
		{
			name: "Switch Threshold",
			numbersToAdd: ROARING_THRESHOLD,
			multiple: 1,
			wantKind: BITMAP,
			pcgInput: 50,
		},
		{
			name: "Five Thousand",
			numbersToAdd: 5000,
			multiple: 3,
			wantKind: BITMAP,
			pcgInput: 32,
		},
		{
			name: "Max Unique",
			numbersToAdd: 1 << 16,
			multiple: 1,
			wantKind: BITMAP,
			pcgInput: 15,
		},
		{
			name: "Repeats 1",
			numbersToAdd: 1 << 17,
			multiple: 1,
			wantKind: BITMAP,
			pcgInput: 15,
		},
		{
			name: "Repeats 2",
			numbersToAdd: 1 << 17,
			multiple: 2,
			wantKind: BITMAP,
			pcgInput: 15,
		},
		{
			name: "Repeats 3",
			numbersToAdd: 1 << 17,
			multiple: 2,
			wantKind: BITMAP,
			pcgInput: 15,
		},
		{
			name: "Repeats But Still Vector",
			numbersToAdd: 1 << 17,
			multiple: 1 << 8,
			wantKind: VECTOR,
			pcgInput: 15,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func (t *testing.T) {
			wantSize := -1
			vec := make([]uint16, tt.numbersToAdd)

			for i := range vec {
				if wantSize == -1 && int(uint16(i) * tt.multiple) < i * int(tt.multiple) &&
				   		(uint16(i) * tt.multiple) % tt.multiple == 0 {
					wantSize = i // first repeat
				} 
				vec[i] = uint16(i) * tt.multiple
			}

			if wantSize == -1 {
				wantSize = tt.numbersToAdd
			}

			shuffled := slices.Clone(vec)

			slices.Sort(vec[:wantSize]) // overflow may mean it wasn't already sorted

			r := rand.New(rand.NewPCG(tt.pcgInput, tt.pcgInput))
			r.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
			
			container := makeContainer()

			for i := range shuffled {
				_, err := container.Add(shuffled[i])
				if err != nil { t.Fatal(err.Error()) }
			}

			if wantSize != container.size {
				t.Errorf("want container size %d, got %d", wantSize, container.size)
			}

			if container.kind != tt.wantKind {
				t.Errorf("want container kind %v, got %v", container.kind, tt.wantKind)
			}

			switch container.kind {
			case VECTOR:
				if wantSize != len(container.vector) {
					t.Errorf("want vector length = %d, got %d", wantSize, len(container.vector))
				}

				for i, got := range container.vector {
					want := vec[i]
					if got != want {
						t.Fatalf("want container vector index %d to be %d, got %d", 
								i, want, got)
					}
				}
			case BITMAP:
				if container.bitmap == nil {
					t.Fatal("bitmap container has nil bitmap")
				}

				oneBits := 0
				for _, word := range container.bitmap {
					oneBits += bits.OnesCount64(uint64(word))
				}

				if container.size != oneBits {
					t.Errorf("want bitmap one bits = %d, got %d", tt.numbersToAdd, oneBits)
				}
			default:
				t.Fatalf("unrecognized kind: %v", container.kind)
			}
		})
	}
}
