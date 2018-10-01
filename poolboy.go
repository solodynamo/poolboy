package poolboy;

import (
	"log"
	"sync"
	"math"
	"runtime"
	"sync/atomic"
	"container/list"
);

type Signal struct{};

type fun func();

type Worker struct {
	pool *Pool
	task chan fun
	exit chan Signal
}

type Pool struct {
	tasks        *DLL
	workers      *DLL
	destroy chan Signal
	mx *sync.Mutex
	wg *sync.WaitGroup
	freeSignal chan Signal
	launchSignal chan Signal
	capacity int32
	running int32
	logFunc func(message string)
};

type DLL struct {
	doublylinkedlist *list.List
	m     sync.Mutex
}

const DEFAULT_ROUTINE_POOL_SIZE = math.MaxInt32;

func logFunc(message string) {
	log.Println(message)
}
var once sync.Once

func getIntOnce() *Pool{
	var gpool * Pool
	once.Do(func() {
		log.Println("Initializing Pool")
		gpool = NewPool(DEFAULT_ROUTINE_POOL_SIZE,logFunc)
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
				w.pool.workers.push(w)
				w.pool.wg.Done()
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

func NewPool(size int32, logFunc func(message string)) *Pool {
	p := &Pool{
		capacity: size,
		tasks: NewDLL(),
		workers: NewDLL(),
		freeSignal:   make(chan Signal, math.MaxInt32),
		launchSignal: make(chan Signal, math.MaxInt32),
		destroy: make(chan Signal, runtime.GOMAXPROCS(-1)),
		wg:           &sync.WaitGroup{},
		logFunc:logFunc,
	};
	p.loop();
	return p;
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
	for i :=0; i < runtime.GOMAXPROCS(-1); i++ {
		go func() {
			for {
				select {
				case <-p.launchSignal:
					p.getWorker().sendTask(p.tasks.pop().(fun))
				case <-p.destroy:
					return
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
	idleGophers := int(atomic.LoadInt32(&p.capacity) - atomic.LoadInt32(&p.running));
	return idleGophers
}
func (p *Pool) Push(task fun) error {
	if len(p.destroy) > 0 {
		return nil
	}
	p.tasks.push(task)
	p.launchSignal <- Signal{}
	p.wg.Add(1)
	return nil
}

func (p *Pool) Capacity() int {
	return int(atomic.LoadInt32(&p.capacity))
}

func (p *Pool) Destroy() error {
	p.mx.Lock()
	defer p.mx.Unlock()
	for i:=0; i < runtime.GOMAXPROCS(-1) +1; i++ {
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
	defer atomic.AddInt32(&p.running,1);
	if w := p.workers.pop(); w != nil {
		return w.(*Worker)
	}
	return p.newWorker()
}

func (p *Pool) PutWorker(worker *Worker) {
	p.workers.push(worker)
	if p.reachLimit() {
		p.freeSignal <- Signal{}
	}
}

 func NewDLL() *DLL {
	q := new(DLL)
	q.doublylinkedlist = list.New()
	return q
}
 func (q *DLL) push(v interface{}) {
	defer q.m.Unlock()
	q.m.Lock()
	q.doublylinkedlist.PushFront(v)
}
 func (q *DLL) pop() interface{} {
	defer q.m.Unlock()
	q.m.Lock()
	if elem := q.doublylinkedlist.Back(); elem != nil{
		return q.doublylinkedlist.Remove(elem)
	}
	return nil
}

func (p *Pool) log(message string) {
	if p.logFunc != nil {
		p.logFunc(message)
	}
}
