package lib

func WeightedDistance(s, t string) int {
	substitutionCost := 21
	deletionCost := 20
	insertionCost := 1
	// degenerate cases
	if s == t {
		return 0
	}

	if len(s) == 0 {
		return len(t) * insertionCost
	}
	if len(t) == 0 {
		return len(s) * deletionCost
	}

	// create two work vectors of integer distances
	v0 := make([]int, len(s)+1)
	v1 := make([]int, len(s)+1)

	// initialize v0 (the previous row of distances)
	// this row is A[0][i]: edit distance for an empty s
	// the distance is just the number of characters to delete from t
	for i := 0; i < len(v0); i++ {
		v0[i] = i * deletionCost
	}

	for i := 0; i < len(t); i++ {
		// calculate v1 (current row distances) from the previous row v0

		// first element of v1 is A[i+1][0]
		//   edit distance is delete (i+1) chars from s to match empty t
		v1[0] = (i + 1) * deletionCost

		// use formula to fill in the rest of the row
		for j := 0; j < len(s); j++ {
			cost := 0
			if t[i] != s[j] {
				cost = substitutionCost
			}
			v1[j+1] = min(v1[j]+deletionCost, v0[j+1]+insertionCost, v0[j]+cost)
		}

		// copy v1 (current row) to v0 (previous row) for next iteration
		for j := 0; j < len(v0); j++ {
			v0[j] = v1[j]
		}
	}

	return v1[len(s)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
	} else {
		if b < c {
			return b
		}
	}
	return c
}
