package util

import (
	"github.com/go-redis/redis"
)

func Subscribe(addr string, db int, channels string, onMessage func(channel string, data []byte) error) error {
	//订阅给定的一个或多个频道的信息
	psc := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       db, // use default DB
	}).Subscribe(channels)

	for {
		msg, err := psc.ReceiveMessage()
		if err != nil {
			return err
		}
		onMessage(msg.Channel, []byte(msg.Payload)) //获取回复信息
	}
}
