package selectserver

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mahjong-connection/core"
	"net/http"
	"strconv"
	"strings"
)

const ()

// Result 选服结果
type Result struct {
	Code   int
	Remote string
	RoomId int64
	RaceId int64
}

// ResultIntact 给网关特供的选服结果
type ResultIntact struct {
	Code       int    `json:"code"`
	Errmsg     string `json:"errmsg"`
	RoomId     int64  `json:"room_id"`
	RoomRemote string `json:"room_remote"`
	RaceId     int64  `json:"race_id"`
	RaceRemote string `json:"race_remote"`
}

// SelectByType 根据类型来选择
// selectType :
// CREATE_ROOM : 自建房间
// 房间号: 加入房间
// RANDOM_ROOM : 随机组局
// KING_ROOM : 雀王榜
// COIN_ROOM : 金币场
// CREATE_TERMINAL_ROOM : 电视端
// LEAGUE : 联赛
// LEAGUE_SERVER : 联赛大厅

func SelectByType(userId int, version string, selectType string) *Result {
	return selectServer(userId, version, selectType)
}

// Create 自建房间
func Create(userId int, version string) *Result {
	return selectServer(userId, version, "CREATE_ROOM")
}

// Join 加入房间
func Join(userId int, version string, number string) *Result {
	return selectServer(userId, version, number)
}

// Random 随机组队
func Random(userId int, version string) *Result {
	return selectServer(userId, version, "RANDOM_ROOM")
}

// Match 雀王榜
func Match(userId int, version string) *Result {
	return selectServer(userId, version, "KING_ROOM")
}

// Coin 金币场
func Coin(userId int, version string) *Result {
	return selectServer(userId, version, "COIN_ROOM")
}

// TV 电视端
func TV(userId int, version string) *Result {
	return selectServer(userId, version, "CREATE_TERMINAL_ROOM")
}

// League 联赛
func League(userId int, version string) *Result {
	return selectServer(userId, version, "LEAGUE")
}

// Rank 排位赛
func Rank(userId int, version string) *Result {
	return selectServer(userId, version, "RANK_ROOM")
}

// Club 俱乐部
func Club(userId int, version string) *Result {
	remote := core.GetAppConfig().ClubRemote
	return &Result{
		Remote: remote,
	}
}

// LeagueServer 联赛大厅
func LeagueServer(userId int, version string) *Result {
	return selectRemote(userId, version, "LEAGUE_SERVER")
}

// Reconnect 重连房间
func Reconnect(userId int, version string) *ResultIntact {
	return selectIntact(userId, version, "RECONNECT_ROOM")
}

// 内网选服
func selectServer(userId int, version string, selectType string) *Result {
	url := fmt.Sprintf("%s/client/selectServerForLocal?user_id=%d&version=%s&select_type=%v",
		core.GetAppConfig().SelectServerUrl, userId, version, selectType)
	core.Logger.Trace("[selectServer]url:%v", url)
	resp, err := http.Get(url)
	if err != nil {
		core.Logger.Error("[httpGet]het.Get, url:%v, error:%v", url, err.Error())
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		core.Logger.Error("[httpGet]het.Get read body, url:%v, error:%v", url, err.Error())
		return nil
	}

	core.Logger.Info("[selectServer]userId:%v, version:%v, selectType:%v, response:%v", userId, version, selectType, string(body))

	result := &Result{}
	data := strings.Split(string(body), "|")
	if len(data) == 1 {
		// 出错了
		result.Code, _ = strconv.Atoi(data[0])
	} else {
		if len(data) >= 1 {
			result.Remote = data[0]
		}
		if len(data) >= 2 {
			result.RoomId, _ = strconv.ParseInt(data[1], 10, 64)
		}
		if len(data) >= 3 {
			result.RaceId, _ = strconv.ParseInt(data[2], 10, 64)
		}
	}

	return result
}

// 完整版 选服
func selectIntact(userId int, version string, selectType string) *ResultIntact {
	url := fmt.Sprintf("%s/selectIntact?user_id=%d&version=%s&select_type=%v",
		core.GetAppConfig().SelectServerUrl, userId, version, selectType)
	core.Logger.Trace("[selectIntact]url:%v", url)
	resp, err := http.Get(url)
	if err != nil {
		core.Logger.Error("[selectIntact.httpGet]het.Get, url:%v, error:%v", url, err.Error())
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		core.Logger.Error("[selectIntact.httpGet]het.Get read body, url:%v, error:%v", url, err.Error())
		return nil
	}

	core.Logger.Info("[selectIntact.selectServer]userId:%v, version:%v, selectType:%v, response:%v", userId, version, selectType, string(body))

	result := ResultIntact{}
	decodeErr := json.Unmarshal(body, &result)
	if err != nil {
		core.Logger.Error("[selectIntact]json decode error, err:%v", decodeErr.Error())
	}
	return &result
}

// 内网选服
func selectRemote(userId int, version string, selectType string) *Result {
	url := fmt.Sprintf("%s/selectServerRemote?version=%s&select_type=%v",
		core.GetAppConfig().SelectServerUrl, version, selectType)
	core.Logger.Trace("[selectRemote]url:%v", url)
	resp, err := http.Get(url)
	if err != nil {
		core.Logger.Error("[httpGet]het.Get, url:%v, error:%v", url, err.Error())
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		core.Logger.Error("[httpGet]het.Get read body, url:%v, error:%v", url, err.Error())
		return nil
	}

	core.Logger.Info("[selectRemote]version:%v, selectType:%v, response:%v", version, selectType, string(body))

	result := &Result{}
	result.Remote = string(body)

	return result
}
