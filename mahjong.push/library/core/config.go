package core

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

var (
	AppConfig map[string]interface{}
)

func init() {
	AppConfig = make(map[string]interface{})
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
		Logger.Warn("%s not exists in app config")
		return nil
	}
}

// GetApnsP12File 读取apns p12文件路径
func GetApnsP12File() string {
	return GetAppConfig("apns_p12_file").(string)
}

// GetApnsP12Passwd 读取apns p12文件解析密码
func GetApnsP12Passwd() string {
	return GetAppConfig("apns_p12_passwd").(string)
}
