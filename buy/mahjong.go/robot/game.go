package robot

import (
	"sort"

	"github.com/fwhappy/util"
	flatbuffers "github.com/google/flatbuffers/go"
	"mahjong.go/fbs/Common"
	"mahjong.go/mi/oc"
	"mahjong.go/mi/protocal"
	"mahjong.go/mi/suggest"
	"mahjong.go/mi/ting"
)

// 判断机器人手牌是否听牌
func (robot *Robot) isTing() bool {
	m := util.SliceToMap(robot.HandTileList)
	if util.IntInSlice(len(robot.HandTileList), []int{2, 5, 8, 11, 14}) {
		// 处于需要打牌状态
		for i := range m {
			m1 := mDeleteTile(m, i)
			if isTing, _ := ting.CanTing(util.MapToSlice(m1), nil); isTing {
				return true
			}
		}
	} else {
		if isTing, _ := ting.CanTing(util.MapToSlice(m), nil); isTing {
			return true
		}
	}
	return false
}

// 在众多决策中选择一个
func (robot *Robot) getChoosenOperation(response *Common.OperationPush) (int, int) {
	// 先循环一趟，找出一些优先级高的操作或者记录一些数据
	// 是否可以出牌
	hasPlay := false
	// suggestTile := 0
	// 是否可以跳过
	hasPass := false
	// 是否可以报听
	hasBaoTing := false
	// 是否可杠
	hasKong := false
	operation := new(Common.Operation)
	for i := 0; i < response.OpLength(); i++ {
		response.Op(operation, i)
		opCode := int(operation.Op())
		if opCode == Common.OperationCodePLAY {
			hasPlay = true
		} else if opCode == Common.OperationCodePASS {
			hasPass = true
		} else if opCode == Common.OperationCodeBAO_TING {
			hasBaoTing = true
			// } else if opCode == Common.OperationCodePLAY_SUGGEST || opCode == Common.OperationCodeROBOT_PLAY_SUGGEST {
			// 	suggestTile = int(operation.TilesBytes()[0])
		} else if oc.IsKongOperation(opCode) {
			hasKong = true
		}
	}
	// 如果有报听，则帮用户报听
	if hasBaoTing {
		baoTingTile := 0
		baoTingLen := 0
		for i := 0; i < response.OpLength(); i++ {
			response.Op(operation, i)
			if operation.Op() == byte(Common.OperationCodeBAO_TING) {
				tiles := operation.TilesBytes()
				if len(tiles) > baoTingLen {
					baoTingTile = int(tiles[0])
					baoTingLen = len(tiles)
				}
			}
		}
		return Common.OperationCodeBAO_TING, baoTingTile
	}

	// 选中的操作、选中的牌
	var choosenOpCode, choosenOpTile int
	switch robot.AILevel {
	case 1:
		choosenOpCode, choosenOpTile = robot.getChoosenOperationByAIBrass(response, hasKong)
	case 2:
		fallthrough
	case 3:
		choosenOpCode, choosenOpTile = robot.getChoosenOperationByAISliver(response, hasKong, robot.isTing())
	case 4:
		fallthrough
	default:
		choosenOpCode, choosenOpTile = robot.getChoosenOperationByAIGold(response)
	}

	// 未找到有效操作
	if choosenOpCode == 0 {
		if hasPlay {
			return Common.OperationCodePLAY, robot.selectTile()
		} else if hasPass {
			return Common.OperationCodePASS, 0
		}
	}

	return choosenOpCode, choosenOpTile

	// 计算用户的手牌，是否处于叫牌状态
	// 这里还是有优化空间
	// if robot.isTing() {
	// 	if hasPlay {
	// 		return Common.OperationCodePLAY, suggestTile
	// 	} else if hasPass {
	// 		return Common.OperationCodePASS, 0
	// 	}
	// }

	// // 选中的操作、选中的牌
	// var choosenOpCode, choosenOpTile int

	// // 计算当前牌型的牌阶、最大一类有效牌
	// currentStep, _ := robot.ms.GetEffects(robot.HandTileList)
	// // 计算当前可进张数
	// robot.ms.SetHandTilesSlice(robot.HandTileList)
	// robot.ms.SetDiscardTilesSlice(robot.DiscardTileList)
	// robot.ms.CalcRemaimTiles()
	// currentValidTileCnt := robot.ms.GetValidTileCnt(robot.HandTileList)

	// for i := 0; i < response.OpLength(); i++ {
	// 	response.Op(operation, i)
	// 	opCode := int(operation.Op())
	// 	opTiles := operation.TilesBytes()

	// 	tiles := util.SliceCopy(robot.HandTileList)

	// 	step := -2
	// 	var validTileCnt int
	// 	switch opCode {
	// 	case Common.OperationCodePONG:
	// 		handTiles := util.SliceDel(tiles, int(opTiles[0]), int(opTiles[0]))
	// 		step, _ = robot.ms.GetEffects(handTiles)
	// 		// 计算当前可进张数
	// 		validTileCnt = robot.ms.GetValidTileCnt(handTiles)
	// 	case Common.OperationCodeKONG:
	// 		handTiles := util.SliceDel(tiles, int(opTiles[0]), int(opTiles[0]), int(opTiles[0]))
	// 		step, _ = robot.ms.GetEffects(handTiles)
	// 		validTileCnt = robot.ms.GetRemainTilesCnt(handTiles)
	// 	case Common.OperationCodeKONG_DARK:
	// 		handTiles := util.SliceDel(tiles, int(opTiles[0]), int(opTiles[0]), int(opTiles[0]), int(opTiles[0]))
	// 		step, _ = robot.ms.GetEffects(handTiles)
	// 		validTileCnt = robot.ms.GetRemainTilesCnt(handTiles)
	// 	case Common.OperationCodeKONG_TURN:
	// 		fallthrough
	// 	case Common.OperationCodeKONG_TURN_FREE:
	// 		handTiles := util.SliceDel(tiles, int(opTiles[0]))
	// 		step, _ = robot.ms.GetEffects(handTiles)
	// 		validTileCnt = robot.ms.GetRemainTilesCnt(handTiles)
	// 	default:
	// 	}

	// 	if step >= currentStep || (step == currentStep-1 && currentValidTileCnt <= validTileCnt) {
	// 		return opCode, int(opTiles[0])
	// 	}
	// }

	// return choosenOpCode, choosenOpTile
}

