package controller

import (
	"mahjong-league/core"
	"mahjong-league/response"
	"mahjong-league/service"
	"net/http"
)

// GetUserRace 获取用户当前参赛的id
func GetUserRace(w http.ResponseWriter, r *http.Request) {
	raceId, err := service.GetUserRace(r)
	if err != nil {
		core.Logger.Error("[GetUserRace], err:%v", err.Error())
	}
	data := make(map[string]interface{})
	data["raceId"] = raceId
	w.Write(response.GenJSONBytes(err, data))
}

// GetRaceList 获取比赛列表，包括报名中、进行中、结算中的
func GetRaceList(w http.ResponseWriter, r *http.Request) {
	raceList, err := service.GetRaceList(r)
	if err != nil {
		core.Logger.Error("[controller.web.GetRaceList]err:%v", err.Error())
	}
	data := make(map[string]interface{})
	data["list"] = raceList

	w.Write(response.GenJSONBytes(err, data))
}

// GetRace 获取比赛信息，包括报名中、进行中、结算中的
func GetRace(w http.ResponseWriter, r *http.Request) {

}

// GetRaceUser 获取比赛用户列表，包括报名中、进行中、结算中的
func GetRaceUser(w http.ResponseWriter, r *http.Request) {
	raceUserList, err := service.GetRaceUserList(r)
	if err != nil {
		core.Logger.Error("[controller.web.GetRaceUserList]err:%v", err.Error())
	}
	data := make(map[string]interface{})
	data["list"] = raceUserList
	w.Write(response.GenJSONBytes(err, data))
}

// GetRaceRooms 获取比赛房间列表，进行中的比赛
func GetRaceRooms(w http.ResponseWriter, r *http.Request) {
	raceRoomList, err := service.GetRaceRoomList(r)
	if err != nil {
		core.Logger.Error("[controller.web.GetRaceRoomList]err:%v", err.Error())
	}
	data := make(map[string]interface{})
	data["list"] = raceRoomList
	w.Write(response.GenJSONBytes(err, data))
}

// FinishRace 强制结束比赛
func FinishRace(w http.ResponseWriter, r *http.Request) {
	err := service.ForceRaceFinish(r)
	if err != nil {
		core.Logger.Error("[controller.web.FinishRace]err:%v", err.Error())
	}
	w.Write(response.GenJSONBytes(err, nil))
}

// FinishRoom 强制房间结束
func FinishRoom(w http.ResponseWriter, r *http.Request) {
	err := service.ForceRoomFinish(r)
	if err != nil {
		core.Logger.Error("[controller.web.FinishRoom]err:%v", err.Error())
	}
	w.Write(response.GenJSONBytes(err, nil))
}

// ClearUserRace 清除用户的『比赛中』状态
func ClearUserRace(w http.ResponseWriter, r *http.Request) {
	err := service.ClearUserRace(r)
	if err != nil {
		core.Logger.Error("[controller.web.ClearUserRace]err:%v", err.Error())
	}
	w.Write(response.GenJSONBytes(err, nil))
}

// HasRaceResult 是否有游戏结果未查看
func HasRaceResult(w http.ResponseWriter, r *http.Request) {
	exists, err := service.HasRaceResult(r)
	if err != nil {
		core.Logger.Error("[controller.web.ClearUserRace]err:%v", err.Error())
	}
	d := map[string]interface{}{
		"exists": exists,
	}
	w.Write(response.GenJSONBytes(err, d))
}
