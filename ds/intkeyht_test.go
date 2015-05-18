package ds

import (
	"fmt"
	"testing"
)

func TestIKhtPutGet(t *testing.T) {
	iht := CreateIntKeyHashTable()
	numItems := uint32(1)
	for i := uint32(0); i < numItems; i++ {
		iht.Put(i, []byte(fmt.Sprintf("some %d", i)))
	}
	for i := uint32(0); i < numItems; i++ {
		v, found := iht.Get(i)
		if !found {
			t.Error("value not found")
		}
		if string(v) != fmt.Sprintf("some %d", i) {
			t.Errorf("expected some but got %s", string(v))
		}
	}
}

// todo: add more tests
