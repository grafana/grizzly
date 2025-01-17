package utils

// Map creates a new list populated with the results of applying the `mapper`
// function to every element from `input`.
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