// 最低级的决策选择逻辑
// 有杠就杠、没杠有碰就碰
func (robot *Robot) getChoosenOperationByAIBrass(response *Common.OperationPush, hasKong bool) (choosenOpCode, choosenOpTile int) {
	operation := new(Common.Operation)
	for i := 0; i < response.OpLength(); i++ {
		response.Op(operation, i)
		opCode := int(operation.Op())
		opTiles := operation.TilesBytes()
		switch opCode {
		case Common.OperationCodePONG:
			if hasKong {
				break
			}
			fallthrough
		case Common.OperationCodeKONG:
			fallthrough
		case Common.OperationCodeKONG_DARK:
			fallthrough
		case Common.OperationCodeKONG_TURN:
			fallthrough
		case Common.OperationCodeKONG_TURN_FREE:
			choosenOpCode = opCode
			choosenOpTile = int(opTiles[0])
			return
		default:
		}
	}
	return
}

// 决策时会考虑会不会叫牌
// 叫牌后，不再做碰、杠操作
// 有杠就杠、没杠有碰就碰
func (robot *Robot) getChoosenOperationByAISliver(response *Common.OperationPush, isTing, hasKong bool) (choosenOpCode, choosenOpTile int) {
	// 已叫牌，不再做碰、杠的决策
	if isTing {
		return
	}

	return robot.getChoosenOperationByAIBrass(response, hasKong)
}

