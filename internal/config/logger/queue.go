package logger

import (
	"sync/atomic"
)

type logOperation struct {
	level   string
	message string
}

var (
	queue       chan logOperation
	queueClosed atomic.Bool
)

func init() {
	queue = make(chan logOperation, 100)
}

func processQueue() {
	for op := range queue {
		sublogWrapper(op.level, op.message)
	}
}
