package cache

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

func MarkCaptcha(mobile string, captcha string) error {
	c := redisPool.Get()
	defer c.Close()

	key := fmt.Sprintf("captcha:%s", mobile)

	_, err := c.Do("SETEX", key, 55, captcha)
	return err
}

func DeleteCaptcha(mobile string) error {
	c := redisPool.Get()
	defer c.Close()

	key := fmt.Sprintf("captcha:%s", mobile)
	_, err := c.Do("DEL", key)
	return err
}

func GetCaptcha(mobile string) (string, error) {
	c := redisPool.Get()
	defer c.Close()

	key := fmt.Sprintf("captcha:%s", mobile)

	reply, err := redis.Bytes(c.Do("GET", key))
	return string(reply), err
}
