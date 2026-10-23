package aggr

// Map applies fn to each element of values and returns a new slice
// of the results. The input slice is not modified. The fn receives
// each element and its index; result[i] = fn(values[i], i).
func Map[I, O any](values []I, fn func(v I, id int) O) []O {
	result := make([]O, len(values))

	for i := range values {
		result[i] = fn(values[i], i)
	}

	return result
}
