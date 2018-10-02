package poolboy

import (
	"math"
	"sync"
	"time"
)

type PoolFunc interface {
	// API
	Running() int
	Free() int
	Push(fn func()) error
	Capacity() int
	Destroy() int
	Wait()

	// Internals
	loop()
	reachLimit() bool
	newWorker() *Worker
	getWorker() *Worker
	log(message string)
}

type Pool struct {
	tasks      chan fun
	workers    chan *Worker
	destroy    chan Signal
	statTime   time.Duration // time to display stats.
	mx         *sync.Mutex
	wg         *sync.WaitGroup
	freeSignal chan Signal
	capacity   int32
	running    int32
	stopped    int32
	logFunc    func(message string)
}

var instance *Pool

func GetNewPool(size int32, statTime time.Duration, logFunc func(message string)) (*Pool, error) {
	if size <= 0 {
		return nil, ErrorInvalidMinRoutines
	}

	if statTime < time.Millisecond {
		return nil, ErrorInvalidStatTime
	}

	p := &Pool{
		capacity:   size,
		tasks:      make(chan fun, size),
		workers:    make(chan *Worker, size),
		freeSignal: make(chan Signal, math.MaxInt32),
		destroy:    make(chan Signal),
		statTime:   statTime,
		wg:         &sync.WaitGroup{},
		logFunc:    logFunc,
		running:    0,
		stopped:    0,
	}
	p.loop()
	return p, nil
}

func NewPool(size int32, statTime time.Duration, logFunc func(message string)) (*Pool, error) {
	if instance == nil {
		instance, _ = GetNewPool(size, statTime, logFunc)
	}

	return instance, nil
}
