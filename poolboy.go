package poolboy

import (
	"fmt"
	"log"
	"runtime"
	"sync/atomic"
	"time"
)

type Signal struct{}

type fun func()

func init() {
	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func limitThreadsAccToCPUCores() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func (p *Pool) Running() int {
	swimmingGophers := int(atomic.LoadInt32(&p.running))
	return swimmingGophers
}

func (p *Pool) Free() int {
	idleGophers := int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running))
	return idleGophers
}
func (p *Pool) Push(task fun) error {
	if len(p.destroy) > 0 {
		return nil
	}
	p.tasks <- task
	return nil
}

func (p *Pool) Capacity() int {
	return int(atomic.LoadInt32(&p.capacity))
}

func (p *Pool) Destroy() error {
	p.mx.Lock()
	defer p.mx.Unlock()
	routines := int(atomic.LoadInt32(&p.running))
	for i := 0; i < routines; i++ {
		p.destroy <- Signal{}
	}
	return nil
}

func (p *Pool) Wait() {
	p.wg.Wait()
}

//---------------Internals------------------//
func (p *Pool) getWorker() *Worker {
	defer atomic.AddInt32(&p.running, 1)
	var worker *Worker
	if p.reachLimit() {
		worker = <-p.workers
	} else {
		select {
		case worker = <-p.workers:
			return worker
		default:
			worker = p.newWorker()
		}
	}
	return worker
}

func (p *Pool) loop() {
	/*
	* Don't alter the no of threads that will be used out of system for running go code by passing -1.
	* Spawn go routines till no of default allocated OS threads reached.
	**/
	for i := 0; i < runtime.GOMAXPROCS(-1); i++ {
		go func() {
			timer := time.NewTimer(p.statTime)
			for {
				select {
				case task := <-p.tasks:
					p.getWorker().sendTask(task)

				case <-p.destroy:
					return
				case <-timer.C:
					// Capture the stats.
					routines := atomic.LoadInt32(&p.running)
					capacity := atomic.LoadInt32(&p.capacity)

					// Display the stats.
					p.log(fmt.Sprintf("Pool : Stats : A[%d] C[%d] ", routines, capacity))

					// Reset the clock.
					timer.Reset(p.statTime)
				}
			}
		}()
	}
}

func (p *Pool) reachLimit() bool {
	return p.Running() >= p.Capacity()
}

func (p *Pool) newWorker() *Worker {
	if p.reachLimit() {
		<-p.freeSignal
		return p.getWorker()
	}
	worker := &Worker{
		pool: p,
		task: make(chan fun),
		exit: make(chan Signal),
	}
	worker.run()
	return worker
}

func (p *Pool) log(message string) {
	if p.logFunc != nil {
		p.logFunc(message)
	}
}
