package robot

import (
	"sort"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	"mahjong.go/fbs/Common"
	"mahjong.go/mi/card"
	"mahjong.go/mi/protocal"
	configService "mahjong.go/service/config"
)

// 处理握手成功
func (this *Robot) handleHandShake(impacket *protocal.ImPacket) {
	js, _ := simplejson.NewJson(impacket.GetMessage())
	roomId, _ := js.Get("roomid").Int64()

	this.debug("success handleHandShake,userId:%d", this.UserId)

	// 发送ack
	this.HandShakeAck()

	// 开启机器人计时器
	this.startTimer()
	// this.HeartBeat()

	if roomId > 0 {
		// 已经处于房间中，等候restore
		this.show("已经处于房间中，等候restore, userId:%d, roomId:%d", this.UserId, roomId)
	} else {
		// 当前不在房间中，开始随机组队
		if configService.IsRandomRoom(this.CType) {
			this.RandomJoin()
		} else if configService.IsMatchRoom(this.CType) {
			this.MatchJoin()
		} else if configService.IsRankRoom(this.CType) {
			this.RankJoin()
		} else if configService.IsLeagueRoom(this.CType) {
			this.LeagueJoin()
		} else if configService.IsCoinRoom(this.CType) {
			this.CoinJoin()
		}
	}
}

// 处理收到心跳
func (this *Robot) handleHeartBeat(impacket *protocal.ImPacket) {
	this.trace("收到心跳, userId: %d.", this.UserId)
}

// 处理加入房间
func (this *Robot) handleJoinRoom(impacket *protocal.ImPacket) {
	response := Common.GetRootAsJoinRoomResponse(impacket.GetBody(), 0)
	result := new(Common.GameResult)
	result = response.S2cResult(result)
	if result.Code() < 0 {
		this.show("加入房间失败, userId:%d", this.UserId)
		this.quit(5)
	} else {
		this.RoomId = int64(response.RoomInfo(nil).RoomId())
		this.RoomNumber = string(response.RoomInfo(nil).Number())
		this.MType = int(response.RoomInfo(nil).GameType())
		this.CType = int(response.RoomInfo(nil).RandomRoom())
		this.TRound = int(response.RoomInfo(nil).Round())
		setting := response.RoomInfo(nil).SettingBytes()
		this.TileCnt = int(setting[10])

		// 初始化选牌器的牌数
		if this.TileCnt == 72 {
			this.ms.SetTiles(card.MahjongCards72)
		} else if this.TileCnt == 108 {
			this.ms.SetTiles(card.MahjongCards108)
		}

		this.debug("加入房间成功,userId:%v,roomId:%v,number:%v,gameType:%v,createType:%v,tileCnt:%v",
			this.UserId, this.RoomId, this.RoomNumber, this.MType, this.CType, this.TileCnt)
	}
}

// 处理房间关闭
func (this *Robot) handleCloseRoom(impacket *protocal.ImPacket) {
	this.debug("房间解散,roomId:%d,userId:%d", this.RoomId, this.UserId)
	this.quit(1)
}

// 处���收���解散房间的消息
func (this *Robot) handleDismissRoom(impacket *protocal.ImPacket) {
	response := Common.GetRootAsDismissRoomPush(impacket.GetBody(), 0)
	if response.Op() == int8(-1) {
		// 收到解散房间请求，自动回应同意申请房间
		time.Sleep(time.Second * time.Duration(this.DismissInterval+util.RandIntn(this.DismissRandom)))

		this.DismissReply()
	}
}

