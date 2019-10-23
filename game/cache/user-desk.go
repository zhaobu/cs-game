package cache

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

const (
	UserDeskList = "UserDeskList"
)

func GetUserDesk(uId uint64) (err error, data []uint64) {
	data = make([]uint64, 0)
	c := redisPool.Get()
	defer c.Close()

	movies, err := redis.String(c.Do("HGET", UserDeskList, uId))
	if err != nil {
		return fmt.Errorf("GetUserDesk 读取缓存哈希表数据失败 key = %d", uId), data
	}
	err = json.Unmarshal([]byte(movies), &data)
	if err != nil {
		return fmt.Errorf("读取Redis Map【%s】【%d】错误 %s", UserDeskList, uId, err.Error()), data
	}
	return
}

func AddUserDesk(uId uint64, deskId uint64) (err error) {
	c := redisPool.Get()
	defer c.Close()
	_, data := GetUserDesk(uId)
	data = append(data, deskId)
	_Value := map[string][]uint64{
		fmt.Sprintf("%d", uId): data,
	}
	Values := []interface{}{}
	Values = append(Values, UserDeskList)
	for k, v := range _Value {
		if b, err := json.Marshal(v); err == nil {
			Values = append(Values, k, string(b))
		}
	}
	_, err = c.Do("HSET", Values...)
	if err != nil {
		return fmt.Errorf("设置Redis Map【%s】【%d】错误 %s", UserDeskList, uId, err.Error())
	}
	return err
}

func DelUserDesk(uId uint64, deskId uint64) (err error) {
	c := redisPool.Get()
	defer c.Close()
	_, data := GetUserDesk(uId)
	for i, v := range data {
		if v == deskId {
			data = append(data[0:i], data[i+1:]...)
		}
	}
	_Value := map[string][]uint64{
		fmt.Sprintf("%d", uId): data,
	}
	Values := []interface{}{}
	Values = append(Values, UserDeskList)
	for k, v := range _Value {
		if b, err := json.Marshal(v); err == nil {
			Values = append(Values, k, string(b))
		}
	}
	_, err = c.Do("HSET", Values...)
	if err != nil {
		return fmt.Errorf("设置Redis Map【%s】【%d】错误 %s", UserDeskList, uId, err.Error())
	}
	return err
}
