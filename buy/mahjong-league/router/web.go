package router

import (
	"mahjong-league/controller"
	"net/http"
)

// WDespatch 分发短链接请求
func WDespatch() {
	// 获取用户当前所在的比赛id
	http.HandleFunc("/getUserRace", controller.GetUserRace)
	// 清除用户参与比赛标志
	http.HandleFunc("/clearUserRace", controller.ClearUserRace)
	// 获取当前比赛列表
	http.HandleFunc("/getRaceList", controller.GetRaceList)
	// 获取比赛用户列表
	http.HandleFunc("/getRaceUser", controller.GetRaceUser)
	// 获取比赛房间列表
	http.HandleFunc("/getRaceRooms", controller.GetRaceRooms)
	// 强制房间完成
	http.HandleFunc("/finishRoom", controller.FinishRoom)
	// 修正比赛状态
	http.HandleFunc("/finishRace", controller.FinishRace)
	// 是否有比赛结果未查看
	http.HandleFunc("/hasRaceResult", controller.HasRaceResult)
}
