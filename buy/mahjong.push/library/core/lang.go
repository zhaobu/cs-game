package core

import (
	"fmt"

	"strings"

	"mahjong.push/config"
)

func GetLang(langId int, argv ...interface{}) string {
	msg, exists := config.LangConfig[langId]
	if !exists {
		Logger.Error("语言包配置未找到:%d", langId)
		return fmt.Sprintf("未知语言包:%d", langId)
	}

	// 解析需要替换的参数个数
	cnt := strings.Count(msg, "%v")
	if cnt == 0 {
		return msg
	} else if cnt > len(argv) {
		Logger.Error("语言包参数个数配置错误:%d", langId)
		return fmt.Sprintf("未知的语言包:%d", langId)
	}

	return fmt.Sprintf(msg, argv[:cnt]...)
}
