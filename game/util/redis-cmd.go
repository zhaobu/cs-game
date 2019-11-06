package util

import (
	"github.com/go-redis/redis"
)

// func Subscribe(addr string, db int, channels string, onMessage func(channel string, data []byte) error) error {
// 	//订阅给定的一个或多个频道的信息
// 	psc := redis.NewClient(&redis.Options{
// 		Addr:     addr,
// 		Password: "", // no password set
// 		DB:       db, // use default DB
// 	}).Subscribe(channels)

// 	for {
// 		msg, err := psc.ReceiveMessage()
// 		if err != nil {
// 			return err
// 		}
// 		onMessage(msg.Channel, []byte(msg.Payload)) //获取回复信息
// 	}
// }

//使用redsi5.0的stream方式读取消息
func RedisXread(addr string, db int, channels string, onMessage func(channel string, msgData []byte) error) error {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       db, // use default DB
	})

	lastID := "$"
	for {
		items, err := c.XRead(&redis.XReadArgs{
			Count:   5, //每次读取处理的消息数量
			Streams: []string{channels, lastID},
			Block:   0, // Wait for new messages without a timeout.
		}).Result()
		if err != nil {
			return err
		}
		for _, msg := range items[0].Messages {
			//处理每一条消息
			onMessage(items[0].Stream, []byte(msg.Values[""].(string)))
			lastID = msg.ID
		}
	}
}

func RedisXadd(c *redis.Client, channels string, msgData []byte) (string, error) {
	return c.XAdd(&redis.XAddArgs{
		Stream:       channels,
		Values:       map[string]interface{}{"": msgData},
		MaxLenApprox: 10, //设置stream保存消息的上限,>=MaxLenApprox,但是不会大很多
	}).Result()
}
