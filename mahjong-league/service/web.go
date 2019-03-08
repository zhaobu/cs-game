package service

import (
	"mahjong-league/config"
	"mahjong-league/core"
	"mahjong-league/hall"
	"mahjong-league/ierror"
	"mahjong-league/model"
	"net/http"
	"strconv"
)

// GetUserRace 获取用户当前比赛id的接口
func GetUserRace(r *http.Request) (int64, *ierror.Error) {
	r.ParseForm()
	userId, _ := strconv.Atoi(r.Form.Get("userId"))
	if userId == 0 {
		return 0, ierror.NewError(-10101, "web.GetUserRace", "userId", userId)
	}
	return model.GetUserRace(userId), nil
}

// ClearUserRace 清除用户当前比赛信息
func ClearUserRace(r *http.Request) *ierror.Error {
	r.ParseForm()
	userId, _ := strconv.Atoi(r.Form.Get("userId"))
	if userId == 0 {
		return ierror.NewError(-10101, "web.ClearUserRace", "userId", userId)
	}
	core.Logger.Debug("[ClearUserRace]parse params, userId:%v", userId)

	model.DelUserRace(userId)

	core.Logger.Info("[ClearUserRace]userId:%v", userId)

	/*
		// 读取用户当前raceId信息
		raceId := model.GetUserRace(userId)
		if raceId == int64(0) {
			return nil
		}

		model.RaceList.Mux.Lock()
		defer model.RaceList.Mux.Unlock()

		raceInfo := model.RaceList.Get(raceId)
		if raceInfo == nil {
			return
		}

		// 如果用户未开始，则将用户
	*/
	return nil
}

// GetRaceList 获取当前比赛列表
func GetRaceList(r *http.Request) ([]*model.Race, *ierror.Error) {
	raceList := make([]*model.Race, 0)
	r.ParseForm()
	leagueId, _ := strconv.Atoi(r.Form.Get("leagueId"))
	core.Logger.Debug("[GetRaceList]parse params, leagueId:%v", leagueId)

	if leagueId == 0 {
		return raceList, ierror.NewError(-10101, "web.GetRaceList", "leagueId", leagueId)
	}

	model.RaceList.Mux.Lock()
	defer model.RaceList.Mux.Unlock()

	for _, raceInfo := range model.RaceList.Data {
		if leagueId > 0 && raceInfo.LeagueId != leagueId {
			continue
		}
		if !raceInfo.IsRunning() {
			continue
		}
		raceList = append(raceList, raceInfo)
	}

	core.Logger.Debug("raceList:%#v", raceList)
	return raceList, nil
}

// GetRaceUserList 获取当前比赛用户列表
func GetRaceUserList(r *http.Request) ([]*model.RaceUser, *ierror.Error) {
	userList := make([]*model.RaceUser, 0)
	r.ParseForm()
	raceId, _ := strconv.ParseInt(r.Form.Get("raceId"), 10, 64)
	if raceId == int64(0) {
		return userList, ierror.NewError(-10101, "web.GetRaceUserList", "raceId", raceId)
	}
	core.Logger.Debug("[GetRaceUserList]parse params, raceId:%v", raceId)

	raceUsers := model.GetRaceUsers(raceId)
	if raceUsers != nil {
		raceUsers.Mux.Lock()
		defer raceUsers.Mux.Unlock()

		for _, raceUserInfo := range raceUsers.Users {
			userList = append(userList, raceUserInfo)
		}
	}

	core.Logger.Info("[GetRaceUserList]raceId:%v", raceId)

	return userList, nil
}

// GetRaceRoomList 获取当前比赛列表
func GetRaceRoomList(r *http.Request) ([]*model.RaceRoom, *ierror.Error) {
	roomList := make([]*model.RaceRoom, 0)
	r.ParseForm()
	raceId, _ := strconv.ParseInt(r.Form.Get("raceId"), 10, 64)
	if raceId == int64(0) {
		return roomList, ierror.NewError(-10101, "web.GetRaceRoomList", "raceId", raceId)
	}
	core.Logger.Debug("[GetRaceRoomList]parse params, raceId:%v", raceId)

	raceRooms := model.GetRaceRooms(raceId)
	if raceRooms != nil {
		raceRooms.Mux.Lock()
		defer raceRooms.Mux.Unlock()

		for _, raceRoom := range raceRooms.Data {
			roomList = append(roomList, raceRoom)
		}
	}

	core.Logger.Info("[GetRaceRoomList]raceId:%v", raceId)

	return roomList, nil
}

// ForceRoomFinish 强制结束房间
func ForceRoomFinish(r *http.Request) *ierror.Error {
	r.ParseForm()
	raceId, _ := strconv.ParseInt(r.Form.Get("raceId"), 10, 64)
	if raceId == int64(0) {
		return ierror.NewError(-10101, "web.ForceRoomFinish", "raceId", raceId)
	}
	raceRoomId, _ := strconv.ParseInt(r.Form.Get("raceRoomId"), 10, 64)
	if raceRoomId == int64(0) {
		return ierror.NewError(-10101, "web.ForceRoomFinish", "raceRoomId", raceRoomId)
	}
	core.Logger.Debug("[ForceRoomFinish]parse params, raceId:%v, raceRoomId:%v", raceId, raceRoomId)

	roomFinish(raceId, raceRoomId, config.RACE_ROOM_STATUS_DISMISS)
	core.Logger.Info("[ForceRoomFinish]raceId:%v, raceRoomId:%v", raceId, raceRoomId)

	return nil
}

// ForceRaceFinish 强制结束比赛
func ForceRaceFinish(r *http.Request) *ierror.Error {
	r.ParseForm()

	raceId, _ := strconv.ParseInt(r.Form.Get("raceId"), 10, 64)
	if raceId == int64(0) {
		return ierror.NewError(-10101, "web.ForceRaceFinish", "raceId", raceId)
	}

	// 获取当前比赛状态
	model.RaceList.Mux.Lock()
	defer model.RaceList.Mux.Unlock()

	raceInfo := model.RaceList.Data[raceId]
	if raceInfo == nil {
		return ierror.NewError(-10505, raceId, 0)
	}

	// 更新比赛状态
	raceInfo.Status = config.RACE_STATUS_DISMISS_FORCE
	_, err := raceInfo.Update(nil, "status")
	if !ierror.MustNil(err) {
		return ierror.NewError(-4, err.Error())
	}
	core.Logger.Debug("[ForceRaceFinish]parse params, raceId:%v", raceId)

	raceUsers := model.GetRaceUsers(raceInfo.Id)
	raceUsers.Mux.Lock()
	defer raceUsers.Mux.Unlock()
	for _, ru := range raceUsers.Users {
		ru.Status = config.RACE_USER_STATUS_DISMISS
		model.DelUserRaceSpecied(ru.UserId, raceInfo.Id)
		core.GetWriter().Update(ru, "status")
	}

	core.Logger.Info("[ForceRaceFinish]raceId:%v", raceId)
	return nil
}

// HasRaceResult 强制结束比赛
func HasRaceResult(r *http.Request) (bool, *ierror.Error) {
	r.ParseForm()
	userId, _ := strconv.Atoi(r.Form.Get("userId"))
	if userId == 0 {
		return false, ierror.NewError(-10101, "web.ClearUserRace", "userId", userId)
	}
	v := hall.GetLastRaceResult(userId)
	return v != nil, nil
}