// 重连
func (this *Robot) handleGameRestore(impacket *protocal.ImPacket) {
	this.debug("机器人重连, roomId:%d, userId:%d", this.RoomId, this.UserId)
	response := Common.GetRootAsGameRestorePush(impacket.GetBody(), 0)
	gamePlayState := new(Common.GamePlayState)
	gamePlayState = response.GameplayState(gamePlayState)

	// 恢复游戏数据
	this.TileCnt = int(gamePlayState.WallTileCount())

	// 恢复房间数据
	roomInfo := new(Common.RoomInfo)
	roomInfo = gamePlayState.RoomInfo(roomInfo)
	this.RoomId = int64(roomInfo.RoomId())
	this.RoomNumber = string(roomInfo.Number())
	this.MType = int(roomInfo.GameType())
	this.CType = int(roomInfo.RandomRoom())
	this.TRound = int(roomInfo.Round())

	// 恢复用户数据
	ru := new(Common.RoomUserInfo)
	for i := 0; i < gamePlayState.RoomUserListLength(); i++ {
		gamePlayState.RoomUserList(ru, i)
		// 找到自己
		if int(ru.UserId()) == this.UserId {
			this.Index = int(ru.Index())
		}
	}

	// 回复用户牌局数据
	mu := new(Common.MahjongUserInfo_v_2_1_0)
	for i := 0; i < gamePlayState.MahjongUserInfoV210Length(); i++ {
		gamePlayState.MahjongUserInfoV210(mu, i)
		if int(mu.UserId()) == this.UserId {
			// 恢复自己的手牌
			for _, tile := range mu.HandTilesBytes() {
				this.HandTileList = append(this.HandTileList, int(tile))
			}
			sort.Ints(this.HandTileList)
			this.debug("恢复用户手牌, roomId:%d,userId: %d, 手牌:%#v:%d", this.RoomId, this.UserId, this.HandTileList, len(this.HandTileList))
		}
		// 恢复所有的弃牌
		for _, tile := range mu.PlayListBytes() {
			this.DiscardTileList = append(this.DiscardTileList, int(tile))
		}
		this.debug("恢复所有人的弃牌, roomId:%d, 弃牌:%#v:%d", this.RoomId, this.DiscardTileList, len(this.DiscardTileList))

		// 恢复所有的明牌
		showCard := new(Common.ShowCard_v_2_1_0)
		for j := 0; j < mu.ShowCardListLength(); j++ {
			mu.ShowCardList(showCard, j)
			for _, tile := range showCard.TilesBytes() {
				this.DiscardTileList = append(this.DiscardTileList, int(tile))
			}
		}
		this.debug("恢复所有人的弃牌(明牌也算), roomId:%d, 弃牌:%#v:%d", this.RoomId, this.DiscardTileList, len(this.DiscardTileList))
	}

	// 游戏状态
	// gameState := int(gamePlayState.GameStatus())
	// 如果已托管，需要取消托管
	for i := 0; i < gamePlayState.HostingUserLength(); i++ {
		if this.UserId == int(gamePlayState.HostingUser(i)) {
			this.debug("机器人重连, 需要取消托管, roomId:%d, userId:%d", this.RoomId, this.UserId)
			// 用户处于托管状态，需要取消托管
			this.CancelHosting()
		}
	}

	// 如果需要准备，帮用户准备
	if gamePlayState.PrepareStatus() == byte(0) {
		this.debug("机器人重连, 需要准备, roomId:%d, userId:%d", this.RoomId, this.UserId)
		this.Prepare(true)
	}

	// 回应操作
	operationPush := new(Common.OperationPush)
	if gamePlayState.OperationPushArrayLength() > 0 {
		this.debug("机器人重连, 需要回应操作, roomId:%d, userId:%d", this.RoomId, this.UserId)
		gamePlayState.OperationPushArray(operationPush, 0)
		this.handleOperationPush(operationPush)
	}

	this.debug("机器人重连完成, roomId:%d, userId:%d", this.RoomId, this.UserId)
}

// 处理报名比赛成功
func (this *Robot) handleLeagueApplyResponse(impacket *protocal.ImPacket) {
	response := Common.GetRootAsLeagueApplyResponse(impacket.GetBody(), 0)
	result := new(Common.GameResult)
	result = response.S2cResult(result)
	this.debug("收到报名比赛结果, userId:%v, leagueId:%v, raceId:%v, code:%v, message:%v", this.UserId, this.GameInfo.LeagueId, this.GameInfo.RaceId, result.Code(), string(result.Msg()))
	this.quit(7)
}
