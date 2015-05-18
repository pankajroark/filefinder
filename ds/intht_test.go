package ds

import (
	"testing"
)

func TestIhtPutGet(t *testing.T) {
	iht := CreateIntHashTable(10)
	numItems := uint32(10)
	for i := uint32(0); i < numItems; i++ {
		iht.Put(i, 2*i)
	}
	for i := uint32(0); i < numItems; i++ {
		v, found := iht.Get(i)
		if !found {
			t.Error("value not found")
		}
		if v != 2*i {
			t.Errorf("expected %d but got %d", 2*i, v)
		}
	}
}

func TestIhtGrow(t *testing.T) {
	iht := CreateIntHashTable(1)
	numItems := uint32(1000)
	for i := uint32(0); i < numItems; i++ {
		iht.Put(i, 2*i)
	}
	for i := uint32(0); i < numItems; i++ {
		v, found := iht.Get(i)
		if !found {
			t.Error("value not found")
		}
		if v != 2*i {
			t.Errorf("expected %d but got %d", 2*i, v)
		}
	}
}

func TestIhtOvewrite(t *testing.T) {
	iht := CreateIntHashTable(1)
	iht.Put(1, 2)
	iht.Put(1, 3)
	v, found := iht.Get(1)
	if !found {
		t.Error("key not found")
	}

	if v != 3 {
		t.Errorf("expected %d but got %d", 3, v)
	}
}
