package cache

import (
	"fmt"
	"time"
)

func MarkCaptcha(mobile string, captcha string) error {
	key := fmt.Sprintf("captcha:%s", mobile)
	_, err := redisCli.Set(key, captcha, time.Second*55).Result()
	return err
}

func DeleteCaptcha(mobile string) error {

	key := fmt.Sprintf("captcha:%s", mobile)
	_, err := redisCli.Del(key).Result()
	return err
}

func GetCaptcha(mobile string) (string, error) {

	key := fmt.Sprintf("captcha:%s", mobile)
	return redisCli.Get(key).Result()
}
