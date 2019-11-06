package cache

import "github.com/go-redis/redis"

func RedisXadd(channels, msgName string, msgData []byte) (string, error) {
	return redisCli.XAdd(&redis.XAddArgs{
		Stream:       channels,
		Values:       map[string]interface{}{msgName: msgData},
		MaxLenApprox: 10, //设置stream保存消息的上限,>=MaxLenApprox,但是不会大很多
	}).Result()
}
