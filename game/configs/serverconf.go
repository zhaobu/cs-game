package configs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// 隐式导入
var (
	Conf *allConfig
)

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

type gameNodeConf struct {
	nodeConf
	TplName  string `json:"TplName"`  //默认建房参数
	GameName string `json:"GameName"` //游戏名称
}
type GameConfs struct {
	MjLibPath string                  `json:"MjLibPath"` //胡牌表加载路径
	GameNode  map[string]gameNodeConf `json:"GameNode"`
}

type allConfig struct {
	ClubConf    *nodeConf `json:"ClubConf"`
	CenterConf  *nodeConf `json:"CenterConf"`
	GateConf    *nodeConf `json:"GateConf"`
	*GlobalConf `json:"GlobalConf"`
	*GameConfs  `json:"GameConfs"`
}

func init() {
	// 初始所有指针数据
	Conf = &allConfig{
		ClubConf:   new(nodeConf),
		CenterConf: new(nodeConf),
		GateConf:   new(nodeConf),
		GlobalConf: new(GlobalConf),
		GameConfs:  new(GameConfs),
	}

	data, err := ioutil.ReadFile("./configs/globalconf.json")
	if err != nil {
		fmt.Println("ReadFile err = ", err)
		return
	}

	err = json.Unmarshal(data, Conf)
	if err != nil {
		fmt.Println("json.Unmarshal err = ", err)
		return
	}
	return
}
