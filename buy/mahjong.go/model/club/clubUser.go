package club

import (
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// IsClubUser 判断userId是否在俱乐部中
func IsClubUser(clubId int, userId int) bool {
	return core.GetWriter().QueryTable("club_user").Filter("club_id", clubId).Filter("user_id", userId).Exist()
}

// GetClubUser 根据id从db读取用户数据
func GetClubUser(clubId int, userId int) *config.ClubUser {
	clubUser := &config.ClubUser{}
	if dberr := core.GetWriter().QueryTable("club_user").Filter("club_id", clubId).Filter("user_id", userId).One(clubUser); dberr != nil {
		core.Logger.Error("sql error, message: %s,", dberr.Error())
		// 将id设置成0
		clubUser.ClubId = 0
		clubUser.UserId = 0
	}
	return clubUser
}
