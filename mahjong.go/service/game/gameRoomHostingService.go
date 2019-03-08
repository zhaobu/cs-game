package game

import (
	"time"

	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/mi/oc"
	"mahjong.go/mi/ting"

	"github.com/fwhappy/util"
	fbsCommon "mahjong.go/fbs/Common"
)

// 判断房间是否支持自动托管
func (room *Room) enableAutoHosting() bool {
	if room.IsCreate() || room.IsClub() || room.IsTV() {
		return false
	}

	// 是否有额外的配置
	if core.AppConfig.EnableAutoHosting == 0 {
		return false
	}

	return true
}

// 用户是否处于托管状态
func (room *Room) userInHosting(userId int) bool {
	return util.IntInSlice(userId, room.HostingUsers)
}

// 监听房间托管
func (r *Room) ListenHosting() {
	// 捕获异常
	defer util.RecoverPanic()
	core.Logger.Debug("监听房间托管,roomId:%v", r.RoomId)
	for {
		// 每秒检查
		time.Sleep(time.Second)

		room, err := RoomMap.GetRoom(r.RoomId)
		if room == nil || err != nil {
			core.Logger.Debug("房间已结束,退出托管,roomId:%v.", r.RoomId)
			break
		}

		// 如果房间已结束，退出执行
		if room.CheckStatus(config.ROOM_STATUS_COMPLETED) {
			core.Logger.Debug("房间已结束,退出托管,roomId:%v.", room.RoomId)
			break
		}

		// 如果房间正在解散中，暂停托管操作
		if room.DismissTime > 0 {
			continue
		}

		// 如果房间支持自动托管，需要将操作超时的用户，自动进入托管状态
		if room.enableAutoHosting() {
			room.autoHosting()
		}

		// 如果无人托管，等候下一次检查
		if len(room.HostingUsers) == 0 {
			continue
		}

		// core.Logger.Debug("[ListenHosting]检测到房间有托管用户, roomId:%v, round:%v, hostingUsers:%v", room.RoomId, room.Round, room.HostingUsers)

		// 自动准备 & 自动定缺 & 自动打牌 & 自动胡牌
		if room.IsReadying() {
			// 用户准备
			room.hostingReady()
		} else {
			room.hostingOperation()
		}
	}
}

// 检测用户操作，自动帮用户进入托管状态
func (room *Room) autoHosting() {
	var startTime int64
	var users []int
	// 获取上次操作到现在的时间间隔
	if room.IsReadying() {
		// 如果是刚刚组局，那么从满员开始算
		// 如果房间处于准备中，需要知道房间上一局的结束时间
		if room.Round == 0 {
			startTime = room.FullTime
		} else {
			startTime = room.LastRoundCompletedTime
		}

		users = room.GetUnReadyUsers()
	} else {
		// 如果牌局还没有初始化，则跳出本次监听
		if room.MI == nil {
			return
		}
		// 需要知道上次回应操作出现到现在的时间
		startTime = room.MI.getReplyInitTime()
		users = room.MI.getUnReplyWaitUsers()
	}

	// 未到自动准备的时间间隔
	if util.GetTime()-startTime <= config.ROOM_AUTO_HOSTING_INTERVAL_SECOND {
		return
	}

	room.Mux.Lock()
	defer room.Mux.Unlock()

	for _, userId := range users {
		if !room.userInHosting(userId) {
			room.HostingUsers = append(room.HostingUsers, userId)
			// 推送消息给用户
			// SendMessageByUserId(userId, GameHostingPush(userId, config.ROOM_USER_HOSTING_YES))
			room.SendMessageToRoomUser(GameHostingPush(userId, config.ROOM_USER_HOSTING_YES), 0)
			core.Logger.Info("[autoHosting]roomId:%v,number:%v,userId:%v", room.RoomId, room.Number, userId)
		}
	}
}

// 托管处理: 用户准备
func (room *Room) hostingReady() {
	// 因为客户端动画的原因，托管准备必须要等5秒
	if util.GetTime()-room.LastRoundCompletedTime < config.HOSTING_OPERATION_WAIT_TIME_5 {
		return
	}
	// 防并发加锁
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 是否全员准备好
	var allReadyed = false
	var err *core.Error
	for k, userId := range room.HostingUsers {
		// 跳过已经准备的用户
		if util.IntInSlice(userId, room.ReadyList) {
			continue
		}

		// 判断房间是否处于回应准备状态
		allReadyed, err = room.userOperationReady(userId)
		if err != nil {
			core.Logger.Error("[GameReady][hosting]roomId:%d, round:%v, userId:%d,error:%s", room.RoomId, room.Round, userId, err.Error())
			continue
		}

		core.Logger.Info("[GameReady][hosting]roomId:%v, round:%v, userId:%v", room.RoomId, room.Round, userId)

		// 多成员的时候，每个等候1秒
		if k != len(room.HostingUsers)-1 {
			room.hostingWait()
		}
	}
	// 准备完成，执行初始化
	if allReadyed {
		room.nextGame()
	}
}

