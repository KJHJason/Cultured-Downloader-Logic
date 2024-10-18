package threadsafe

import (
	"sync"
	"testing"
)

func TestDoublyLinkedList(t *testing.T) {
	list := NewDoublyLinkedList[int]()

	// Test Append
	list.Append(10)
	list.Append(20)
	list.Append(30)

	// Test Length
	if length := list.LenUnsafe(); length != 3 {
		t.Errorf("Expected length 3, got %d", length)
	}

	// Test Front
	if value := list.FrontUnsafe().Value; value != 10 {
		t.Errorf("Expected front value 10, got %v", value)
	}

	// Test Back
	if value := list.BackUnsafe().Value; value != 30 {
		t.Errorf("Expected back value 30, got %v", value)
	}

	// Test Prepend
	list.Prepend(5)
	if value := list.FrontUnsafe().Value; value != 5 {
		t.Errorf("Expected front value 5 after prepend, got %v", value)
	}

	// Test Remove
	removed := list.RemoveUnsafe(20, TraverseFromHead, func(curValue, valueToRemove int) bool {
		return curValue == valueToRemove
	})
	if !removed {
		t.Error("Expected to remove value 20, but it was not removed")
	}
	if length := list.LenUnsafe(); length != 3 {
		t.Errorf("Expected length 3 after removing, got %d", length)
	}

	// Test Clear
	list.Clear()
	if length := list.LenUnsafe(); length != 0 {
		t.Errorf("Expected length 0 after clearing, got %d", length)
	}
	if node := list.FrontUnsafe(); node != nil {
		t.Error("Expected front to be empty after clearing")
	}
	if node := list.BackUnsafe(); node != nil {
		t.Error("Expected back to be empty after clearing")
	}
}

// Test for Remove non-existent element
func TestRemoveNonExistent(t *testing.T) {
	list := NewDoublyLinkedList[int]()
	list.Append(10)
	list.Append(20)

	removed := list.RemoveUnsafe(30, TraverseFromHead, func(curValue, valueToRemove int) bool {
		return curValue == valueToRemove
	})

	if removed {
		t.Error("Expected not to remove non-existent value 30, but it was removed")
	}
}

// Test for multiple removes
func TestMultipleRemoves(t *testing.T) {
	list := NewDoublyLinkedList[int]()
	list.Append(10)
	list.Append(20)
	list.Append(30)

	// Remove 20
	list.RemoveUnsafe(20, TraverseFromHead, func(curValue, valueToRemove int) bool {
		return curValue == valueToRemove
	})
	if length := list.LenUnsafe(); length != 2 {
		t.Errorf("Expected length 2 after removing, got %d", length)
	}

	// Remove 10
	list.RemoveUnsafe(10, TraverseFromHead, func(curValue, valueToRemove int) bool {
		return curValue == valueToRemove
	})
	if length := list.LenUnsafe(); length != 1 {
		t.Errorf("Expected length 1 after removing, got %d", length)
	}

	// Remove 30
	list.RemoveUnsafe(30, TraverseFromHead, func(curValue, valueToRemove int) bool {
		return curValue == valueToRemove
	})
	if length := list.LenUnsafe(); length != 0 {
		t.Errorf("Expected length 0 after removing, got %d", length)
	}
}

// TestConcurrentAppend tests concurrent appending to the list.
func TestConcurrentAppend(t *testing.T) {
	list := NewDoublyLinkedList[int]()
	var wg sync.WaitGroup
	numGoroutines := 10
	numElementsPerGoroutine := 100

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numElementsPerGoroutine; j++ {
				list.Append(goroutineID*100 + j)
			}
		}(i)
	}

	wg.Wait()

	// Check length
	expectedLength := numGoroutines * numElementsPerGoroutine
	if length := list.Len(); length != uint32(expectedLength) {
		t.Errorf("Expected length %d, got %d", expectedLength, length)
	}
}

// TestConcurrentRemove tests concurrent removal from the list.
func TestConcurrentRemove(t *testing.T) {
	list := NewDoublyLinkedList[int]()
	numGoroutines := 10
	numElementsPerGoroutine := 5
	totalElements := numGoroutines * numElementsPerGoroutine

	toDelete := 17
	if toDelete > totalElements {
		t.Fatal(
			"Too little elements for testing, please either increase the number of elements to add or decrease the number of elements to delete",
		)
	}

	for i := 0; i < totalElements; i++ {
		list.AppendUnsafe(i)
	}

	var wg sync.WaitGroup
	wg.Add(toDelete)
	for i := 0; i < toDelete; i++ {
		go func() {
			defer wg.Done()
			list.Remove(i, TraverseFromHead, func(curValue, valueToRemove int) bool {
				return curValue == valueToRemove
			})
		}()
	}
	wg.Wait()

	// Check remaining elements
	expectedLength := totalElements - toDelete
	if length := list.Len(); length != uint32(expectedLength) {
		t.Errorf("Expected length %d after removal, got %d", expectedLength, length)
	}
}
