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
	}
	Logger.Warn("[GetAppConfig]%s not exists in app config")
	return nil
}

// LoadConfig 读取配置文件内容
func LoadConfig(cfgFile string) (map[string]interface{}, error) {
	content, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}
	data := make(map[string]interface{})
	if _, err := toml.Decode(string(content), &data); err != nil {
		return nil, err
	}
	return data, nil
}
