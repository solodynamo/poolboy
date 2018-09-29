package poolboy;

import (
	"sync"
	"math"
	"runtime"
	"sync/atomic"
);

type Signal struct{};

type fun func();

type Worker struct {
	pool *Pool
	task chan fun
	exit chan Signal
}

type Pool struct {
	workers chan *Worker
	destroy chan Signal
	mx sync.Mutex
	tasks chan fun
	capacity int32
	running int32
};


const DEFAULT_ROUTINE_POOL_SIZE = math.MaxInt32;
var gopherPool = NewPool(DEFAULT_ROUTINE_POOL_SIZE);


func (w *Worker) run() {
	go func() {
		for {
			select {
			case f := <-w.task:
				f();
				w.pool.workers <- w;
			case <-w.exit:
				atomic.AddInt32(&w.pool.running, -1);
				return;
			}
		}
	}()
}

func (w *Worker) stop() {
	w.exit <- Signal{};
}

func (w *Worker) sendTask(task fun) {
	w.task <- task;
}

func Push(task fun) error {
	return gopherPool.Push(task);
}

func SwimmingGophers() int {
	return gopherPool.Running();
}

func PoolCapacity() int {
	return gopherPool.Capacity();
}

func FreeGopherSwimmers() int {
	return gopherPool.Free();
}

func NewPool(size int32) *Pool {
	p := &Pool{
		capacity: size,
		tasks: make(chan fun, math.MaxInt32),
		workers: make(chan *Worker, size),
		destroy: make(chan Signal, runtime.GOMAXPROCS(-1)),
	};
	p.loop(); 
	return p;
}

func (p *Pool) loop() {
	for i :=0; i < runtime.GOMAXPROCS(-1); i++ {
		go func() {
			for {
				select {
				case task := <-p.tasks:
					p.getWorker().sendTask(task)
				case <-p.destroy:
					return;
				}
			}
		}()
	}
}

func (p *Pool) Running() int {
	swimmingGophers := int(atomic.LoadInt32(&p.running));
	return swimmingGophers;
}

func (p *Pool) Free() int {
	idleGophers := int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running));
	return idleGophers;
}
func (p *Pool) Push(task fun) error {
	if len(p.destroy) > 0 {
		return nil;
	}
	p.tasks <- task;
	return nil
}

func (p *Pool) Capacity() int {
	return int(atomic.LoadInt32(&p.capacity));
}

func (p *Pool) Destroy() error {
	p.mx.Lock();
	defer p.mx.Unlock();
	for i:=0; i < runtime.GOMAXPROCS(-1) +1; i++ {
		p.destroy <- Signal{};
	}
	return nil;
}

func (p *Pool) reachLimit() bool {
	return p.Running() >= p.Capacity();
}

func (p *Pool) newWorker() *Worker {
	worker := &Worker{
		pool: p,
		task: make(chan fun),
		exit: make(chan Signal),
	}
	worker.run();
	return worker;
}

func (p *Pool) getWorker() *Worker {
	var worker *Worker;
	defer atomic.AddInt32(&p.running,1);
	if p.reachLimit() {
		worker = <-p.workers;
	} else {
		select {
		case worker = <-p.workers:
			return worker;
		default:
			worker = p.newWorker();
		}
	}
	return worker;
}
