package internal

import "github.com/samber/lo"

type Lookup[T comparable] map[T]struct{}

func NewLookup[T comparable](items ...T) Lookup[T] {
	return lo.SliceToMap(items, func(item T) (T, struct{}) {
		return item, struct{}{}
	})
}

func (l Lookup[T]) Contains(item T) bool {
	if l == nil {
		return false
	}

	_, ok := l[item]

	return ok
}
