package gmover

//goland:noinspection GoUnusedParameter
func noop(...any) {}

func StringSlice[T ~string](tt []T) (ss []string) {
	ss = make([]string, len(tt))
	for i := range tt {
		ss[i] = string(tt[i])
	}
	return ss
}

// SlicesIntersect returns true if any element in slice1 is also in slice2.
func SlicesIntersect[S ~string](slice1 []S, slice2 []S) bool {
	var found bool
	// Build a map for one of the slices to allow O(1) lookups.
	lookup := make(map[string]struct{}, len(slice1))
	for _, s := range slice1 {
		lookup[string(s)] = struct{}{}
	}

	for _, s := range slice2 {
		_, found = lookup[string(s)]
		if found {
			goto end
		}
	}
end:
	return found
}
