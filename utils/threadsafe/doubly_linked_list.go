package threadsafe

import (
	"sync"
)

type DoublyLinkedListNode[T any] struct {
	Value T
	next  *DoublyLinkedListNode[T]
	prev  *DoublyLinkedListNode[T]
}

func (n *DoublyLinkedListNode[T]) Next() *DoublyLinkedListNode[T] {
	return n.next
}

func (n *DoublyLinkedListNode[T]) Prev() *DoublyLinkedListNode[T] {
	return n.prev
}

type DoublyLinkedList[T any] struct {
	head *DoublyLinkedListNode[T]
	tail *DoublyLinkedListNode[T]
	len  uint32
	mu   sync.RWMutex
}

func NewDoublyLinkedList[T any]() *DoublyLinkedList[T] {
	return &DoublyLinkedList[T]{}
}

// Clear removes all elements from the list.
func (l *DoublyLinkedList[T]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.ClearUnsafe()
}

// ClearUnsafe removes all elements from the list without locking.
func (l *DoublyLinkedList[T]) ClearUnsafe() {
	l.head = nil
	l.tail = nil
	l.len = 0
}

// Len returns the number of elements in the list.
func (l *DoublyLinkedList[T]) Len() uint32 {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.LenUnsafe()
}

// LenUnsafe returns the number of elements in the list without locking.
func (l *DoublyLinkedList[T]) LenUnsafe() uint32 {
	return l.len
}

// Front returns the first element in the list.
func (l *DoublyLinkedList[T]) Front() *DoublyLinkedListNode[T] {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.FrontUnsafe()
}

// FrontUnsafe returns the first element in the list without locking.
func (l *DoublyLinkedList[T]) FrontUnsafe() *DoublyLinkedListNode[T] {
	if l.len == 0 {
		return nil
	}
	return l.head
}

// Back returns the last element in the list.
func (l *DoublyLinkedList[T]) Back() *DoublyLinkedListNode[T] {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.BackUnsafe()
}

// BackUnsafe returns the last element in the list without locking.
func (l *DoublyLinkedList[T]) BackUnsafe() *DoublyLinkedListNode[T] {
	if l.len == 0 {
		return nil
	}
	return l.tail
}

// Append adds a new element to the end of the list.
func (l *DoublyLinkedList[T]) Append(value T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.AppendUnsafe(value)
}

// AppendUnsafe adds a new element to the end of the list without locking.
func (l *DoublyLinkedList[T]) AppendUnsafe(value T) {
	newNode := &DoublyLinkedListNode[T]{Value: value}

	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		l.tail.next = newNode
		newNode.prev = l.tail
		l.tail = newNode
	}
	l.len++
}

// Prepend adds a new element to the beginning of the list.
func (l *DoublyLinkedList[T]) Prepend(value T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.PrependUnsafe(value)
}

// PrependUnsafe adds a new element to the beginning of the list without locking.
func (l *DoublyLinkedList[T]) PrependUnsafe(value T) {
	newNode := &DoublyLinkedListNode[T]{Value: value}

	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		newNode.next = l.head
		l.head.prev = newNode
		l.head = newNode
	}
	l.len++
}

type TraversalDirection int8

const (
	TraverseFromHead TraversalDirection = iota
	TraverseFromTail
)

// removeNode removes a node from the list and returns true if successful.
func (l *DoublyLinkedList[T]) removeNode(node *DoublyLinkedListNode[T]) bool {
	if node == nil {
		return false
	}

	// Adjust the pointers of adjacent nodes
	if node.prev != nil {
		node.prev.next = node.next
	} else {
		// We're removing the head
		l.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		// We're removing the tail
		l.tail = node.prev
	}

	l.len--
	return true // Node was found and removed
}

// Remove deletes the first occurrence of the specified value from the list.
func (l *DoublyLinkedList[T]) Remove(value T, direction TraversalDirection, cmp func(curValue T, valueToRemove T) bool) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.RemoveUnsafe(value, direction, cmp)
}

// RemoveUnsafe deletes the first occurrence of the specified value from the list without locking.
func (l *DoublyLinkedList[T]) RemoveUnsafe(value T, direction TraversalDirection, cmp func(curValue T, valueToRemove T) bool) bool {
	fromHead := direction == TraverseFromHead
	var current *DoublyLinkedListNode[T]
	if fromHead {
		current = l.head
	} else {
		current = l.tail
	}

	for current != nil {
		if cmp(current.Value, value) {
			return l.removeNode(current) // Remove the found node
		}

		if fromHead {
			current = current.next
		} else {
			current = current.prev
		}
	}
	return false // Value not found
}

func (l *DoublyLinkedList[T]) RemoveViaFn(direction TraversalDirection, deleteCheck func(curValue T) bool) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.RemoveViaFnUnsafe(direction, deleteCheck)
}

// RemoveViaFnUnsafe deletes the first occurrence of the node that matches the condition specified by deleteCheck without locking.
func (l *DoublyLinkedList[T]) RemoveViaFnUnsafe(direction TraversalDirection, deleteCheck func(curValue T) bool) bool {
	fromHead := direction == TraverseFromHead
	var current *DoublyLinkedListNode[T]
	if fromHead {
		current = l.head
	} else {
		current = l.tail
	}

	for current != nil {
		if deleteCheck(current.Value) {
			return l.removeNode(current) // Remove the found node
		}

		if fromHead {
			current = current.next
		} else {
			current = current.prev
		}
	}
	return false // Value not found
}
