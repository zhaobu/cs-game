package user

import (
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// 根据id从db读取用户数据
func GetUserInfoList(userId int) config.UserInfoList {
	o := core.GetWriter()

	var result []config.UserInfo
	userInfoList := config.UserInfoList{}
	o.QueryTable(config.TABLE_NAME_USER_INFO).Filter("userId", userId).All(&result)

	// 解析成map格式
	for _, v := range result {
		userInfoList[v.InfoType] = v
	}

	return userInfoList
}
