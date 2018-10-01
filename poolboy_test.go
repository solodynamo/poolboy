package poolboy_test;

import (
	"testing"
	"github.com/solodynamo/poolboy"
	"runtime"
	"fmt"
)

var n = 500000;

func doLotOfWorkFun() {
	var n int
	for i := 0; i < 10000; i++ {
		n += i
	}
}

func TestGopherPool(t *testing.T) {
	for i := 0; i < n; i++ {
		poolboy.Push(doLotOfWorkFun)
	}
	fmt.Println("____Stats____");
 	t.Log("PoolCapacity", poolboy.PoolCapacity())
	t.Log("SwimmingGophers", poolboy.SwimmingGophers())
	t.Log("FreeGopherSwimmers", poolboy.FreeGopherSwimmers())
	poolboy.WaitingGophers();
 	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	t.Log("memory usage:", mem.TotalAlloc/1024)
}