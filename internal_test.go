package roaring

import (
	"math/rand/v2"
	"testing"
	"math/bits"
	//"fmt"
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

func TestVectorAddMany(t *testing.T) {
	tests := []struct {
		name string 
		numbersToAdd int 
		multiple int
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func (t *testing.T) {
			vec := make([]uint16, tt.numbersToAdd)
			for i := range vec {
				vec[i] = uint16(i * tt.multiple)
			}
			r := rand.New(rand.NewPCG(tt.pcgInput, tt.pcgInput))
			r.Shuffle(tt.numbersToAdd, func(i, j int) { vec[i], vec[j] = vec[j], vec[i] })
			
			container := makeContainer()

			for i := range vec {
				gotAdded, err := container.Add(vec[i])
				if err != nil { t.Fatal(err.Error()) }
				if !gotAdded { t.Fatalf("failed to add %d (i = %d)", vec[i], i) }
			}

			if container.size != tt.numbersToAdd {
				t.Errorf("want container size %d, got %d", container.size, tt.numbersToAdd)
			}

			if container.kind != tt.wantKind {
				t.Errorf("want container kind %v, got %v", container.kind, tt.wantKind)
			}

			switch container.kind {
			case VECTOR:
				if tt.numbersToAdd != len(container.vector) {
					t.Errorf("want vector length = %d, got %d", tt.numbersToAdd, len(container.vector))
				}

				for i, got := range container.vector {
					want := uint16(i * tt.multiple)
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