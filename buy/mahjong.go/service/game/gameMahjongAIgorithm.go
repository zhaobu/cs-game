package game

import (
	"sort"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/core"
	"mahjong.go/mi/card"
	"mahjong.go/mi/oc"
	"mahjong.go/mi/ting"
	"mahjong.go/mi/win"
	configService "mahjong.go/service/config"
)

// 判断用户能不能胡牌
func (m *Mahjong) canWin(opList []*Operation, user *MahjongUser, player *MahjongUser, tile, opCode int, repaoFlag, allowPingHu bool) ([]*Operation, bool) {
	var win = false
	if user.MTC.IsTing() {
		for _, p := range user.MTC.GetTingTiles() {
			// 过胡不胡
			if util.IntInSlice(tile, user.SkipWin) {
				SendMessageByUserId(user.UserId, GameSkipOperateNoticePush(fbsCommon.OperationCodeWIN, tile))
				core.Logger.Info("[canWin]用户处于过胡状态，不允许胡牌,roomId:%v,round:%v,userId:%v", m.RoomId, m.Round, user.UserId)
				break
			}
			if p == tile {
				if allowPingHu || m.isNotPiHu(user, player, tile, opCode, repaoFlag) {
					if oc.IsKongTurnOperation(opCode) {
						// 被抢杠
						opList = append(opList, NewOperation(fbsCommon.OperationCodeWIN_AFTER_KONG_TURN, []int{tile}))
					} else if repaoFlag {
						// 热炮
						opList = append(opList, NewOperation(fbsCommon.OperationCodeWIN_AFTER_KONG_PLAY, []int{tile}))
					} else {
						// 点炮
						opList = append(opList, NewOperation(fbsCommon.OperationCodeWIN, []int{tile}))
					}

					return opList, true
				}
				break
			}
		}
	}
	return opList, win
}

// 判断是不是屁胡, 天胡都在winself中，所以这里不需要判断这些类型
func (m *Mahjong) isNotPiHu(user *MahjongUser, player *MahjongUser, tile int, playerOpCode int, repaoFlag bool) bool {
	// 热炮直接返回
	if repaoFlag {
		return true
	}
	// 有没有人报听
	if player.MTC.IsBaoTing() || user.MTC.IsBaoTing() {
		return true
	}
	if user.canDihu() {
		return true
	}

	var pai = append(user.HandTileList.ToSlice(), tile)
	if qingCheck(pai, user.ShowCardList.GetAll()) {
		return true
	}

	// 自己手上有杠牌(明杠，暗杠， 转弯杠，都可以)
	for _, j := range user.ShowCardList.GetAll() {
		if oc.IsKongOperation(j.GetOpCode()) {
			// 部分麻将，明杠不算通行证
			if m.setting.EnableKongTXZ || j.GetOpCode() != fbsCommon.OperationCodeKONG {
				return true
			}
		}
	}

	// player转弯杠或者憨包杠的话，可以抢胡
	if oc.IsKongTurnOperation(playerOpCode) {
		return true
	}

	var t = m.winType(pai, user.ShowCardList.GetAll(), tile)
	if t != config.HU_TYPE_PI {
		return true
	}
	return false
}

// 计算胡牌类型(认为已经是胡牌)
func (m *Mahjong) winType(pai []int, sCard []*card.ShowCard, tile int) int {
	var tmp = util.SliceCopy(pai)
	sort.Ints(tmp)

	if m.setting.EnableShuangLongQiDui && shuangLong7Dui(tmp, tile) {
		// 双龙七对肯定是龙七对，所以这个判断，必须在龙七对的判断之前
		return config.HU_TYPE_SHUANG_LONG_7DUI
	} else if m.setting.EnableHePu7Dui && hepu7Dui(tmp) {
		return config.HU_TYPE_HEPU_7DUI
	} else if long7Dui(tmp, tile) {
		return config.HU_TYPE_LONG_7DUI
	} else if m.setting.EnableDi7Dui && di7Dui(tmp, sCard, tile) {
		return config.HU_TYPE_DIQIDUI
	} else if allDui(tmp, sCard) {
		return config.HU_TYPE_7DUI
	} else if m.setting.EnableDanDiao && danDiao(tmp) {
		return config.HU_TYPE_DANDIAO
	} else if daDui(tmp) {
		return config.HU_TYPE_DADUI
	} else if m.setting.EnableBianKaDiao && m.bianKaDiao(tmp, tile) {
		return config.HU_TYPE_BIANKADIAO
	} else if m.setting.EnableDaKuanZhang && m.daKuanZhang(tmp, tile) {
		return config.HU_TYPE_DAKUANZHANG
	}
	return config.HU_TYPE_PI
}

