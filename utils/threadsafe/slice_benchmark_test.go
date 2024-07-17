package threadsafe

import (
	"sync"
	"testing"
)

func BenchmarkThreadsafeSliceBig(b *testing.B) {
	for range b.N {
		threadSafeSliceBenchmarkLogic(getBigN())
	}
}

func BenchmarkThreadsafeSliceWithCapBig(b *testing.B) {
	for range b.N {
		threadSafeSliceWithCapBenchmarkLogic(getBigN())
	}
}

func BenchmarkBufferedChannelBig(b *testing.B) {
	for range b.N {
		bufferedChannelBenchmarkLogic(getBigN())
	}
}

func BenchmarkThreadsafeSliceSmall(b *testing.B) {
	for range b.N {
		threadSafeSliceBenchmarkLogic(getSmallN())
	}
}

func BenchmarkThreadsafeSliceWithCapSmall(b *testing.B) {
	for range b.N {
		threadSafeSliceWithCapBenchmarkLogic(getSmallN())
	}
}

func BenchmarkBufferedChannelSmall(b *testing.B) {
	for range b.N {
		bufferedChannelBenchmarkLogic(getSmallN())
	}
}

func getSmallN() int {
	return 100
}

func getBigN() int {
	return 1_000_000
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
