package poolboy

import (
	"fmt"
	"sync/atomic"
)

type WorkerFunc interface {
	// Only Internals
	run()
	stop()
	sendTask(task func())
}

type Worker struct {
	pool *Pool
	task chan fun
	exit chan Signal
}

func (w *Worker) run() {
	go func() {
		for {
			select {
			case f := <-w.task:
				f()
				w.pool.workers <- w
				atomic.AddInt32(&w.pool.running, 1)
				if atomic.AddInt32(&w.pool.running, 0) == w.pool.capacity {
					fmt.Println(ErrCapacity)
				}
			case <-w.exit:
				atomic.AddInt32(&w.pool.running, -1)
				return
			}
		}
	}()
}

func (w *Worker) stop() {
	w.exit <- Signal{}
}

func (w *Worker) sendTask(task fun) {
	w.task <- task
}