func (m *Mahjong) checkHu(hCard []int, sCard []*card.ShowCard) bool {
	showTiles := []int{}
	// 不支持地龙，直接不给showTiles
	if m.setting.EnableDi7Dui && sCard != nil {
		for _, showCard := range sCard {
			showTiles = append(showTiles, showCard.GetTiles()...)
		}
	}
	return win.CanWin(hCard, showTiles)
}

// 检查用户能不能自摸
func (m *Mahjong) canWinSelf(opList []*Operation, user *MahjongUser, tile int) ([]*Operation, bool) {
	// 天胡的情况
	if user.MTC.IsBad() && m.checkHu(user.HandTileList.ToSlice(), user.ShowCardList.GetAll()) {
		opList = append(opList, NewOperation(fbsCommon.OperationCodeWIN_SELF, []int{tile}))
		user.MTC.SetStatus(ting.STATUS_TING)
		return opList, true
	}

	if !user.MTC.IsNormal() {
		for _, p := range user.MTC.GetTingTiles() {
			if p == tile {
				if user.DrowAfterKongFlag {
					opList = append(opList, NewOperation(fbsCommon.OperationCodeWIN_AFTER_KONG_DRAW, []int{tile}))
				} else {
					opList = append(opList, NewOperation(fbsCommon.OperationCodeWIN_SELF, []int{tile}))
				}
				return opList, true
			}
		}
	}
	return opList, false
}

// 检查用户能不能听牌和报听(一定是抓牌的时候判断能否听),将结果存入user.TingMap中减少后续计算
func (m *Mahjong) canTingOrBaoTing(opList []*Operation, user *MahjongUser, lackTile int) ([]*Operation, bool) {
	var ting = false
	var pai = user.HandTileList.ToSlice()

	// 可能听的牌只能是手牌，或者手牌的临牌
	var maybeTing = m.getMaybeTing(user.HandTileList, user.ShowCardList.GetAll())
	user.MTC.Clean()

	core.Logger.Debug("[canTingOrBaoTing]maybeTing:%#v", maybeTing)
	for i, play := range pai {
		// remove same
		if i != 0 && pai[i] == pai[i-1] {
			continue
		}
		var tmp = util.SliceCopy(pai)
		var t = []int{play}
		for p := range maybeTing {
			tmp[i] = p
			// 跳过有缺的
			if lackTile > 0 && getLackCount(lackTile, tmp) > 0 {
				continue
			}
			if m.checkHu(tmp, user.ShowCardList.GetAll()) {
				t = append(t, p)
				ting = true
			}
		}
		// 可以听牌的消息
		if len(t) > 1 {
			// 如果还没打牌可以报听
			if user.canBaoTing() {
				opList = append(opList, NewOperation(fbsCommon.OperationCodeBAO_TING, t))
			}
			// 告诉用户听的消息
			opList = append(opList, NewOperation(fbsCommon.OperationCodeTING, t))
			user.MTC.AppendTingTiles(t)
			core.Logger.Debug("[canTingOrBaoTing]算到用户听牌规则，userId:%v,ting:%#v", user.UserId, t)
		}
	}
	return opList, ting
}

// 杠后更新用户的听牌状态
func (m *Mahjong) setTingAfterKong(user *MahjongUser) {
	var tiles = m.getTingSlice(user)
	if len(tiles) == 0 {
		user.MTC.SetNormal()
	} else {
		user.MTC.SetTingTiles(tiles)
	}

}

// 获取所有能听的牌
// fixme， 这里需要仔细看一下，为什么需要先在slice中放一个0，后面再判断长度是否==1
func (m *Mahjong) getTingSlice(user *MahjongUser) []int {
	var ting = []int{0}
	var pai = user.HandTileList.ToSlice()
	var maybeTing = m.getMaybeTing(user.HandTileList, user.ShowCardList.GetAll())
	for i := range maybeTing {
		if m.checkHu(append(pai, i), user.ShowCardList.GetAll()) {
			ting = append(ting, i)
		}
	}
	if len(ting) == 1 {
		return []int{}
	}
	return ting
}

