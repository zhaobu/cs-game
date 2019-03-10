package core

import (
	"fmt"

	"mahjong.go/config"
)

func GetLang(langId int, argv ...interface{}) string {
	msg, exists := config.LangConfig[langId]
	if !exists {
		return fmt.Sprintf("语言配置项未找到:%d", langId)
	} else {
		return fmt.Sprintf(msg, argv...)
	}
}
