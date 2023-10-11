package redisq

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

type Producer struct {
	client *asynq.Client
	// key is the name of the task type, and value is the name of the queue
	queueNames map[string]string
}

func NewProducer() RedisProducer {
	nodes := strings.Split("config.Addr", ",")
	client := asynq.NewClient(asynq.RedisClusterClientOpt{
		Addrs: nodes,
		// DialTimeout:  config.DialTimeout,
		// ReadTimeout:  config.ReadTimeout,
		// WriteTimeout: config.WriteTimeout,
	})
	return NewProducerWithClient(client, map[string]string{"sss": "sss"})
	// config.Queues
}

func NewProducerWithClient(client *asynq.Client, queueNames map[string]string) RedisProducer {
	producer := &Producer{client, queueNames}
	return producer
}

func (p *Producer) Enqueue(taskType string, payload []byte, retry int) (*string, error) {
	// log := monitoring.Logger()

	task := asynq.NewTask(taskType, payload)

	queueName, found := p.queueNames[taskType]
	if !found {
		return nil, errors.New("no queue set for task type: " + taskType)
	}
	info, err := p.client.EnqueueContext(context.Background(), task, asynq.Queue(queueName), asynq.MaxRetry(retry))
	if err != nil {
		// log.Error("could not enqueue task", zap.Error(err), zap.Any("task", task))
		return nil, fmt.Errorf("could not enqueue task: %w", err)
	}
	// log.Debug("enqueued task", zap.Any("queue", info.Queue), zap.Any("task", info.Type), zap.String("taskID", info.ID))
	return &info.ID, nil
}

func (p *Producer) Schedule(taskType string, payload []byte, t time.Time) error {
	// log := monitoring.Logger()

	queueName, found := p.queueNames[taskType]
	if !found {
		return errors.New("no queue set for task type: " + taskType)
	}

	task := asynq.NewTask(taskType, payload)

	_, err := p.client.Enqueue(task, asynq.ProcessAt(t), asynq.Queue(queueName))
	// info, err := p.client.Enqueue(task, asynq.ProcessAt(t), asynq.Queue(queueName))

	if err != nil {
		// log.Error("could not schedule task", zap.Error(err), zap.Any("task", task))
		return fmt.Errorf("could not schedule task: %w", err)
	}
	// log.Debug("enqueued task", zap.String("taskID", info.ID), zap.Any("queue", info.Queue))

	return nil
}

func (p *Producer) Shutdown() error {
	return p.client.Close()
}
