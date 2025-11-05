package async

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
)

const (
	queuePrefix = "asynq:queues:"
	scheduled   = "asynq:scheduled"
	allQueues   = "asynq:queues"
)

type Client struct {
	rdb *redis.Client
}

type Task struct {
	Handler string
	Args    []interface{}
}

type delayedTask struct {
	ID    string
	Queue string
	Task  *Task
}

type RedisOpt struct {
	Addr     string
	Password string
}

func NewClient(opt *RedisOpt) *Client {
	rdb := redis.NewClient(&redis.Options{Addr: opt.Addr, Password: opt.Password})
	return &Client{rdb: rdb}
}

func (c *Client) Enqueue(queue string, task *Task, delay time.Duration) error {
	if delay == 0 {
		bytes, err := json.Marshal(task)
		if err != nil {
			return err
		}
		return c.rdb.RPush(queuePrefix+queue, string(bytes)).Err()
	}
	executeAt := time.Now().Add(delay).Unix()
	delayed := &delayedTask{
		ID:    uuid.New().String(),
		Queue: queue,
		Task:  task,
	}
	bytes, err := json.Marshal(delayed)
	if err != nil {
		return err
	}
	return c.rdb.ZAdd(scheduled, &redis.Z{
		Score:  float64(executeAt),
		Member: string(bytes),
	}).Err()
}

//-------------------- Workers --------------------

type Workers struct {
	rdb        *redis.Client
	poolTokens chan struct{}
	handlers   map[string]func(msg string)
}

func NewWorkers(poolSize int, opt *RedisOpt) *Workers {
	rdb := redis.NewClient(&redis.Options{Addr: opt.Addr, Password: opt.Password})
	return &Workers{
		rdb:        rdb,
		poolTokens: make(chan struct{}, poolSize),
		handlers:   make(map[string]func(string)),
	}
}

// Handle registers a handler function for a given queue.
func (w *Workers) Handle(q string, fn func(msg string)) {
	w.handlers[q] = fn
}

// Run starts the workers and scheduler.
func (w *Workers) Run() {
	go w.pollScheduledTasks()
	for {
		res, err := w.rdb.BLPop(0, "asynq:queues:test").Result()
		if err != nil {
			if err != redis.Nil {
				log.Printf("error when BLPOP from %s: %v\n", "aysnq:queues:test", err)
			}
			continue
		}
		q, msg := res[0], res[1]
		fmt.Printf("perform task %v from %s\n", msg, q)
		handler, ok := w.handlers[strings.TrimPrefix(q, queuePrefix)]
		if !ok {
			log.Printf("no handler found for queue %q\n", strings.TrimPrefix(q, queuePrefix))
			continue
		}
		w.poolTokens <- struct{}{}
		go func(msg string) {
			handler(msg)
			<-w.poolTokens
		}(msg)
	}
}

func (w *Workers) pollScheduledTasks() {
	for {
		now := time.Now().Unix()
		jobs, err := w.rdb.ZRangeByScore(scheduled,
			&redis.ZRangeBy{
				Min: "-inf",
				Max: strconv.Itoa(int(now)),
			}).Result()
		fmt.Printf("len(jobs) = %d\n", len(jobs))
		if err != nil {
			log.Printf("radis command ZRANGEBYSCORE failed: %v\n", err)
			continue
		}
		if len(jobs) == 0 {
			fmt.Println("jobs empty")
			time.Sleep(5 * time.Second)
			continue
		}

		for _, j := range jobs {
			var job delayedTask
			err = json.Unmarshal([]byte(j), &job)
			if err != nil {
				fmt.Println("unmarshal failed")
				continue
			}

			// TODO(hibiken): Acquire lock for job.ID
			pipe := w.rdb.TxPipeline()
			pipe.ZRem(scheduled, j)
			// Do we need to encode this again?
			// Can we skip this entirely by defining Task field to be a string field?
			bytes, err := json.Marshal(job.Task)
			if err != nil {
				log.Printf("could not marshal job.Task %v: %v\n", job.Task, err)
				pipe.Discard()
				continue
			}
			pipe.RPush(queuePrefix+job.Queue, string(bytes))
			_, err = pipe.Exec()
			if err != nil {
				log.Printf("could not execute pipeline: %v\n", err)
				continue
			}
			// TODO(hibiken): Release lock for job.ID
		}
	}
}
