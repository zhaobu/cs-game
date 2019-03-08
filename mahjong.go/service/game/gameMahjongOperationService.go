package game

import (
	"fmt"
	"reflect"
	"sort"
	"sync"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/mi/card"
	"mahjong.go/mi/oc"
	"mahjong.go/mi/ting"
	configService "mahjong.go/service/config"

	fbsCommon "mahjong.go/fbs/Common"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 结构定义
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 操作
type Operation struct {
	OperationCode int   // 操作码
	Tiles         []int // 操作的牌
}

// 用户操作
type UserOperation struct {
	UserId int
	Op     *Operation
}

// 系统操作
type ClientOperation struct {
	UserId int
	Op     *Operation
}

// 用户当前可进行的操作容器
type WaitInfo struct {
	// 用户可进行的操作
	OpList []*Operation

	// 用户回应的动作
	Reply *Operation

	// 回应时间
	ReplyTime int64
}

// WaitMap 所有用户当前可进行的操作信息
type WaitMap struct {
	Maps *sync.Map
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 结构初始化
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// NewOperation 生成一个新的操作结构
func NewOperation(opCode int, tiles []int) *Operation {
	if tiles == nil {
		tiles = []int{0}
	}
	return &Operation{opCode, tiles}
}

// NewUserOperation 生成一个新的用户操作结构
func NewUserOperation(userId int, op *Operation) *UserOperation {
	return &UserOperation{userId, op}
}

// NewClientOperation 生成一个新的系统操作结构
func NewClientOperation(userId int, op *Operation) *ClientOperation {
	return &ClientOperation{userId, op}
}

// NewWaitInfo 生成一个新的WaitInfo
func NewWaitInfo(opList []*Operation) *WaitInfo {
	return &WaitInfo{OpList: opList}
}

// NewWaitMap 生成一个新的WaitInfo
func NewWaitMap() *WaitMap {
	return &WaitMap{
		Maps: &sync.Map{},
	}
}

// 取长度
func (wm *WaitMap) Len() int {
	return util.SMapLen(wm.Maps)
}

// GetWaitInfo 取信息
func (wm *WaitMap) GetWaitInfo(userId int) *WaitInfo {
	if v, ok := wm.Maps.Load(userId); ok {
		return v.(*WaitInfo)
	}
	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 能进行的操作计算汇总
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 计算用户可进行的操作
func (this *Mahjong) calcOperation(userId int, cType int, tile int) []*Operation {
	// 可进行的操作列表
	var opList = []*Operation{}

	switch cType {
	case config.MAHJONG_OPERATION_CALC_PLAY:
		// 用户出牌后，其他人能做什么操作
		opList = this.calcOperationAfterPlay(userId, tile)
	case config.MAHJONG_OPERATION_CALC_KONG_TURN:
		// 用户转弯杠之后，其他人能做什么
		opList = this.calcOperationAfterKongTurn(userId, tile)
	case config.MAHJONG_OPERATION_CALC_DROW:
		// 计算用户抓牌后自己能进行什么操作
		opList = this.calcOperationAfterDraw(userId, tile)
	case config.MAHJONG_OPERATION_CALC_BEFORE_DRAW:
		// 计算用户抓牌前自己能进行什么操作
		opList = this.calcOperationBeforeDraw(userId)
	default:
	}

	core.Logger.Debug("[calcOperation]roomId:%v,round:%v,userId:%v,cType:%v,tile:%v,oplist:%+v", this.RoomId, this.Round, userId, cType, tile, opList)
	return opList
}

// 计算用户抓牌后或者庄家刚刚初始化的时候，能做什么操作
func (this *Mahjong) calcOperationAfterDraw(userId int, tile int) []*Operation {
	// 可进行的操作列表
	opList := []*Operation{}
	// 被检查操作的用户
	user := this.getUser(userId)

	if !user.hasLackTile() {
		// 如果用户无缺的话，检测能不能自摸
		opList, _ = this.canWinSelf(opList, user, tile)
	}
	if user.MTC.IsBaoTing() {
		// 如果玩法支持报听状态下暗杠的话，需要检查是否可以暗杠
		if this.setting.EnableBaoTingKong {
			if user.HandTileList.GetTileCnt(tile) == 4 {
				currentTingTiles := user.MTC.GetTingTiles()
				remainTiles := util.SliceDel(user.HandTileList.ToSlice(), tile, tile, tile, tile)
				_, nowTingTiles := ting.CanTing(remainTiles, nil)
				sort.Ints(currentTingTiles)
				sort.Ints(nowTingTiles)
				if len(currentTingTiles) > 0 && reflect.DeepEqual(currentTingTiles, nowTingTiles) {
					opList = append(opList, NewOperation(fbsCommon.OperationCodeKONG_DARK, []int{tile}))
				}
			}
		}
		if len(opList) > 0 {
			// 如果用户处于报听状态且有其他操作，给用户添加一个pass
			opList = append(opList, NewOperation(fbsCommon.OperationCodePASS, []int{0}))
		}
	} else {
		// 检查暗杠
		opList, _ = canKongDark(opList, user.HandTileList, this.getLackTile(userId))
		// 检查是否可以转弯杠、憨包杠
		opList, _ = canKongTurn(opList, user.ShowCardList.GetAll(), user.HandTileList)

		lackCnt := 0
		canTing := false
		if user.LackTile > 0 {
			lackCnt = getLackCount(user.LackTile, this.getUser(userId).HandTileList.ToSlice())
		}
		if lackCnt <= 1 {
			// 检查报听
			opList, canTing = this.canTingOrBaoTing(opList, user, user.LackTile)
		}
		// 如果真实用户，且听牌了，则给机器人的推荐code
		suggestCode := fbsCommon.OperationCodePLAY_SUGGEST
		if canTing || configService.IsRobot(user.UserId) {
			suggestCode = fbsCommon.OperationCodeROBOT_PLAY_SUGGEST
		}
		opList = this.getSuggestPlayTile(opList, user, suggestCode)
		// 根据操作列表，酌情给用户添加pass_cancel操作所
		opList = canPassCancel(opList)
		// 给用户添加一个出牌操作
		opList = append(opList, NewOperation(fbsCommon.OperationCodePLAY, []int{0}))
	}

	core.Logger.Debug("[ting_draw]roomId:%d, round:%d, userId:%d, ting map%#v, ready:%d", this.RoomId, this.Round, user.UserId, user.MTC.GetMaps(), user.MTC.GetStatus())
	return opList
}

// 计算用户转弯杠之后，其他人能做什么
func (this *Mahjong) calcOperationAfterKongTurn(userId int, tile int) []*Operation {
	// 可进行的操作列表
	opList := []*Operation{}
	// 被检查操作的用户
	user := this.getUser(userId)
	// 目前正在操作的用户
	player := this.getUser(this.LastOperator)
	// 计算别人是否可以截胡
	opList, _ = this.canWin(opList, user, player, tile, this.LastOperation.OperationCode, false, this.setting.IsEnablePinghu())

	// 抢杠必胡，不再添加pass
	/*
		// 添加一个pass
		if len(opList) > 0 {
			opList = append(opList, NewOperation(fbsCommon.OperationCodePASS, nil))
		}
	*/
	return opList
}

// 计算用户出牌后，其他人能做什么
// 部分麻将类型需要计算屁胡
func (this *Mahjong) calcOperationAfterPlay(userId int, tile int) []*Operation {
	// 可进行的操作列表
	var opList = []*Operation{}

	// 如果打的是缺，无任何可进行操作
	if this.tileIsQue(userId, tile) {
		return opList
	}
	// 被检查操作的用户
	user := this.getUser(userId)
	// 打牌用户
	player := this.getUser(this.LastOperator)

	if !user.MTC.IsBaoTing() {
		// 计算能不能明杠
		opList, _ = canKong(opList, user.HandTileList, tile)
		// 计算能不能碰
		opList, _ = user.canPong(opList, user.HandTileList, tile)
	} else {
		// 必须要杠之前和杠之后听的牌是一样的，才可以杠
		if this.setting.EnableBaoTingKong &&
			user.HandTileList.GetTileCnt(tile) == 3 {
			currentTingTiles := user.MTC.GetTingTiles()
			remainTiles := util.SliceDel(user.HandTileList.ToSlice(), tile, tile, tile)
			_, nowTingTiles := ting.CanTing(remainTiles, nil)
			sort.Ints(currentTingTiles)
			sort.Ints(nowTingTiles)
			if len(currentTingTiles) > 0 && reflect.DeepEqual(currentTingTiles, nowTingTiles) {
				opList = append(opList, NewOperation(fbsCommon.OperationCodeKONG, []int{tile}))
			}
		}
	}
	// 计算能不能胡
	repaoFlag := player.DrowAfterKongFlag
	if repaoFlag {
		// 明杠是否算热炮
		if this.setting.EnableKongHotPao == false && player.KongCode == fbsCommon.OperationCodeKONG {
			repaoFlag = false
		}
	}
	winFlag := false
	opList, winFlag = this.canWin(opList, user, player, tile, this.LastOperation.OperationCode, repaoFlag, this.setting.IsEnablePinghu())

	// 添加一个pass
	// 热炮必胡
	if len(opList) > 0 &&
		!(winFlag && repaoFlag) {
		opList = append(opList, NewOperation(fbsCommon.OperationCodePASS, []int{0}))
	}
	return opList
}

// 抓牌前能进行什么操作
// 补花或者暗杠红中
func (m *Mahjong) calcOperationBeforeDraw(userId int) []*Operation {
	// 可进行的操作列表
	var opList = []*Operation{}
	mu := m.getUser(userId)
	flowerRedCnt := mu.HandTileList.GetTileCnt(card.MAHJONG_RED_FLOWER)
	if flowerRedCnt == 4 {
		opList = append(opList, NewOperation(fbsCommon.OperationCodeKONG_DARK, []int{card.MAHJONG_RED_FLOWER}))
		// 添加一个pss
		opList = append(opList, NewOperation(fbsCommon.OperationCodePASS, []int{0}))
	} else if flowerRedCnt > 0 {
		// 执行补花操作
		m.flowerExchange(mu, false)
	}
	return opList
}

// 计算用户操作后，其他人能做什么操作
// 如果其他人不能进行操作，则继续执行run函数，给下一个人发牌
func (this *Mahjong) calcAfterUserOperation(cType int) bool {
	// 是否有其他人操作
	var noWait = true
	// 是否有胡的操作
	var hasHu = false

	for _, userId := range this.Index {
		// 跳过用户自己
		if userId == this.LastOperator {
			continue
		}
		// 最后操作的牌
		var tile int
		if this.LastOperation.Tiles == nil {
			tile = 0
		} else {
			tile = this.LastOperation.Tiles[0]
		}
		var opList []*Operation = this.calcOperation(userId, cType, tile)
		if len(opList) > 0 {
			noWait = false
			// 添加到waitqueue
			waitInfo := NewWaitInfo(opList)
			this.setWait(userId, waitInfo)
			if waitInfo.hasHu() {
				hasHu = true
			}
			// 推送操作通知
			this.getUser(userId).SendOperationPush(opList)
		}
	}
	// fixme，这段逻辑放在这里不合适，需要优化
	// 检查并设置上下级
	if !hasHu {
		this.setChikenRock()
	}
	return noWait
}

// 检查用户是否可以进行这个操作
func (this *Mahjong) checkUserOperation(userId int, op *Operation) *core.Error {
	// 检测用户是否可以进行操作
	waitInfo := this.WaitQueue.GetWaitInfo(userId)
	if waitInfo == nil {
		return core.NewError(-313, userId, this.RoomId, op.OperationCode, op.Tiles)
	}
	// 检测是否已经回应过了
	if waitInfo.Reply != nil {
		return core.NewError(-315, userId, this.RoomId, op.OperationCode)
	}

	// 收集选择的操作列表
	opList := []*Operation{}
	for _, v := range waitInfo.OpList {
		if v.OperationCode == op.OperationCode {
			opList = append(opList, v)
		}
	}

	// 不存在这个操作
	if len(opList) == 0 {
		return core.NewError(-314, userId, this.RoomId, op.OperationCode)
	}

	// 用户数据
	u := this.getUser(userId)
	errorFlag := false
	switch op.OperationCode {
	case fbsCommon.OperationCodePLAY: // 判断用户有没有这张手牌
		if u.HandTileList.GetTileCnt(op.Tiles[0]) == 0 {
			errorFlag = true
		}
		/*
			// 如果用户处于报听状态，必须打上一次抓的牌
			if u.MTC.IsBaoTing() {
				if op.Tiles[0] != this.LastOperation.Tiles[0] {
					errorFlag = true
				}
			}
		*/
		// 如果有缺，必须先打缺的那一门
		var count = getLackCount(this.getLackTile(userId), u.HandTileList.ToSlice())
		if !this.tileIsQue(userId, op.Tiles[0]) && count > 0 {
			errorFlag = true
		}
	case fbsCommon.OperationCodeBAO_TING: // 判断是不是打了可以报听的牌
		fallthrough
	case fbsCommon.OperationCodeKONG_DARK: // 判断暗杠了可以暗杠的牌
		errorFlag = true
		for _, v := range opList {
			if op.Tiles[0] == v.Tiles[0] {
				errorFlag = false
				break
			}
		}
	case fbsCommon.OperationCodeNEED_LACK_TILE: // 定缺，需要验证选牌的范围(必须万、条、筒)
		tile := op.Tiles[0]
		if !card.IsCrak(tile) && !card.IsBAM(tile) && !card.IsDot(tile) {
			errorFlag = true
		}
	// 以下这些是支持的、且没有额外验证逻辑的用户操作
	case fbsCommon.OperationCodeNEED_EXCHANGE_TILE:
		// todo 判断手牌是否有这些
	case fbsCommon.OperationCodeCHOW:
	case fbsCommon.OperationCodePONG:
	case fbsCommon.OperationCodeKONG:
	case fbsCommon.OperationCodeKONG_TURN:
	case fbsCommon.OperationCodeKONG_TURN_FREE:
	case fbsCommon.OperationCodeWIN:
	case fbsCommon.OperationCodeWIN_AFTER_KONG_TURN:
	case fbsCommon.OperationCodeWIN_SELF:
	case fbsCommon.OperationCodeWIN_AFTER_KONG_DRAW:
	case fbsCommon.OperationCodeWIN_AFTER_KONG_PLAY:
	case fbsCommon.OperationCodePASS:
	case fbsCommon.OperationCodePASS_CANCEL: // 用户不能选择取消操作
	// 不支持的用户操作，直接报错
	// return core.NewError(-316, userId, this.RoomId, op.OperationCode)
	default:
		return core.NewError(-317, userId, this.RoomId, op.OperationCode)
	}

	// 操作检测不通过
	if errorFlag == true {
		return core.NewError(-318, userId, this.RoomId, op.OperationCode)
	}
	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 服务器操作
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 补花
// 支持是否跳过补花
func (m *Mahjong) flowerExchange(mu *MahjongUser, skipKong bool) {
	// 如果抓到了补花的牌，则抓不到牌，认为牌局已结束
	isFinish := false
	for {
		tileCnt := mu.HandTileList.GetTileCnt(card.MAHJONG_RED_FLOWER)
		if tileCnt == 0 || (skipKong && tileCnt == 4) {
			break
		}
		// 删除用户补花的牌
		mu.HandTileList.DelTile(card.MAHJONG_RED_FLOWER, tileCnt)
		// 添加用户明牌
		mu.ShowCardList.AddFlower(mu.UserId, card.MAHJONG_RED_FLOWER, tileCnt)
		// 需要考虑一下最后一张，正好抓到花的情况，这时候无花可补
		opList := make([]*Operation, 0)
		showTiles := make([]int, 2*tileCnt)
		hideTiles := make([]int, 2*tileCnt)
		for i := 0; i < tileCnt; i++ {
			showTiles[2*i] = card.MAHJONG_RED_FLOWER
			hideTiles[2*i] = card.MAHJONG_RED_FLOWER
			// 已抓完
			if !m.TileWall.IsAllDrawn() {
				// 从后面抓一张牌
				tile := m.TileWall.ForwardDraw()
				// 给用户增加手牌
				mu.HandTileList.AddTile(tile, 1)
				showTiles[2*i+1] = tile
			} else {
				isFinish = true
			}
		}
		core.Logger.Debug("补花之后用户的手牌,roomId:%v,userId:%v,tiles:%v", m.RoomId, mu.UserId, mu.HandTileList.ToSlice())
		core.Logger.Infof("[flowerExchange]roomId:%v,round:%v,userId:%v,isFinish:%v,tiles:%v", m.RoomId, m.Round, mu.UserId, isFinish, showTiles)

		// 消息推送
		showOperation := NewOperation(fbsCommon.OperationCodeFLOWER_CHANGE, showTiles)
		hideOperation := NewOperation(fbsCommon.OperationCodeFLOWER_CHANGE, hideTiles)
		// 给用户发送补花消息
		opList = append(opList, showOperation)
		// 给用户自己发operationPush(FlowerExchange, tile, exchangeTile...)
		mu.SendOperationPush(opList)

		// 给其他用户发userOperationPush(userId, FlowerExchange, tile, tileCnt)
		m.SendUserOperationPush(NewUserOperation(mu.UserId, hideOperation), mu.UserId)
		showUserOperation := NewUserOperation(mu.UserId, showOperation)

		// 推送用户操作给观察者
		m.Ob.sendMessage(UserOperationPush(showUserOperation), 0)
		// 记录回放：补花
		m.playback.appendUserOperation(showUserOperation)

		// 如果牌已经抓完了，则跳出循环
		if isFinish {
			break
		}
	}

	// 更新用户的听牌状态
	if !mu.MTC.IsBaoTing() && !mu.HandTileList.IsPlayStatus() {
		m.setTingAfterKong(mu)
	}

	// 牌不够抓了，直接结束游戏
	if isFinish {
		m.next()
	}
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 用户操作
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 换牌
func (m *Mahjong) userOperationExchange(userId int, tiles []int) (bool, *core.Error) {
	core.Logger.Debug("[userOperationExchange]userId:%v, tile:%v", userId, tiles)
	userIndex := m.getUser(userId).Index
	m.ExchangeList.Store(userIndex, tiles)
	// 回放：添加换牌操作
	userOperation := NewUserOperation(userId, NewOperation(fbsCommon.OperationCodeEXCHANGE_TILE_OK, tiles))
	m.playback.appendUserOperation(userOperation)
	// 推送换牌操作给观察者
	m.Ob.sendMessage(UserOperationPush(userOperation), 0)
	// 未登记所有人的换牌结果，等候下一个操作
	if util.SMapLen(m.ExchangeList) < util.SMapLen(m.Users) {
		return false, nil
	}
	// 所有人都已经选择了换牌，执行换牌操作
	var directions []int
	// 计算换牌的方向，非偶数人数，不支持和对面换牌
	if m.setting.GetSettingPlayerCnt()%2 == 0 {
		directions = []int{config.EXCHANGE_DIRECTION_OPPOSITE, config.EXCHANGE_DIRECTION_CLOCKWISE, config.EXCHANGE_DIRECTION_COUNTERCLOCKWISE}
	} else {
		directions = []int{config.EXCHANGE_DIRECTION_CLOCKWISE, config.EXCHANGE_DIRECTION_COUNTERCLOCKWISE}
	}

	// 随机一个换牌顺序
	direction := directions[util.RandIntn(len(directions))]
	core.Logger.Debug("[userOperationExchange]roomId:%v, round:%v, direction:%v", m.RoomId, m.Round, direction)

	m.Users.Range(func(k, v interface{}) bool {
		mu := v.(*MahjongUser)

		// 执行换牌
		// 计算换牌对象
		var exchangeIndex int
		exchangeIndex = m.getExchangeTargetIndex(direction, mu.Index)
		target := m.getUser(m.Index[exchangeIndex])

		// 获取用户换到的牌
		inTiles, _ := m.ExchangeList.Load(exchangeIndex)
		mu.ExchangeInTiles = inTiles.([]int)

		// 删除用户要换的手牌
		for _, tile := range mu.ExchangeOutTiles {
			mu.HandTileList.DelTile(tile, 1)
		}
		// 添加用户换过来的牌
		for _, tile := range mu.ExchangeInTiles {
			mu.HandTileList.AddTile(tile, 1)
		}

		// 给用户发布换牌消息
		tiles := append(append([]int{direction}, mu.ExchangeOutTiles...), mu.ExchangeInTiles...)
		userOperation := NewUserOperation(mu.UserId, NewOperation(fbsCommon.OperationCodeEXCHANGE_TILE_RESULT, tiles))
		mu.SendUserOperationPush(userOperation)

		// 添加回放
		m.playback.appendUserOperation(userOperation)

		core.Logger.Debug("[userOperationExchange]roomId:%v, round:%v, userId:%v, index:%v, target:%v, target index:%v, from tiles:%v, to tiles:%v, handTiles:%v",
			m.RoomId, m.Round, mu.UserId, mu.Index, target.UserId, exchangeIndex, mu.ExchangeOutTiles, mu.ExchangeInTiles, mu.HandTileList.ToSlice())

		return true
	})

	return true, nil
}

// 定缺
func (this *Mahjong) userOperationLack(userId int, tile int) bool {
	userIndex := this.getUser(userId).Index
	if _, ok := this.LackList.Load(userIndex); ok {
		// 已经回应过定缺了
		return false
	}
	this.LackList.Store(userIndex, tile)
	// 回放：添加定缺操作
	userOperation := NewUserOperation(userId, NewOperation(fbsCommon.OperationCodeLACK_TILE_RESULT, []int{tile}))
	this.playback.appendUserOperation(userOperation)
	// 推送定缺消息消息给观察者
	this.Ob.sendMessage(UserOperationPush(userOperation), 0)
	if util.SMapLen(this.LackList) < util.SMapLen(this.Users) {
		return false
	}
	// 给房间其他用户发送用户已定缺的通知
	// 未全员定缺，需要通知其他人，自己已定缺
	this.SendUserOperationPush(NewUserOperation(userId, NewOperation(fbsCommon.OperationCodeLACK_TILE_OK, nil)), userId)
	// 如果已经全员选择了，需要通知所有人定缺情况并发送push
	// 发送定缺结果，这里因为客户端显示逻辑的原因，始终要把用户自己的消息放在第一位
	for k, userId := range this.Index {
		// 先发用户自己的
		lactTile, _ := this.LackList.Load(k)
		userOperation := NewUserOperation(userId, NewOperation(fbsCommon.OperationCodeLACK_TILE_RESULT, []int{lactTile.(int)}))
		this.getUser(userId).SendUserOperationPush(userOperation)
	}
	// 将用户的定缺结果发给其他人，不包括自己
	this.LackList.Range(func(k, v interface{}) bool {
		index := k.(int)
		lactTile := v.(int)
		userOperation := NewUserOperation(this.Index[index], NewOperation(fbsCommon.OperationCodeLACK_TILE_RESULT, []int{lactTile}))
		this.SendUserOperationPush(userOperation, this.Index[index])
		return true
	})
	core.Logger.Info("[userOperationLack]roomId:%d, userId:%d, tile:%d", this.RoomId, userId, tile)
	return true
}

// 吃
func (this *Mahjong) userOperationChow(userId int, tiles []int) (bool, *core.Error) {
	return true, nil
}

// 碰
func (this *Mahjong) userOperationPong(userId int) *core.Error {
	// 碰牌前，需要判断是否需要变鸡
	this.setChikenRock()
	// 读取上次用户操作的牌
	var tile = this.LastOperation.Tiles[0]
	if tile == 0 {
		return core.NewError(-553, userId)
	}
	core.Logger.Debug("需要碰的牌：%d", tile)
	mahjongUser := this.getUser(userId)
	lastOperatorUser := this.getUser(this.LastOperator)
	core.Logger.Debug("碰之前的手牌: %#v", mahjongUser.HandTileList.ToSlice())
	core.Logger.Debug("碰之前的明牌: %#v", mahjongUser.ShowCardList)
	core.Logger.Debug("碰之前打牌者的弃牌: %#v", lastOperatorUser.DiscardTileList.GetTiles())
	// 记录责任鸡相关参数
	responsibilityOperator := this.LastOperator
	responsibilityTile := this.LastOperation.Tiles[0]
	// 从手牌删除两张
	mahjongUser.HandTileList.DelTile(tile, 2)
	// 从打牌者的弃牌移除
	lastOperatorUser.DiscardTileList.DelLastTile()
	// 生成明牌
	var tiles = []int{tile, tile, tile}
	var showCard = card.NewShowCard(fbsCommon.OperationCodePONG, this.LastOperator, tiles, false)
	this.finishOperation(userId, fbsCommon.OperationCodePONG, []int{tile})
	// 设置责任鸡
	this.setChikenResponsibility(responsibilityOperator, responsibilityTile)
	// 添加到明牌
	if tile == card.MAHJONG_BAM1 && this.getChikenResponsibility() > 0 {
		showCard.SetResponsibility()
	}
	mahjongUser.ShowCardList.Add(showCard)
	core.Logger.Debug("碰之后的手牌: %#v", mahjongUser.HandTileList.ToSlice())
	for _, showCard := range mahjongUser.ShowCardList.GetAll() {
		core.Logger.Debug("碰之后的明牌: %v", showCard)
	}
	core.Logger.Debug("碰之后打牌者的弃牌: %#v", lastOperatorUser.DiscardTileList.GetTiles())
	core.Logger.Info("[userOperationPong]roomId:%d, userId:%d, tile:%d", this.RoomId, userId, tile)
	return nil
}

// 处理碰后操作，通知用户出牌
func (this *Mahjong) userOperationAfterPong(userId int) {
	mu := this.getUser(userId)
	// 碰牌之后，通知用户出牌
	opList := []*Operation{NewOperation(fbsCommon.OperationCodePLAY, []int{0})}
	// 手里有一张缺时，打掉缺有可能听
	lackCnt := 0
	canTing := false
	if mu.LackTile > 0 {
		lackCnt = getLackCount(mu.LackTile, mu.HandTileList.ToSlice())
	}
	if lackCnt <= 1 {
		opList, canTing = this.canTingOrBaoTing(opList, mu, mu.LackTile)
	}
	// 如果真实用户，且听牌了，则给机器人的推荐code
	suggestCode := fbsCommon.OperationCodePLAY_SUGGEST
	if canTing || configService.IsRobot(mu.UserId) {
		suggestCode = fbsCommon.OperationCodeROBOT_PLAY_SUGGEST
	}
	opList = this.getSuggestPlayTile(opList, mu, suggestCode)

	this.setWait(userId, NewWaitInfo(opList))
	mu.SendOperationPush(opList)
	core.Logger.Debug("[ting_pong]roomId:%d, round:%d, userId:%d, ting map%#v, ready:%d", this.RoomId, this.Round, userId, mu.MTC.GetMaps(), mu.MTC.GetStatus())
	core.Logger.Info("[userOperationAfterPong]roomId:%d, userId:%d", this.RoomId, userId)
}

// 明杠
func (this *Mahjong) userOperationKong(userId int) *core.Error {
	// 明杠前，需要判断是否需要变鸡
	this.setChikenRock()
	var user = this.getUser(userId)
	// 打出来的牌
	var tile = this.LastOperation.Tiles[0]
	// 添加明牌
	var tiles = []int{tile, tile, tile, tile}
	var showCard = card.NewShowCard(fbsCommon.OperationCodeKONG, this.LastOperator, tiles, false)
	if tile == card.MAHJONG_BAM1 {
		showCard.SetResponsibility()
	}
	user.ShowCardList.Add(showCard)
	// 删除手牌
	user.HandTileList.DelTile(tile, 3)
	// 移除last operator 的弃牌
	this.getUser(this.LastOperator).DiscardTileList.DelLastTile()
	// 设置责任鸡
	this.setChikenResponsibility(this.LastOperator, this.LastOperation.Tiles[0])
	// 更新听牌状态
	this.setTingAfterKong(user)
	this.finishOperation(userId, fbsCommon.OperationCodeKONG, []int{tile})
	core.Logger.Debug("[ting_kong]roomId:%d, round:%d, userId:%d, ting map:%#v, ready:%d", this.RoomId, this.Round, userId, user.MTC.GetMaps(), user.MTC.GetStatus())
	core.Logger.Info("[userOperationKong]roomId:%d, userId:%d, tile:%v", this.RoomId, userId, tile)
	return nil
}

// 暗杠
func (this *Mahjong) userOperationKongDark(userId int, tile int) *core.Error {
	var user = this.getUser(userId)
	//删除手牌
	user.HandTileList.DelTile(tile, 4)
	// 添加明牌
	var tiles = []int{tile, tile, tile, tile}
	var showCard = card.NewShowCard(fbsCommon.OperationCodeKONG_DARK, this.LastOperator, tiles, false)
	user.ShowCardList.Add(showCard)
	// 更新听牌状态
	this.setTingAfterKong(user)
	this.finishOperation(userId, fbsCommon.OperationCodeKONG_DARK, []int{tile})
	// 因为庄家第一手操作，可能是暗杠，暗杠后也需要发牌，所以如果有暗杠，需要将庄家首次操作标志设置为false
	this.firstOperateFlag = false
	core.Logger.Debug("[ting_kong_dark]roomId:%d, round:%d, userId:%d, ting map:%#v, ready:%d", this.RoomId, this.Round, userId, user.MTC.GetMaps(), user.MTC.GetStatus())
	core.Logger.Debug("用户手牌:%#v", user.HandTileList.ToSlice())
	core.Logger.Debug("用户明牌:%#v", user.ShowCardList.GetLast())
	core.Logger.Info("[userOperationKongDark]roomId:%d, userId:%d, tile:%v", this.RoomId, userId, tile)
	return nil
}

// 转弯杠或憨包杠
func (this *Mahjong) userOperationKongTurn(userId int, tile int, opCode int) *core.Error {
	user := this.getUser(userId)
	// 将碰牌转为转弯杠
	user.ShowCardList.SetPongToKongTurn(tile, opCode)
	// 删除手牌
	user.HandTileList.DelTile(tile, 1)
	// 更新用户听牌状态
	this.setTingAfterKong(user)
	this.finishOperation(userId, opCode, []int{tile})
	core.Logger.Debug("[ting_kong_turn]roomId:%d, round:%d, userId:%d, ting map:%#v, ready:%d", this.RoomId, this.Round, userId, user.MTC.GetMaps(), user.MTC.GetStatus())
	core.Logger.Info("[userOperationKongTurn]roomId:%d, userId:%d, tile:%v, opcode:%v", this.RoomId, userId, tile, opCode)
	return nil
}

// 报听
func (this *Mahjong) userOperationBaoTing(userId, tile int) *core.Error {
	mu := this.getUser(userId)
	mu.MTC.SetBaoTing(tile)
	// 删除用户手牌
	mu.HandTileList.DelTile(tile, 1)
	// 添加到弃牌中
	mu.DiscardTileList.AppendTile(tile)
	// 如果上次操作不是抓牌的话，更新用户的过胡、过碰状态
	if this.LastOperation.OperationCode != fbsCommon.OperationCodePLAY {
		mu.SkipWin = []int{}
		mu.SkipPong = []int{}
	}
	this.finishOperation(userId, fbsCommon.OperationCodeBAO_TING, []int{tile})
	core.Logger.Info("[userOperationBaoTing]roomId:%d, userId:%d, tile:%v", this.RoomId, userId, tile)
	return nil
}

// 胡
func (this *Mahjong) userOperationWin(userId int, opCode int) *core.Error {
	var user = this.getUser(userId)
	var tile = this.LastOperation.Tiles[0]
	var lastIndex = this.getUser(this.LastOperator).Index
	//从点炮者弃牌中拿出放入手牌中,一炮多响delLastDiscard 不做操作
	this.getUser(this.LastOperator).DiscardTileList.DelLastTile()
	user.HandTileList.AddTile(tile, 1)
	// 设置胡牌信息
	this.setHuInfo(user, opCode, tile)
	// 记录用户最后的牌
	user.WinTile = tile
	this.finishOperation(userId, opCode, []int{tile, lastIndex})
	core.Logger.Info("[userOperationWin]roomId:%d, userId:%d, opCode:%v", this.RoomId, userId, opCode)
	return nil
}

// 抢杠胡
func (this *Mahjong) userOperationWinAfterKongTurn(userId int) *core.Error {
	var user = this.getUser(userId)
	var tile = this.LastOperation.Tiles[0]
	var lastIndex = this.getUser(this.LastOperator).Index
	//从点炮者明牌中拿出放入手牌中,一炮多响delLastDiscard 不做操作
	this.getUser(this.LastOperator).ShowCardList.DelKongTurnTile(tile)
	user.HandTileList.AddTile(tile, 1)
	// 设置胡牌信息
	this.setHuInfo(user, fbsCommon.OperationCodeWIN_AFTER_KONG_TURN, tile)
	// 记录用户最后的牌
	user.WinTile = tile
	this.finishOperation(userId, fbsCommon.OperationCodeWIN_AFTER_KONG_TURN, []int{tile, lastIndex})
	core.Logger.Info("[userOperationWinAfterKongTurn]roomId:%d, userId:%d", this.RoomId, userId)
	return nil
}

// 自摸
func (this *Mahjong) userOperationWinSelf(userId int, opCode int) *core.Error {
	var user = this.getUser(userId)
	tile := user.HandTileList.GetLastAdd()
	// 设置胡牌信息
	this.setHuInfo(user, opCode, tile)
	// 记录用户最后的牌
	user.WinTile = tile
	this.finishOperation(userId, opCode, []int{tile})
	core.Logger.Info("[userOperationWinSelf]roomId:%d, userId:%d, opCode:%v", this.RoomId, userId, opCode)
	return nil
}

// 跳过
func (this *Mahjong) userOperationPass(userId int) int {
	// 给用户发一个过的消息
	op := NewOperation(fbsCommon.OperationCodePASS, nil)
	userOperation := NewUserOperation(userId, op)
	this.getUser(userId).SendUserOperationPush(userOperation)
	// 给观察者发一个过的消息
	this.Ob.sendMessage(UserOperationPush(userOperation), 0)
	// 记录到回放
	this.playback.appendUserOperation(userOperation)

	core.Logger.Info("[userOperationPass]roomId:%d, userId:%d, opCode:%v", this.RoomId, userId)
	return 0
}

// pass_cancel 操作
func (this *Mahjong) userOperationPassCancel(userId int) int {
	// 删除用户的可选操作
	opList := make([]*Operation, 0)
	for _, operation := range this.getWait(userId).OpList {
		// 跳过所有被pass_cancel的操作
		if !oc.IsPassCancelRemain(operation.OperationCode) {
			continue
		}
		opList = append(opList, operation)
	}
	this.overideWait(userId, NewWaitInfo(opList))

	// 给用户发一个PASS_CANCEL的消息
	op := NewOperation(fbsCommon.OperationCodePASS_CANCEL, nil)
	userOperation := NewUserOperation(userId, op)
	this.getUser(userId).SendUserOperationPush(userOperation)

	// 给观察者发一个过的消息
	this.Ob.sendMessage(UserOperationPush(userOperation), 0)

	// 记录到回放
	this.playback.appendUserOperation(userOperation)

	core.Logger.Info("[userOperationPassCancel]roomId:%d, userId:%d, opCode:%v", this.RoomId, userId)
	return 0
}

// 跳过之后补花
func (m *Mahjong) userOperationFlowerExchange(userId int) int {
	m.flowerExchange(m.getUser(userId), false)
	return 0
}

// 出牌
func (this *Mahjong) userOperationPlay(userId int, tile int) *core.Error {
	u := this.getUser(userId)
	core.Logger.Debug("打牌前, roomId:%v, round:%v, userId:%v, handTiles:%v, discardTiles:%v", this.RoomId, this.Round, userId, u.HandTileList.ToSlice(), u.DiscardTileList.GetTiles())
	if !u.MTC.IsBaoTing() {
		u.MTC.SetTingMap(tile)
	}
	core.Logger.Debug("[ting_play]roomId:%d, round:%d, userId:%d, ting map:%#v, ready:%d", this.RoomId, this.Round, userId, u.MTC.GetMaps(), u.MTC.GetStatus())

	// 删除用户手牌
	if !u.HandTileList.DelTile(tile, 1) {
		core.Logger.Warn("[userOperationPlay][DelTile]roomId:%v, round:%v, userId:%v, tile:%v", this.RoomId, this.Round, userId, tile)
	}
	// 添加到弃牌中
	u.DiscardTileList.AppendTile(tile)
	// 如果上次操作不是抓牌的话，更新用户的过胡、过碰状态
	if this.LastOperation.OperationCode != fbsCommon.OperationCodePLAY {
		u.SkipWin = []int{}
		u.SkipPong = []int{}
	}
	this.finishOperation(userId, fbsCommon.OperationCodePLAY, []int{tile})

	core.Logger.Debug("打牌后, roomId:%v, round:%v, userId:%v, playTile:%v, handTiles:%v, discardTiles:%v", this.RoomId, this.Round, userId, tile, u.HandTileList.ToSlice(), u.DiscardTileList.GetTiles())
	core.Logger.Info("[userOperationPlay]roomId:%d, userId:%d, tile: %d", this.RoomId, userId, tile)

	return nil
}

func (this *Mahjong) finishOperation(userId, opCode int, tiles []int) {
	// 更新最后操作者
	var op = NewOperation(opCode, tiles)
	if !oc.IsWinOperation(opCode) {
		// 胡牌时候，不改变最后操作者
		this.setLastOperate(userId, op)
	}

	// 回放：添加用户操作
	userOperation := NewUserOperation(userId, op)
	this.playback.appendUserOperation(userOperation)

	// 推送用户操作给观察者
	this.Ob.sendMessage(UserOperationPush(userOperation), 0)

	// 如果是抓牌，需要把抓到的牌替换掉
	var shielder = 0
	if oc.IsDrawOperation(opCode) {
		op = NewOperation(opCode, nil)
		shielder = userId
	}

	// 给用户发消息
	this.SendUserOperationPush(NewUserOperation(userId, op), shielder)
	// core.Logger.Info("[finishOperation]opCode:%d,roomId:%d, userId:%d, tile:%d", opCode, this.RoomId, userId, tiles)
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 操作列表相关
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 将wait信息抓成
func (waitInfo *WaitInfo) String() string {
	var s string
	for _, v := range waitInfo.OpList {
		s = s + fmt.Sprintf("[code:%d, tiles:%v]", v.OperationCode, v.Tiles)
	}
	return s
}

// 设置用户操作队列
func (this *Mahjong) setWait(userId int, waitInfo *WaitInfo) {
	// 防并发加锁
	this.Mux.Lock()
	defer this.Mux.Unlock()

	// 判断用户是否有未回应的操作
	if existsWaitInfo := this.WaitQueue.GetWaitInfo(userId); existsWaitInfo != nil {
		core.Logger.Warn("用户有尚未回应的操作,roomId:%v,round:%v,userId: %d, wait: %s", this.RoomId, this.Round, userId, existsWaitInfo.String())
	}

	this.WaitQueue.Maps.Store(userId, waitInfo)
	this.replayInitTime = util.GetTime()

	// 保存至回放数据
	this.playback.appendOperationPush(userId, waitInfo.OpList)

	core.Logger.Info("[setWait]roomId:%v,round:%v,userId:%d,wait:%s", this.RoomId, this.Round, userId, waitInfo.String())
}

// 覆盖用户操作队列
func (this *Mahjong) overideWait(userId int, waitInfo *WaitInfo) {
	// 防并发加锁
	this.Mux.Lock()
	defer this.Mux.Unlock()

	this.WaitQueue.Maps.Store(userId, waitInfo)
	this.replayInitTime = util.GetTime()

	core.Logger.Info("[overideWait]roomId:%v,round:%v,userId:%d,wait:%s", this.RoomId, this.Round, userId, waitInfo.String())
}

// 读取用户操作队列
func (this *Mahjong) getWait(userId int) *WaitInfo {
	// 防并发加锁
	this.Mux.Lock()
	defer this.Mux.Unlock()
	return this.WaitQueue.GetWaitInfo(userId)
}

// 读取所有未操作的用户id
func (this *Mahjong) getUnReplyWaitUsers() []int {
	// 防并发加锁
	this.Mux.Lock()
	defer this.Mux.Unlock()

	users := []int{}
	this.WaitQueue.Maps.Range(func(k, v interface{}) bool {
		waitInfo := v.(*WaitInfo)
		if waitInfo.ReplyTime == 0 {
			users = append(users, k.(int))
		}
		return true
	})
	return users
}

// 清空操作队列
func (this *Mahjong) cleanWait() {
	// 防并发加锁
	this.Mux.Lock()
	defer this.Mux.Unlock()
	this.WaitQueue.Maps.Range(func(k, v interface{}) bool {
		this.WaitQueue.Maps.Delete(k.(int))
		return true
	})
}

// 用户回应操作
func (this *Mahjong) replyWait(userId int, op *Operation) *WaitInfo {
	// 防并发加锁
	this.Mux.Lock()
	defer this.Mux.Unlock()

	// 判断是否有可供选择的决策
	waitInfo := this.WaitQueue.GetWaitInfo(userId)
	if waitInfo == nil {
		core.Logger.Warn("[replyWait]waitInfo is nil, roomId:%v, round:%v, userId:%v", this.RoomId, this.Round, userId)
		return nil
	}

	// 用户信息
	mu := this.getUser(userId)

	if op.OperationCode == fbsCommon.OperationCodePASS {
		// 如果用户选择了“过”、需要标记过胡、过碰状态
		if waitInfo.hasHu() {
			mu.SkipWin = append(mu.SkipWin, this.LastOperation.Tiles[0])
			core.Logger.Debug("append SkipWin, roomId:%v, round:%v, userId:%v, tile:%v", this.RoomId, this.Round, userId, this.LastOperation.Tiles[0])
		} else if waitInfo.hasPong() {
			mu.SkipPong = append(mu.SkipPong, this.LastOperation.Tiles[0])
			core.Logger.Debug("append SkipPong, roomId:%v, round:%v, userId:%v, tile:%v", this.RoomId, this.Round, userId, this.LastOperation.Tiles[0])
		}
	} else if op.OperationCode == fbsCommon.OperationCodeNEED_LACK_TILE {
		// 用户定缺之后，操作会在wait中，等候所有人定缺完成
		// 这时候如果用户重连，服务端不好知道用户是否已经点过重连
		// 所以加一个额外的逻辑，只要用户点击了定缺，就在用户属性中记录一下
		// 只用到重连中，不做任何其他的逻辑
		mu.LackTile = op.Tiles[0]

		// 用户处理了定缺，需要通知用户自己和其他人，用户已定缺
		// 给用户自己发的消息
		userOperation := NewUserOperation(userId, NewOperation(fbsCommon.OperationCodeLACK_TILE_RESULT, op.Tiles))
		// 给其他人发的消息
		otherOperation := NewUserOperation(userId, NewOperation(fbsCommon.OperationCodeLACK_TILE_OK, nil))
		this.Users.Range(func(k, v interface{}) bool {
			_mu := v.(*MahjongUser)
			if _mu.UserId == userId {
				_mu.SendUserOperationPush(userOperation)
			} else {
				_mu.SendUserOperationPush(otherOperation)
			}
			return true
		})
	} else if op.OperationCode == fbsCommon.OperationCodeNEED_EXCHANGE_TILE {
		// 换牌和定缺一样的逻辑
		mu.ExchangeOutTiles = op.Tiles

		// 通知其他人，已换牌, 换几张就给几张空白的牌
		emptyTiles := make([]int, len(op.Tiles))

		// 给用户自己发的消息
		userOperation := NewUserOperation(userId, NewOperation(fbsCommon.OperationCodeEXCHANGE_TILE_OK, mu.ExchangeOutTiles))
		// 给其他人发的消息
		otherOperation := NewUserOperation(userId, NewOperation(fbsCommon.OperationCodeEXCHANGE_TILE_OK, emptyTiles))
		this.Users.Range(func(k, v interface{}) bool {
			_mu := v.(*MahjongUser)
			if _mu.UserId == userId {
				_mu.SendUserOperationPush(userOperation)
			} else {
				_mu.SendUserOperationPush(otherOperation)
			}
			return true
		})

		core.Logger.Debug("[replywait]用户选择了换牌, roomId:%v, round:%v, tiles:%v", this.RoomId, this.Round, op.Tiles)
	}

	// 如果用户选择了“关闭决策框”，需要删除waitInfo中数据选择的决策，然后等候用户下一步的选择，这时候不能设置reply
	if op.OperationCode == fbsCommon.OperationCodePASS_CANCEL {
		// do nothing
	} else {
		waitInfo.Reply = op
		waitInfo.ReplyTime = util.GetTime()
		mu.ReplyTimeCnt += int(util.GetTime() - this.replayInitTime)
	}

	return waitInfo
}

// 是否已全部进行回应
func (this *Mahjong) isAllReply() bool {
	// 防并发加锁
	this.Mux.Lock()
	defer this.Mux.Unlock()

	flag := true
	this.WaitQueue.Maps.Range(func(k, v interface{}) bool {
		waitInfo := v.(*WaitInfo)
		if waitInfo.Reply == nil {
			flag = false
			return false
		}
		return true
	})
	return flag
}

// 根据所有用户选择的操作，返回一个用户可进行的操作列表
// 整体优先级，胡 > 碰、杠 > 吃
// 如果有用户选择了胡，则所有可以胡的人，都直接胡
func (this *Mahjong) getReplyResult() *WaitMap {
	if this.WaitQueue.Len() == 1 {
		return this.WaitQueue
	}

	// 是否有人选择了胡
	hasHu := false
	winCode := fbsCommon.OperationCodeWIN
	// 换牌
	exchangeMap := NewWaitMap()
	// 定缺操作
	lackMap := NewWaitMap()
	// 可胡牌的操作列表
	winMap := NewWaitMap()
	// 中间优先级的操作列表，包括碰、杠
	middleMap := NewWaitMap()
	// 低优先级的操作列表，
	chowMap := NewWaitMap()

	// 检查有没有人选择了胡
	// 只有要一个人胡牌了，其他人不管做了什么选择，都跟着胡牌
	this.WaitQueue.Maps.Range(func(k, v interface{}) bool {
		waitInfo := v.(*WaitInfo)
		if waitInfo.Reply != nil && oc.IsHuOperation(waitInfo.Reply.OperationCode) {
			hasHu = true
			winCode = waitInfo.Reply.OperationCode
		}
		// 记录所有能胡的人
		if waitInfo.hasHu() {
			winMap.Maps.Store(k.(int), waitInfo)
		}
		return true
	})

	// 当不能胡时，才去判断其他操作
	if !hasHu {
		this.WaitQueue.Maps.Range(func(k, v interface{}) bool {
			userId := k.(int)
			waitInfo := v.(*WaitInfo)
			if waitInfo.Reply.OperationCode == fbsCommon.OperationCodeNEED_EXCHANGE_TILE {
				// 换牌操作
				exchangeMap.Maps.Store(userId, waitInfo)
			} else if waitInfo.Reply.OperationCode == fbsCommon.OperationCodeNEED_LACK_TILE {
				// 定缺操作
				lackMap.Maps.Store(userId, waitInfo)
			} else if waitInfo.Reply.OperationCode == fbsCommon.OperationCodePONG ||
				waitInfo.Reply.OperationCode == fbsCommon.OperationCodeKONG {
				// 所有能碰、能明杠的用户以及操作
				middleMap.Maps.Store(userId, waitInfo)
			} else if waitInfo.Reply.OperationCode == fbsCommon.OperationCodeCHOW {
				chowMap.Maps.Store(userId, waitInfo)
			}
			return true
		})
	}

	// 如果有人选择了胡，则所有能胡且选择了过的人，都需要变成胡操作
	if hasHu {
		winMap.Maps.Range(func(k, v interface{}) bool {
			waitInfo := v.(*WaitInfo)
			waitInfo.Reply = NewOperation(winCode, nil)
			winMap.Maps.Store(k.(int), waitInfo)
			return true
		})
		return winMap
	} else if exchangeMap.Len() > 0 {
		return exchangeMap
	} else if lackMap.Len() > 0 {
		return lackMap
	} else if middleMap.Len() > 0 {
		return middleMap
	} else {
		return chowMap
	}
}

func (this *WaitInfo) hasWin() bool {
	for _, operation := range this.OpList {
		if oc.IsWinOperation(operation.OperationCode) {
			return true
		}
	}
	return false
}

func (this *WaitInfo) hasPong() bool {
	for _, operation := range this.OpList {
		if oc.IsPongOperation(operation.OperationCode) {
			return true
		}
	}
	return false
}

// 是否拥有“胡”或者“自摸”操作
// 是否拥有“胡”这个操作
func (this *WaitInfo) hasHu() bool {
	for _, operation := range this.OpList {
		if oc.IsHuOperation(operation.OperationCode) {
			return true
		}
	}
	return false
}

// 是否拥有“选择”操作
// 选择包括: 报听、吃、碰、杠、所有的胡
func (this *WaitInfo) hasSelect() bool {
	for _, operation := range this.OpList {
		if oc.IsHuOperation(operation.OperationCode) ||
			oc.IsKongOperation(operation.OperationCode) ||
			operation.OperationCode == fbsCommon.OperationCodeBAO_TING ||
			operation.OperationCode == fbsCommon.OperationCodeCHOW ||
			operation.OperationCode == fbsCommon.OperationCodePONG {
			return true
		}
	}
	return false
}

// HasKongFlower 是否拥有杠补花操作
func (w *WaitInfo) HasKongFlower() bool {
	for _, operation := range w.OpList {
		if operation.OperationCode == fbsCommon.OperationCodeKONG_DARK &&
			len(operation.Tiles) > 0 && operation.Tiles[0] == card.MAHJONG_RED_FLOWER {
			return true
		}
	}
	return false
}

// 获取用户重连时可进行的操作列表
// 和正常发的操作列表的区别是，删除掉抓牌的消息，防止客户端重复抓牌
// 抓牌包括从前和从后两种
func (this *Mahjong) getRestoreOpreationlist(userId int) []*Operation {
	opList := []*Operation{}
	if waitInfo := this.getWait(userId); waitInfo != nil && waitInfo.Reply == nil {
		for _, op := range waitInfo.OpList {
			if !oc.IsDrawOperation(op.OperationCode) {
				opList = append(opList, op)
			}
		}
	}
	return opList
}

// 设置最后操作
func (this *Mahjong) setLastOperate(userId int, op *Operation) {
	this.LastOperator = userId
	this.LastOperation = op

	// 用户杠了之后，记录是什么杠
	// 用户杠了之后，抓牌，记录杠后抓标志
	// 用户杠后抓之后，如果做的不是打牌的操作，清除杠后抓标志
	u := this.getUser(userId)
	if oc.IsKongOperation(op.OperationCode) {
		u.KongCode = op.OperationCode
	} else if op.OperationCode == fbsCommon.OperationCodeDRAW_AFTER_KONG {
		u.DrowAfterKongFlag = true
	} else if op.OperationCode != fbsCommon.OperationCodePLAY {
		u.DrowAfterKongFlag = false
		u.KongCode = 0
	}

	// 设置最后打牌的人，打牌、报听之后设置，碰、明杠之后消除，牌局结束之后消除
	switch op.OperationCode {
	case fbsCommon.OperationCodePLAY:
		fallthrough
	case fbsCommon.OperationCodeBAO_TING:
		this.LastPlayerId = userId
	case fbsCommon.OperationCodePONG:
		fallthrough
	case fbsCommon.OperationCodeKONG:
		this.LastPlayerId = 0
	default:
		break
	}
}

// 用户操作回滚
// 当有多家用户同时有决策时
// 有用户选择了“碰”或“杠”，其他用户选择了“胡”
// 通知选择了“碰”或“杠”的用户回滚操作
func (m *Mahjong) userOperationRollback() {
	m.WaitQueue.Maps.Range(func(k, v interface{}) bool {
		userId := k.(int)
		waitInfo := v.(*WaitInfo)
		if waitInfo.Reply == nil {
			return true
		}
		var replyTile, replyOpCode int
		switch waitInfo.Reply.OperationCode {
		case fbsCommon.OperationCodePONG:
			replyOpCode = fbsCommon.OperationCodeROOLBACK_PONG
		case fbsCommon.OperationCodeCHOW:
			replyOpCode = fbsCommon.OperationCodeROOLBACK_CHOW
		case fbsCommon.OperationCodeKONG:
			replyOpCode = fbsCommon.OperationCodeROOLBACK_KONG
		default:
			return true
		}
		if waitInfo.Reply.Tiles != nil && len(waitInfo.Reply.Tiles) > 0 {
			replyTile = waitInfo.Reply.Tiles[0]
		}
		// 推送消息给用户
		// 推送消息，告诉所有人
		clientOperation := NewClientOperation(userId, NewOperation(replyOpCode, []int{replyTile}))
		SendMessageByUserId(userId, ClientOperationPush(clientOperation))
		core.Logger.Info("[userOperationRollback]roomId:%v,userId:%v,replyCode:%v,replyTile:%v", m.RoomId, userId, replyOpCode, replyTile)
		return true
	})
}

// 加倍
// incr=0时，不加倍，仅通知，用于牌局初始化时的通知
func (m *Mahjong) incrMultipleRound(incr int) {
	m.setting.MultipleRound += incr
	// 通知客户端
	tiles := []int{m.setting.MultipleRound, incr}
	clientOperation := NewClientOperation(0, NewOperation(fbsCommon.OperationCodeMULTIPLE_ROUND, tiles))
	m.SendClientOperationPush(clientOperation)
	// 观察者：系统消息
	pushPacket := ClientOperationPush(clientOperation)
	m.Ob.sendMessage(pushPacket, 0)
	// 回放：添加系统操作
	m.playback.appendClientOperation(clientOperation)

	core.Logger.Info("[incrMultipleRound]roomId:%v, round:%v, multiple:%v", m.RoomId, m.Round, m.setting.MultipleRound)
}
