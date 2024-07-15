package threadsafe

import (
	"sync"
	"testing"
)

func BenchmarkThreadsafeSliceBig(b *testing.B) {
	elLen := 1_000_000
	for range b.N {
		threadSafeSliceBenchmarkLogic(elLen)
	}
}

func BenchmarkThreadsafeSliceWithCapBig(b *testing.B) {
	elLen := 1_000_000
	for range b.N {
		threadSafeSliceWithCapBenchmarkLogic(elLen)
	}
}

func BenchmarkBufferedChannelBig(b *testing.B) {
	elLen := 1_000_000
	for range b.N {
		bufferedChannelBenchmarkLogic(elLen)
	}
}

func BenchmarkThreadsafeSliceSmall(b *testing.B) {
	elLen := 100
	for range b.N {
		threadSafeSliceBenchmarkLogic(elLen)
	}
}

func BenchmarkThreadsafeSliceWithCapSmall(b *testing.B) {
	elLen := 100
	for range b.N {
		threadSafeSliceWithCapBenchmarkLogic(elLen)
	}
}

func BenchmarkBufferedChannelSmall(b *testing.B) {
	elLen := 100
	for range b.N {
		bufferedChannelBenchmarkLogic(elLen)
	}
}

func threadSafeSliceBenchmarkLogic(elLen int) {
	slice := NewSlice[int]()
	wg := sync.WaitGroup{}
	for i := range elLen {
		wg.Add(1)
		go func() {
			slice.Append(i)
			wg.Done()
		}()
	}
	wg.Wait()
}

func threadSafeSliceWithCapBenchmarkLogic(elLen int) {
	slice := NewSliceWithCapacity[int](elLen)
	wg := sync.WaitGroup{}
	for i := range elLen {
		wg.Add(1)
		go func() {
			slice.Append(i)
			wg.Done()
		}()
	}
	wg.Wait()
}

func bufferedChannelBenchmarkLogic(elLen int) {
	ch := make(chan int, elLen)
	wg := sync.WaitGroup{}
	for i := range elLen {
		wg.Add(1)
		go func() {
			ch <- i
			wg.Done()
		}()
	}
	wg.Wait()
	close(ch)
}
