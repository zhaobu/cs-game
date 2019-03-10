package game

import (
	"github.com/fwhappy/util"
	"mahjong.go/config"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/core"
	"mahjong.go/mi/card"
	"mahjong.go/mi/oc"
)

// 翻牌捉鸡
func (this *Mahjong) setChikenDraw() {
	// 翻鸡
	this.chiken.GenChikenDraw(this.TileWall)
	tile := this.chiken.GetChikenDraw()
	if tile == 0 {
		core.Logger.Info("所有的牌都抓完了，退出捉鸡, roomId:%d, round: %d.", this.RoomId, this.Round)
		return
	}

	// 如果翻牌鸡和前后鸡重复了，则不告诉客户端具体的鸡是什么
	var operation *Operation
	if this.TileWall.GetForward() > 0 && this.TileWall.GetForward() == this.getChikenFBIndex() {
		operation = &Operation{}
		operation.OperationCode = fbsCommon.OperationCodeCHIKEN_DRAW
	} else {
		operation = NewOperation(fbsCommon.OperationCodeCHIKEN_DRAW, []int{tile})
	}

	// 推送消息给局内用户
	clientOperation := NewClientOperation(0, operation)
	this.SendClientOperationPush(clientOperation)
	// 观察者：系统操作
	pushPacket := ClientOperationPush(clientOperation)
	this.Ob.sendMessage(pushPacket, 0)

	// 回放：添加系统操作
	this.playback.appendClientOperation(clientOperation)

	core.Logger.Info("[setChikenDraw]roomId:%d, round: %d, tile:%d", this.RoomId, this.Round, tile)
}

// 获取翻牌鸡，根被被翻出来的牌面值和setting，计算哪些牌是“鸡”
// 如果是上下鸡，则翻牌鸡为翻到的上下张，如果不是，则为抓到的牌的下一张
func (m *Mahjong) getChikenDraw() []int {
	return m.chiken.GetChikensByGivenTiles(m.setting, m.chiken.GetChikenDraw())
}

// 刷新前后鸡
func (m *Mahjong) setChikenFB() {
	// 读取当前前后鸡的位置和牌面值
	currentFBIndex, currentFBTile := m.chiken.GetChikenFB(m.TileWall)
	// 刷新前后鸡
	m.chiken.GenChikenFB(m.TileWall)
	newFBIndex, newFBTile := m.chiken.GetChikenFB(m.TileWall)

	// 如果存在前后鸡，则推送前后鸡消息
	// fixme 这里如果前面一栋的上面的牌正好被抓走了，用户杠成功了, 这个逻辑，现在是不发消息给客户端, 客户端就不知道将牌盖上
	if newFBIndex > 0 && !m.TileWall.IsDrawn(newFBIndex) {
		var tiles = []int{newFBTile, newFBIndex - m.TileWall.GetForward()}
		if currentFBIndex > 0 {
			tiles = append(tiles, currentFBTile, currentFBIndex-m.TileWall.GetForward())
		}

		// 通知客户端
		clientOperation := NewClientOperation(0, NewOperation(fbsCommon.OperationCodeCHIKEN_ROLLER, tiles))
		m.SendClientOperationPush(clientOperation)
		// 观察者：系统消息
		pushPacket := ClientOperationPush(clientOperation)
		m.Ob.sendMessage(pushPacket, 0)
		// 回放：添加系统操作
		m.playback.appendClientOperation(clientOperation)

		core.Logger.Info("[setChikenFB]roomId:%d, round: %d, tile:%d, index:%d", m.RoomId, m.Round, newFBTile, newFBIndex)

		// 设置牌墙这张牌不能被expect
		m.TileWall.SetExpectDisabled(newFBIndex)
	}
}

// 刷新滚筒鸡
func (m *Mahjong) setChikenTumbling() {
	// 追加滚筒鸡
	appendIndex := m.chiken.AppendChikenTumbling(m.TileWall)
	if appendIndex > 0 {
		var tiles = []int{m.TileWall.GetTile(appendIndex), appendIndex - m.TileWall.GetForward()}

		// 通知客户端
		clientOperation := NewClientOperation(0, NewOperation(fbsCommon.OperationCodeCHIKEN_TUMBLING, tiles))
		m.SendClientOperationPush(clientOperation)
		// 观察者：系统消息
		pushPacket := ClientOperationPush(clientOperation)
		m.Ob.sendMessage(pushPacket, 0)
		// 回放：添加系统操作
		m.playback.appendClientOperation(clientOperation)

		// 设置牌墙这张牌不能被expect
		m.TileWall.SetExpectDisabled(appendIndex)

		core.Logger.Info("[setChikenTumbling]roomId:%d, round: %d, tile:%d, index:%d", m.RoomId, m.Round, appendIndex, tiles[0])
	}
}

// 刷新滚筒鸡或前后鸡
func (m *Mahjong) setChikenRock() {
	// 未处于预变鸡状态，需要退出执行
	if !m.preChangeChikenRock {
		return
	}

	// 支持前后鸡
	if m.setting.IsSettingChikenFB() {
		m.setChikenFB()
	} else if m.setting.IsSettingChikenTumbling() {
		m.setChikenTumbling()
	}

	// 还原预变鸡状态
	m.preChangeChikenRock = false
}

