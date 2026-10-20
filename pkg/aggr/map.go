package aggr

// Map applies predicate to each element of values and returns a new slice
// of the results. The input slice is not modified. The predicate receives
// each element and its index; result[i] = predicate(values[i], i).
func Map[I, O any](values []I, predicate func(v I, id int) O) []O {
	result := make([]O, len(values))

	for i := range values {
		result[i] = predicate(values[i], i)
	}

	return result
}
