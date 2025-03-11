package rejson

type Stack[T any] struct {
	items []T
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{items: make([]T, 0, 32)}
}

//nolint:ireturn
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var t T
		return t, false
	}

	return s.items[len(s.items)-1], true
}

func (s *Stack[T]) Push(el T) {
	s.items = append(s.items, el)
}

//nolint:ireturn
func (s *Stack[T]) Pop() (T, bool) {
	var t T
	if len(s.items) == 0 {
		return t, false
	}

	t, s.items = s.items[len(s.items)-1], s.items[:len(s.items)-1]

	return t, true
}

func (s *Stack[T]) Empty() bool {
	return len(s.items) == 0
}