// 是否是自摸
func (m *Mahjong) isZimoHu() bool {
	return m.HInfo.Hu && oc.IsZMOperation(m.HInfo.HuOperationCode)
}

// 可能听的牌只能是几种情况：
// 1、手牌
// 2、手牌的临牌
// 3、有且只碰过一次的时候，碰的那张牌（地龙）
func (m *Mahjong) getMaybeTing(tileMap *card.CMap, sCard []*card.ShowCard) map[int]int {
	var maybeTing = map[int]int{}
	relationTiles := card.GetSelfAndNeighborCards(tileMap.GetUnique()...)
	for _, i := range relationTiles {
		maybeTing[i] = 1
	}
	// 地7对情况
	if m.setting.EnableDi7Dui && len(sCard) == 1 && sCard[0].IsPong() {
		maybeTing[sCard[0].GetTile()] = 1
	}
	return maybeTing
}

// 获取当前推荐的出牌
// 推荐给用户的
// 需要给不同的code，用于机器人使用和正常用户的使用
func (m *Mahjong) getSuggestPlayTile(opList []*Operation, mu *MahjongUser, suggestCode int) []*Operation {
	// 设置手牌
	m.Selector.Clean()
	m.Selector.SetHandTilesSlice(mu.HandTileList.ToSlice())
	// 如果是非机器人，且处于托管状态，降低推荐出牌ai级别
	if !configService.IsRobot(mu.UserId) {
		var room, _ = RoomMap.GetRoom(m.RoomId)
		if room.userInHosting(mu.UserId) {
			m.Selector.SetAILevel(0)
		}
	}

	// 设置定缺
	m.Selector.SetLack(mu.LackTile)
	// 追加个用户的明牌和弃牌
	for _, tmpMu := range m.getUsers() {
		m.Selector.AddShowTilesSlice(tmpMu.ShowCardList.GetAllTiles())
		m.Selector.AddDiscardTilesSlice(tmpMu.DiscardTileList.GetTiles())
	}
	// 计算剩余的牌
	m.Selector.CalcRemaimTiles()

	var tile int
	var effectTiles []int
	// if !core.IsProduct() && mu.UserId <= 2000 {
	tile, effectTiles = m.Selector.GetSuggestOld()
	// } else {
	// 	tile, effectTiles = m.Selector.GetSuggest()
	// }
	sd := append(append([]int{}, tile), effectTiles...)
	opList = append(opList, NewOperation(suggestCode, sd))

	// 计算听牌状态下的推荐列表
	// 如果opList里面有ting这个操作，需要找出听剩余最多的牌
	tingRemainCntList := map[int]int{}
	maxRemainCnt := 0
	for _, op := range opList {
		if op.OperationCode == fbsCommon.OperationCodeTING {
			remainCnt := m.Selector.GetRemainTilesCnt(op.Tiles[1:])
			tingRemainCntList[op.Tiles[0]] = remainCnt
			if maxRemainCnt < remainCnt {
				maxRemainCnt = remainCnt
			}
		}
	}
	if len(tingRemainCntList) > 0 {
		tingSuggestTiles := []int{}
		for tile, cnt := range tingRemainCntList {
			if cnt == maxRemainCnt {
				tingSuggestTiles = append(tingSuggestTiles, tile)
			}
		}
		if len(tingSuggestTiles) != len(tingRemainCntList) {
			opList = append(opList, NewOperation(fbsCommon.OperationCodeTING_PLAY_SUGGEST, tingSuggestTiles))
		}
	}

	return opList
}

// 获取用户的一类有效牌，需要排除定缺的
func (m *Mahjong) getEffectTiles(mu *MahjongUser) []int {
	// 设置手牌
	m.Selector.Clean()
	m.Selector.SetHandTilesSlice(mu.HandTileList.ToSlice())
	// 设置定缺
	m.Selector.SetLack(mu.LackTile)
	return m.Selector.CalcEffects()
}

