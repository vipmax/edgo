package main

func max(x, y int) int {
	if x < y { return y }
	return x
}

func min(x, y int) int {
	if x <= y { return x }
	return y
}

func insert[T any](a []T, index int, value T) []T {
	n := len(a)
	if index < 0 {
		index = (index%n + n) % n
	}
	switch {
	case index == n: // nil or empty slice or after last element
		return append(a, value)

	case index < n: // index < len(a)
		a = append(a[:index+1], a[index:]...)
		a[index] = value
		return a

	case index < cap(a): // index > len(a)
		a = a[:index+1]
		var zero T
		for i := n; i < index; i++ {
			a[i] = zero
		}
		a[index] = value
		return a

	default:
		b := make([]T, index+1) // malloc
		if n > 0 {
			copy(b, a)
		}
		b[index] = value
		return b
	}
}

var matched = []rune{' ', '.', ',', '=', '+', '-', '[', '(', '{', '"'}

func findNextWord(chars []rune, from int) int {
	// Find the next word index after the specified index
	for i := from; i < len(chars); i++ {
		if contains(matched, chars[i]) {
			return i
		}
	}

	return len(chars)
}

func findPrevWord(chars []rune, from int) int {
	// Find the previous word index before the specified index
	for i := from - 1; i >= 0; i-- {
		if contains(matched, chars[i]) {
			return i + 1
		}
	}

	return 0
}

func contains[T comparable](slice []T, e T) bool {
	for _, val := range slice {
		if val == e {
			return true
		}
	}
	return false
}

func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}


func GreaterThan(x, y, x1, y1 int) bool {
	if y > y1 { return true }
	return y == y1 && x > x1
}

func LessThan(x, y, x1, y1 int) bool {
	if y < y1 { return true }
	return y == y1 && x < x1
}

func GreaterEqual(x, y, x1, y1 int) bool {
	if y > y1 { return true }
	if y == y1 && x >= x1 {
		return true
	}
	return false
}