// 托管处理: 回应用户操作
// 如果用户已报听，等候3秒
// 用户第一次操作，需要等5秒
// 如果只有摸打，等1秒
// 有选择的情况下, 等5秒
func (room *Room) hostingOperation() {
	// 当前时间、等候间隔
	currentTime := util.GetTime()
	// 判断托管回应时间间隔
	if room.MI.getReplyInitTime() == 0 {
		return
	}
	// 开局5秒内不进行托管
	if currentTime-room.MI.getRoundCreateTime() < config.HOSTING_OPERATION_WAIT_TIME_5 {
		return
	}
	// 遍历用户当前的操作，根据操作来决定最终该等候多少秒
	for k, userId := range room.HostingUsers {
		timeInterval := int64(0)
		waitInfo := room.MI.getWait(userId)
		// 如果用户无可进行的操作，或者已经回复，则跳过
		if waitInfo == nil || waitInfo.ReplyTime > 0 {
			continue
		}

		// 选择等候时长
		if waitInfo.hasSelect() {
			timeInterval = config.HOSTING_OPERATION_WAIT_TIME_3
		} else {
			timeInterval = config.HOSTING_OPERATION_WAIT_TIME_1
		}

		// 报听状态，如果没有决策，让报听程序自动出牌
		// 有决策，也是等3秒后操作
		mu := room.MI.getUsers()[userId]
		if mu.MTC.IsBaoTing() && timeInterval < config.HOSTING_OPERATION_WAIT_TIME_3 {
			timeInterval = config.HOSTING_OPERATION_WAIT_TIME_3
		}

		if currentTime-room.MI.getReplyInitTime() < timeInterval {
			break
		}

		handTiles := mu.HandTileList.ToSlice()
		lack := mu.LackTile
		// 读取用户最后抓的牌
		lastDrawTile := 0
		lastOperator := room.MI.getLastOperator()
		lastOperation := room.MI.getLastOperation()

		if lastOperator == userId && oc.IsDrawOperation(lastOperation.OperationCode) {
			lastDrawTile = mu.HandTileList.GetLastAdd()
		}

		// 帮用户选择一个操作, 若选择失败，则跳过
		operation := room.selectHostingOperation(userId, waitInfo, handTiles, lack, lastDrawTile)
		if operation.OperationCode == 0 {
			continue
		}

		// 执行此操作
		err := handlerUserReply(room, userId, operation)
		if err != nil {
			core.Logger.Error("[hostingOperation]Error,roomId:%v, userId:%v, operation:%+v", room.RoomId, userId, operation)
			continue
		}

		core.Logger.Info("[hostingOperation]roomId:%v, userId:%v, operation:%+v", room.RoomId, userId, operation)

		// 多成员的时候，每个等候1秒
		if k != len(room.HostingUsers)-1 {
			room.hostingWait()
		}
	}
}

