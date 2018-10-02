package poolboy_test

import (
	"fmt"
	"github.com/solodynamo/poolboy"
	"log"
	"runtime"
	"testing"
	"time"
)

var n = 1000000

func doLotOfWorkFun() {
	var n int
	for i := 0; i < 10000; i++ {
		n += i
	}
}

func logFunc(message string) {
	log.Println(message)
}

func TestGopherPool(t *testing.T) {
	poolboyPool, _ := poolboy.NewPool(poolboy.DefaultRoutinePoolSize, time.Second, logFunc)
	for i := 0; i < n; i++ {
		poolboyPool.Push(doLotOfWorkFun)
	}
	fmt.Println("____Stats____")
	t.Log("PoolCapacity", poolboyPool.Capacity())
	t.Log("SwimmingGophers", poolboyPool.Running())
	t.Log("FreeGopherSwimmers", poolboyPool.Free())
	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Log("memory usage:", mem.TotalAlloc/1024)
}
