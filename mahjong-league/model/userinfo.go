package model

import (
	"encoding/json"
	"mahjong-league/config"
	"mahjong-league/core"
	"strconv"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
)

func init() {
	orm.RegisterModel(new(UserInfo))
}

// UserInfo 用户扩展信息表
type UserInfo struct {
	UserId   int `orm:"pk"`
	InfoType int
	Info     string
	Time     int64
}

// TableName 数据库表名
func (ui *UserInfo) TableName() string {
	return "user_info"
}

// GetUserInfo 获取用户信息
func GetUserInfo(userId, infoType int) *UserInfo {
	ui := &UserInfo{}
	err := core.GetWriter().QueryTable("user_info").Filter("user_id", userId).Filter("info_type", infoType).One(ui)
	if err != nil {
		core.Logger.Error("[GetUserInfo]userId:%v, infoType:%v, err:%v", userId, infoType, err.Error())
	}
	return ui
}

// UpdateUserInfo 获取用户信息
func UpdateUserInfo(userId, infoType int, params orm.Params) error {
	_, err := core.GetWriter().QueryTable("user_info").Filter("user_id", userId).Filter("info_type", infoType).Update(params)
	if err != nil {
		core.Logger.Error("[UpdateUserInfo]userId:%v, infoType:%v, err:%v", userId, infoType, err.Error())
	}
	return nil
}

// GetUserInfoGold 获取用户金币余额
func GetUserInfoGold(userId int) int {
	ui := GetUserInfo(userId, config.USER_INFO_TYPE_SCORE_COIN)
	if ui == nil {
		return 0
	}
	amount, err := strconv.Atoi(ui.Info)
	if err != nil {
		core.Logger.Emergency("[GetUserInfoGold]strconv.Atoi, userId:%v, err:%v", userId, err.Error())
		return 0
	}
	return amount
}

// UpdateUserInfoGold 更新用户金币余额
func UpdateUserInfoGold(userId int, amount int) error {
	var params = orm.Params{}
	params["info"] = amount
	_, err := core.GetWriter().QueryTable("user_info").Filter("user_id", userId).Filter("info_type", config.USER_INFO_TYPE_SCORE_COIN).Update(params)
	return err
}

// GetUserInfoItems 获取用户道具列表
func GetUserInfoItems(userId int) map[string]int {
	items := make(map[string]int)
	ui := GetUserInfo(userId, config.USER_INFO_TYPE_ITEMS)
	if ui != nil {
		if len(ui.Info) > 0 {
			err := json.Unmarshal([]byte(ui.Info), &items)
			if err != nil {
				core.Logger.Error("[GetUserInfoItems] json unmarshal error, string:%v, err:%v", ui.Info, err.Error())
			}
		}
	}
	return items
}

// GetUserInfoItemCount 获取用户道具数量
func GetUserInfoItemCount(userId int, itemId int) int {
	items := GetUserInfoItems(userId)
	return items[strconv.Itoa(itemId)]
}

// UpdateUserInfoItem 更新道具数量
func UpdateUserInfoItem(userId int, itemId int, count int) error {
	items := GetUserInfoItems(userId)
	itemIdStr := strconv.Itoa(itemId)
	items[itemIdStr] += count
	had := items[itemIdStr]
	if had < 0 {
		core.Logger.Error("[UpdateUserInfoItem]用户道具被扣成了负数, userId:%v, itemId:%v, had:%v", userId, itemId, had)
		delete(items, itemIdStr)
	} else if had == 0 {
		delete(items, itemIdStr)
	}
	var params = orm.Params{}
	params["info"], _ = util.InterfaceToJsonString(items)
	_, err := core.GetWriter().QueryTable("user_info").Filter("user_id", userId).Filter("info_type", config.USER_INFO_TYPE_ITEMS).Update(params)
	return err
}
