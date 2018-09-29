package poolboy_test;

import (
	"testing"
	"github.com/solodynamo/poolboy"
	"fmt"
	"runtime"
)

var n = 100000;

func doLotOfWorkFun() {
	for i := 0; i < 1000000; i++ {

	}
}

func TestGopherPool(t *testing.T) {
	for i := 0; i < n; i++ {
		poolboy.Push(doLotOfWorkFun)
	}
 	t.Logf("PoolCapacity:%d", poolboy.PoolCapacity())
	t.Logf("SwimmingGophers:%d", poolboy.SwimmingGophers())
	t.Logf("FreeGopherSwimmers:%d", poolboy.FreeGopherSwimmers())
 	mem := runtime.MemStats{}
	runtime.ReadMemStats(&mem)
	fmt.Println("memory usage:", mem.TotalAlloc/1024)
}