// 最高级的决策选择逻辑
func (robot *Robot) getChoosenOperationByAIGold(response *Common.OperationPush) (choosenOpCode, choosenOpTile int) {
	// 计算当前牌型的牌阶、最大一类有效牌
	currentStep, _ := robot.ms.GetEffects(robot.HandTileList)
	// 计算当前可进张数
	robot.ms.SetHandTilesSlice(robot.HandTileList)
	robot.ms.SetDiscardTilesSlice(robot.DiscardTileList)
	robot.ms.CalcRemaimTiles()
	currentValidTileCnt := robot.ms.GetValidTileCnt(robot.HandTileList)

	operation := new(Common.Operation)
	for i := 0; i < response.OpLength(); i++ {
		response.Op(operation, i)
		opCode := int(operation.Op())
		opTiles := operation.TilesBytes()
		tiles := util.SliceCopy(robot.HandTileList)

		step := -2
		var validTileCnt int
		switch opCode {
		case Common.OperationCodePONG:
			handTiles := util.SliceDel(tiles, int(opTiles[0]), int(opTiles[0]))
			step, _ = robot.ms.GetEffects(handTiles)
			// 计算当前可进张数
			validTileCnt = robot.ms.GetValidTileCnt(handTiles)
		case Common.OperationCodeKONG:
			handTiles := util.SliceDel(tiles, int(opTiles[0]), int(opTiles[0]), int(opTiles[0]))
			step, _ = robot.ms.GetEffects(handTiles)
			validTileCnt = robot.ms.GetRemainTilesCnt(handTiles)
		case Common.OperationCodeKONG_DARK:
			handTiles := util.SliceDel(tiles, int(opTiles[0]), int(opTiles[0]), int(opTiles[0]), int(opTiles[0]))
			step, _ = robot.ms.GetEffects(handTiles)
			validTileCnt = robot.ms.GetRemainTilesCnt(handTiles)
		case Common.OperationCodeKONG_TURN:
			fallthrough
		case Common.OperationCodeKONG_TURN_FREE:
			handTiles := util.SliceDel(tiles, int(opTiles[0]))
			step, _ = robot.ms.GetEffects(handTiles)
			validTileCnt = robot.ms.GetRemainTilesCnt(handTiles)
		default:
		}

		// 如果是杠
		if oc.IsKongOperation(opCode) {
			// 三阶及以下，必杠
			// 三阶以上，不降阶就杠
			if currentStep <= 3 || step >= currentStep-1 {
				choosenOpCode = opCode
				choosenOpTile = int(opTiles[0])
				return
			}
		} else {
			if step >= currentStep || (step == currentStep-1 && currentValidTileCnt <= validTileCnt) {
				return opCode, int(opTiles[0])
			}
		}
	}

	return

}

// 计算用户现在该打什么牌
func (this *Robot) selectTile() int {
	// 设置选牌器的定缺
	this.ms.SetLack(this.Lack)
	// 设置AI级别
	this.ms.SetAILevel(this.AILevel)
	// this.ms.Clean()
	// 设置当前手牌
	this.ms.SetHandTilesSlice(this.HandTileList)
	// 设置当前弃牌
	this.ms.SetDiscardTilesSlice(this.DiscardTileList)
	// 计算剩余的牌
	this.ms.CalcRemaimTiles()
	suggestTile, _ := this.ms.GetSuggest()
	return suggestTile
}

// 删除手牌
func (this *Robot) delHandTiles(tile, cnt int) {
	m := util.SliceToMap(this.HandTileList)
	m[tile] -= cnt
	this.HandTileList = util.MapToSlice(m)
	sort.Ints(this.HandTileList)
}

// 从map中删除一张牌，生成一个新的map
func mDeleteTile(m map[int]int, tile int) map[int]int {
	m1 := map[int]int{}
	for k, v := range m {
		if k == tile {
			if v > 1 {
				m1[k] = v - 1
			}
		} else {
			m1[k] = v
		}
	}
	return m1
}

// 处理牌局中的消息
func (this *Robot) handleOperation(impacket *protocal.ImPacket) {
	response := Common.GetRootAsOperationPush(impacket.GetBody(), 0)
	this.handleOperationPush(response)
}

