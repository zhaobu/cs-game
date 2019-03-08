package model

import (
	"fmt"
	"mahjong-league/core"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
)

func init() {
	orm.RegisterModel(new(UserAccountLog))
}

// UserAccountLog 用户消费日志
type UserAccountLog struct {
	LogId      int `orm:"pk"`
	UserId     int
	Money      int
	GiftMoney  int
	CreateTime int64
	Sn         string
	ChangeType string
	OrderId    int
}

// GenSn YYMMDDHHIISS + userId(固定8位，左补0) + 4位随机数字
func GenSn(userId int) string {
	arr := []string{time.Now().Format("060102150405"), fmt.Sprintf("%08d", userId), util.GetRandString(4)}
	return strings.Join(arr, "")
}

// LogMoney 消费日志
func LogMoney(userId, money, giftMoney int, change_type string, sn string) error {
	moneyLog := new(UserAccountLog)
	moneyLog.UserId = userId
	moneyLog.Money = money
	moneyLog.GiftMoney = giftMoney
	moneyLog.CreateTime = util.GetTime()
	moneyLog.Sn = sn
	moneyLog.ChangeType = change_type

	if _, dberr := core.GetWriter().Insert(moneyLog); dberr != nil {
		core.Logger.Error("[LogMoney]入库失败,err:%v", dberr.Error())
		return dberr
	}
	return nil
}
