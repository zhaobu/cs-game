package core

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

// AppConfig 项目配置
type AppConfig struct {
	Env               string   `toml:"env"`
	SelectServerUrl   string   `toml:"select_server_url"`
	SelectMyServerUrl string   `toml:"get_my_server_url"`
	RewardsNotifyUrl  string   `toml:"rewards_notify_url"`
	GatewayRemotes    []string `toml:"gateway_remotes"`
}

// 项目配置
var cfg *AppConfig

func init() {
	cfg = &AppConfig{}
}

// GetConfigFile 读取配置文件路径
func GetConfigFile(filename, env, confDir string) string {
	return fmt.Sprintf("%s/env/%s/%s", confDir, env, filename)
}

// GetSharedConfigFile 读取共享配置文件路径
func GetSharedConfigFile(filename, confDir string) string {
	return fmt.Sprintf("%s/shared/%s", confDir, filename)
}

// LoadAppConfig 加载app配置
func LoadAppConfig(cfgFile string) {
	content, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		panic(err)
	}
	if _, err := toml.Decode(string(content), &cfg); err != nil {
		panic(err)
	}
}

// GetAppConfig 读取app配置
func GetAppConfig() *AppConfig {
	return cfg
}

// IsLocal 是否本地环境
func IsLocal() bool {
	return cfg.Env == "local"
}

// IsTest 是否测试环境
func IsTest() bool {
	return cfg.Env == "qa"
}

// IsProduct 是否生产环境
func IsProduct() bool {
	return cfg.Env == "product"
}
