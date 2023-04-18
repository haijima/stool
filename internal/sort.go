package internal

import (
	"fmt"
	"strings"
)

type ComparisonFunc[T any] func(i, j T) bool

type SortOrder int

const (
	Asc  SortOrder = iota // Ascending
	Desc                  // Descending
)

type Sortable[T any] struct {
	values     []T
	mappers    map[string]ComparisonFunc[T]
	sortKeys   []string
	sortOrders []SortOrder
}

func NewSortable[T any](values []T) *Sortable[T] {
	return &Sortable[T]{
		values:  values,
		mappers: make(map[string]ComparisonFunc[T]),
	}
}

func (s *Sortable[T]) MustSetSortOption(sortKeys []string, sortOrders []SortOrder) {
	err := s.SetSortOption(sortKeys, sortOrders)
	if err != nil {
		panic(err)
	}
}

func (s *Sortable[T]) SetSortOption(sortKeys []string, sortOrders []SortOrder) error {
	if len(sortKeys) != len(sortOrders) {
		return fmt.Errorf("sortKeys and sortOrders must have the same length")
	}
	s.sortKeys = sortKeys
	s.sortOrders = sortOrders
	return nil
}

func (s *Sortable[T]) AddMapper(key string, mapper ComparisonFunc[T]) {
	s.mappers[key] = mapper
}

func (s *Sortable[T]) Len() int {
	return len(s.values)
}

func (s *Sortable[T]) Swap(i, j int) {
	s.values[i], s.values[j] = s.values[j], s.values[i]
}

func (s *Sortable[T]) Less(i, j int) bool {
	for k, key := range s.sortKeys {
		key = strings.ToLower(key)
		comparisonFunc, exists := s.mappers[key]
		if !exists {
			continue
		}
		vi, vj := s.values[i], s.values[j]
		if comparisonFunc(vi, vj) == comparisonFunc(vj, vi) {
			continue
		}

		if s.sortOrders[k] == Asc {
			return comparisonFunc(vi, vj)
		}
		return comparisonFunc(vj, vi)
	}
	return false
}

func (s *Sortable[T]) Values() []T {
	return s.values
}