func (this *Robot) handleOperationPush(response *Common.OperationPush) {
	// 第一趟循环，处理优先级别的操作
	// 处理抓牌、胡牌、定缺
	operation := new(Common.Operation)
	for i := 0; i < response.OpLength(); i++ {
		response.Op(operation, i)
		opCode := int(operation.Op())
		tiles := operation.TilesBytes()
		intTiles := []int{}
		for _, tile := range tiles {
			intTiles = append(intTiles, int(tile))
		}
		switch opCode {
		case Common.OperationCodeDRAW: // 抓牌
			fallthrough
		case Common.OperationCodeDRAW_AFTER_KONG: // 杠牌
			this.handleOperationDraw(opCode, int(tiles[0]))
		case Common.OperationCodeFLOWER_CHANGE: // 补花
			this.handleOperationFlowerChange(tiles)
		case Common.OperationCodeWIN: // 点炮胡
			fallthrough
		case Common.OperationCodeWIN_AFTER_KONG_TURN: // 抢杠胡
			fallthrough
		case Common.OperationCodeWIN_AFTER_KONG_DRAW: // 杠上开花
			fallthrough
		case Common.OperationCodeWIN_AFTER_KONG_PLAY: // 热炮
			fallthrough
		case Common.OperationCodeWIN_SELF: // 自摸
			this.handleOperationWin(opCode, int(tiles[0]))
			return
		case Common.OperationCodeNEED_LACK_TILE: // 定缺
			this.handleOperationLack()
			return
		case Common.OperationCodeNEED_EXCHANGE_TILE: // 换牌
			this.handleOperationExchange(intTiles)
		default:
		}
	}
	choosenOpCode, choosenOpTile := this.getChoosenOperation(response)

	switch choosenOpCode {
	case Common.OperationCodePASS:
		this.handleOperationPass()
	case Common.OperationCodeBAO_TING:
		this.handleOperationBaoTing(choosenOpTile)
	case Common.OperationCodePLAY:
		this.handleOperationPlay(choosenOpTile)
	case Common.OperationCodePONG:
		this.handleOperationPong(choosenOpTile)
	case Common.OperationCodeKONG:
		this.handleOperationKong(choosenOpTile)
	case Common.OperationCodeKONG_DARK:
		this.handleOperationKongDark(choosenOpTile)
	case Common.OperationCodeKONG_TURN:
		this.handleOperationKongTurn(choosenOpTile)
	case Common.OperationCodeKONG_TURN_FREE:
		this.handleOperationKongTurnFree(choosenOpTile)
	case 0:
		// 因为现在抓牌的消息是独立发的，所以opCode可能是0
	default:
		this.show("未支持的操作,robotId:%v,opCode:%v,choosenOpTile:%v", this.UserId, choosenOpCode, choosenOpTile)
	}
}

// 处理用户抓牌
func (this *Robot) handleOperationDraw(opCode, tile int) {
	this.HandTileList = append(this.HandTileList, tile)
	sort.Ints(this.HandTileList)
	this.debug("抓牌, roomId:%d,userId: %d, 抓了:%d, 手牌:%#v:%d", this.RoomId, this.UserId, tile, this.HandTileList, len(this.HandTileList))
}

// 处理用户补花
func (this *Robot) handleOperationFlowerChange(tiles []byte) {
	for k, v := range tiles {
		if k%2 == 0 { // 换掉的牌
			this.delHandTiles(int(v), 1)
		} else { // 补的牌
			this.HandTileList = append(this.HandTileList, int(v))
		}
	}
	this.debug("补花, roomId:%d,userId: %d, tiles:%#v", this.RoomId, this.UserId, tiles)
}

// 处理用户胡牌
func (this *Robot) handleOperationWin(opCode, tile int) {
	this.Operation(opCode, tile)
	this.show("胡牌, roomId:%d,userId: %d", this.RoomId, this.UserId)
}

// 处理用户定缺
func (this *Robot) handleOperationLack() {
	ms := suggest.NewMSelector()
	ms.SetHandTilesSlice(this.HandTileList)
	this.Lack = ms.GetSuggestLack()
	this.debug("定缺, userId:%v, 选择了:%v", this.UserId, this.Lack)

	this.Operation(Common.OperationCodeNEED_LACK_TILE, this.Lack)
}

