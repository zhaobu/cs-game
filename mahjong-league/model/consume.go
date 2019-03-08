package model

import (
	"mahjong-league/core"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
)

func init() {
	orm.RegisterModel(new(UserConsumeInfo))
}

// UserConsumeInfo 消费日志
type UserConsumeInfo struct {
	Id         int `orm:"pk"`
	Sn         string
	UserId     int
	Num        int
	GiftNum    int
	Note       string
	Ctype      int // 消耗类型
	CreateTime int64
}

// LogConsumeInfo 记录消费日志
func LogConsumeInfo(userId, money, giftMoney int, sn, note string, ctype int) error {
	userConsumeInfo := new(UserConsumeInfo)
	userConsumeInfo.Sn = sn
	userConsumeInfo.UserId = userId
	userConsumeInfo.Num = money
	userConsumeInfo.GiftNum = giftMoney
	userConsumeInfo.Note = note
	userConsumeInfo.Ctype = ctype
	userConsumeInfo.CreateTime = util.GetTime()

	if _, dberr := core.GetWriter().Insert(userConsumeInfo); dberr != nil {
		core.Logger.Error("[LogConsumeInfo]入库失败,err:%v", dberr.Error())
		return dberr
	}
	return nil
}
