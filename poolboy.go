package poolboy

import (
	"container/list"
	"errors"
	"fmt"
	"log"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Signal struct{}

type fun func()

type Worker struct {
	pool *Pool
	task chan fun
	exit chan Signal
}

type Pool struct {
	tasks        chan fun
	workers      chan *Worker
	destroy      chan Signal
	statTime    time.Duration  // time to display stats.
	mx           *sync.Mutex
	wg           *sync.WaitGroup
	freeSignal   chan Signal
	capacity     int32
	running      int32
	stopped      int32
	logFunc      func(message string)
}

type DLL struct {
	doublylinkedlist *list.List
	m                sync.Mutex
}

const DEFAULT_ROUTINE_POOL_SIZE = math.MaxInt32

var (
	ErrorInvalidMinRoutines = errors.New("Invalid minimum number of routines")
	ErrorInvalidStatTime = errors.New("Invalid duration for stat time")
	ErrNoTaskInQ = errors.New("No task to run")
	ErrCapacity = errors.New("Thread Pool At Capacity")
)

func logFunc(message string) {
	log.Println(message)
}
var once sync.Once

func init() {
	log.SetPrefix("TRACE: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func limitThreadsAccToCPUCores() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func getIntOnce() *Pool {
	var gpool *Pool
	limitThreadsAccToCPUCores()
	once.Do(func() {
		log.Println("Initializing Pool")
		gpool, _ = NewPool(DEFAULT_ROUTINE_POOL_SIZE, time.Second, logFunc)
	})
	return gpool
}

var gopherPool = getIntOnce()

func (w *Worker) run() {
	go func() {
		for {
			select {
			case f := <-w.task:
				f()
				w.pool.workers <- w
				atomic.AddInt32(&w.pool.running,1)
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

func NewPool(size int32, statTime time.Duration, logFunc func(message string)) (*Pool, error) {
	if size <= 0 {
		return nil, ErrorInvalidMinRoutines
	}

	if statTime < time.Millisecond {
		return nil, ErrorInvalidStatTime
	}

	p := &Pool{
		capacity:     size,
		tasks:        make(chan fun, size),
		workers:      make(chan *Worker, size),
		freeSignal:   make(chan Signal, math.MaxInt32),
		destroy:      make(chan Signal),
		statTime:    statTime,
		wg:           &sync.WaitGroup{},
		logFunc:      logFunc,
		running: 0,
		stopped: 0,
	}
	p.loop()
	return p,nil
}

func Push(task fun) error {
	return gopherPool.Push(task)
}

func SwimmingGophers() int {
	return gopherPool.Running()
}

func WaitingGophers() {
	gopherPool.Wait()
}

func PoolCapacity() int {
	return gopherPool.Capacity()
}

func FreeGopherSwimmers() int {
	return gopherPool.Free()
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

func (p *Pool) log(message string) {
	if p.logFunc != nil {
		p.logFunc(message)
	}
}
