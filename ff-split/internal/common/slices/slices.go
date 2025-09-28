package slices

import "slices"

func DiffUsingMap[T comparable](a, b []T) []T {
	m := make(map[T]struct{}, len(b))
	for _, v := range b {
		m[v] = struct{}{}
	}
	var result []T
	for _, v := range a {
		if _, found := m[v]; !found {
			result = append(result, v)
		}
	}
	return result
}

func DiffNestedLoops[T comparable](a, b []T) []T {
	var result []T
	for _, v := range a {
		if !slices.Contains(b, v) {
			result = append(result, v)
		}
	}
	return result
}

func SmartDiff[T comparable](a, b []T) []T {
	// Эвристика: если один из слайсов короткий (например, <=10), используем вложенные циклы
	if len(a) <= 10 || len(b) <= 10 {
		return DiffNestedLoops(a, b)
	}
	return DiffUsingMap(a, b)
}
