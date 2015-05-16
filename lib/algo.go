package lib

import "sort"

func MergeSortedIntArray(xs, ys []uint32) []uint32 {
	i := 0
	j := 0
	k := 0
	lx := len(xs)
	ly := len(ys)
	merged := make([]uint32, lx+ly)
	for i < lx || j < ly {
		if i == lx {
			merged[k] = ys[j]
			j++
		} else if j == ly {
			merged[k] = xs[i]
			i++
		} else {
			if xs[i] == ys[j] {
				merged[k] = xs[i]
				i++
				j++
			} else if xs[i] < ys[j] {
				merged[k] = xs[i]
				i++
			} else {
				merged[k] = ys[j]
				j++
			}
		}
		k++
	}
	return merged[0:k]
}

type UInt32ByValue []uint32

func (a UInt32ByValue) Len() int           { return len(a) }
func (a UInt32ByValue) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a UInt32ByValue) Less(i, j int) bool { return a[i] < a[j] }

func MergeIndices(idx1, idx2 map[string][]uint32) map[string][]uint32 {
	indices := make([]map[string][]uint32, 0)
	indices = append(indices, idx1)
	indices = append(indices, idx2)

	newIdx := make(map[string][]uint32)
	for _, idx := range indices {
		for trigram, paths := range idx {
			sort.Sort(UInt32ByValue(paths))
			newIdx[trigram] = MergeSortedIntArray(paths, newIdx[trigram])
		}
	}
	return newIdx
}
