package poolboy_test

import (
	"github.com/solodynamo/poolboy"
	"sync"
	"testing"
	"time"
)

const RunTimes = 10000000

func chillout() {
	time.Sleep(time.Millisecond)
}

func BenchmarkGoroutineWithoutPoolBoy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			go func() {
				chillout()
				wg.Done()
			}()
		}
		wg.Wait()
	}
}

func BenchmarkGoroutineWithPoolBoy(b *testing.B) {
	poolboyPool, _ := poolboy.NewPool(poolboy.DefaultRoutinePoolSize, time.Second, logFunc)
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < RunTimes; j++ {
			wg.Add(1)
			poolboyPool.Push(func() {
				chillout()
				wg.Done()
			})
		}
		wg.Wait()
	}
}
