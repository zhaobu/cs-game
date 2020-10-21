package cache

import (
	"encoding/json"
	"fmt"
	"strconv"
)

const (
	UserDeskList = "UserDeskList"
)

func GetUserDesk(uId uint64) (err error, data []uint64) {
	data = make([]uint64, 0)
	movies, err := redisCli.HGet(UserDeskList, strconv.FormatUint(uId, 10)).Result()
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
	_, data := GetUserDesk(uId)
	data = append(data, deskId)
	_, err = redisCli.HSet(UserDeskList, fmt.Sprintf("%d", uId), data).Result()
	if err != nil {
		return fmt.Errorf("设置Redis Map【%s】【%d】错误 %s", UserDeskList, uId, err.Error())
	}
	return err
}

func DelUserDesk(uId uint64, deskId uint64) (err error) {
	_, data := GetUserDesk(uId)
	for i, v := range data {
		if v == deskId {
			data = append(data[0:i], data[i+1:]...)
		}
	}

	_, err = redisCli.HSet(UserDeskList, fmt.Sprintf("%d", uId), data).Result()
	if err != nil {
		return fmt.Errorf("设置Redis Map【%s】【%d】错误 %s", UserDeskList, uId, err.Error())
	}
	return err
}
