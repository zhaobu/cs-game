package user

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// 根据id从db读取用户数据
func GetUser(userId int) *config.User {
	o := core.GetWriter()

	user := config.User{UserId: userId}
	dberr := o.Read(&user)

	// 判断用户是否存在
	if dberr != nil {
		errsql := fmt.Sprintf("SELECT * FROM user WHERE userId = %d limit 1", userId)
		core.Logger.Error("sql error, message: %s, errsql: %s", dberr.Error(), errsql)

		// 将userId设置成0
		user.UserId = 0
	}

	return &user
}

// 读取用户钻石数
func GetMoney(user *config.User) int {
	return user.Money + user.GiftMoney
}

// 执行扣钱或加钱操作
// 扣钱时优先扣除免费的钻石，扣完之后再扣收费的钻石
func UpdateMoney(ormObj orm.Ormer, user *config.User, amount int) (m map[int]int, err *core.Error) {
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
	_, dberr := core.GetWriter().QueryTable("user").Filter("user_id", user.UserId).Update(params)
	if dberr != nil {
		err = core.NewError(-4, dberr.Error())
	}
	return
}

// 插入userInfo数据
func InsertInfo(userId int, infoType int, infoValue string) (*config.UserInfo, *core.Error) {
	userInfo := &config.UserInfo{}
	userInfo.UserId = userId
	userInfo.InfoType = infoType
	userInfo.Info = infoValue
	userInfo.Time = util.GetTime()

	o := core.GetWriter()
	if _, dberr := o.Insert(userInfo); dberr != nil {
		return nil, core.NewError(-4, dberr.Error())
	}

	return userInfo, nil
}

// 更新userInfo数据
func UpdateInfo(userInfo config.UserInfo) *core.Error {
	o := core.GetWriter()
	_, dberr := o.QueryTable("user_info").Filter("user_id", userInfo.UserId).Filter("info_type", userInfo.InfoType).Update(orm.Params{"info": userInfo.Info})
	if dberr != nil {
		return core.NewError(-4, dberr.Error())
	}

	return nil
}
