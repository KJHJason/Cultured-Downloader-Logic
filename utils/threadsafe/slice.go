package threadsafe

import (
	"slices"
	"sort"
	"sync"
)

type Slice[T any] struct {
	items []T
	mu    sync.RWMutex
}

func NewSlice[T any]() *Slice[T] {
	return &Slice[T]{
		items: make([]T, 0),
	}
}

func NewSliceWithCapacity[T any](capacity int) *Slice[T] {
	return &Slice[T]{
		items: make([]T, 0, capacity),
	}
}

func (s *Slice[T]) withinBound(idx int) bool {
	return idx >= 0 && idx < s.LenUnsafe()
}

func (s *Slice[T]) Append(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.AppendUnsafe(v)
}

func (s *Slice[T]) AppendUnsafe(v T) {
	s.items = append(s.items, v)
}

func (s *Slice[T]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ClearUnsafe()
}

func (s *Slice[T]) ClearUnsafe() {
	s.items = make([]T, 0)
}

func (s *Slice[T]) At(idx int) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.AtUnsafe(idx)
}

func (s *Slice[T]) AtUnsafe(idx int) (T, bool) {
	if !s.withinBound(idx) {
		var defaultVal T
		return defaultVal, false
	}
	return s.items[idx], true
}

func (s *Slice[T]) Pop(idx int) (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.PopUnsafe(idx)
}

func (s *Slice[T]) PopUnsafe(idx int) (T, bool) {
	if !s.withinBound(idx) {
		var defaultVal T
		return defaultVal, false
	}
	v := s.items[idx]
	s.items = append(s.items[:idx], s.items[idx+1:]...)
	return v, true
}

func (s *Slice[T]) Remove(target T, equals func(a, b T) bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RemoveUnsafe(target, equals)
}

func (s *Slice[T]) RemoveUnsafe(target T, equals func(a, b T) bool) {
	for i, v := range s.items {
		if equals(target, v) {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return
		}
	}
}

func (s *Slice[T]) CopyItems() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.CopyItemsUnsafe()
}

func (s *Slice[T]) CopyItemsUnsafe() []T {
	return append([]T(nil), s.items...)
}

func (s *Slice[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LenUnsafe()
}

func (s *Slice[T]) LenUnsafe() int {
	return len(s.items)
}

func (s *Slice[T]) Reverse() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ReverseUnsafe()
}

func (s *Slice[T]) ReverseUnsafe() {
	slices.Reverse(s.items)
}

type LessFunc[T any] func(valAtI, valAtJ T) bool

func (s *Slice[T]) Sort(less LessFunc[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SortUnsafe(less)
}

func (s *Slice[T]) SortUnsafe(less LessFunc[T]) {
	sort.Slice(s.items, func(i, j int) bool {
		return less(s.items[i], s.items[j])
	})
}

func (s *Slice[T]) SortStable(less LessFunc[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SortStableUnsafe(less)
}

func (s *Slice[T]) SortStableUnsafe(less LessFunc[T]) {
	sort.SliceStable(s.items, func(i, j int) bool {
		return less(s.items[i], s.items[j])
	})
}

func (s *Slice[T]) IsSorted(less LessFunc[T]) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.IsSortedUnsafe(less)
}

func (s *Slice[T]) IsSortedUnsafe(less LessFunc[T]) bool {
	return sort.SliceIsSorted(s.items, func(i, j int) bool {
		return less(s.items[i], s.items[j])
	})
}

const SLICE_ITERATOR_INIT_VAL = -1

// Note that SliceIterator is not a thread-safe iterator by default!
//
// Why? As of now, I don't see a need for it to be thread-safe.
// If you need a thread-safe iterator, use the functions denoted with the "Safe" suffix.
//
// Additionally, you would usually read the contents of the slice
// after some operations that modifies the slice in multiple goroutines.
type SliceIterator[T any] struct {
	slice *Slice[T]
	index int
}

func (s *Slice[T]) NewIter() *SliceIterator[T] {
	return &SliceIterator[T]{
		slice: s,
		index: SLICE_ITERATOR_INIT_VAL,
	}
}

func (it *SliceIterator[T]) Reset() {
	it.index = SLICE_ITERATOR_INIT_VAL
}

func (it *SliceIterator[T]) ResetSafe() {
	it.slice.mu.Lock()
	defer it.slice.mu.Unlock()
	it.Reset()
}

func (it *SliceIterator[T]) Next() bool {
	it.index++
	if sliceLen := len(it.slice.items); it.index >= sliceLen {
		it.index = sliceLen - 1
		return false
	}
	return true
}

func (it *SliceIterator[T]) NextSafe() bool {
	it.slice.mu.Lock()
	defer it.slice.mu.Unlock()
	return it.Next()
}

func (it *SliceIterator[T]) Prev() bool {
	if it.index == SLICE_ITERATOR_INIT_VAL {
		it.index = len(it.slice.items)
	}

	it.index--
	if it.index < 0 {
		it.index = 0
		return false
	}
	return true
}

func (it *SliceIterator[T]) PrevSafe() bool {
	it.slice.mu.Lock()
	defer it.slice.mu.Unlock()
	return it.Prev()
}

func (it *SliceIterator[T]) Item() T {
	if it.index == SLICE_ITERATOR_INIT_VAL {
		panic("Iterator not started, did you forget to call Next/Prev?")
	}
	return it.slice.items[it.index]
}

func (it *SliceIterator[T]) ItemSafe() T {
	it.slice.mu.RLock()
	defer it.slice.mu.RUnlock()
	return it.Item()
}
