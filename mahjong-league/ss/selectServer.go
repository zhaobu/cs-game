package ss

import (
	"io/ioutil"
	"mahjong-league/core"
	"net/http"
	"strconv"
	"strings"
)

// Select 选服
func Select() string {
	if core.IsLocal() {
		return "0.0.0.0:9090"
	}

	url := core.GetAppConfig().SelectServerUrl
	resp, err := http.Get(url)
	if err != nil {
		core.Logger.Error("[Select]het.Get, url:%v, error:%v", url, err.Error())
		return ""
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		core.Logger.Error("[Select]het.Get read body, url:%v, error:%v", url, err.Error())
	}
	return string(body)
}

// GetUserRoomId 获取用户的房间id
func GetUserRoomId(userId int) int64 {
	url := core.GetAppConfig().SelectMyServerUrl + "?user_id=" + strconv.Itoa(userId)
	core.Logger.Trace("[GetUserRoomId]url:%v", url)
	resp, err := http.Get(url)
	if err != nil {
		core.Logger.Error("[ss.GetUserRoomId]het.Get, url:%v, error:%v", url, err.Error())
		return 0
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		core.Logger.Error("[ss.GetUserRoomId]het.Get read body, url:%v, error:%v", url, err.Error())
	}
	result := string(body)
	core.Logger.Trace("[GetUserRoomId]result:%v", result)
	info := strings.Split(result, "|")
	if len(info) > 1 {
		roomId, _ := strconv.ParseInt(info[1], 10, 64)
		return roomId
	}
	return 0
}
