package core

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

var (
	AppConfig map[string]interface{}
)

func init() {
	AppConfig = make(map[string]interface{})
}

// 读取配置文件路径
func GetConfigFile(filename, env, confDir string) string {
	return fmt.Sprintf("%s/%s/%s", confDir, env, filename)
}

func LoadAppConfig(cfg_file string) {
	content, err := ioutil.ReadFile(cfg_file)
	if err != nil {
		panic(err)
	}
	if _, err := toml.Decode(string(content), &AppConfig); err != nil {
		panic(err)
	}
}

func GetAppConfig(key string) interface{} {
	if value, ok := AppConfig[key]; ok {
		return value
	} else {
		Logger.Warn("[GetAppConfig]%s not exists in app config")
		return nil
	}
}

// 判断是否本地环境
func IsLocal() bool {
	if value, ok := AppConfig["env"]; ok && value == "local" {
		return true
	}

	return false
}

// 判断是否本地环境
func IsProduct() bool {
	if value, ok := AppConfig["env"]; ok && value == "product" {
		return true
	}

	return false
}

// 判断是否送审环境
// 暂时放在qa上
func IsForReview() bool {
	if value, ok := AppConfig["env"]; ok && value == "qa" {
		return true
	}

	return false
}
