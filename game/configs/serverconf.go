package configs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type MjHuLibConfig struct {
	MjPath string `json:"mjPath"`
}

//公共配置
type GlobalConf struct {
	ConsulAddr string `json:"ConsulAddr"`
	Release    bool   `json:"Release"`
	RedisAddr  string `json:"RedisAddr"`
	RedisDb    int    `json:"RedisDb"`
	MgoURI     string `json:"MgoURI"`
}

//节点配置
type nodeConf struct {
	Addr string `json:"Addr"`
}

type AllConfig struct {
	ClubConf     nodeConf `json:"ClubConf"`
	CenterConf   nodeConf `json:"CenterConf"`
	GateConf     nodeConf `json:"GateConf"`
	ChangShuConf nodeConf `json:"ChangShuConf"`
	GlobalConf   `json:"GlobalConf"`
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
