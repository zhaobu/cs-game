package gateway

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"mahjong.go/library/core"
)

// HTTPPost 发送一个http请求
func httpPost(url string, b []byte) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(b))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		core.Logger.Error("[gateway.HTTPPost]url:%v", url)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	core.Logger.Info("[gateway.HTTPPost]url:%v, response:%v", string(body))
}