// 当多人同时托管时，可能有并发操作，为了避免可能出现的问题，每个操作时间，有个时间间隔
// 暂定 0.5秒
func (room *Room) hostingWait() {
	time.Sleep(time.Second)
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 用户托管状态切换
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func UserHosting(userId, hostingStatus int) *core.Error {
	// 判断用户是否已连接
	_, room, err := getUserRoom(userId)
	if err != nil {
		return err
	}

	// 判断房间状态
	if !room.CheckStatus(config.ROOM_STATUS_PALYING) {
		return core.NewError(-323, room.RoomId, userId)
	}

	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 设置用户托管状态
	if hostingStatus == config.ROOM_USER_HOSTING_YES {
		if !room.userInHosting(userId) {
			room.HostingUsers = append(room.HostingUsers, userId)
		}
	} else {
		if room.userInHosting(userId) {
			room.HostingUsers = util.SliceDel(room.HostingUsers, userId)
		}
	}

	// 推送消息给用户
	// user.AppendMessage(GameHostingPush(userId, hostingStatus))
	room.SendMessageToRoomUser(GameHostingPush(userId, hostingStatus), 0)

	core.Logger.Info("[UserHosting]roomId:%v, userId:%v, hostingStatus:%v", room.RoomId, userId, hostingStatus)

	return nil
}

// 帮用户选择一个操作
func (room *Room) selectHostingOperation(userId int, waitInfo *WaitInfo, handTiles []int, lack, lastDrawTile int) *Operation {
	var operation *Operation
	var hasPlay, hasPass, hasBaoTing bool
	// 推荐出牌
	var suggusetTile int

	// 第一趟运算，有优先级最高的操作，直接选择
	// 如果没有，记录出一些后面需要的变量
	for _, op := range waitInfo.OpList {
		if oc.IsWinOperation(op.OperationCode) {
			// 可以胡的话，直接帮用户胡牌
			operation = NewOperation(op.OperationCode, nil)
			break
		} else if op.OperationCode == fbsCommon.OperationCodeNEED_LACK_TILE {
			// 定缺, 帮用户选择最少一门，如果两门人数相同，则按照万=>条=>筒的顺序选择
			// 有定缺的时候，无需考虑其他动作
			lack := GetSuggestLack(handTiles)
			operation = NewOperation(op.OperationCode, []int{lack})
			break
		} else if op.OperationCode == fbsCommon.OperationCodeNEED_EXCHANGE_TILE {
			// 换牌，自动选服务器推荐的牌
			operation = NewOperation(fbsCommon.OperationCodeNEED_EXCHANGE_TILE, op.Tiles)
			break
		} else if op.OperationCode == fbsCommon.OperationCodePLAY {
			hasPlay = true
		} else if op.OperationCode == fbsCommon.OperationCodePASS {
			hasPass = true
		} else if op.OperationCode == fbsCommon.OperationCodeBAO_TING {
			hasBaoTing = true
		} else if op.OperationCode == fbsCommon.OperationCodePLAY_SUGGEST || op.OperationCode == fbsCommon.OperationCodeROBOT_PLAY_SUGGEST {
			suggusetTile = op.Tiles[0]
		}
	}

	// 如果有报听，选择报听牌最多的那张
	if operation == nil && hasBaoTing {
		baoTingTile := 0
		baoTingLen := 0
		for _, op := range waitInfo.OpList {
			if op.OperationCode == fbsCommon.OperationCodeBAO_TING {
				if len(op.Tiles) > baoTingLen {
					baoTingTile = op.Tiles[0]
					baoTingLen = len(op.Tiles)
				}
			}
		}
		operation = NewOperation(fbsCommon.OperationCodeBAO_TING, []int{baoTingTile})
	}

	// 如果用户已经听牌
	// 那么有出牌请求，直接让出牌
	// 其他操作，直接让pass
	if operation == nil && hostingUserIsTing(handTiles) {
		if hasPlay {
			operation = NewOperation(fbsCommon.OperationCodePLAY, []int{suggusetTile})
		} else if hasPass {
			operation = NewOperation(fbsCommon.OperationCodePASS, nil)
		}
	}

	// 高级的AI用这段逻辑
	// 第二趟循环，确认是否进行吃碰杠的操作
	/* 注释留着备用
	if operation == nil {
		mselector := room.MI.getSelector()
		// 计算当前牌型的牌阶、最大一类有效牌
		currentStep, _ := mselector.GetEffects(handTiles)

		for _, op := range waitInfo.OpList {
			step := -2
			tiles := util.SliceCopy(handTiles)
			switch op.OperationCode {
			case fbsCommon.OperationCodePONG:
				step, _ = mselector.GetEffects(util.SliceDel(tiles, op.Tiles[0], op.Tiles[0]))
			case fbsCommon.OperationCodeKONG:
				step, _ = mselector.GetEffects(util.SliceDel(tiles, op.Tiles[0], op.Tiles[0], op.Tiles[0]))
			case fbsCommon.OperationCodeKONG_DARK:
				step, _ = mselector.GetEffects(util.SliceDel(tiles, op.Tiles[0], op.Tiles[0], op.Tiles[0], op.Tiles[0]))
			case fbsCommon.OperationCodeKONG_TURN:
				fallthrough
			case fbsCommon.OperationCodeKONG_TURN_FREE:
				step, _ = mselector.GetEffects(util.SliceDel(tiles, op.Tiles[0]))
			default:
			}
			if step >= currentStep-1 {
				operation = op
				break
			}
		}
	}
	*/

	if operation == nil {
		for _, op := range waitInfo.OpList {
			switch op.OperationCode {
			case fbsCommon.OperationCodePONG:
				fallthrough
			case fbsCommon.OperationCodeKONG:
				fallthrough
			case fbsCommon.OperationCodeKONG_DARK:
				fallthrough
			case fbsCommon.OperationCodeKONG_TURN:
				fallthrough
			case fbsCommon.OperationCodeKONG_TURN_FREE:
				operation = op
				break
			default:
			}
		}
	}

	// 最后，再选择是过还是出牌
	if operation == nil {
		if hasPlay {
			operation = NewOperation(fbsCommon.OperationCodePLAY, []int{suggusetTile})
		} else if hasPass {
			operation = NewOperation(fbsCommon.OperationCodePASS, nil)
		}
	}

	core.Logger.Info("[selectHostingOperation]roomId:%v, round:%v, userId:%v, waitinfo:%+v, handTiles:%v, lack:%v, operation:%+v", room.RoomId, room.Round, userId, waitInfo, handTiles, lack, operation)

	return operation
}

func hostingUserIsTing(handTiles []int) bool {
	if util.IntInSlice(len(handTiles), []int{2, 5, 8, 11, 14}) {
		return len(ting.GetTingMap(handTiles, nil)) > 0
	}
	isTing, _ := ting.CanTing(handTiles, nil)
	return isTing
}
