package async

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
)

type poller struct {
	rdb         *redis.Client
	done        chan struct{}
	avgInterval time.Duration
	zsets       []string
}

func (p *poller) start() {
	// ticker := time.NewTicker(p.avgInterval)
	go func() {
		for {
			select {
			// case <-ticker.C:
			// 	p.enqueue()
			case <-p.done:
				p.shutdown()
			default:
				p.enqueue()
				time.Sleep(p.avgInterval)
			}
		}
	}()
}

func (p *poller) enqueue() {
	for _, zset := range p.zsets {
		now := time.Now().Unix()
		fmt.Printf("[DEBUG] polling ZSET %q at %d\n", zset, now)
		jobs, err := p.rdb.ZRangeByScore(zset, &redis.ZRangeBy{Min: "-inf", Max: strconv.Itoa(int(now))}).Result()
		fmt.Printf("len(jobs) = %d\n", len(jobs))
		if err != nil {
			fmt.Printf("redis command ZRANGEBYSCORE failed: %v\n", err)
			continue
		}
		if len(jobs) == 0 {
			fmt.Println("jobs empty")
			continue
		}
		for _, j := range jobs {
			fmt.Printf("[DEBUG] j=%v\n", j)
			var msg taskMessage
			err = json.Unmarshal([]byte(j), &msg)
			if err != nil {
				fmt.Println("unmarshal failed")
				continue
			}
			fmt.Println("[debug] ZREM")
			if p.rdb.ZRem(zset, j).Val() > 0 {
				err = push(p.rdb, &msg)
				if err != nil {
					log.Printf("could not push task to queue %q: %v", msg.Queue, err)
					// TODO(vinh): Handle this error properly. Add back to scheduled ZSET?
					continue
				}
			}
		}
	}
}

func (p *poller) shutdown() {
	// TODO(vinh): implement this. Gracefully shutdown all active goroutines.
	fmt.Println("-------------[Poller]---------------")
	// Stop the ticker
	// wait for active goroutines to finish
	// close the done channel
	// log shutdown complete
	fmt.Println("Poller shutting down...")
	fmt.Println("------------------------------------")

}
