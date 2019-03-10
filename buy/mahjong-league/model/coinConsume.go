package model

import (
	"mahjong-league/core"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
)

func init() {
	orm.RegisterModel(new(CoinConsumeLog))
}

// CoinConsumeLog 金币场消费日志
type CoinConsumeLog struct {
	Id         int64 `orm:"pk"`
	MatchType  int
	RoomId     int64 // 房间id
	UserId     int   // 用户id
	Coin       int   // 消耗的金币
	CoinType   int   // 消耗类型
	CreateTime int64 // 房间创建时间
}

// LogCoinConsumeLog 记录消费日志
func LogCoinConsumeLog(userId int, coin int) error {
	defer util.RecoverPanic()

	coinConsumeLog := new(CoinConsumeLog)
	coinConsumeLog.UserId = userId
	coinConsumeLog.Coin = coin
	coinConsumeLog.CoinType = 4
	coinConsumeLog.CreateTime = util.GetTime()
	if _, dberr := core.GetWriter().Insert(coinConsumeLog); dberr != nil {
		core.Logger.Error("[LogCoinConsumeLog]入库失败,err:%v", dberr.Error())
		return dberr
	}
	return nil
}
