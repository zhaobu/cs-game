package core

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/BurntSushi/toml"
)

// RobotConfig 机器人配置
type RobotConfig struct {
	LogLevel           int     `toml:"log_level"`
	JoinIntervalBase   int     `toml:"join_interval_base"`
	JoinIntervalRand   int     `toml:"join_interval_rand"`
	JoinInterverTimes  []int   `toml:"join_interver_times"`
	JoinInterverGrades []int   `toml:"join_interver_grades"`
	PrepareInterval    []int   `toml:"prepare_interval"`
	WinInterval        int     `toml:"win_interval"`
	LackInterval       []int   `toml:"lack_interval"`
	ReplyInterval      [][]int `toml:"reply_interval"`
	ChatRate           int     `toml:"chat_rate"`
	ChatInterval       int     `toml:"chat_interval"`
	ChatCheckInterval  int     `toml:"chat_check_interval"`
	ChatIds            []int   `toml:"chat_ids"`
	UrgeRate           int     `toml:"urge_rate"`
	UrgeInterval       int     `toml:"urge_interval"`
	UrgeChatId         int     `toml:"urge_chat_id"`
	NetworkInterval    []int   `toml:"network_interval"`
}

var (
	// RobotCfg 机器人配置
	RobotCfg *RobotConfig
)

func init() {
	RobotCfg = &RobotConfig{}
}

// LoadRobotConfig 载入机器人配置
func LoadRobotConfig(cfgFile string) {
	go func() {
		for {
			loadRobotConfigFile(cfgFile)
			time.Sleep(time.Minute)
			// time.Sleep(time.Second)
		}
	}()
}

func loadRobotConfigFile(cfgFile string) {
	content, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		fmt.Println("loadRobotConfigFile error:%v", err.Error())
		return
	}
	if _, err := toml.Decode(string(content), &RobotCfg); err != nil {
		fmt.Println("loadRobotConfigFile error:%v", err.Error())
		return
	}

	fmt.Println(fmt.Sprintf("loadRobotConfigFile completed:%#v", *RobotCfg))
}
