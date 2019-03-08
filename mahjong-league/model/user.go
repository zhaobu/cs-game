package model

import (
	"mahjong-league/config"
	"mahjong-league/core"

	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(new(User))
}

// User 用户基本信息表
type User struct {
	UserId        int `orm:"pk"`
	Money         int
	GiftMoney     int
	Status        int
	LastLoginTime int64
	RegisterTime  int64
	Nickname      string
	IconUrl       string
	TypeId        int
	Unionid       string
}

// GetUser 根据id从db读取用户数据
func GetUser(userId int) *User {
	user := User{UserId: userId}
	if dberr := core.GetWriter().Read(&user); dberr != nil {
		// 将userId设置成0
		user.UserId = 0
		core.Logger.Error("sql error, message: %s", dberr.Error())
	}
	return &user
}

// GetMoney 读取用户钻石数(金钻+银钻)
func (user *User) GetMoney() int {
	return user.Money + user.GiftMoney
}

// UpdateMoney 执行扣钱或加钱操作
// 扣钱时优先扣除免费的钻石，扣完之后再扣收费的钻石
func (user *User) UpdateMoney(amount int) (m map[int]int, err error) {
	m = make(map[int]int)
	var params = orm.Params{}
	if amount < 0 {
		amountAbs := -1 * amount
		if user.GiftMoney > 0 {
			if user.GiftMoney >= amountAbs {
				params["gift_money"] = orm.ColValue(orm.ColMinus, amountAbs)
				m[config.ENTITY_MODULE_DIAMOND_FREE] = amountAbs
				amountAbs = 0
			} else {
				amountAbs -= user.GiftMoney
				params["gift_money"] = orm.ColValue(orm.ColMinus, user.GiftMoney)
				m[config.ENTITY_MODULE_DIAMOND_FREE] = user.GiftMoney
			}
		}
		if amountAbs > 0 {
			params["money"] = orm.ColValue(orm.ColMinus, amountAbs)
			m[config.ENTITY_MODULE_DIAMOND] = amountAbs
		}
	} else {
		params["gift_money"] = orm.ColValue(orm.ColAdd, amount)
		m[config.ENTITY_MODULE_DIAMOND_FREE] = amount
	}
	_, err = core.GetWriter().QueryTable("user").Filter("user_id", user.UserId).Update(params)
	return
}

// UpdateDiamond 执行扣除或者增加钻石操作
// 扣钱时优先扣除免费的钻石，扣完之后再扣收费的钻石
func (user *User) UpdateDiamond(amount int) (m map[int]int, err error) {
	m = make(map[int]int)
	var params = orm.Params{}
	if amount < 0 {
		params["money"] = orm.ColValue(orm.ColMinus, -1*amount)
		m[config.ENTITY_MODULE_DIAMOND] = -1 * amount
	} else {
		params["money"] = orm.ColValue(orm.ColAdd, amount)
		m[config.ENTITY_MODULE_DIAMOND] = amount
	}
	_, err = core.GetWriter().QueryTable("user").Filter("user_id", user.UserId).Update(params)
	return
}
