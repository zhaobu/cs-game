package util

import (
	"github.com/gomodule/redigo/redis"
)

func Subscribe(addr string, db int, channels string, onMessage func(channel string, data []byte) error) error {

	c, err := redis.Dial("tcp", addr, redis.DialDatabase(db))
	if err != nil {
		return err
	}

	psc := redis.PubSubConn{Conn: c}
	defer psc.Close()

	if err := psc.Subscribe(channels); err != nil { //订阅给定的一个或多个频道的信息
		return err
	}

	for {
		switch t := psc.Receive().(type) { //获取回复信息
		case redis.Message:
			onMessage(t.Channel, t.Data)
		case error:
			return t
		}
	}
}
