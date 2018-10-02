package poolboy

import (
	"math"
	"time"
)

const (
	DefaultRoutinePoolSize = math.MaxInt32
	DefaultWorkerPurgeTime = 30 * time.Second
)