// 获取哪些牌是前后鸡，不是翻出来的那张，而是前后张
// 需要考虑上下鸡的情况
func (m *Mahjong) getChikenFB() []int {
	tiles := []int{}
	fbIndex, fbTile := m.chiken.GetChikenFB(m.TileWall)
	if fbIndex > 0 && fbTile > 0 {
		tiles = m.chiken.GetChikensByGivenTiles(m.setting, fbTile)
	}
	return tiles
}

// 获取所有的滚筒鸡
// 不是翻出来的那张，而是前后张
func (m *Mahjong) getChikenTumbling() []int {
	tiles := []int{}
	tumblingMap := m.chiken.GetChikenTumblingMap(m.TileWall)
	if len(tumblingMap) > 0 {
		tiles = m.chiken.GetChikensByGivenTiles(m.setting, util.GetMapValues(tumblingMap)...)
	}
	return tiles
}

// 获取前后鸡的牌面值
func (m *Mahjong) getChikenFBTile() int {
	_, fbTile := m.chiken.GetChikenFB(m.TileWall)
	return fbTile
}

// 获取前后鸡的索引
func (m *Mahjong) getChikenFBIndex() int {
	fbIndex, _ := m.chiken.GetChikenFB(m.TileWall)
	return fbIndex
}

// 获取冲锋幺鸡用户Id
func (m *Mahjong) getChikenChargeBam1() int {
	return m.chiken.GetChargeBam1()
}

// 获取冲锋乌骨鸡用户Id
func (m *Mahjong) getChikenChargeDot8() int {
	return m.chiken.GetChargeDot8()
}

// 获取责任鸡用户Id, 打了冲锋鸡的人
func (m *Mahjong) getChikenResponsibility() int {
	return m.chiken.GetResponsibility()
}

// 获取翻牌鸡的牌
func (m *Mahjong) getChikenDrawTile() int {
	return m.chiken.GetChikenDraw()
}

// 设置冲锋鸡
func (m *Mahjong) setChikenCharge() {
	// 是否支持冲锋鸡
	if !m.setting.EnableCharge {
		return
	}
	// 读取上一次操作的牌
	if m.LastOperation.Tiles == nil || !oc.IsPlayOperation(m.LastOperation.OperationCode) {
		return
	}
	tile := m.LastOperation.Tiles[0]

	// 设置冲锋幺鸡
	if tile == card.MAHJONG_BAM1 && m.chiken.GetChargeBam1() == 0 && m.chiken.GetResponsibility() == 0 && m.getTilePlayedCnt(card.MAHJONG_BAM1) <= 1 {
		m.chiken.SetChargeBam1(m.LastOperator)

		// 客户端: 系统消息
		clientOperation := NewClientOperation(m.LastOperator, NewOperation(fbsCommon.OperationCodeCHIKEN_CHARGE_BAM1, nil))
		m.SendClientOperationPush(clientOperation)

		// 观察者：系统消息
		pushPacket := ClientOperationPush(clientOperation)
		m.Ob.sendMessage(pushPacket, 0)

		// 回放：添加系统操作
		m.playback.appendClientOperation(clientOperation)

		core.Logger.Info("[setChikenChargeBam1]设置冲锋幺鸡，roomId:%v, round:%v, userId:%d.", m.RoomId, m.Round, m.LastOperator)
	}

	// 设置冲锋乌骨鸡
	if m.setting.IsSettingChikenDot8() && tile == card.MAHJONG_DOT8 && m.chiken.GetChargeDot8() == 0 && m.getTilePlayedCnt(card.MAHJONG_DOT8) <= 1 {
		m.chiken.SetChargeDot8(m.LastOperator)

		// 客户端: 系统消息
		clientOperation := NewClientOperation(m.LastOperator, NewOperation(fbsCommon.OperationCodeCHIKEN_CHARGE_DOT8, nil))
		m.SendClientOperationPush(clientOperation)

		// 观察者：系统消息
		pushPacket := ClientOperationPush(clientOperation)
		m.Ob.sendMessage(pushPacket, 0)

		// 回放：添加系统操作
		m.playback.appendClientOperation(clientOperation)

		core.Logger.Info("[setChikenChargeDot8]设置冲锋乌骨鸡，roomId:%v, round:%v, userId:%d.", m.RoomId, m.Round, m.LastOperator)
	}
}

// 设置责任鸡
func (m *Mahjong) setChikenResponsibility(userId, tile int) {
	// 判断操作的牌是不是1条
	if tile != card.MAHJONG_BAM1 {
		return
	}
	// 判断是否支持责任鸡
	if !m.setting.EnableResponsibilityBam1 {
		return
	}
	// 如果有了冲锋鸡， 则不能再成为责任鸡了
	if m.chiken.GetChargeBam1() > 0 {
		return
	}
	m.chiken.SetResponsibility(userId)

	// 推送责任鸡消息
	clientOperation := NewClientOperation(userId, NewOperation(fbsCommon.OperationCodeCHIKEN_RESPONSIBILITY, nil))
	m.SendClientOperationPush(clientOperation)

	// 观察者：系统消息
	pushPacket := ClientOperationPush(clientOperation)
	m.Ob.sendMessage(pushPacket, 0)

	// 回放：添加系统操作
	m.playback.appendClientOperation(clientOperation)

	core.Logger.Info("[setChikenResponsibility]设置责任鸡，roomId:%v, round:%v, userId:%d, tile: %d.", m.RoomId, m.Round, userId, tile)
}

