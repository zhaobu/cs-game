package service

import (
	"mahjong-league/config"
	"mahjong-league/core"
	"mahjong-league/ierror"
	"mahjong-league/model"

	"github.com/fwhappy/util"
)

// ChargeEntity 扣费
func ChargeEntity(userId int, entityId int, price int, changeType string, moneyType int) (consume map[int]int, err *ierror.Error) {
	consume = make(map[int]int)
	m := make(map[int]int)
	// 如果eneityId=0，默认是金银钻
	if entityId == 0 {
		entityId = util.BuildEntityID(config.ENTITY_ID_CONSUME, config.ENTITY_MODULE_DIAMOND_ALL, 0)
		core.Logger.Warn("[ChargeEntity]entityId=0, userId:%v, price:%v", userId, price)
	}

	// 解析entityID
	_, module, subId := util.ParseEntityID(entityId)
	switch module {
	case config.ENTITY_MODULE_DIAMOND_ALL:
		m, err = UpdateMoney(model.GetUser(userId), price*-1, false, changeType, moneyType)
		for k, v := range m {
			consumeEntityId := util.BuildEntityID(config.ENTITY_ID_CONSUME, k, 0)
			consume[consumeEntityId] = v
		}
	case config.ENTITY_MODULE_DIAMOND:
		m, err = UpdateMoney(model.GetUser(userId), price*-1, true, changeType, moneyType)
		for k, v := range m {
			consumeEntityId := util.BuildEntityID(config.ENTITY_ID_CONSUME, k, 0)
			consume[consumeEntityId] = v
		}
	case config.ENTITY_MODULE_GOLD:
		m, err = UpdateGold(userId, price*-1)
		for k, v := range m {
			consumeEntityId := util.BuildEntityID(config.ENTITY_ID_CONSUME, k, 0)
			consume[consumeEntityId] = v
		}
	case config.ENTITY_MODULE_ITEM:
		m, err = UpdateItem(userId, subId, price*-1)
		for k, v := range m {
			consumeEntityId := util.BuildEntityID(config.ENTITY_ID_CONSUME, config.ENTITY_MODULE_ITEM, k)
			consume[consumeEntityId] = v
		}
	default:
		core.Logger.Error("[ChargeEntity]未支持的entityId, userId:%v, entityId:%v, price:%v", userId, entityId, price)
	}
	return
}

// ChargeReturnEntities 退费
func ChargeReturnEntities(userId int, entities map[int]int, changeType string, moneyType int) (consume map[int]int, err *ierror.Error) {
	consume = make(map[int]int)
	m := make(map[int]int)
	user := model.GetUser(userId)
	for entityId, price := range entities {
		_, module, subId := util.ParseEntityID(entityId)
		switch module {
		case config.ENTITY_MODULE_DIAMOND_ALL:
			fallthrough
		case config.ENTITY_MODULE_DIAMOND_FREE:
			m, err = UpdateMoney(user, price, false, changeType, moneyType)
			for k, v := range m {
				consumeEntityId := util.BuildEntityID(config.ENTITY_ID_GET, k, 0)
				consume[consumeEntityId] += v
			}
		case config.ENTITY_MODULE_DIAMOND:
			m, err = UpdateMoney(user, price, true, changeType, moneyType)
			for k, v := range m {
				consumeEntityId := util.BuildEntityID(config.ENTITY_ID_GET, k, 0)
				consume[consumeEntityId] += v
			}
		case config.ENTITY_MODULE_GOLD:
			m, err = UpdateGold(userId, price)
			for k, v := range m {
				consumeEntityId := util.BuildEntityID(config.ENTITY_ID_GET, k, 0)
				consume[consumeEntityId] += v
			}
		case config.ENTITY_MODULE_ITEM:
			m, err = UpdateItem(userId, subId, price)
			for k, v := range m {
				consumeEntityId := util.BuildEntityID(config.ENTITY_ID_GET, config.ENTITY_MODULE_ITEM, k)
				consume[consumeEntityId] += v
			}
		default:
			core.Logger.Error("[ChargeReturnEntities]未支持的entityId, userId:%v, entityId:%v, price:%v", userId, entityId, price)
		}
		if err != nil {
			core.Logger.Emergency("[ChargeReturnEntities]error, userId:%v, entityId:%v, price:%v, err:%v", userId, entityId, price, err.Error())
		}
	}
	return
}

