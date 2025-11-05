package async

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/google/uuid"
)

const (
	queuePrefix = "asynq:queues:"
	scheduled   = "asynq:scheduled"
)

type Client struct {
	rdb *redis.Client
}

type Task struct {
	Handler string
	Args    []interface{}
}

type deleyedTask struct {
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
	deleyed := &deleyedTask{
		ID:    uuid.New().String(),
		Queue: queue,
		Task:  task,
	}
	bytes, err := json.Marshal(deleyed)
	if err != nil {
		return err
	}
	return c.rdb.ZAdd(scheduled, &redis.Z{
		Score:  float64(executeAt),
		Member: string(bytes),
	}).Err()
}
