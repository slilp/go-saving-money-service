package redisq

import (
	"context"
	"time"
)

// Worker processes jobs from a queue.
type RedisWorker interface {
	Register(queueName string, handler func(c context.Context, payload []byte) error)
	Start()
	Shutdown()
}

// Producer sends jobs to a queue. User is responsible of encoding message to bytes.
type RedisProducer interface {
	// Enqueue a job to run ASAP
	Enqueue(jobType string, payload []byte, retry int) (*string, error)
	// Schedule for later processing
	Schedule(jobType string, payload []byte, t time.Time) error
	Shutdown() error
}
