package web

type Queue[T any] []T

func (q *Queue[T]) Push(t ...T) {
	*q = append(*q, t...)
}

func (q *Queue[T]) Pop() T {

	var res T

	if len(*q) > 0 {
		res = (*q)[0]
		*q = (*q)[1:]
	}

	return res
}

func (q *Queue[T]) Len() int {
	return len(*q)
}

func (q *Queue[T]) IsEmpty() bool {
	return q.Len() == 0
}
