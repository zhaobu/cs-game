package mqredis

import (
	"fmt"
	"sync"

	"github.com/gomodule/redigo/redis"
)

type Redis struct {
	c    redis.Conn
	lkey string
	once sync.Once
}

func (r *Redis) Conn(network, address, lKey string) (err error) {
	r.once.Do(func() {
		r.c, err = redis.Dial(network, address)
		r.lkey = lKey
	})
	return err
}

func (r *Redis) Push(data []byte) (err error) {
	_, err = r.c.Do("lpush", r.lkey, data)
	return
}

func (r *Redis) Pop() ([]byte, error) {
	reply, err := redis.ByteSlices(r.c.Do("brpop", r.lkey, 1)) // wait 1s
	if err != nil {
		return nil, err
	}
	if len(reply) != 2 {
		return nil, fmt.Errorf("len(%d) err", len(reply))
	}
	if string(reply[0]) != r.lkey {
		return nil, fmt.Errorf("key(%s) err, want %s", reply[0], r.lkey)
	}
	return reply[1], nil
}
