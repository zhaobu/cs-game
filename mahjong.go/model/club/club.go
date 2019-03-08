package club

import (
	"github.com/astaxie/beego/orm"
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// GetClub 根据id从db读取用户数据
func GetClub(clubId int) *config.Club {
	club := &config.Club{Id: clubId}
	if dberr := core.GetWriter().Read(club); dberr != nil {
		core.Logger.Error("sql error, message: %s,", dberr.Error())
		// 将id设置成0
		club.Id = 0
	}
	return club
}

// GetFund 读取俱乐部钻石数
func GetFund(club *config.Club) int {
	return club.Fund
}

// UpdateFund 增加或者扣除俱乐部基金
func UpdateFund(clubId int, amount int) *core.Error {
	var params = orm.Params{}
	if amount < 0 {
		params["fund"] = orm.ColValue(orm.ColMinus, -1*amount)
	} else {
		params["fund"] = orm.ColValue(orm.ColAdd, amount)
	}
	_, dberr := core.GetWriter().QueryTable("club").Filter("id", clubId).Update(params)
	if dberr != nil {
		return core.NewError(-4, dberr.Error())
	}
	return nil
}
