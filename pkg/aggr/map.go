package aggr

func Map[I, O any](values []I, predicate func(v I, id int) O) []O {
	result := make([]O, len(values))

	for i := range values {
		result[i] = predicate(values[i], i)
	}

	return result
}