// 处理用户换牌
func (this *Robot) handleOperationExchange(tiles []int) {
	this.debug("换牌, userId:%v, 选择了:%v", this.UserId, tiles)

	this.Operation(Common.OperationCodeNEED_EXCHANGE_TILE, tiles...)
}

// 处理用户成功
func (this *Robot) handleOperationExchangeResult(tiles []int) {
	this.debug("换牌结果, userId:%v, 牌:%v", this.UserId, tiles)

	exchangeCnt := (len(tiles) - 1) / 2
	outTiles := tiles[1 : exchangeCnt+1]
	inTiles := tiles[exchangeCnt+1:]
	this.debug("换牌结果, userId:%v, 方向:%v, 换出的牌:%v, 换入的牌:%v", this.UserId, tiles[0], outTiles, inTiles)
	for _, v := range outTiles {
		this.delHandTiles(v, 1)
	}
	for _, v := range inTiles {
		this.HandTileList = append(this.HandTileList, v)
	}

	this.Operation(Common.OperationCodeEXCHANGE_TILE_RESULT, tiles...)
}

// 处理用户打牌
func (this *Robot) handleOperationPlay(choosenOpTile int) {
	// 打牌
	// tile := this.getPlayTile()
	tile := this.selectTile()
	// tile := choosenOpTile
	this.debug("选牌系统选择了:%v,手牌:%v", tile, this.HandTileList)

	if tile == 0 {
		this.debug("选牌系统选牌错误,手牌:%v", this.HandTileList)
		tile = this.HandTileList[0]
	}

	this.delHandTiles(tile, 1)
	this.Operation(Common.OperationCodePLAY, tile)
	this.debug("出牌, userId: %d, 出了:%d, 手牌:%#v:%d", this.UserId, tile, this.HandTileList, len(this.HandTileList))

	if !util.IntInSlice(len(this.HandTileList), []int{1, 4, 7, 10, 13}) {
		this.show("用户手牌数错误, roomId:%d, userId:%d, 手牌数:%#v", this.RoomId, this.UserId, len(this.HandTileList))
	}
}

// 处理用户报听
func (this *Robot) handleOperationBaoTing(tile int) {
	this.delHandTiles(tile, 1)
	this.Operation(Common.OperationCodeBAO_TING, tile)
	this.debug("报听, userId: %d, 出了:%d, 手牌:%#v:%d", this.UserId, tile, this.HandTileList, len(this.HandTileList))

	if !util.IntInSlice(len(this.HandTileList), []int{1, 4, 7, 10, 13}) {
		this.show("用户手牌数错误, roomId:%d, userId:%d, 手牌数:%#v", this.RoomId, this.UserId, len(this.HandTileList))
	}
}

// 处理用户pass
func (this *Robot) handleOperationPass() {
	this.Operation(Common.OperationCodePASS, 0)
	this.debug("过, userId:%v", this.UserId)
}

// 处理用户碰牌
func (this *Robot) handleOperationPong(tile int) {
	this.delHandTiles(tile, 2)
	this.Operation(Common.OperationCodePONG, tile)
	this.debug("碰牌,userId: %d,tile :%d,手牌:%#v%d", this.UserId, tile, this.HandTileList, len(this.HandTileList))
}

// 处理用户明杠
func (this *Robot) handleOperationKong(tile int) {
	this.delHandTiles(tile, 3)
	this.Operation(Common.OperationCodeKONG, tile)
	this.debug("明杠,userId: %d,tile :%d,手牌:%#v:%d", this.UserId, tile, this.HandTileList, len(this.HandTileList))
}

// 处理用户暗杠
func (this *Robot) handleOperationKongDark(tile int) {
	this.delHandTiles(tile, 4)
	this.Operation(Common.OperationCodeKONG_DARK, tile)
	this.debug("暗杠,userId: %d,tile :%d,手牌:%#v:%d", this.UserId, tile, this.HandTileList, len(this.HandTileList))
}

