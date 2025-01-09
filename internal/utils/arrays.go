package utils

func Map[T any, O any](input []T, mapper func(T) O) []O {
	if input == nil {
		return nil
	}

	output := make([]O, len(input))

	for i := range input {
		output[i] = mapper(input[i])
	}

	return output
}