// UpdateMoney 更新用户钻石, amount为负，表示扣钻石
// 优先扣除免费钻石，免费钻石不够扣的时候，再扣付费钻石
func UpdateMoney(user *model.User, amount int, forceDiamond bool, changeType string, moneyType int) (m map[int]int, err *ierror.Error) {
	m = make(map[int]int)
	// 判断用户金额是否足够支付
	if amount < 0 {
		var hasMoney int
		if forceDiamond {
			hasMoney = user.Money
		} else {
			hasMoney = user.GetMoney()
		}
		if hasMoney < -1*amount {
			err = ierror.NewError(-204, -1*amount)
			return
		}
	}

	var dbErr error
	if forceDiamond {
		m, dbErr = user.UpdateDiamond(amount)
	} else {
		m, dbErr = user.UpdateMoney(amount)
	}
	if dbErr != nil {
		core.Logger.Error("更新用户钻石错误, userId: %d, amount: %d, err:%s", user.UserId, amount, dbErr.Error())
		err = ierror.NewError(-4, dbErr.Error())
	}

	// 钻石数量
	money := m[config.ENTITY_MODULE_DIAMOND]
	// 免费钻石数量
	giftMoney := m[config.ENTITY_MODULE_DIAMOND_FREE]
	// 正负值
	symbal := 1
	if amount < 0 {
		symbal = -1
	}

	// 记录操作日志
	sn := model.GenSn(user.UserId)
	model.LogMoney(user.UserId, money*symbal, giftMoney*symbal, changeType, sn)

	if amount > 0 {
		// 记录收入日志
		if money > 0 {
			model.LogUserTransInfo(0, user.UserId, money, sn, moneyType, config.DIAMOND_TYPE_MONEY)
		}
		if giftMoney > 0 {
			model.LogUserTransInfo(0, user.UserId, giftMoney, sn, moneyType, config.DIAMOND_TYPE_GIFT_MONEY)
		}
	} else {
		// 记录消耗日志
		model.LogConsumeInfo(user.UserId, money, giftMoney, sn, "比赛场", moneyType)
	}
	return
}

// UpdateGold 更新用户金币
func UpdateGold(userId int, amount int) (m map[int]int, err *ierror.Error) {
	m = make(map[int]int)
	remain := amount + model.GetUserInfoGold(userId)
	if remain < 0 {
		err = ierror.NewError(-701)
		return
	}

	dbErr := model.UpdateUserInfoGold(userId, remain)
	if dbErr != nil {
		core.Logger.Error("[UpdateGold]userId:%v, amount:%v dbErr:%v", userId, amount, dbErr.Error())
		err = ierror.NewError(-4, dbErr.Error())
		return
	}

	// 记录金币日志
	if amount > 0 {
		// 退费日志
		model.LogCoinAddedLog(userId, amount)
	} else {
		// 消耗日志
		model.LogCoinConsumeLog(userId, amount*-1)
	}
	m[config.ENTITY_MODULE_GOLD] = util.Abs(amount)
	return
}

// UpdateItem 更新用户道具
func UpdateItem(userId, itemId, amount int) (m map[int]int, err *ierror.Error) {
	m = make(map[int]int)
	if amount < 0 {
		had := model.GetUserInfoItemCount(userId, itemId)
		if had+amount < 0 {
			err = ierror.NewError(-300)
			return
		}
	}

	dbErr := model.UpdateUserInfoItem(userId, itemId, amount)
	if dbErr != nil {
		core.Logger.Error("[UpdateItem]userId:%v, amount:%v dbErr:%v", userId, amount, dbErr.Error())
		err = ierror.NewError(-4, dbErr.Error())
		return
	}

	// 记录日志
	var entityId int
	if amount < 0 {
		entityId = util.BuildEntityID(config.ENTITY_ID_CONSUME, config.ENTITY_MODULE_ITEM, itemId)
	} else {
		entityId = util.BuildEntityID(config.ENTITY_ID_GET, config.ENTITY_MODULE_ITEM, itemId)
	}
	amount = util.Abs(amount)
	m[itemId] = amount
	model.LogPropsLog(userId, entityId, amount)

	return
}
