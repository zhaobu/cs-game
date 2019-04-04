package configs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type AllConfig struct {
	ConsulAddr string `json:"consulAddr"`
	Release    bool   `json:"release"`
	RedisAddr  string `json:"redisAddr"`
	RedisDb    int    `json:"redisDb"`
	MgoURI     string `json:"mgoURI"`
}

func GetConfig(filename string) *AllConfig {
	var tmp AllConfig

	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("ReadFile err = ", err)
		return nil
	}

	err = json.Unmarshal(data, &tmp)
	if err != nil {
		fmt.Println("json.Unmarshal err = ", err)
		return nil
	}
	return &tmp
}
