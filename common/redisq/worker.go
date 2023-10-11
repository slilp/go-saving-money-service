package redisq

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	"github.com/hibiken/asynq"
)

func NewWorker(conf WorkerConfig) RedisWorker {
	var logLevel asynq.LogLevel
	logLevel.Set(conf.LogLevel)

	queues := GetQueueNames(conf.Queues)

	queueConf := asynq.Config{
		Concurrency:              conf.Concurrency,
		StrictPriority:           conf.StrictPriority,
		ShutdownTimeout:          conf.ShutdownTimeout,
		HealthCheckInterval:      conf.HealthCheckInterval,
		DelayedTaskCheckInterval: conf.DelayedTaskCheckInterval,
		GroupGracePeriod:         conf.GroupGracePeriod,
		GroupMaxDelay:            conf.GroupMaxDelay,
		GroupMaxSize:             conf.GroupMaxSize,
		// Logger:                   monitoring.Logger().Sugar(),
		LogLevel: logLevel,
		Queues:   queues,
	}

	// TODO cluster option
	// asynq api has pretty poor DX

	redisOpts := asynq.RedisClusterClientOpt{
		Addrs: strings.Split(conf.Addr, ","),
		// Username:     opt.Username,
		Password: conf.Password,
		// DB:           opt.DB,
		DialTimeout:  conf.DialTimeout,
		ReadTimeout:  conf.ReadTimeout,
		WriteTimeout: conf.WriteTimeout,
		// PoolSize:     opt.PoolSize,
	}

	srv := asynq.NewServer(
		redisOpts,
		queueConf,
	)

	w := &Worker{
		handlers:    map[string]func(context.Context, *asynq.Task) error{},
		asyncServer: srv,
	}

	return w
}

type Worker struct {
	handlers    map[string]func(context.Context, *asynq.Task) error
	asyncServer *asynq.Server
}

func (w *Worker) Register(queueName string, handler func(c context.Context, payload []byte) error) {
	w.handlers[queueName] = func(c context.Context, t *asynq.Task) error {
		err := handler(c, t.Payload())
		return err
	}
}

func (w *Worker) Start() {
	// log := monitoring.Logger()
	// mux maps a type to a handler
	mux := asynq.NewServeMux()

	for pattern, handler := range w.handlers {
		// log.Info("registered Redis workers: ", zap.Any("name", pattern))
		mux.HandleFunc(pattern, handler)
	}

	// log.Info("starting Redis workers...")

	w.asyncServer.Start(mux)

	// if err := srv.Run(mux); err != nil {
	// 	log.Fatalf("could not run server: %v", err)
	// }
}

func (w *Worker) Shutdown() {
	w.asyncServer.Shutdown()
}

func GetQueueNames(jobs2queues map[string]string) map[string]int {
	// make a set of unique values from jobs2queues
	queues := make(map[string]int)
	for _, queue := range jobs2queues {
		// we don't support priority at the moment, all queues are equal
		queues[queue] = 1
	}
	return queues
}

type ClientConfig struct {
	// // Network type to use, either tcp or unix.
	// // Default is tcp.
	// Network string

	Addr string `envconfig:"REDIS_URL" required:"true"`

	// Username to authenticate the current connection when Redis ACLs are used.
	// See: https://redis.io/commands/auth.
	Username string `envconfig:"REDIS_USERNAME"`

	// Password to authenticate the current connection.
	// See: https://redis.io/commands/auth.
	Password string `envconfig:"REDIS_PASSWORD"`

	// Redis DB to select after connecting to a server.
	// See: https://redis.io/commands/select.
	DB int `envconfig:"REDIS_DB" default:"0"`

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration `envconfig:"REDIS_DIAL_TIMEOUT" default:"5s"`

	// Timeout for socket reads.
	// If timeout is reached, read commands will fail with a timeout error
	// instead of blocking.
	//
	// Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration `envconfig:"REDIS_READ_TIMEOUT" default:"3s"`

	// Timeout for socket writes.
	// If timeout is reached, write commands will fail with a timeout error
	// instead of blocking.
	//
	// Use value -1 for no timeout and 0 for default.
	// Default is ReadTimout.
	WriteTimeout time.Duration `envconfig:"REDIS_WRITE_TIMEOUT" default:"3s"`

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int `envconfig:"REDIS_POOL_SIZE" default:"10"`

	// TLS Config used to connect to a server.
	// TLS will be negotiated only if this field is set.
	TLSConfig *tls.Config

	// Mapping of jobs and queues
	// key is the name of the task type, and value is the name of the queue
	Queues map[string]string `envconfig:"REDIS_QUEUES" required:"true"`
}

type WorkerConfig struct {
	ClientConfig

	// Maximum number of concurrent processing of tasks.
	//
	// If set to a zero or negative value, NewServer will overwrite the value
	// to the number of CPUs usable by the current process.
	Concurrency int

	// StrictPriority indicates whether the queue priority should be treated strictly.
	//
	// If set to true, tasks in the queue with the highest priority is processed first.
	// The tasks in lower priority queues are processed only when those queues with
	// higher priorities are empty.
	StrictPriority bool

	// LogLevel specifies the minimum log level to enable.
	//
	// If unset, InfoLevel is used by default.
	LogLevel string `envconfig:"REDIS_QUEUES_LOG_LEVEL" default:"info"`

	// ShutdownTimeout specifies the duration to wait to let workers finish their tasks
	// before forcing them to abort when stopping the server.
	//
	// If unset or zero, default timeout of 8 seconds is used.
	ShutdownTimeout time.Duration `envconfig:"REDIS_QUEUES_SHUTDOWN_TIMEOUT" default:"8s"`

	// HealthCheckInterval specifies the interval between healthchecks.
	HealthCheckInterval time.Duration `envconfig:"REDIS_QUEUES_SHUTDOWN_TIMEOUT" default:"15s"`

	// DelayedTaskCheckInterval specifies the interval between checks run on 'scheduled' and 'retry'
	// tasks, and forwarding them to 'pending' state if they are ready to be processed.
	DelayedTaskCheckInterval time.Duration `envconfig:"REDIS_QUEUES_DELAYED_TASK_CHECK_INTERVAL" default:"5s"`

	// GroupGracePeriod specifies the amount of time the server will wait for an incoming task before aggregating
	// the tasks in a group. If an incoming task is received within this period, the server will wait for another
	// period of the same length, up to GroupMaxDelay if specified.
	//
	// If unset or zero, the grace period is set to 1 minute.
	// Minimum duration for GroupGracePeriod is 1 second. If value specified is less than a second, the call to
	// NewServer will panic.
	GroupGracePeriod time.Duration `envconfig:"REDIS_QUEUES_GROUP_GRACE_PERIOD" default:"1m"`

	// GroupMaxDelay specifies the maximum amount of time the server will wait for incoming tasks before aggregating
	// the tasks in a group.
	GroupMaxDelay time.Duration `envconfig:"REDIS_QUEUES_GROUP_MAX_DELAY" default:"0"`

	// GroupMaxSize specifies the maximum number of tasks that can be aggregated into a single task within a group.
	// If GroupMaxSize is reached, the server will aggregate the tasks into one immediately.
	//
	// If unset or zero, no size limit is used.
	GroupMaxSize int `envconfig:"REDIS_QUEUES_GROUP_MAX_SIZE" default:"0"`
}
