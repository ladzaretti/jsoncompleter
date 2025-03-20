package jsoncompleter

type stack[T any] struct {
	items []T
}

//nolint:ireturn
func (s *stack[T]) peek() (T, bool) {
	if len(s.items) == 0 {
		var t T
		return t, false
	}

	return s.items[len(s.items)-1], true
}

func (s *stack[T]) push(t T) {
	s.items = append(s.items, t)
}

//nolint:ireturn
func (s *stack[T]) pop() (T, bool) {
	var t T
	if len(s.items) == 0 {
		return t, false
	}

	t, s.items = s.items[len(s.items)-1], s.items[:len(s.items)-1]

	return t, true
}

func (s *stack[T]) empty() bool {
	return len(s.items) == 0
}