// dk:弃牌鸡;hk:手牌鸡;sk:明牌鸡
func (this *Mahjong) findKitchen(userId int, ji int) (dk, sk, hk int) {
	// 获取的鸡未设置
	if ji != 0 {
		// 只有幺鸡，乌骨鸡在未设置满堂鸡的情况下计算弃牌
		if this.setting.IsSettingAllChikenDraw() ||
			ji == card.MAHJONG_BAM1 ||
			(ji == card.MAHJONG_DOT8 && this.setting.IsSettingChikenDot8()) {
			dk = this.getUser(userId).DiscardTileList.GetTileCnt(ji)
		}
		// 手牌鸡
		hk = this.getUser(userId).HandTileList.GetTileCnt(ji)
		// 明牌鸡
		sk = this.getUser(userId).ShowCardList.GetTileCnt(ji)
	}
	return
}

// 1 赢鸡，2 不赢不包，3包鸡
func (this *Mahjong) getKitchenAndTarget(userId int) (target []int, t int) {
	if this.getUser(userId).MTC.IsNormal() {
		// 包鸡 未上听, 无论是否黄牌
		target = this.getBeiBaoKitchenUserId(userId)
		t = config.USER_SETTLEMENT_CHIKEN_STATUS_LOSE
	} else if this.HInfo.Hu {
		var can bool = this.canWinKongOrKitchen(userId)
		if can {
			// 可以赢鸡的玩家
			target = this.getOtherUserId(userId)
			t = config.USER_SETTLEMENT_CHIKEN_STATUS_WIN
		}
	} else {
		t = config.USER_SETTLEMENT_CHIKEN_STATUS_NO
	}
	return
}

// 获取被包鸡的对象
// 如果设置了必须要叫牌才包鸡，则对象为“叫牌了且没有烧鸡”的用户，未设置，则包给其他人
func (this *Mahjong) getBeiBaoKitchenUserId(userId int) []int {
	if this.setting.BaoChikenNeedTing {
		return this.getTingAndNoshaoJiId()
	}
	return this.getNoshaoJiId(userId)
}

// 读取谁碰了或者杠了责任鸡
func (m *Mahjong) getChikenResponsibilityTarget() int {
	if rbUserId := m.chiken.GetResponsibility(); rbUserId > 0 {
		for _, u := range m.getUsers() {
			if u.UserId != rbUserId && u.ShowCardList.HasPongOrKongTile(card.MAHJONG_BAM1) {
				return u.UserId
			}
		}
	}
	return 0
}

// 本局是否是金鸡
func (this *Mahjong) isGoldBam1() bool {
	chikenType, exists := this.Chikens[card.MAHJONG_BAM1]
	return exists && chikenType > config.CHIKEN_TYPE_BAM1
}

// 本局是否是金乌骨
func (this *Mahjong) isGoldDot8() bool {
	chikenType, exists := this.Chikens[card.MAHJONG_DOT8]
	return exists && chikenType&config.CHIKEN_TYPE_DOT8 > 0 && chikenType > config.CHIKEN_TYPE_DOT8
}

// 获取钻石鸡以及翻倍次数
// 钻石鸡有两种，所以需要通过传入的牌来避免冗余
func (m *Mahjong) getDiamondChikenTimes(tile int) int {
	times := 0
	if m.getChikenFBTile() == tile {
		times++
	}
	if m.getChikenDrawTile() == tile {
		times++
	}
	return times
}

// 获取爬坡鸡
func (m *Mahjong) getPaPoChiken() int {
	drawTile := m.getChikenDrawTile()
	// 非普通牌，没有爬坡鸡
	if !card.IsSuit(drawTile) {
		return 0
	}
	return card.GetBehindCardCycle(drawTile)
}

// 获取所有的与翻牌有关的鸡
// 如果开启了本鸡，自身也算
// 未去重
func (m *Mahjong) getDrawRelationChikens() []int {
	chikens := []int{}
	// 读取翻牌鸡
	chikens = append(chikens, m.getChikenDraw()...)
	// 读取前后鸡
	chikens = append(chikens, m.getChikenFB()...)
	// 读取滚筒鸡
	chikens = append(chikens, m.getChikenTumbling()...)
	// 如果支持本鸡，需要计算自身
	if m.setting.IsSettingChikenSelf() {
		if drawTile := m.getChikenDrawTile(); drawTile > 0 {
			chikens = append(chikens, drawTile)
		}
		if fbTile := m.getChikenFBTile(); fbTile > 0 {
			chikens = append(chikens, fbTile)
		}
		// TODO 暂不支持滚筒鸡
	}
	return chikens
}
