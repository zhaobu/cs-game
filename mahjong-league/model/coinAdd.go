package model

import (
	"mahjong-league/core"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
)

func init() {
	orm.RegisterModel(new(CoinAddedLog))
}

// CoinAddedLog 退费日志
type CoinAddedLog struct {
	Id         int64 `orm:"pk"`
	UserId     int   // 用户id
	Coin       int   // 退费
	CoinType   int   // 消耗类型
	CreateTime int64 // 房间创建时间
}

// LogCoinAddedLog 记录退费日志
func LogCoinAddedLog(userId int, amount int) error {
	defer util.RecoverPanic()

	log := new(CoinAddedLog)
	log.UserId = userId
	log.Coin = amount
	log.CoinType = 8
	log.CreateTime = util.GetTime()
	if _, dberr := core.GetWriter().Insert(log); dberr != nil {
		core.Logger.Error("[LogCoinAddedLog]入库失败,err:%v", dberr.Error())
		return dberr
	}
	return nil
}
