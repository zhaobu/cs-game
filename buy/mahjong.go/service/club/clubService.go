package club

import (
	"github.com/astaxie/beego/orm"

	"mahjong.go/library/core"
	clubModel "mahjong.go/model/club"
	logService "mahjong.go/service/log"
)

func UpdateFund(roomId int64, clubId int, amount int, changeType string, createTime int64) *core.Error {
	// 更新DB
	if err := clubModel.UpdateFund(clubId, amount); err != nil {
		core.Logger.Error("更新俱乐部基金错误, clubId: %d, amount: %d, err:%s", clubId, amount, err.Error())
		return err
	}
	logService.LogClubConsumeInfo(roomId, clubId, amount, changeType, createTime)
	return nil
}

// 更新用户的俱乐部淘汰赛积分
func UpdateUserScore(clubId, userId, value int) *core.Error {
	var params = orm.Params{
		"score": orm.ColValue(orm.ColAdd, value),
	}
	_, dberr := core.GetWriter().QueryTable("club_user").Filter("club_id", clubId).Filter("user_id", userId).Update(params)
	if dberr != nil {
		return core.NewError(-4, dberr.Error())
	}
	return nil
}
