package lib

import "testing"

func compareSlices(xs, ys []uint32) bool {
	lx := len(xs)
	ly := len(ys)

	if lx != ly {
		return false
	}

	for i := 0; i < lx; i++ {
		if xs[i] != ys[i] {
			return false
		}
	}
	return true
}

func verify(xs, ys, expected []uint32, t *testing.T) {
	merged := MergeSortedIntArray(xs, ys)
	if !compareSlices(merged, expected) {
		t.Errorf("Error: %v != %v", merged, expected)
	}
	merged = MergeSortedIntArray(ys, xs)
	if !compareSlices(merged, expected) {
		t.Errorf("Error: %v != %v", merged, expected)
	}
}

func TestMergeSortedIntArrayOneElement(t *testing.T) {
	verify([]uint32{1}, []uint32{2}, []uint32{1, 2}, t)
}

func TestMergedPathsMultipleElements(t *testing.T) {
	verify([]uint32{1, 2}, []uint32{2, 3}, []uint32{1, 2, 3}, t)
}

func TestMergedPathsOneEmpty(t *testing.T) {
	verify([]uint32{1, 2}, []uint32{}, []uint32{1, 2}, t)
}

// todo write tests for mergeIndices
