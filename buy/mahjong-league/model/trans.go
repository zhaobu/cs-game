package model

import (
	"mahjong-league/core"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
)

func init() {
	orm.RegisterModel(new(UserTransInfo))
}

// UserTransInfo 收入日志
type UserTransInfo struct {
	Id           int `orm:"pk"`
	Sn           string
	UserId       int
	TargetUserId int
	Num          int
	CreateTime   int64
	TransType    int
	DiamondType  int
}

// LogUserTransInfo 记录收入日志
func LogUserTransInfo(userId, targetUserId, amount int, sn string, transType int, diamondType int) error {
	userTransInfo := new(UserTransInfo)
	userTransInfo.Sn = sn
	userTransInfo.UserId = userId
	userTransInfo.TargetUserId = targetUserId
	userTransInfo.Num = amount
	userTransInfo.CreateTime = util.GetTime()
	userTransInfo.TransType = transType
	userTransInfo.DiamondType = diamondType
	if _, dberr := core.GetWriter().Insert(userTransInfo); dberr != nil {
		core.Logger.Error("[LogUserTransInfo]入库失败,err:%v", dberr.Error())
		return dberr
	}
	return nil
}
