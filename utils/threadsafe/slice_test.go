package threadsafe

import (
	"sort"
	"sync"
	"testing"
)

func TestSliceGeneral(t *testing.T) {
	s := NewSliceWithCapacity[int](10)
	for i := range 10 {
		s.Append(i)
	}
	if s.Len() != 10 {
		t.Errorf("Expected length of 10, got %d", s.Len())
	}

	expectedVal := 5
	if valAtIdx, ok := s.At(expectedVal); !ok || valAtIdx != expectedVal {
		t.Errorf("Expected %d, got %d", expectedVal, valAtIdx)
	}

	if _, ok := s.At(1000); ok {
		t.Errorf("Expected false, got true")
	}

	s.Clear()
	if s.Len() != 0 {
		t.Errorf("Expected length of 0, got %d", s.Len())
	}
}

func TestSlicePop(t *testing.T) {
	s := NewSlice[int]()
	for i := range 10 {
		s.Append(i)
	}
	if s.Len() != 10 {
		t.Errorf("Expected length of 10, got %d", s.Len())
	}

	idxToRemove := 5
	removedVal, _ := s.Pop(idxToRemove)
	if s.Len() != 9 {
		t.Errorf("Expected length of 9, got %d", s.Len())
	}

	if valAtIdx, _ := s.At(idxToRemove); valAtIdx == removedVal {
		t.Errorf("Expected %d to be removed", removedVal)
	}

	if _, ok := s.Pop(1000); ok {
		t.Errorf("Expected false, got true")
	}
}

func TestSliceRemove(t *testing.T) {
	s := NewSlice[int]()
	for i := range 10 {
		s.Append(i)
	}
	if s.Len() != 10 {
		t.Errorf("Expected length of 10, got %d", s.Len())
	}

	target := 5
	s.Remove(target, func(a, b int) bool {
		return a == b
	})
	if s.Len() != 9 {
		t.Errorf("Expected length of 9, got %d", s.Len())
	}

	it := s.NewIter()
	for it.Next() {
		if it.Item() == target {
			t.Errorf("Expected %d to be removed", target)
		}
	}
}

func TestSliceIterator(t *testing.T) {
	s := NewSlice[int]()
	testLen := 10
	for i := range testLen {
		s.Append(i)
	}
	if s.Len() != testLen {
		t.Errorf("Expected length of 10, got %d", s.Len())
	}

	expectedVal := 0
	it := s.NewIter()
	for it.Next() {
		if it.Item() != expectedVal {
			t.Errorf("Expected %d, got %d", expectedVal, it.Item())
		}
		expectedVal++
	}

	// Test out of bounds
	lastIdx := testLen - 1
	if it.Next(); it.index != lastIdx {
		t.Errorf("Expected index of %d, got %d", lastIdx, it.index)
	}

	it.Reset()
	expectedVal = testLen - 1
	for it.Prev() {
		if it.Item() != expectedVal {
			t.Errorf("Expected %d, got %d", expectedVal, it.Item())
		}
		expectedVal--
	}

	// Test out of bounds
	if it.Prev(); it.index != 0 {
		t.Errorf("Expected index of 0, got %d", it.index)
	}
}

func TestSliceReverse(t *testing.T) {
	s := NewSlice[int]()
	testLen := 10
	for i := range testLen {
		s.Append(i)
	}
	if s.Len() != testLen {
		t.Errorf("Expected length of 10, got %d", s.Len())
	}

	s.Reverse()
	expectedVal := testLen - 1
	it := s.NewIter()
	for it.Next() {
		if it.Item() != expectedVal {
			t.Errorf("Expected %d, got %d", expectedVal, it.Item())
		}
		expectedVal--
	}
}

func TestSliceSort(t *testing.T) {
	s := NewSlice[int]()
	testLen := 10
	for i := range testLen {
		s.Append(i)
	}
	if s.Len() != testLen {
		t.Errorf("Expected length of 10, got %d", s.Len())
	}

	s.Sort(func(i, j int) bool {
		return i < j
	})
	expectedVal := 0
	it := s.NewIter()
	for it.Next() {
		if it.Item() != expectedVal {
			t.Errorf("Expected %d, got %d", expectedVal, it.Item())
		}
		expectedVal++
	}
}

func TestSliceSortStable(t *testing.T) {
	s := NewSlice[int]()
	testLen := 10
	for i := range testLen {
		s.Append(i)
	}
	if s.Len() != testLen {
		t.Errorf("Expected length of 10, got %d", s.Len())
	}

	s.SortStable(func(i, j int) bool {
		return i < j
	})
	expectedVal := 0
	it := s.NewIter()
	for it.Next() {
		if it.Item() != expectedVal {
			t.Errorf("Expected %d, got %d", expectedVal, it.Item())
		}
		expectedVal++
	}
}

func TestSliceCopy(t *testing.T) {
	s := NewSlice[int]()
	testLen := 10
	for i := range testLen {
		s.Append(i)
	}
	if s.Len() != testLen {
		t.Errorf("Expected length of 10, got %d", s.Len())
	}

	itemsCopy := s.CopyItems()
	if len(itemsCopy) != testLen {
		t.Errorf("Expected length of 10, got %d", len(itemsCopy))
	}

	for i := range testLen {
		if itemsCopy[i] != i {
			t.Errorf("Expected %d, got %d", i, itemsCopy[i])
		}
	}
}

func TestThreadSafety(t *testing.T) {
	values := []string{
		"test",
		"kjhjason",
		"hello",
		"world",
		"cultured",
		"cdl",
	}
	s := NewSliceWithCapacity[string](len(values))

	wg := sync.WaitGroup{}
	for _, v := range values {
		wg.Add(1)
		go func() {
			s.Append(v)
			wg.Done()
		}()
	}
	wg.Wait()

	if s.Len() != len(values) {
		t.Errorf("Expected length of %d, got %d", len(values), s.Len())
	}

	less := func(i, j string) bool {
		return i < j
	}
	s.Sort(less)
	if !s.IsSorted(less) {
		t.Errorf("Expected slice to be sorted")
	}

	sort.Slice(values, func(i, j int) bool {
		return less(values[i], values[j])
	})
	for i, v := range values {
		if item, _ := s.At(i); item != v {
			t.Errorf("Expected %s, got %s", v, item)
		}
	}
}
