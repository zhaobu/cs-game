package sensitive

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"mahjong.club/core"
)

// Replace 屏蔽词过滤
func Replace(content string) string {
	sensitiveURL := core.GetAppConfig("sensitive_url").(string)
	if sensitiveURL == "" {
		core.Logger.Warn("sensitive_url未配置，返回原字符串")
		return content
	}

	resp, err := http.Post(sensitiveURL, "application/x-www-form-urlencoded",
		strings.NewReader(fmt.Sprintf("s=%s", content)))
	if err != nil {
		core.Logger.Warn("屏蔽词过滤失败,content:%v,err:%v", content, err.Error())
		return content
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		core.Logger.Warn("屏蔽词过滤失败,content:%v,err:%v", content, err.Error())
		return content
	}
	return string(body)
}
