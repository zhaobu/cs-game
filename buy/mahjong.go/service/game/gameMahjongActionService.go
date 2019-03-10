package game

import (
	"mahjong.go/config"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/core"
	"mahjong.go/mi/oc"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 对应action的接口
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 用户操作
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func UserOperationAction(userId int, operation *fbsCommon.Operation) *core.Error {
	_, room, err := getUserRoom(userId)
	if err != nil {
		return err
	}

	// 解析operation
	op := int(operation.Op())       // 进行什么操作
	tiles := operation.TilesBytes() // 有什么牌

	// 将旧格式的operation转成新格式的operation
	// 需要先将byte类型的tiles转换成int类型
	intTiles := []int{}
	if len(tiles) > 0 {
		for _, v := range tiles {
			intTiles = append(intTiles, int(v))
		}
	}
	newOp := NewOperation(op, intTiles)

	core.Logger.Debug("[UserOperationAction]用户操作,userId:%v, roomId:%v, op:%#v, tiles:%#v", userId, room.RoomId, op, intTiles)

	return handlerUserReply(room, userId, newOp)
}

// 处理用户回应
func handlerUserReply(room *Room, userId int, operation *Operation) *core.Error {
	room.UserOperationMux.Lock()
	defer room.UserOperationMux.Unlock()

	// 这里的check需要加强，因为用户操作改成了异步，如果这里出现错误，会导致后面的数据异常且无法回滚
	if err := mahjongActionCheck(userId, room, config.ROOM_STATUS_PALYING, operation); err != nil {
		// 不支持的用户操作或者用户已经操作过了，需要提示用户重载数据
		// SendMessageByUserId(userId, GameRestorePush(userId, room))

		// 忽略掉重复操作
		if err.GetCode() == -315 {
			// do nothing
		} else {
			user, _ := UserMap.GetUser(userId)
			if user != nil {
				room.restoreIntact(user)
			}
		}
		return err
	}

	core.Logger.Debug("[handlerUserReply]userId:%d,roomId:%d,number:%s, op:%d, tiles:%#v", userId, room.RoomId, room.Number, operation.OperationCode, operation.Tiles)

	// 用户回应操做
	waitInfo := room.MI.replyWait(userId, operation)
	// 是否有暗杠花牌
	hasKongFlower := waitInfo.HasKongFlower()

	if operation.OperationCode == fbsCommon.OperationCodePASS_CANCEL {
		// 如果用户选择了“关闭决策框”，需要将wait中需要选择的决策删除，并且此次操作不写reply
		handlerUserOperation(room.MI, userId, operation)
	} else if room.MI.isAllReply() || oc.IsWinOperation(operation.OperationCode) {
		// 如果所有人都做了操作或者是有用户进行了胡的操作，则直接进入胡牌操作

		waitMap := room.MI.getReplyResult()
		lastOpCode := fbsCommon.OperationCodePASS
		waitMap.Maps.Range(func(k, v interface{}) bool {
			waitInfo := v.(*WaitInfo)
			handlerUserOperation(room.MI, k.(int), waitInfo.Reply)
			lastOpCode = waitInfo.Reply.OperationCode
			return true
		})
		core.Logger.Debug("lastOpCode:%v, hasKongFlower:%v, roomId:%v, round:%v", lastOpCode, hasKongFlower, room.RoomId, room.Round)

		// 提示用户回滚操作
		if operation.OperationCode == fbsCommon.OperationCodeWIN {
			room.MI.userOperationRollback()
		}

		// 清空waitQueue
		room.MI.cleanWait()

		// 处理用户操作之后的逻辑
		if lastOpCode == fbsCommon.OperationCodePLAY || lastOpCode == fbsCommon.OperationCodeBAO_TING { // 出牌后
			// 打牌或者软报之后，计算其他人能进行什么操作，不能进行操作的话，则通知下一次抓牌
			if room.MI.calcAfterUserOperation(config.MAHJONG_OPERATION_CALC_PLAY) == true {
				room.MI.run()
			}
		} else if oc.IsKongTurnOperation(lastOpCode) { // 转弯杠后
			// 转弯杠或憨包杠之后，需要判断其他人能不能胡牌，不能胡牌，则通知用户进行杠牌
			if room.MI.calcAfterUserOperation(config.MAHJONG_OPERATION_CALC_KONG_TURN) == true {
				room.MI.run()
			}
		} else if lastOpCode == fbsCommon.OperationCodeCHOW || lastOpCode == fbsCommon.OperationCodePONG { // 碰或者吃了后
			room.MI.userOperationAfterPong(room.MI.getLastOperator())
		} else if oc.IsWinOperation(lastOpCode) { // 胡牌后
			// 胡牌之后，需要走finish的逻辑去了
			room.MI.next()
		} else if lastOpCode == fbsCommon.OperationCodePASS && hasKongFlower { // 跳过补花的暗杠后
			// 如果选择了过，且是暗杠补花红中的话，帮用户执行补花操作
			// 这里必须要用userId，不能用lastOperator
			room.MI.userOperationFlowerExchange(userId)
			room.MI.run()
		} else if lastOpCode == fbsCommon.OperationCodePASS && room.MI.getLastOperation().OperationCode == fbsCommon.OperationCodeDRAW { // 报听后抓牌且选择了过
			// 抓牌之后，选择了过，说明是报听之后，提示了自摸，用户选择了过，需要自动帮用户打牌
			handlerUserOperation(room.MI, room.MI.getLastOperator(), NewOperation(fbsCommon.OperationCodePLAY, room.MI.getLastOperation().Tiles))
			// 帮用户打牌之后，还需要继续计算其他用户可进行的操作
			if room.MI.calcAfterUserOperation(config.MAHJONG_OPERATION_CALC_PLAY) == true {
				room.MI.run()
			}
		} else if lastOpCode == fbsCommon.OperationCodeNEED_LACK_TILE { // 定缺后
			// 定缺之后，完成当前步骤
			room.MI.next()
		} else if lastOpCode == fbsCommon.OperationCodeNEED_EXCHANGE_TILE { // 换牌后
			// 换牌后，完成当前步骤
			room.MI.next()
		} else {
			// 执行下一步、发牌
			room.MI.run()
		}
	} else {
		// 如果还有人没有进行回应，则不做任何操作，等候其他用户回应
		// 以后可以扩展成发送消息，通知用户正在等候其他玩家操作
		core.Logger.Debug("[handlerUserReply]操作完成之后,等待其他用户操作,roomId:%v,userId:%v, operation:%+v", room.RoomId, userId, operation)
	}

	return nil
}

// 处理用户操作
func handlerUserOperation(mi MahjongInterface, userId int, operation *Operation) *core.Error {
	var handlerError *core.Error

	switch operation.OperationCode {
	case fbsCommon.OperationCodePLAY: // 用户出牌
		handlerError = handlerUserOperationPlay(mi, userId, operation.Tiles[0])
		break
	case fbsCommon.OperationCodePONG: // 碰牌
		handlerError = handlerUserOperationPong(mi, userId)
		break
	case fbsCommon.OperationCodeKONG: // 明杠
		handlerError = handlerUserOperationKong(mi, userId)
		break
	case fbsCommon.OperationCodeKONG_DARK: // 暗杠
		handlerError = handlerUserOperationKongDark(mi, userId, operation.Tiles[0])
		break
	case fbsCommon.OperationCodeKONG_TURN: // 转弯杠
		fallthrough
	case fbsCommon.OperationCodeKONG_TURN_FREE: // 憨包杠
		handlerError = handlerUserOperationKongTurn(mi, userId, operation.Tiles[0], operation.OperationCode)
		break
	case fbsCommon.OperationCodeBAO_TING: // 报听
		handlerError = handlerUserOperationBaoTing(mi, userId, operation.Tiles[0])
		break
	case fbsCommon.OperationCodeWIN: // 胡
		handlerError = handlerUserOperationWin(mi, userId)
		break
	case fbsCommon.OperationCodeWIN_AFTER_KONG_TURN: // 抢杠胡
		handlerError = handlerUserOperationWinAfterKongTurn(mi, userId)
		break
	case fbsCommon.OperationCodeWIN_AFTER_KONG_PLAY: // 热炮
		handlerError = handlerUserOperationWinAfterKongPlay(mi, userId)
	case fbsCommon.OperationCodeWIN_SELF: // 自摸
		handlerError = handlerUserOperationWinSelf(mi, userId)
		break
	case fbsCommon.OperationCodeWIN_AFTER_KONG_DRAW: // 杠上开花
		handlerError = handlerUserOperationWinAfterKongDraw(mi, userId)
		break
	case fbsCommon.OperationCodePASS: // 跳过
		handlerUserOperationPass(mi, userId)
		break
	case fbsCommon.OperationCodePASS_CANCEL: // 跳过
		handlerUserOperationPassCancel(mi, userId)
		break
	case fbsCommon.OperationCodeNEED_LACK_TILE: // 定缺
		handlerError = handlerUserOperationLack(mi, userId, operation.Tiles[0])
		break
	case fbsCommon.OperationCodeNEED_EXCHANGE_TILE: // 换牌
		handlerError = handlerUserOperationExchange(mi, userId, operation.Tiles)
		break
	default:
		core.Logger.Error("未支持的用户操作, op:%d", operation.OperationCode)
		break
	}

	return handlerError
}

// 处理用户操作：出牌
func handlerUserOperationPlay(mi MahjongInterface, userId int, tile int) *core.Error {
	// 出牌
	if err := mi.userOperationPlay(userId, tile); err != nil {
		return err
	}

	return nil
}

// 处理用户操作：碰
func handlerUserOperationPong(mi MahjongInterface, userId int) *core.Error {

	// 执行碰牌操作
	if err := mi.userOperationPong(userId); err != nil {
		return err
	}

	return nil
}

// 处理用户操作：明杠(用户打一张)
func handlerUserOperationKong(mi MahjongInterface, userId int) *core.Error {
	if err := mi.userOperationKong(userId); err != nil {
		return err
	}

	return nil
}

// 处理用户操作：暗杠(自己抓一张, 抓这张不需要是杠的牌)
func handlerUserOperationKongDark(mi MahjongInterface, userId int, tile int) *core.Error {
	if err := mi.userOperationKongDark(userId, tile); err != nil {
		return err
	}

	return nil
}

// 处理用户操作：转弯杠(抓到一张名牌中碰过的牌,只有抓到这一张可以)
func handlerUserOperationKongTurn(mi MahjongInterface, userId int, tile int, opcode int) *core.Error {
	if err := mi.userOperationKongTurn(userId, tile, opcode); err != nil {
		return err
	}
	return nil
}

// 报听
func handlerUserOperationBaoTing(mi MahjongInterface, userId, tile int) *core.Error {
	if err := mi.userOperationBaoTing(userId, tile); err != nil {
		return err
	}

	return nil
}

// 点炮胡
func handlerUserOperationWin(mi MahjongInterface, userId int) *core.Error {
	if err := mi.userOperationWin(userId, fbsCommon.OperationCodeWIN); err != nil {
		return err
	}

	return nil
}

// 抢杠胡
func handlerUserOperationWinAfterKongTurn(mi MahjongInterface, userId int) *core.Error {
	if err := mi.userOperationWinAfterKongTurn(userId); err != nil {
		return err
	}

	return nil
}

// 热炮胡
func handlerUserOperationWinAfterKongPlay(mi MahjongInterface, userId int) *core.Error {
	if err := mi.userOperationWin(userId, fbsCommon.OperationCodeWIN_AFTER_KONG_PLAY); err != nil {
		return err
	}

	return nil
}

// 处理用户操作：自摸
func handlerUserOperationWinSelf(mi MahjongInterface, userId int) *core.Error {
	if err := mi.userOperationWinSelf(userId, fbsCommon.OperationCodeWIN_SELF); err != nil {
		return err
	}
	return nil
}

// 处理用户操作：杠上开花
func handlerUserOperationWinAfterKongDraw(mi MahjongInterface, userId int) *core.Error {
	if err := mi.userOperationWinSelf(userId, fbsCommon.OperationCodeWIN_AFTER_KONG_DRAW); err != nil {
		return err
	}
	return nil
}

// 跳过
func handlerUserOperationPass(mi MahjongInterface, userId int) {
	mi.userOperationPass(userId)
}

// pass_cancel
func handlerUserOperationPassCancel(mi MahjongInterface, userId int) {
	mi.userOperationPassCancel(userId)
}

// 处理用户操作：定缺
func handlerUserOperationLack(mi MahjongInterface, userId int, tile int) *core.Error {
	// 定缺
	mi.userOperationLack(userId, tile)
	return nil

}

// 处理用户操作：换牌
func handlerUserOperationExchange(mi MahjongInterface, userId int, tiles []int) *core.Error {
	// 定缺
	suc, err := mi.userOperationExchange(userId, tiles)
	if !suc && err != nil {
		core.Logger.Error("[userOperationExchange]userId:%v, tiles:%v, err:%v", userId, tiles, err.Error())
	}
	return nil
}

// action的公共检测函数
// 检测用户是否连接
// 房间是否存在
// 是否可以进行opCode这个操作
func mahjongActionCheck(userId int, room *Room, status int, op *Operation) *core.Error {
	// 判断房间是否处于解散中
	if room.IsDismissing() {
		return core.NewError(-307, room.RoomId)
	}

	// 游戏是否正在进行中
	if !room.CheckStatus(status) {
		return core.NewError(-312, room.Status, status)
	}

	// 判断是否允许这个操作
	if err := room.MI.checkUserOperation(userId, op); err != nil {
		return err
	}

	return nil
}
