package async

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v7"
)

type manager struct {
	rdb     *redis.Client
	handler TaskHandler
	sema    chan struct{}
	done    chan struct{}
}

func newManager(rdb *redis.Client, numWorkers int, handler TaskHandler) *manager {
	return &manager{
		rdb:     rdb,
		handler: handler,
		sema:    make(chan struct{}, numWorkers),
		done:    make(chan struct{}),
	}
}

func (m *manager) terminate() {
	m.done <- struct{}{}
}

func (m *manager) start() {
	go func() {
		for {
			select {
			case <-m.done:
				m.shutdown()
			default:
				m.processTasks()
			}
		}
	}()
}

func (m *manager) processTasks() {
	res, err := m.rdb.BLPop(5*time.Second, listQueues(m.rdb)...).Result()
	if err != nil {
		if err != redis.Nil {
			log.Printf("redis BLPop failed: %v\n", err)
		}
		return
	}
	q, data := res[0], res[1]
	fmt.Printf("perform task %v from %s\n", data, q)
	var msg taskMessage
	err = json.Unmarshal([]byte(data), &msg)
	if err != nil {
		log.Printf("failed to unmarshal task message: %v\n", err)
		return
	}
	t := &Task{Type: msg.Type, Payload: msg.Payload}
	m.sema <- struct{}{} // acquire semaphore locks
	go func(task *Task) {
		defer func() { <-m.sema }() // release semaphore locks
		err := m.handler(task)
		if err != nil {
			log.Printf("task handler error: %v", err)
			if msg.Retried >= msg.Retry {
				fmt.Println("Retry exhausted!!!")
			}
		}
		fmt.Println("RETRY!!!")
		retryAt := time.Now().Add(delaySeconds((msg.Retried)))
		fmt.Printf("[DEBUG] retying the task in %v\n", retryAt.Sub(time.Now()))
		msg.Retried++
		msg.ErrorMsg = err.Error()
		if err := zadd(m.rdb, retry, float64(retryAt.Unix()), &msg); err != nil {
			// TODO(vinh): finding way to ensure how to handle this error
			log.Printf("[SEVERE ERROR] could not add msg %+v to 'retry' set: %v\n", msg, err)
			return
		}
	}(t)
}

func (m *manager) shutdown() {
	// TODO(vinh): implement this. Gracefully shutdown all active goroutines.
	fmt.Println("-------------[Manager]---------------")
	fmt.Println("Manager shutting down...")
	fmt.Println("------------------------------------")
}
