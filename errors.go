package poolboy

import "errors"

var (
	ErrorInvalidMinRoutines = errors.New("Invalid minimum number of routines")
	ErrorInvalidStatTime    = errors.New("Invalid duration for stat time")
	ErrNoTaskInQ            = errors.New("No task to run")
	ErrCapacity             = errors.New("Thread Pool At Capacity")
)
