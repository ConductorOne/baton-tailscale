package utils

import (
	_ "golang.org/x/exp/slices"
)

// TODO(marcos): Copy the slices helpers from C1 to baton-sdk.

// Convert accepts a list of T and returns a list of R based on the input func.
func Convert[T any, R any](slice []T, f func(in T) R) []R {
	output := make([]R, 0, len(slice))
	for _, t := range slice {
		output = append(output, f(t))
	}
	return output
}

// Unique returns a list of de-duplicated items. The first matched item is returned.
func Unique[T comparable](slice []T) []T {
	var output []T
	dupeTrack := make(map[T]bool)

	for _, o := range slice {
		if _, ok := dupeTrack[o]; ok {
			continue
		}

		dupeTrack[o] = true
		output = append(output, o)
	}

	return output
}