// 处理用户转弯杠
func (this *Robot) handleOperationKongTurn(tile int) {
	this.delHandTiles(tile, 1)
	this.Operation(Common.OperationCodeKONG_TURN, tile)
	this.debug("转弯杠,userId: %d,tile :%d,手牌:%#v:%d", this.UserId, tile, this.HandTileList, len(this.HandTileList))
}

// 处理用户憨包杠
func (this *Robot) handleOperationKongTurnFree(tile int) {
	this.delHandTiles(tile, 1)
	this.Operation(Common.OperationCodeKONG_TURN_FREE, tile)
	this.debug("转弯杠,userId: %d,tile :%d,手牌:%#v:%d", this.UserId, tile, this.HandTileList, len(this.HandTileList))
}

// 处理用户消息
// 用来维护明牌、弃牌
func (this *Robot) handleUserOperation(impacket *protocal.ImPacket) {
	this.LastOtherOperationTime = util.GetTime()
	response := Common.GetRootAsUserOperationPush(impacket.GetBody(), 0)
	operation := new(Common.Operation)
	operation = response.Op(operation)
	opCode := operation.Op()
	tiles := operation.TilesBytes()
	intTiles := []int{}
	for _, tile := range tiles {
		intTiles = append(intTiles, int(tile))
	}
	// this.debug("[userOperationCode],userId:%v,opCode:%v,tiles:%v", response.UserId(), opCode, tiles)

	var discardAddFlag = true

	switch opCode {
	case Common.OperationCodeBAO_TING: // 报听
		fallthrough
	case Common.OperationCodePLAY: // 出牌: 增加一张弃牌
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
	case Common.OperationCodePONG: // 碰: 增加两张弃牌
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
	case Common.OperationCodeKONG_DARK: // 暗杠: 增加4张弃牌
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
	case Common.OperationCodeKONG: // 明: 增加3张弃牌
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
	case Common.OperationCodeKONG_TURN: // 转弯杠: 增加一张弃牌
		fallthrough
	case Common.OperationCodeKONG_TURN_FREE: // 憨包杠
		this.DiscardTileList = append(this.DiscardTileList, int(tiles[0]))
	case Common.OperationCodeEXCHANGE_TILE_RESULT:
		this.handleOperationExchangeResult(intTiles)
		return
	default:
		discardAddFlag = false
		break
	}

	if discardAddFlag {
		// this.debug("[当前弃牌]opCode:%v,tile:%v,discardTiles:%v", opCode, tiles[0], this.DiscardTileList)
	}
}

// 处理服务器操作
func (this *Robot) handleClientOperation(impacket *protocal.ImPacket) {
	this.LastOtherOperationTime = util.GetTime()
	op := new(Common.Operation)
	opCode := Common.GetRootAsClientOperationPush(impacket.GetBody(), 0).Op(op).Op()
	this.trace("收到client operation:%d", opCode)
}

// 处理单局结算
func (this *Robot) handleGameSettlement(impacket *protocal.ImPacket) {
	this.debug("单局完成,roomId:%d,userId:%d", this.RoomId, this.UserId)

	/*
		// 局数+1
		roundTimes++
		//是否胡牌
		if Common.GetRootAsGameSettlementPush(impacket.GetBody(), 0).IsHuangPai() == int8(0) {
			huTimes++
		}

		if roundTimes%4 == 0 {
			this.show("胡牌统计: %d / %d", huTimes/4, roundTimes/4)
		}
	*/

	// 打完4局, 直接退出
	if this.TRound > 0 && this.Round >= this.TRound {
		return
	}

	// 让用户准备
	this.Prepare(true)
}

// 处理游戏完成
func (this *Robot) handleGameResult(impacket *protocal.ImPacket) {
	this.debug("游戏完成,roomId:%d,userId:%d", this.RoomId, this.UserId)
}