// 牌型判断：大宽张
func (m *Mahjong) daKuanZhang(tmp []int, tile int) bool {
	var pai = util.SliceCopy(tmp)
	pai = util.SliceDel(pai, tile)

	var mayBeTing = ting.GetMaybeTing(pai, nil)
	var hu []int
	for _, i := range mayBeTing {
		if m.checkHu(append(pai, i), nil) {
			hu = append(hu, i)
		}
	}
	if len(hu) != 3 || !card.IsSameSuit(hu...) {
		return false
	}
	sort.Ints(hu)
	if hu[2]-hu[1] != 3 || hu[1]-hu[0] != 3 {
		return false
	}
	return true
}

// 牌型判断：边卡吊
func (m *Mahjong) bianKaDiao(tmp []int, tile int) bool {
	var pai = util.SliceDel(util.SliceCopy(tmp), tile)
	var mayBeTing = ting.GetMaybeTing(pai, nil)

	for _, i := range mayBeTing {
		if i != tile && m.checkHu(append(pai, i), nil) {
			return false
		}
	}
	return true
}

func (m *Mahjong) getExchangeTargetIndex(direction int, fromIndex int) int {
	var index int
	switch direction {
	case config.EXCHANGE_DIRECTION_OPPOSITE: // 跟对面换
		var exchangeTaget = map[int]int{
			0: 2,
			1: 3,
			2: 0,
			3: 1,
		}
		index = exchangeTaget[fromIndex]
	case config.EXCHANGE_DIRECTION_COUNTERCLOCKWISE: // 逆时针，0给3，3给2，2给1，1给0
		if fromIndex == 0 {
			index = len(m.Index) - 1
		} else {
			index = fromIndex - 1
		}
	case config.EXCHANGE_DIRECTION_CLOCKWISE: // 顺时针，0给1,1给2,2给3,3给0
		if fromIndex == len(m.Index)-1 {
			index = 0
		} else {
			index = fromIndex + 1
		}
	default:
		core.Logger.Error("[getExchangeTargetIndex]未支持的换牌方向, roomId:%v, round:%v,direction:%v", m.RoomId, m.Round, direction)
		// 做一下保护，跟自己换，至少牌不会出错
		index = fromIndex
	}
	core.Logger.Debug("[getExchangeTargetIndex]roomId:%v, round:%v,direction:%v, from index:%v, index:%v", m.RoomId, m.Round, direction, fromIndex, index)
	return index
}

// 用户是否可以报听
// 只抓过14张且没有出牌没有明牌
func (mu *MahjongUser) canBaoTing() bool {
	return mu.HandTileList.Len() == 14 && mu.HandTileList.GetDrawTileCnt() == 14 && mu.DiscardTileList.GetPlayedLen() == 0 && mu.ShowCardList.Len() == 0
}

// 用户是否满足天胡的条件
func (mu *MahjongUser) canTianhu() bool {
	return mu.IsDealer && mu.HandTileList.Len() == 14 && mu.HandTileList.GetDrawTileCnt() == 14 && mu.DiscardTileList.GetPlayedLen() == 0 && mu.ShowCardList.Len() == 0
}

// 用户是否满足地胡的条件
func (mu *MahjongUser) canDihu() bool {
	return !mu.IsDealer && (mu.HandTileList.Len() == 14 || mu.HandTileList.Len() == 13) &&
		(mu.HandTileList.GetDrawTileCnt() == 14 || mu.HandTileList.GetDrawTileCnt() == 13) &&
		mu.DiscardTileList.GetPlayedLen() == 0 && mu.ShowCardList.Len() == 0
}

// 判断能不能碰, 调用此方法, 需要在外围添加非缺的判断
func (mu *MahjongUser) canPong(opList []*Operation, tileMap *card.CMap, tile int) ([]*Operation, bool) {
	// 过碰不碰
	if util.IntInSlice(tile, mu.SkipPong) {
		SendMessageByUserId(mu.UserId, GameSkipOperateNoticePush(fbsCommon.OperationCodePONG, tile))
		core.Logger.Info("[canPong]用户处于过碰状态，不允许碰牌,userId:%v", mu.UserId)
		return opList, false
	}

	if tileMap.GetTileCnt(tile) < 2 {
		return opList, false
	}

	opList = append(opList, NewOperation(fbsCommon.OperationCodePONG, []int{tile}))
	return opList, true
}