// 处理牌局初始化
func (this *Robot) handleGameInit(impacket *protocal.ImPacket) {
	push := Common.GetRootAsGameInitPush(impacket.GetBody(), 0)
	this.Round = int(push.CurrentRound())
	this.HandTileList = []int{}
	this.DiscardTileList = []int{}
	this.ShowTileList = [][]int{}
	for _, tile := range Common.GetRootAsGameInitPush(impacket.GetBody(), 0).TilesBytes() {
		this.HandTileList = append(this.HandTileList, int(tile))
	}
	// 记录用户手牌
	sort.Ints(this.HandTileList)
	// 设置用户定缺
	this.Lack = 0
	this.debug("初始化用户手牌, roomId:%d, userId:%d, 手牌：%#v,手牌数:%d", this.RoomId, this.UserId, this.HandTileList, len(this.HandTileList))
}

// 机器人准备
func (this *Robot) Prepare(needsleep bool) {
	if needsleep {
		this.prepareSleep()
	}
	mType := protocal.MSG_TYPE_NOTIFY
	builder := flatbuffers.NewBuilder(0)
	Common.GameReadyNotifyStart(builder)
	Common.GameReadyNotifyAddAgree(builder, byte(1))
	Common.GameReadyNotifyAddReadying(builder, byte(1))
	orc := Common.GameReadyNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	message := protocal.NewImMessage(uint16(Common.CommandGameReadyNotify), uint16(mType), uint16(0), uint16(0), buf)
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())

	this.debug("Prepare,userId:%d,roomId:%d", this.UserId, this.RoomId)
}

// 机器人取消托管
func (this *Robot) CancelHosting() {
	mType := protocal.MSG_TYPE_NOTIFY
	builder := flatbuffers.NewBuilder(0)
	Common.GameHostingNotifyStart(builder)
	Common.GameHostingNotifyAddHostingStatus(builder, byte(0))
	orc := Common.GameHostingNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	message := protocal.NewImMessage(uint16(Common.CommandGameHostingNotify), uint16(mType), uint16(0), uint16(0), buf)
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())

	this.debug("CancelHosting,userId:%d,roomId:%d", this.UserId, this.RoomId)
}

// 机器人回应操作
func (this *Robot) Operation(opCode int, tiles ...int) {
	if oc.IsWinOperation(opCode) {
		this.winSleep()
	} else if oc.IsLackOperation(opCode) {
		this.lackSleep()
	} else {
		this.replySleep()
	}

	builder := flatbuffers.NewBuilder(0)
	// 构建一个operation
	var tilesBinary flatbuffers.UOffsetT
	tileCnt := len(tiles)
	Common.OperationStartTilesVector(builder, tileCnt)
	for i := tileCnt - 1; i >= 0; i-- {
		builder.PrependByte(byte(tiles[i]))
	}
	tilesBinary = builder.EndVector(tileCnt)
	// 构建对象
	Common.OperationStart(builder)
	Common.OperationAddOp(builder, byte(opCode))
	Common.OperationAddTiles(builder, tilesBinary)
	op := Common.OperationEnd(builder)

	Common.UserOperationNotifyStart(builder)
	Common.UserOperationNotifyAddOp(builder, op)
	orc := Common.UserOperationNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	message := protocal.NewImMessage(uint16(Common.CommandUserOperationNotify), uint16(protocal.MSG_TYPE_NOTIFY), uint16(0), uint16(0), buf)
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())
}

// 机器人同意解散房间
func (this *Robot) DismissReply() {
	mType := protocal.MSG_TYPE_NOTIFY
	builder := flatbuffers.NewBuilder(0)
	Common.DismissRoomNotifyStart(builder)
	Common.DismissRoomNotifyAddOp(builder, int8(0))

	orc := Common.DismissRoomNotifyEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	message := protocal.NewImMessage(uint16(Common.CommandDismissRoomNotify), uint16(mType), uint16(0), uint16(0), buf)

	// 发送消息给服务器
	imPacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_DATA, message)
	this.Conn.Write(imPacket.Serialize())

	this.debug("同意解散,userId:%d,roomId:%d", this.UserId, this.RoomId)
}
