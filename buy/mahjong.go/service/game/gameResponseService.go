package game

import (
	"encoding/json"
	"sort"

	"github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/library/response"
	"mahjong.go/mi/card"
	"mahjong.go/mi/oc"
	"mahjong.go/mi/protocal"
	"mahjong.go/model"

	flatbuffers "github.com/google/flatbuffers/go"
	fbsCommon "mahjong.go/fbs/Common"
	configService "mahjong.go/service/config"
)

// 构建一个fbs coommonresult
func genGameResult(builder *flatbuffers.Builder, err *core.Error) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	code := 0
	msg := ""
	if err != nil {
		code = err.GetCode()
		msg = err.Error()
	}
	errmsg := builder.CreateString(msg)
	fbsCommon.GameResultStart(builder)
	fbsCommon.GameResultAddCode(builder, int32(code))
	fbsCommon.GameResultAddMsg(builder, errmsg)
	commonResult := fbsCommon.GameResultEnd(builder)

	return builder, commonResult
}

// 构建一个room
func genRoom(builder *flatbuffers.Builder, room *Room) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var roomInfoBinary flatbuffers.UOffsetT
	var roomUsersBinary flatbuffers.UOffsetT
	var status int

	if room.Status > config.ROOM_STATUS_CREATING {
		status = 1
	}
	builder, roomInfoBinary = genRoomInfo(builder, room)
	builder, roomUsersBinary = genRoomUsers(builder, room.GetUsers())

	fbsCommon.RoomStart(builder)
	fbsCommon.RoomAddRoomInfo(builder, roomInfoBinary)
	fbsCommon.RoomAddRoomUsers(builder, roomUsersBinary)
	fbsCommon.RoomAddStatus(builder, int8(status))
	fbsCommon.RoomAddCreateTime(builder, int64(room.CreateTime))
	fbsCommon.RoomAddStartTime(builder, int64(room.StartTime))
	fbsCommon.RoomAddCurrentRound(builder, uint8(room.Round))
	return builder, fbsCommon.RoomEnd(builder)
}

// 构建一个roomInfo
func genRoomInfo(builder *flatbuffers.Builder, room *Room) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	number := builder.CreateString(room.Number)

	// 构建settingBinary
	var settingBinary flatbuffers.UOffsetT
	setting := room.setting.GetSetting()
	length := len(setting)
	fbsCommon.RoomInfoStartSettingVector(builder, length)
	for i := length - 1; i >= 0; i-- {
		builder.PrependByte(byte(setting[i]))
	}
	settingBinary = builder.EndVector(length)

	fbsCommon.RoomInfoStart(builder)
	fbsCommon.RoomInfoAddGameType(builder, uint16(room.MType))
	fbsCommon.RoomInfoAddRoomId(builder, uint64(room.RoomId))
	fbsCommon.RoomInfoAddRound(builder, uint8(room.TRound))
	fbsCommon.RoomInfoAddNumber(builder, number)
	fbsCommon.RoomInfoAddPlayerCount(builder, uint8(room.setting.GetSettingPlayerCnt()))
	fbsCommon.RoomInfoAddSetting(builder, settingBinary)
	fbsCommon.RoomInfoAddRandomRoom(builder, byte(room.CType))
	roomInfo := fbsCommon.RoomInfoEnd(builder)

	return builder, roomInfo
}

// 构建一个房间用户列表
func genRoomUsers(builder *flatbuffers.Builder, roomUsers map[int]*RoomUser) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var roomUserInfo flatbuffers.UOffsetT

	userCnt := len(roomUsers)
	users := make([]flatbuffers.UOffsetT, 0, userCnt)

	for _, roomUser := range roomUsers {
		builder, roomUserInfo = genRoomUserInfo(builder, roomUser)
		users = append(users, roomUserInfo)
	}

	fbsCommon.JoinRoomResponseStartRoomUserListVector(builder, userCnt)
	for _, userInfo := range users {
		builder.PrependUOffsetT(userInfo)
	}
	roomUserList := builder.EndVector(userCnt)

	return builder, roomUserList
}

// 构建一个房间用户信息
func genRoomUserInfo(builder *flatbuffers.Builder, roomUser *RoomUser) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 判断是否在线
	var online uint8
	if UserMap.IsUserExists(roomUser.UserId) || configService.IsRobot(roomUser.UserId) {
		online = uint8(1)
	} else {
		online = uint8(0)
	}

	// 生成对应的字符串
	nickname := builder.CreateString(roomUser.Info.Nickname)
	avatar := builder.CreateString(roomUser.Info.Avatar)
	ip := builder.CreateString(roomUser.Info.Ip)
	area := builder.CreateString(roomUser.Info.Area)

	fbsCommon.RoomUserInfoStart(builder)
	fbsCommon.RoomUserInfoAddUserId(builder, uint32(roomUser.UserId))
	fbsCommon.RoomUserInfoAddIndex(builder, uint8(roomUser.Index))
	fbsCommon.RoomUserInfoAddNickname(builder, nickname)
	fbsCommon.RoomUserInfoAddAvatar(builder, avatar)
	fbsCommon.RoomUserInfoAddIp(builder, ip)
	fbsCommon.RoomUserInfoAddArea(builder, area)
	fbsCommon.RoomUserInfoAddOnline(builder, online)
	// index=0的就是创建者
	if roomUser.Index == 0 {
		fbsCommon.RoomUserInfoAddIsHost(builder, byte(1))
	} else {
		fbsCommon.RoomUserInfoAddIsHost(builder, byte(0))
	}
	// 不显示总积分
	if configService.IsCreateRoom(roomUser.CType) ||
		configService.IsTVRoom(roomUser.CType) ||
		configService.IsClubRoom(roomUser.CType) {
		fbsCommon.RoomUserInfoAddScore(builder, int32(0))
	} else {
		fbsCommon.RoomUserInfoAddScore(builder, int32(roomUser.GetAccumulativeScore()))
	}

	fbsCommon.RoomUserInfoAddGender(builder, uint8(roomUser.Info.Gender))
	fbsCommon.RoomUserInfoAddMoney(builder, int32(roomUser.Info.Money))
	fbsCommon.RoomUserInfoAddAvatarBox(builder, int32(roomUser.Info.AvatarBox))
	roomUserInfo := fbsCommon.RoomUserInfoEnd(builder)

	return builder, roomUserInfo
}

// 构建一个operation对象
func genOperation(builder *flatbuffers.Builder, operation *Operation) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 组织tiles对象
	tileLen := len(operation.Tiles)

	var tilesBinary flatbuffers.UOffsetT
	if tileLen > 0 {
		fbsCommon.OperationStartTilesVector(builder, tileLen)
		for i := tileLen - 1; i >= 0; i-- {
			builder.PrependByte(byte(operation.Tiles[i]))
		}
		tilesBinary = builder.EndVector(tileLen)
	}

	// 构建对象
	fbsCommon.OperationStart(builder)
	fbsCommon.OperationAddOp(builder, byte(operation.OperationCode))
	if tileLen > 0 {
		fbsCommon.OperationAddTiles(builder, tilesBinary)
	}

	return builder, fbsCommon.OperationEnd(builder)
}

// 构建推送用户可进行的操作的push
func genOperationList(builder *flatbuffers.Builder, opList []*Operation) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	operationBinarys := make([]flatbuffers.UOffsetT, 0, len(opList))
	var operationBinary flatbuffers.UOffsetT
	for _, operation := range opList {
		builder, operationBinary = genOperation(builder, operation)
		operationBinarys = append(operationBinarys, operationBinary)
	}

	fbsCommon.OperationPushStartOpVector(builder, len(opList))
	for i := len(opList) - 1; i >= 0; i-- {
		builder.PrependUOffsetT(operationBinarys[i])
	}
	// for _, operationBinary := range operationBinarys {
	// 	builder.PrependUOffsetT(operationBinary)
	// }
	operationList := builder.EndVector(len(opList))

	return builder, operationList
}

// 构建手牌
func genShowCard(builder *flatbuffers.Builder, showCard *card.ShowCard) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 构建tiles二进制流
	var tilesBinary flatbuffers.UOffsetT
	tiles := showCard.GetTiles()
	tilesLen := len(tiles)
	if tilesLen > 0 {
		fbsCommon.ShowCardStartTilesVector(builder, tilesLen)
		for i := tilesLen - 1; i >= 0; i-- {
			builder.PrependByte(byte(tiles[i]))
		}
		tilesBinary = builder.EndVector(tilesLen)
	}

	// 这里需要做一个判断，如果是转弯杠，只有三张，则告诉客户端这个是碰
	opCode := showCard.GetOpCode()
	if oc.IsKongTurnOperation(opCode) && len(tiles) == 3 {
		opCode = fbsCommon.OperationCodePONG
	}

	// 构建数据
	fbsCommon.ShowCardStart(builder)
	fbsCommon.ShowCardAddOperationCode(builder, byte(opCode))
	fbsCommon.ShowCardAddTiles(builder, tilesBinary)

	return builder, fbsCommon.ShowCardEnd(builder)
}

// 构建手牌
func genShowCard_v_2_1_0(builder *flatbuffers.Builder, showCard *card.ShowCard) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 构建tiles二进制流
	var tilesBinary flatbuffers.UOffsetT
	tiles := showCard.GetTiles()
	tilesLen := len(tiles)
	if tilesLen > 0 {
		fbsCommon.ShowCardStartTilesVector(builder, tilesLen)
		for i := tilesLen - 1; i >= 0; i-- {
			builder.PrependByte(byte(tiles[i]))
		}
		tilesBinary = builder.EndVector(tilesLen)
	}

	// 这里需要做一个判断，如果是转弯杠，只有三张，则告诉客户端这个是碰
	opCode := showCard.GetOpCode()
	if oc.IsKongTurnOperation(opCode) && len(tiles) == 3 {
		opCode = fbsCommon.OperationCodePONG
	}

	// 构建数据
	fbsCommon.ShowCard_v_2_1_0Start(builder)
	fbsCommon.ShowCard_v_2_1_0AddOperationCode(builder, byte(opCode))
	fbsCommon.ShowCard_v_2_1_0AddTiles(builder, tilesBinary)
	fbsCommon.ShowCard_v_2_1_0AddTarget(builder, uint32(showCard.GetTarget()))
	if showCard.IsResponsibility() {
		fbsCommon.ShowCard_v_2_1_0AddResponsibility(builder, byte(1))
	} else {
		fbsCommon.ShowCard_v_2_1_0AddResponsibility(builder, byte(0))
	}
	return builder, fbsCommon.ShowCard_v_2_1_0End(builder)
}

// 构建手牌列表
func genShowCardList(builder *flatbuffers.Builder, sc *card.ShowCardList) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	showCardBinarys := make([]flatbuffers.UOffsetT, 0, sc.Len())
	var showCardBinary flatbuffers.UOffsetT
	for _, showCard := range sc.GetAll() {
		builder, showCardBinary = genShowCard(builder, showCard)
		showCardBinarys = append(showCardBinarys, showCardBinary)
	}
	fbsCommon.MahjongUserInfoStartShowCardListVector(builder, sc.Len())
	if sc.Len() > 0 {
		for i := sc.Len() - 1; i >= 0; i-- {
			builder.PrependUOffsetT(showCardBinarys[i])
		}
	}
	operationList := builder.EndVector(sc.Len())
	return builder, operationList
}

// 构建手牌列表
func genShowCardList_v_2_1_0(builder *flatbuffers.Builder, sc *card.ShowCardList) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	showCardBinarys := make([]flatbuffers.UOffsetT, 0, sc.Len())
	var showCardBinary flatbuffers.UOffsetT
	for _, showCard := range sc.GetAll() {
		builder, showCardBinary = genShowCard_v_2_1_0(builder, showCard)
		showCardBinarys = append(showCardBinarys, showCardBinary)
	}
	fbsCommon.MahjongUserInfoStartShowCardListVector(builder, sc.Len())
	if sc.Len() > 0 {
		for i := sc.Len() - 1; i >= 0; i-- {
			builder.PrependUOffsetT(showCardBinarys[i])
		}
	}
	operationList := builder.EndVector(sc.Len())
	return builder, operationList
}

// 构建mahjongUserInfo，仅含开局时的局内积分
func genMahjongUserInfoOnlyGameScore(builder *flatbuffers.Builder, userId int, score int) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	fbsCommon.MahjongUserInfoStart(builder)
	fbsCommon.MahjongUserInfoAddUserId(builder, uint32(userId))
	fbsCommon.MahjongUserInfoAddGameScore(builder, int32(score))

	return builder, fbsCommon.MahjongUserInfoEnd(builder)
}

func genMahjongUserInfo(builder *flatbuffers.Builder, room *Room, userId int, showHandTile bool, showGameDetail bool) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 打出去的牌列表
	var playList flatbuffers.UOffsetT
	var showCardList flatbuffers.UOffsetT
	var handTiles flatbuffers.UOffsetT
	var baoTing = 0
	var lackTile = 0
	var handTilesLen = 0
	// 鸡
	var chikenChargeBam1, chikenChargeDot8, chikenResponsibility int

	// 将用户手牌转成slice
	mUser := room.MI.getUsers()[userId]
	// 是否显示游戏详情
	if showGameDetail {
		// 读取用户手牌
		// 如果用户胡牌了, 将最后的牌换到牌的最后一个位置
		// 如果最后操作是抓牌，需要将抓牌用户抓的最后一张牌放到最后
		rightTile := mUser.WinTile
		if mUser.WinTile == 0 {
			if room.MI.getLastOperator() == userId &&
				(room.MI.getLastOperation().OperationCode == fbsCommon.OperationCodeDRAW ||
					room.MI.getLastOperation().OperationCode == fbsCommon.OperationCodeDRAW_AFTER_KONG) {
				rightTile = mUser.HandTileList.GetLastAdd()
			}
		}
		handTilesSlice := mUser.sortHandTile(rightTile)
		core.Logger.Debug("[genMahjongUserInfoOnlyGameScore]userId:%v, handTiles:%v", userId, handTilesSlice)
		handTilesLen = len(handTilesSlice)

		// 打出去的牌列表
		discardTiles := mUser.DiscardTileList.GetTiles()
		core.Logger.Debug("[genMahjongUserInfoOnlyGameScore]userId:%v, discardTiles:%v", userId, discardTiles)
		discardLen := len(discardTiles)
		if discardLen > 0 {
			fbsCommon.MahjongUserInfoStartPlayListVector(builder, discardLen)
			for i := discardLen - 1; i >= 0; i-- {
				builder.PrependByte(byte(discardTiles[i]))
			}
			playList = builder.EndVector(discardLen)
		}
		// 明牌列表
		if mUser.ShowCardList.Len() > 0 {
			builder, showCardList = genShowCardList(builder, mUser.ShowCardList)
		}

		// 显示用户的手牌
		// 如果已经打完了，则显示所有人的手牌
		if showHandTile {
			fbsCommon.MahjongUserInfoStartHandTilesVector(builder, handTilesLen)
			for i := handTilesLen - 1; i >= 0; i-- {
				builder.PrependByte(byte(handTilesSlice[i]))
			}
			handTiles = builder.EndVector(handTilesLen)
		}

		// 是否报听
		if mUser.MTC.IsBaoTing() {
			baoTing = 1
		}

		// 显示定缺
		// 如果所有人都定缺了，则显示所有人的定缺
		// 如果还有人没有定缺，则用户只显示自己的定缺情况
		if room.setting.IsEnableLack() {
			lackList := room.MI.getLackList()
			lackLen := len(lackList)
			if lackLen > 0 {
				lackTile = lackList[mUser.Index]
			} else if showHandTile {
				lackTile = mUser.LackTile
			}
		}

		if userId == room.MI.getChikenChargeBam1() {
			chikenChargeBam1 = 1
		}
		if userId == room.MI.getChikenChargeDot8() {
			chikenChargeDot8 = 1
		}
		if userId == room.MI.getChikenResponsibility() {
			chikenResponsibility = 1
		}
	}

	fbsCommon.MahjongUserInfoStart(builder)
	fbsCommon.MahjongUserInfoAddUserId(builder, uint32(mUser.UserId))
	fbsCommon.MahjongUserInfoAddGameScore(builder, int32(room.GetScoreInfo(userId).Score))
	fbsCommon.MahjongUserInfoAddHandTilesCount(builder, int32(handTilesLen))
	fbsCommon.MahjongUserInfoAddPlayList(builder, playList)
	fbsCommon.MahjongUserInfoAddShowCardList(builder, showCardList)
	fbsCommon.MahjongUserInfoAddHandTiles(builder, handTiles)
	fbsCommon.MahjongUserInfoAddBaoTing(builder, uint8(baoTing))
	fbsCommon.MahjongUserInfoAddLackTile(builder, uint8(lackTile))
	fbsCommon.MahjongUserInfoAddChikenChargeBam1(builder, uint8(chikenChargeBam1))
	fbsCommon.MahjongUserInfoAddChikenChargeDot8(builder, uint8(chikenChargeDot8))
	fbsCommon.MahjongUserInfoAddChikenResponsibility(builder, uint8(chikenResponsibility))

	return builder, fbsCommon.MahjongUserInfoEnd(builder)
}

func genMahjongUserInfo_v_2_1_0(builder *flatbuffers.Builder, room *Room, userId int, showHandTile bool, showGameDetail bool, self bool) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 打出去的牌列表
	var playList flatbuffers.UOffsetT
	var showCardList flatbuffers.UOffsetT
	var handTiles flatbuffers.UOffsetT
	// 选择换的牌
	var exchangeTiles flatbuffers.UOffsetT
	var baoTing = 0
	var lackTile = 0
	var handTilesLen = 0
	// 鸡
	var chikenChargeBam1, chikenChargeDot8, chikenResponsibility int

	// 将用户手牌转成slice
	mUser := room.MI.getUsers()[userId]
	// 是否显示游戏详情
	if showGameDetail {
		// 读取用户手牌
		// 如果用户胡牌了, 将最后的牌换到牌的最后一个位置
		// 如果最后操作是抓牌，需要将抓牌用户抓的最后一张牌放到最后
		rightTile := mUser.WinTile
		if mUser.WinTile == 0 {
			if room.MI.getLastOperator() == userId &&
				(room.MI.getLastOperation().OperationCode == fbsCommon.OperationCodeDRAW ||
					room.MI.getLastOperation().OperationCode == fbsCommon.OperationCodeDRAW_AFTER_KONG) {
				rightTile = mUser.HandTileList.GetLastAdd()
			}
		}
		handTilesSlice := mUser.sortHandTile(rightTile)
		handTilesLen = len(handTilesSlice)

		// 打出去的牌列表
		discardTiles := mUser.DiscardTileList.GetTiles()
		discardLen := len(discardTiles)
		if discardLen > 0 {
			fbsCommon.MahjongUserInfo_v_2_1_0StartPlayListVector(builder, discardLen)
			for i := discardLen - 1; i >= 0; i-- {
				builder.PrependByte(byte(discardTiles[i]))
			}
			playList = builder.EndVector(discardLen)
		}
		// 明牌列表
		if mUser.ShowCardList.Len() > 0 {
			builder, showCardList = genShowCardList_v_2_1_0(builder, mUser.ShowCardList)
		}

		// 显示用户的手牌
		// 如果已经打完了，则显示所有人的手牌
		if showHandTile {
			fbsCommon.MahjongUserInfo_v_2_1_0StartHandTilesVector(builder, handTilesLen)
			for i := handTilesLen - 1; i >= 0; i-- {
				builder.PrependByte(byte(handTilesSlice[i]))
			}
			handTiles = builder.EndVector(handTilesLen)
		}

		// 是否报听
		if mUser.MTC.IsBaoTing() {
			baoTing = 1
		}

		// 显示换牌
		// if room.MI.isExchanging() {
		core.Logger.Debug("=========,room.setting.IsEnableExchange():%v", room.setting.IsEnableExchange())
		if room.setting.IsEnableExchange() {
			tiles := []int{}
			exchangeCnt := len(mUser.ExchangeOutTiles)
			core.Logger.Debug("=========,exchangeCnt:%v, tiles:%v", exchangeCnt, mUser.ExchangeOutTiles)

			if exchangeCnt > 0 {
				if self {
					tiles = mUser.ExchangeOutTiles
				} else {
					for i := 0; i < exchangeCnt; i++ {
						tiles = append(tiles, 0)
					}
				}
				fbsCommon.MahjongUserInfo_v_2_1_0StartExchangeTilesVector(builder, exchangeCnt)
				for i := exchangeCnt - 1; i >= 0; i-- {
					builder.PrependByte(byte(tiles[i]))
				}
				exchangeTiles = builder.EndVector(exchangeCnt)
			}
		}

		// 显示定缺
		// 如果所有人都定缺了，则显示所有人的定缺
		// 如果还有人没有定缺，则用户只显示自己的定缺情况
		if room.setting.IsEnableLack() {
			lackList := room.MI.getLackList()
			lackLen := len(lackList)
			if lackLen > 0 {
				lackTile = lackList[mUser.Index]
			} else if showHandTile {
				lackTile = mUser.LackTile
			}
		}

		if userId == room.MI.getChikenChargeBam1() {
			chikenChargeBam1 = 1
		}
		if userId == room.MI.getChikenChargeDot8() {
			chikenChargeDot8 = 1
		}
		if userId == room.MI.getChikenResponsibility() {
			chikenResponsibility = 1
		}
	}

	fbsCommon.MahjongUserInfo_v_2_1_0Start(builder)
	fbsCommon.MahjongUserInfo_v_2_1_0AddUserId(builder, uint32(mUser.UserId))
	fbsCommon.MahjongUserInfo_v_2_1_0AddGameScore(builder, int32(room.GetScoreInfo(userId).Score))
	fbsCommon.MahjongUserInfo_v_2_1_0AddHandTilesCount(builder, int32(handTilesLen))
	fbsCommon.MahjongUserInfo_v_2_1_0AddPlayList(builder, playList)
	fbsCommon.MahjongUserInfo_v_2_1_0AddShowCardList(builder, showCardList)
	fbsCommon.MahjongUserInfo_v_2_1_0AddHandTiles(builder, handTiles)
	fbsCommon.MahjongUserInfo_v_2_1_0AddBaoTing(builder, uint8(baoTing))
	fbsCommon.MahjongUserInfo_v_2_1_0AddLackTile(builder, uint8(lackTile))
	fbsCommon.MahjongUserInfo_v_2_1_0AddChikenChargeBam1(builder, uint8(chikenChargeBam1))
	fbsCommon.MahjongUserInfo_v_2_1_0AddChikenChargeDot8(builder, uint8(chikenChargeDot8))
	fbsCommon.MahjongUserInfo_v_2_1_0AddChikenResponsibility(builder, uint8(chikenResponsibility))
	fbsCommon.MahjongUserInfo_v_2_1_0AddExchangeTiles(builder, exchangeTiles)

	return builder, fbsCommon.MahjongUserInfo_v_2_1_0End(builder)
}

// 构建MahjongUserList
func genMahjongUserList(builder *flatbuffers.Builder, userId int, gameStatus int, showGameDetail bool, room *Room) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var mahjongUserInfo flatbuffers.UOffsetT

	userCnt := len(room.MI.getUsers())
	users := make([]flatbuffers.UOffsetT, 0, userCnt)

	for _, mUser := range room.MI.getUsers() {
		builder, mahjongUserInfo = genMahjongUserInfo(builder, room, mUser.UserId, gameStatus == 3 || userId == mUser.UserId || room.Ob.hasUser(userId), showGameDetail || room.Ob.hasUser(userId))
		users = append(users, mahjongUserInfo)
	}

	fbsCommon.GamePlayStateStartMahjongUserInfoVector(builder, userCnt)
	for _, userInfo := range users {
		builder.PrependUOffsetT(userInfo)
	}
	mahjongUserList := builder.EndVector(userCnt)

	return builder, mahjongUserList
}

func genMahjongUserList_v_2_1_0(builder *flatbuffers.Builder, userId int, gameStatus int, showGameDetail bool, room *Room) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var mahjongUserInfo flatbuffers.UOffsetT

	userCnt := len(room.MI.getUsers())
	users := make([]flatbuffers.UOffsetT, 0, userCnt)

	for _, mUser := range room.MI.getUsers() {
		builder, mahjongUserInfo = genMahjongUserInfo_v_2_1_0(builder, room, mUser.UserId, gameStatus == 3 || userId == mUser.UserId || room.Ob.hasUser(userId), showGameDetail || room.Ob.hasUser(userId), mUser.UserId == userId)
		users = append(users, mahjongUserInfo)
	}

	fbsCommon.GamePlayStateStartMahjongUserInfoVector(builder, userCnt)
	for _, userInfo := range users {
		builder.PrependUOffsetT(userInfo)
	}
	mahjongUserList := builder.EndVector(userCnt)

	return builder, mahjongUserList
}

// 构建gameplayState
func genGameplayState(builder *flatbuffers.Builder, userId int, room *Room) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 游戏状态，0：组队中；1：准备中（检查作弊、或者需要用户准备）2：游戏中;3: 单局结算中
	var gameStatus int
	// 房间信息
	var roomInfo flatbuffers.UOffsetT
	// 房间用户信息
	var roomUserList flatbuffers.UOffsetT
	// 骰子数
	var dice flatbuffers.UOffsetT
	// 用户麻将信息
	var mahjongUserList flatbuffers.UOffsetT
	// 用户麻将信息
	var mahjongUserList_v_2_1_0 flatbuffers.UOffsetT
	// 准备状态
	var prepareStatus = 2
	// 定缺状态
	var lackStatus int
	// 定缺用户列表
	var lackedUsers flatbuffers.UOffsetT
	// 已准用户列表
	var prepareUsers flatbuffers.UOffsetT
	// 已同意解用户列表
	var dismissUsers flatbuffers.UOffsetT
	// 解散房间剩余同意时间
	var dismissRemainTime uint32
	// 进入消息
	var gameEnterPush flatbuffers.UOffsetT
	// 初始化操作
	var operationPush flatbuffers.UOffsetT
	var operationPushArray flatbuffers.UOffsetT
	var dealer int
	// 结算信息
	var settlementPush flatbuffers.UOffsetT
	var hostingUsers flatbuffers.UOffsetT
	// 已听的牌的信息
	var tingTiles flatbuffers.UOffsetT
	var seq int
	// 比赛信息
	var raceBinary flatbuffers.UOffsetT
	var raceRank int
	// 换牌状态
	var exchangeStatus int

	// 构建数据
	// 游戏状态
	if !room.IsFull() {
		// 组队中
		gameStatus = 0
	} else if room.IsReadying() {
		if room.Round == 0 || room.MI == nil {
			// 未开始
			gameStatus = 1
		} else {
			// 准备中
			gameStatus = 3
		}
	} else {
		// 游戏中
		gameStatus = 2
	}

	// 房间信息、房间用户信息
	builder, roomInfo = genRoomInfo(builder, room)
	builder, roomUserList = genRoomUsers(builder, room.GetUsers())
	// 计算庄家
	if gameStatus == 0 {
		for i := 0; i < room.setting.GetSettingPlayerCnt(); i++ {
			if v, exists := room.Index.Load(i); exists {
				dealer = v.(int)
				break
			}
		}
	} else if gameStatus == 1 {
		dealer = room.GetIndexUserId(0)
	} else {
		if room.MI == nil {
			core.Logger.Error("room.MI is nil, roomId:%v, round:%v", room.RoomId, room.Round)
			dealer = room.Dealer
		} else {
			dealer, _ = room.MI.getDealer()
		}
	}

	// 组队完成之后，需要检测解散状态
	if room.DismissTime > 0 {
		dismissRemainTime = uint32(room.DismissTime + int64(config.ROOM_DISMISS_AOTO_ALLOW_INTERVAL) - util.GetTime())
		dismissUserSlice := room.GetDismissUsers()
		length := len(dismissUserSlice)
		fbsCommon.GamePlayStateStartDismissUsersVector(builder, length)
		for i := length - 1; i >= 0; i-- {
			builder.PrependUint32(uint32(dismissUserSlice[i]))
		}
		dismissUsers = builder.EndVector(length)
	}

	// 牌局结束状态和未开始的时候，需要返回用户的准备状态
	if gameStatus == 1 || gameStatus == 3 {
		// 判断用户是否需要准备
		if room.IsReadying() {
			length := len(room.ReadyList)
			if length > 0 {
				// 返回所有已准备的用户
				fbsCommon.GamePlayStateStartPrepareUsersVector(builder, length)
				for _, v := range room.ReadyList {
					builder.PrependUint32(uint32(v))
				}
				prepareUsers = builder.EndVector(length)
			}
			// 检查用户是否已准备
			if room.Ob.isReady(userId) || util.IntInSlice(userId, room.ReadyList) {
				prepareStatus = 1
			} else {
				prepareStatus = 0
			}
		}
	}

	// 准备数据
	if gameStatus == 0 {
		// 暂时无逻辑
	} else if gameStatus == 1 {
	} else {
		mu := room.MI.getUser(userId)
		// 骰子
		fbsCommon.GamePlayStateStartDiceVector(builder, 2)
		builder.PrependUint8(byte(room.MI.getDiceList()[1]))
		builder.PrependUint8(byte(room.MI.getDiceList()[0]))
		dice = builder.EndVector(2)
		// 定缺状态
		lackList := room.MI.getLackList()
		if room.setting.IsEnableLack() && len(lackList) == 0 {
			if !room.Ob.hasUser(userId) && mu.LackTile > 0 {
				// 用户已定缺，等候其他人定缺
				lackStatus = 1
			}
			// 读取已定缺的用户列表
			lackedUserIds := room.MI.getLackedUsers()
			if lackedUserLen := len(lackedUserIds); lackedUserLen > 0 {
				fbsCommon.GamePlayStateStartLackedUsersVector(builder, lackedUserLen)
				for _, userId := range lackedUserIds {
					builder.PrependUint32(uint32(userId))
				}
				lackedUsers = builder.EndVector(lackedUserLen)
			}
		}
		// 换牌状态
		if room.setting.IsEnableExchange() && room.MI.isExchanging() {
			exchangeStatus = 1
		}

		if gameStatus == 2 {
			// 计算用户可进行的操作
			opList := room.MI.getRestoreOpreationlist(userId)
			if len(opList) > 0 {
				builder, operationPush = genOperationPush(builder, opList)
				fbsCommon.GamePlayStateStartOperationPushArrayVector(builder, 1)
				builder.PrependUOffsetT(operationPush)
				operationPushArray = builder.EndVector(1)
			}
			// 计算用户听牌状态
			if mu.MTC.IsTing() && len(mu.MTC.GetMaps()) > 0 {
				tiles := mu.MTC.GetTingTiles()
				fbsCommon.GamePlayStateStartTingTilesVector(builder, len(tiles))
				for _, tile := range tiles {
					builder.PrependByte(uint8(tile))
				}
				tingTiles = builder.EndVector(len(tiles))
			}
		} else if gameStatus == 3 {
			// 结算信息
			if prepareStatus == 0 {
				builder, settlementPush = genGameSettlementPush(builder, room)
			}
		}

		// 组织房间用户数据
		builder, mahjongUserList = genMahjongUserList(builder, userId, gameStatus, gameStatus == 2 || prepareStatus == 0, room)
		builder, mahjongUserList_v_2_1_0 = genMahjongUserList_v_2_1_0(builder, userId, gameStatus, gameStatus == 2 || prepareStatus == 0, room)

		// 最后消息编号
		seq = mu.MSC.GetSeq()
	}

	// 返回所有已托管的用户
	hostingUserCnt := len(room.HostingUsers)
	if hostingUserCnt > 0 {
		fbsCommon.GamePlayStateStartHostingUserVector(builder, hostingUserCnt)
		for _, v := range room.HostingUsers {
			builder.PrependUint32(uint32(v))
		}
		hostingUsers = builder.EndVector(hostingUserCnt)
	}
	if room.RaceInfo != nil && room.LeagueInfo != nil {
		builder, raceBinary = buildRace(builder, room.RaceInfo, room.LeagueInfo)
		raceRank = model.GetRaceUserRank(room.RaceInfo.Id, userId)
	}

	// 根据游戏有没有开始来组织数据
	fbsCommon.GamePlayStateStart(builder)
	fbsCommon.GamePlayStateAddGameStatus(builder, uint8(gameStatus))
	fbsCommon.GamePlayStateAddRoomInfo(builder, roomInfo)
	fbsCommon.GamePlayStateAddRoomUserList(builder, roomUserList)
	fbsCommon.GamePlayStateAddCurrentRound(builder, uint8(room.Round))
	fbsCommon.GamePlayStateAddDealer(builder, uint32(dealer))
	if room.DismissTime > 0 {
		fbsCommon.GamePlayStateAddDismissRemainTime(builder, dismissRemainTime)
		fbsCommon.GamePlayStateAddDismissUsers(builder, dismissUsers)
	}
	// 追加准备数据
	fbsCommon.GamePlayStateAddPrepareStatus(builder, uint8(prepareStatus))
	if prepareStatus != 2 {
		fbsCommon.GamePlayStateAddPrepareUsers(builder, prepareUsers)
	}
	// 追加定缺状态
	fbsCommon.GamePlayStateAddLackStatus(builder, uint8(lackStatus))
	fbsCommon.GamePlayStateAddLackedUsers(builder, lackedUsers)

	if gameStatus == 0 {

	} else if gameStatus == 1 {
		fbsCommon.GamePlayStateAddGameEnterPush(builder, gameEnterPush)
	} else {
		fbsCommon.GamePlayStateAddMahjongUserInfo(builder, mahjongUserList)
		fbsCommon.GamePlayStateAddMahjongUserInfoV210(builder, mahjongUserList_v_2_1_0)

		if gameStatus == 2 || prepareStatus == 0 {
			// 游戏中或者未准备，需要发牌局相关的数据
			fbsCommon.GamePlayStateAddDice(builder, dice)
			fbsCommon.GamePlayStateAddDealCount(builder, uint8(room.DealCount))
			fbsCommon.GamePlayStateAddWallTileCount(builder, uint8(room.MI.getWallTileCount()))
			fbsCommon.GamePlayStateAddChikenRoller(builder, uint8(room.MI.getChikenFBTile()))
			if room.MI.getChikenFBIndex() > 0 {
				fbsCommon.GamePlayStateAddChikenRollerIndex(builder, uint8(room.MI.getChikenFBIndex()-room.MI.getForward()))
			}
			fbsCommon.GamePlayStateAddLastPlayerId(builder, uint32(room.MI.getLastPlayerId()))
			fbsCommon.GamePlayStateAddCurrentPlayerId(builder, uint32(room.MI.getLastOperator()))
			fbsCommon.GamePlayStateAddDrawFront(builder, uint8(room.MI.getForward()))
			fbsCommon.GamePlayStateAddDrawBehind(builder, uint8(room.MI.getBackward()))
			if gameStatus == 2 {
				// 游戏中发用户可进行的操作
				fbsCommon.GamePlayStateAddOperationPushArray(builder, operationPushArray)
				// 游戏中发用户听的牌
				fbsCommon.GamePlayStateAddTingTiles(builder, tingTiles)
			} else if gameStatus == 3 {
				// 设置翻牌鸡
				fbsCommon.GamePlayStateAddChikenDraw(builder, uint8(room.MI.getChikenDrawTile()))
				// 游戏结算信息
				fbsCommon.GamePlayStateAddGameSettlementPush(builder, settlementPush)
			}
		}
	}
	// 用户托管状态
	fbsCommon.GamePlayStateAddHostingUser(builder, hostingUsers)
	// 最后消息序号
	fbsCommon.GamePlayStateAddStep(builder, uint16(seq))
	// 比赛信息
	fbsCommon.GamePlayStateAddRaceInfo(builder, raceBinary)
	fbsCommon.GamePlayStateAddRaceRank(builder, int32(raceRank))
	// 房间底分
	if room.IsCoin() {
		fbsCommon.GamePlayStateAddConsumeCoin(builder, int32(room.CoinConfig.ConsumeCoin))
	}
	// 本局倍数
	fbsCommon.GamePlayStateAddMultipleRound(builder, int32(room.setting.MultipleRound))
	// 换牌状态
	fbsCommon.GamePlayStateAddExchangeStatus(builder, byte(exchangeStatus))

	return builder, fbsCommon.GamePlayStateEnd(builder)
}

// 构建比赛信息
func buildRace(builder *flatbuffers.Builder, raceInfo *model.Race, leagueInfo *model.League) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var name, img, rounds, settingBinary, rewardsBinary flatbuffers.UOffsetT
	var setting []byte
	name = builder.CreateString(leagueInfo.Name)
	img = builder.CreateString(leagueInfo.Img)
	rounds = builder.CreateString(leagueInfo.Rounds)
	json.Unmarshal([]byte(leagueInfo.Setting), &setting)

	length := len(setting)
	fbsCommon.RaceStartSettingVector(builder, length)
	for i := length - 1; i >= 0; i-- {
		builder.PrependByte(byte(setting[i]))
	}
	settingBinary = builder.EndVector(length)

	rewards := model.GetFrontLeagueRewards(leagueInfo.Id)
	rewardsLen := len(rewards)
	rewardsBinaryList := make([]flatbuffers.UOffsetT, 0, rewardsLen)
	for _, reward := range rewards {
		rewardsBinaryList = append(rewardsBinaryList, builder.CreateString(reward))
	}
	fbsCommon.LeagueStartRewardsVector(builder, rewardsLen)
	for i := rewardsLen - 1; i >= 0; i-- {
		builder.PrependUOffsetT(rewardsBinaryList[i])
	}
	rewardsBinary = builder.EndVector(rewardsLen)

	fbsCommon.RaceStart(builder)
	fbsCommon.RaceAddRaceId(builder, raceInfo.Id)
	fbsCommon.RaceAddLeagueId(builder, int32(raceInfo.LeagueId))
	fbsCommon.RaceAddName(builder, name)
	fbsCommon.RaceAddImg(builder, img)
	fbsCommon.RaceAddIcon(builder, int32(leagueInfo.Icon))
	fbsCommon.RaceAddGameType(builder, uint16(leagueInfo.GameType))
	fbsCommon.RaceAddSetting(builder, settingBinary)
	fbsCommon.RaceAddRounds(builder, rounds)
	fbsCommon.RaceAddRequireUserCount(builder, int32(leagueInfo.RequireUserCount))
	if raceInfo.SignupUserCount > leagueInfo.RequireUserCount {
		fbsCommon.RaceAddSignupUserCount(builder, int32(leagueInfo.RequireUserCount))
	} else {
		fbsCommon.RaceAddSignupUserCount(builder, int32(raceInfo.SignupUserCount))
	}
	fbsCommon.RaceAddPrice(builder, int32(leagueInfo.Price))
	fbsCommon.RaceAddLeagueType(builder, int32(leagueInfo.LeagueType))
	fbsCommon.RaceAddSignupTime(builder, raceInfo.SignTime)
	fbsCommon.RaceAddGiveupTime(builder, raceInfo.GiveupTime)
	fbsCommon.RaceAddStartTime(builder, raceInfo.StartTime)
	fbsCommon.RaceAddStatus(builder, int32(raceInfo.Status))
	fbsCommon.RaceAddRound(builder, int32(raceInfo.Round))
	fbsCommon.RaceAddRewards(builder, rewardsBinary)
	fbsCommon.RaceAddRequireUserMin(builder, int32(leagueInfo.RequireUserMin))
	fbsCommon.RaceAddCategory(builder, uint8(leagueInfo.Category))
	raceBinary := fbsCommon.RaceEnd(builder)

	return builder, raceBinary
}

// 构建比赛用户
func buildRaceUser(builder *flatbuffers.Builder, raceUserInfo *model.RaceUser) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	fbsCommon.RaceUserStart(builder)
	fbsCommon.RaceUserAddRaceId(builder, raceUserInfo.RaceId)
	fbsCommon.RaceUserAddUserId(builder, int32(raceUserInfo.UserId))
	fbsCommon.RaceUserAddRound(builder, int32(raceUserInfo.Round))
	fbsCommon.RaceUserAddStatus(builder, int32(raceUserInfo.Status))
	fbsCommon.RaceUserAddScore(builder, int32(raceUserInfo.Score))
	fbsCommon.RaceUserAddSignupTime(builder, raceUserInfo.SignTime)
	fbsCommon.RaceUserAddGiveupTime(builder, raceUserInfo.GiveupTime)
	fbsCommon.RaceUserAddRank(builder, int32(raceUserInfo.Rank))
	raceUserBinary := fbsCommon.RaceUserEnd(builder)
	return builder, raceUserBinary
}

// 构建ScoreItem
func genScoreItem(builder *flatbuffers.Builder, item *FrontItem) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 构建tiles二进制流
	var tilesBinary flatbuffers.UOffsetT
	if item == nil {
		item = &FrontItem{}
	}
	tilesLen := len(item.Tiles)
	if tilesLen > 0 {
		fbsCommon.ScoreItemStartTilesVector(builder, tilesLen)
		for i := tilesLen - 1; i >= 0; i-- {
			builder.PrependByte(byte(item.Tiles[i]))
		}
		tilesBinary = builder.EndVector(tilesLen)
	}

	fbsCommon.ScoreItemStart(builder)
	fbsCommon.ScoreItemAddTypeId(builder, uint16(item.TypeId))
	fbsCommon.ScoreItemAddCount(builder, uint32(item.Count))
	fbsCommon.ScoreItemAddTiles(builder, tilesBinary)
	fbsCommon.ScoreItemAddScore(builder, int32(item.Score))
	fbsCommon.ScoreItemAddScoreCount(builder, uint8(item.ScoreCount))

	return builder, fbsCommon.ScoreItemEnd(builder)
}

// 构建ScoreItemV230
func genScoreItemV230(builder *flatbuffers.Builder, item *FrontItem) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 构建tiles二进制流
	var tilesBinary flatbuffers.UOffsetT
	if item == nil {
		item = &FrontItem{}
	}
	tilesLen := len(item.Tiles)
	if tilesLen > 0 {
		fbsCommon.ScoreItem_v_2_3_0StartTilesVector(builder, tilesLen)
		for i := tilesLen - 1; i >= 0; i-- {
			builder.PrependByte(byte(item.Tiles[i]))
		}
		tilesBinary = builder.EndVector(tilesLen)
	}

	fbsCommon.ScoreItem_v_2_3_0Start(builder)
	fbsCommon.ScoreItem_v_2_3_0AddTypeId(builder, uint16(item.TypeId))
	fbsCommon.ScoreItem_v_2_3_0AddCount(builder, uint32(item.Count))
	fbsCommon.ScoreItem_v_2_3_0AddTiles(builder, tilesBinary)
	fbsCommon.ScoreItem_v_2_3_0AddScore(builder, int32(item.Score))
	fbsCommon.ScoreItem_v_2_3_0AddGroup(builder, item.Group)
	fbsCommon.ScoreItem_v_2_3_0AddScoreCount(builder, uint8(item.ScoreCount))

	return builder, fbsCommon.ScoreItem_v_2_3_0End(builder)
}

// 构建ScoreItem集合
func genScoreItems(builder *flatbuffers.Builder, items []*FrontItem) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	positiveItemBinarys := make([]flatbuffers.UOffsetT, 0)
	nagaviteItemBinarys := make([]flatbuffers.UOffsetT, 0)
	var scoreItemBinary flatbuffers.UOffsetT
	// 先按typeId排序
	// 再将加分放在前面，扣分放在后面
	sort.Sort(FrontItemSorted(items))
	for _, item := range items {
		builder, scoreItemBinary = genScoreItem(builder, item)
		if item.Score > 0 {
			positiveItemBinarys = append(positiveItemBinarys, scoreItemBinary)
		} else {
			nagaviteItemBinarys = append(nagaviteItemBinarys, scoreItemBinary)
		}
	}
	fbsCommon.SettlementInfoStartScoreItemsVector(builder, len(items))
	for _, scoreItemBinary := range nagaviteItemBinarys {
		builder.PrependUOffsetT(scoreItemBinary)
	}
	for _, scoreItemBinary := range positiveItemBinarys {
		builder.PrependUOffsetT(scoreItemBinary)
	}
	return builder, builder.EndVector(len(items))
}

// 构建ScoreItemV230集合
func genScoreItemsV230(builder *flatbuffers.Builder, items []*FrontItem) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	positiveItemBinarys := make([]flatbuffers.UOffsetT, 0)
	nagaviteItemBinarys := make([]flatbuffers.UOffsetT, 0)
	var scoreItemBinary flatbuffers.UOffsetT
	// 先按typeId排序
	// 再将加分放在前面，扣分放在后面
	sort.Sort(FrontItemSorted(items))
	for _, item := range items {
		builder, scoreItemBinary = genScoreItemV230(builder, item)
		if item.Score > 0 {
			positiveItemBinarys = append(positiveItemBinarys, scoreItemBinary)
		} else {
			nagaviteItemBinarys = append(nagaviteItemBinarys, scoreItemBinary)
		}
	}
	fbsCommon.SettlementInfo_v_2_3_0StartScoreItemsVector(builder, len(items))
	for _, scoreItemBinary := range nagaviteItemBinarys {
		builder.PrependUOffsetT(scoreItemBinary)
	}
	for _, scoreItemBinary := range positiveItemBinarys {
		builder.PrependUOffsetT(scoreItemBinary)
	}
	return builder, builder.EndVector(len(items))
}

// 构建SettlementInfo
func genSettlementInfo(builder *flatbuffers.Builder, settlementInfo *FrontScoreInfo) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 构建结算明细
	var scoreItemsBinary flatbuffers.UOffsetT
	builder, scoreItemsBinary = genScoreItems(builder, settlementInfo.Item)

	fbsCommon.SettlementInfoStart(builder)
	fbsCommon.SettlementInfoAddUserId(builder, uint32(settlementInfo.UserId))
	fbsCommon.SettlementInfoAddWinWay(builder, uint8(settlementInfo.WinWay))
	fbsCommon.SettlementInfoAddBaoTing(builder, uint8(settlementInfo.BaoTing))
	fbsCommon.SettlementInfoAddWinStatus(builder, uint8(settlementInfo.WinStatus))
	fbsCommon.SettlementInfoAddTotalScore(builder, int32(settlementInfo.Total))
	fbsCommon.SettlementInfoAddGameScore(builder, int32(settlementInfo.GameScore))
	fbsCommon.SettlementInfoAddScoreItems(builder, scoreItemsBinary)
	fbsCommon.SettlementInfoAddTingStatus(builder, uint8(settlementInfo.TingStatus))
	fbsCommon.SettlementInfoAddHuStatus(builder, uint8(settlementInfo.HuStatus))
	fbsCommon.SettlementInfoAddPaoStatus(builder, uint8(settlementInfo.PaoStatus))

	return builder, fbsCommon.SettlementInfoEnd(builder)
}

// 构建SettlementInfoV230
func genSettlementInfoV230(builder *flatbuffers.Builder, settlementInfo *FrontScoreInfo) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 构建结算明细
	var scoreItemsBinary flatbuffers.UOffsetT
	builder, scoreItemsBinary = genScoreItemsV230(builder, settlementInfo.Item)

	fbsCommon.SettlementInfo_v_2_3_0Start(builder)
	fbsCommon.SettlementInfo_v_2_3_0AddUserId(builder, uint32(settlementInfo.UserId))
	fbsCommon.SettlementInfo_v_2_3_0AddWinWay(builder, uint8(settlementInfo.WinWay))
	fbsCommon.SettlementInfo_v_2_3_0AddBaoTing(builder, uint8(settlementInfo.BaoTing))
	fbsCommon.SettlementInfo_v_2_3_0AddWinStatus(builder, uint8(settlementInfo.WinStatus))
	fbsCommon.SettlementInfo_v_2_3_0AddTotalScore(builder, int32(settlementInfo.Total))
	fbsCommon.SettlementInfo_v_2_3_0AddGameScore(builder, int32(settlementInfo.GameScore))
	fbsCommon.SettlementInfo_v_2_3_0AddScoreItems(builder, scoreItemsBinary)
	fbsCommon.SettlementInfo_v_2_3_0AddTingStatus(builder, uint8(settlementInfo.TingStatus))
	fbsCommon.SettlementInfo_v_2_3_0AddHuStatus(builder, uint8(settlementInfo.HuStatus))
	fbsCommon.SettlementInfo_v_2_3_0AddPaoStatus(builder, uint8(settlementInfo.PaoStatus))

	return builder, fbsCommon.SettlementInfo_v_2_3_0End(builder)
}

// 构建鸡排信息
func genChikenInfo(builder *flatbuffers.Builder, chiken *FrontChikenInfo) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	extraString := builder.CreateString(chiken.Extra)
	fbsCommon.ChikenInfoStart(builder)
	fbsCommon.ChikenInfoAddTile(builder, uint8(chiken.Tile))
	fbsCommon.ChikenInfoAddIsRecharge(builder, byte(chiken.IsRecharge))
	fbsCommon.ChikenInfoAddIsBao(builder, byte(chiken.IsBao))
	fbsCommon.ChikenInfoAddIsGold(builder, byte(chiken.IsGlod))
	fbsCommon.ChikenInfoAddChikenType(builder, uint16(chiken.ChikenType))
	fbsCommon.ChikenInfoAddExtra(builder, extraString)
	return builder, fbsCommon.ChikenInfoEnd(builder)
}

// 构建结算鸡牌集合
func genChikenInfos(builder *flatbuffers.Builder, chikens []*FrontChikenInfo) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	length := len(chikens)

	var chikenBinary flatbuffers.UOffsetT
	chikenBinarys := make([]flatbuffers.UOffsetT, 0, length)

	for _, chiken := range chikens {
		builder, chikenBinary = genChikenInfo(builder, chiken)
		chikenBinarys = append(chikenBinarys, chikenBinary)
	}

	fbsCommon.SettlementChikensStartPlayChikensVector(builder, length)
	for i := length - 1; i >= 0; i-- {
		builder.PrependUOffsetT(chikenBinarys[i])
	}
	return builder, builder.EndVector(length)
}

// 构建结算鸡牌信息
func genSettlementChikenInfo(builder *flatbuffers.Builder, settlementInfo *FrontScoreInfo) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var playChikensBinary flatbuffers.UOffsetT
	var handChikensBinary flatbuffers.UOffsetT
	var showCardChikensBinary flatbuffers.UOffsetT

	if len(settlementInfo.PlayChikens) > 0 {
		builder, playChikensBinary = genChikenInfos(builder, settlementInfo.PlayChikens)
	}
	if len(settlementInfo.HandChikens) > 0 {
		builder, handChikensBinary = genChikenInfos(builder, settlementInfo.HandChikens)
	}
	if len(settlementInfo.ShowCardChikens) > 0 {
		builder, showCardChikensBinary = genChikenInfos(builder, settlementInfo.ShowCardChikens)
	}

	fbsCommon.SettlementChikensStart(builder)
	fbsCommon.SettlementChikensAddUserId(builder, uint32(settlementInfo.UserId))
	fbsCommon.SettlementChikensAddPlayChikens(builder, playChikensBinary)
	fbsCommon.SettlementChikensAddHandChikens(builder, handChikensBinary)
	fbsCommon.SettlementChikensAddShowCardChikens(builder, showCardChikensBinary)
	return builder, fbsCommon.SettlementChikensEnd(builder)
}

// 构建一个json的成功response
func GenJsonSuccess(packageId uint8) *protocal.ImPacket {
	js := simplejson.New()
	js.Set("code", 0)
	js.Set("message", "")

	return response.GenJson(packageId, js)
}

// 构建一个Json的错误response
func GenJsonError(packageId uint8, err *core.Error) *protocal.ImPacket {
	js := simplejson.New()
	js.Set("code", err.GetCode())
	js.Set("message", err.Error())

	return response.GenJson(packageId, js)
}

// 构建握手成功回复消息
func HandShakeResponse(responseMap map[string]interface{}) *protocal.ImPacket {
	js := simplejson.New()
	js.Set("code", 0)
	js.Set("message", "")
	for k, v := range responseMap {
		js.Set(k, v)
	}

	return response.GenJson(protocal.PACKAGE_TYPE_HANDSHAKE, js)
}

// 构建心跳回复
func HeartBeatResponse() *protocal.ImPacket {
	return response.GenEmpty(protocal.PACKAGE_TYPE_HEARTBEAT)
}

// 构建一个closeRoomPush
func CloseRoomPush(code int, msg string) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	msgBinary := builder.CreateString(msg)
	fbsCommon.CloseRoomPushStart(builder)
	fbsCommon.CloseRoomPushAddCode(builder, int32(code))
	fbsCommon.CloseRoomPushAddMsg(builder, msgBinary)
	orc := fbsCommon.CloseRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandCloseRoomPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构建一个dismissRoomPush
func DismissRoomPush(userId int, op int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.DismissRoomPushStart(builder)
	fbsCommon.DismissRoomPushAddUserId(builder, uint32(userId))
	fbsCommon.DismissRoomPushAddOp(builder, int8(op))
	orc := fbsCommon.DismissRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandDismissRoomPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构建一个quitRoomPush
func QuitRoomPush(userId int, index int, code int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.QuitRoomPushStart(builder)
	fbsCommon.QuitRoomPushAddUserId(builder, uint32(userId))
	fbsCommon.QuitRoomPushAddIndex(builder, byte(index))
	fbsCommon.QuitRoomPushAddCode(builder, int32(code))
	orc := fbsCommon.QuitRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandQuitRoomPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构建一个randomJoin push
func JoinRoomPush(roomUser *RoomUser) *protocal.ImPacket {
	var roomUserInfo flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	// 生成roomUserInfo
	builder, roomUserInfo = genRoomUserInfo(builder, roomUser)

	// 生成fbs buf
	fbsCommon.JoinRoomPushStart(builder)
	fbsCommon.JoinRoomPushAddRoomUserInfo(builder, roomUserInfo)
	orc := fbsCommon.JoinRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandJoinRoomPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构建一个JoinRoom成功的response
func JoinRoomResponse(room *Room, mNumber uint16) *protocal.ImPacket {
	var commonResult flatbuffers.UOffsetT
	var roomInfo flatbuffers.UOffsetT
	var roomUserList flatbuffers.UOffsetT
	// 已准用户列表
	var prepareUsers flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	// 生成commonResult
	builder, commonResult = genGameResult(builder, nil)
	// 生成roomInfo
	builder, roomInfo = genRoomInfo(builder, room)
	// 生成roomUserList
	builder, roomUserList = genRoomUsers(builder, room.GetUsers())
	// 生成已准备用户列表
	length := len(room.ReadyList)
	if room.IsFull() && length > 0 {
		fbsCommon.JoinRoomResponseStartPrepareUsersVector(builder, length)
		for _, v := range room.ReadyList {
			builder.PrependUint32(uint32(v))
		}
		prepareUsers = builder.EndVector(length)
	}

	// 构建对象
	fbsCommon.JoinRoomResponseStart(builder)
	fbsCommon.JoinRoomResponseAddS2cResult(builder, commonResult)
	fbsCommon.JoinRoomResponseAddRoomInfo(builder, roomInfo)
	fbsCommon.JoinRoomResponseAddRoomUserList(builder, roomUserList)
	fbsCommon.JoinRoomResponseAddPrepareUsers(builder, prepareUsers)
	orc := fbsCommon.JoinRoomResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandJoinRoomResponse, protocal.MSG_TYPE_RESPONSE, uint16(0), mNumber, buf)
}

// 构建一个JoinRoom失败的response
func JoinRoomFailedResponse(err *core.Error, mNumber uint16) *protocal.ImPacket {
	var commonResult flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	// 生成commonResult
	builder, commonResult = genGameResult(builder, err)

	// 构建对象
	fbsCommon.JoinRoomResponseStart(builder)
	fbsCommon.JoinRoomResponseAddS2cResult(builder, commonResult)
	orc := fbsCommon.JoinRoomResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandJoinRoomResponse, protocal.MSG_TYPE_RESPONSE, uint16(0), mNumber, buf)
}

func genGameInitPush(builder *flatbuffers.Builder, dealer int, dealCount int, dice [2]int, tiles []int, round int) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 构建骰子数据
	var diceBinary flatbuffers.UOffsetT
	fbsCommon.GameInitPushStartDiceVector(builder, 2)
	builder.PrependUint8(byte(dice[1]))
	builder.PrependUint8(byte(dice[0]))
	diceBinary = builder.EndVector(2)

	// 构建手牌数据
	var tileBinary flatbuffers.UOffsetT
	fbsCommon.GameInitPushStartTilesVector(builder, len(tiles))
	// 用户手牌toSlice的时候已排序，这个设计不太好，但是用的地方比较多，影响太大，暂不修改
	// 在给用户发手牌的时候，重新生成一个乱序slice
	// tiles = util.ShuffleSliceInt(tiles)
	for i := len(tiles) - 1; i >= 0; i-- {
		builder.PrependUint8(byte(tiles[i]))
	}
	tileBinary = builder.EndVector(len(tiles))

	// 构建对象
	fbsCommon.GameInitPushStart(builder)
	// 当前局数
	fbsCommon.GameInitPushAddCurrentRound(builder, uint8(round))
	// 庄家
	fbsCommon.GameInitPushAddDealer(builder, uint32(dealer))
	// 连庄数
	fbsCommon.GameInitPushAddDealCount(builder, uint8(dealCount))
	// 骰子数
	fbsCommon.GameInitPushAddDice(builder, diceBinary)
	// 手牌
	fbsCommon.GameInitPushAddTiles(builder, tileBinary)

	orc := fbsCommon.GameInitPushEnd(builder)

	return builder, orc
}

// 构建一个游戏初始化的push
func GameInitPush(dealer int, dealCount int, dice [2]int, tiles []int, round int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	builder, orc := genGameInitPush(builder, dealer, dealCount, dice, tiles, round)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameInitPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构建 一个作弊确认的builder
func genGameEnterPush(builder *flatbuffers.Builder) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	fbsCommon.GameEnterPushStart(builder)
	orc := fbsCommon.GameEnterPushEnd(builder)
	return builder, orc
}

// 构建一个作弊确认的push
func GameEnterPush() *protocal.ImPacket {
	var orc flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, orc = genGameEnterPush(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameEnterPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构建一个用户准备好的消息
func GameReadyPush(userId int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.GameReadyPushStart(builder)
	fbsCommon.GameReadyPushAddReadying(builder, uint8(1))
	fbsCommon.GameReadyPushAddUserId(builder, uint32(userId))
	orc := fbsCommon.GameReadyPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameReadyPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构建用户消息
func genUserOperation(builder *flatbuffers.Builder, operation *UserOperation) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 生成operation的二进制流
	var operationBinary flatbuffers.UOffsetT
	builder, operationBinary = genOperation(builder, operation.Op)

	// 构建返回的对象
	fbsCommon.UserOperationPushStart(builder)
	fbsCommon.UserOperationPushAddUserId(builder, uint32(operation.UserId))
	fbsCommon.UserOperationPushAddOp(builder, operationBinary)
	return builder, fbsCommon.UserOperationPushEnd(builder)
}

// UserOperationPush 构建发送用户操作的消息（不带seq）
func UserOperationPush(operation *UserOperation) *protocal.ImPacket {
	return UserOperationPushWithSeq(operation, 0)
}

// UserOperationPushWithSeq 构建发送用户操作的消息（带seq）
func UserOperationPushWithSeq(operation *UserOperation, seq int) *protocal.ImPacket {
	var orc flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, orc = genUserOperation(builder, operation)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandUserOperationPush, protocal.MSG_TYPE_PUSH, uint16(seq), uint16(0), buf)
}

// 构建推送用户可进行的操作的builder
func genOperationPush(builder *flatbuffers.Builder, opList []*Operation) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var operationList flatbuffers.UOffsetT
	builder, operationList = genOperationList(builder, opList)

	fbsCommon.OperationPushStart(builder)
	fbsCommon.OperationPushAddOp(builder, operationList)
	orc := fbsCommon.OperationPushEnd(builder)

	return builder, orc
}

// OperationPush 构建推送用户可进行的操作的push(不带seq)
func OperationPush(opList []*Operation) *protocal.ImPacket {
	return OperationPushWithSeq(opList, 0)
}

// OperationPushWithSeq 构建推送用户可进行的操作的push(带seq)
func OperationPushWithSeq(opList []*Operation, seq int) *protocal.ImPacket {
	var orc flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, orc = genOperationPush(builder, opList)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandOperationPush, protocal.MSG_TYPE_PUSH, uint16(seq), uint16(0), buf)
}

// 构建房间聊天的push
func RoomChatPush(userId int, chatId int16, memberId uint8, content string) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	pContent := builder.CreateString(content)
	fbsCommon.RoomChatPushStart(builder)
	fbsCommon.RoomChatPushAddUserId(builder, uint32(userId))
	fbsCommon.RoomChatPushAddChatId(builder, chatId)
	fbsCommon.RoomChatPushAddMemberId(builder, memberId)
	fbsCommon.RoomChatPushAddContent(builder, pContent)
	orc := fbsCommon.RoomChatPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandRoomChatPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 推送用户上下线的消息
func UserOnlinePush(userId int, online int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.UserOnlinePushStart(builder)
	fbsCommon.UserOnlinePushAddUserId(builder, uint32(userId))
	fbsCommon.UserOnlinePushAddOnline(builder, uint8(online))
	orc := fbsCommon.UserOnlinePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandUserOnlinePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构建客户端操作
func genClientOperation(builder *flatbuffers.Builder, operation *ClientOperation) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 生成operation的二进制流
	var operationBinary flatbuffers.UOffsetT
	builder, operationBinary = genOperation(builder, operation.Op)

	// 构建返回的对象
	fbsCommon.ClientOperationPushStart(builder)
	fbsCommon.ClientOperationPushAddUserId(builder, uint32(operation.UserId))
	fbsCommon.ClientOperationPushAddOp(builder, operationBinary)
	return builder, fbsCommon.ClientOperationPushEnd(builder)
}

// ClientOperationPush 推送客户端操作的消息(不带seq)
func ClientOperationPush(operation *ClientOperation) *protocal.ImPacket {
	return ClientOperationPushWithSeq(operation, 0)
}

// ClientOperationPushWithSeq 推送客户端操作的消息(带seq)
func ClientOperationPushWithSeq(operation *ClientOperation, seq int) *protocal.ImPacket {
	var orc flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, orc = genClientOperation(builder, operation)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClientOperationPush, protocal.MSG_TYPE_PUSH, uint16(seq), uint16(0), buf)
}

// 推送付����消息
func UpdateMoneyPush(amount int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.UpdateMoneyPushStart(builder)
	fbsCommon.UpdateMoneyPushAddAmount(builder, int32(amount))
	orc := fbsCommon.UpdateMoneyPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandUpdateMoneyPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 重连失败
func GameRestoreFailpush(err *core.Error) *protocal.ImPacket {
	var gameResult flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	// 生成gameResult
	builder, gameResult = genGameResult(builder, err)

	fbsCommon.GameRestorePushStart(builder)
	fbsCommon.GameRestorePushAddS2cResult(builder, gameResult)
	orc := fbsCommon.GameRestorePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameRestorePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// GameRestorePush 重连成功
func GameRestorePush(userId int, room *Room) *protocal.ImPacket {
	var gameplayState flatbuffers.UOffsetT
	var commonResult flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	// 生成commonResult
	builder, commonResult = genGameResult(builder, nil)
	// 生成gameplayState
	builder, gameplayState = genGameplayState(builder, userId, room)

	fbsCommon.GameRestorePushStart(builder)
	fbsCommon.GameRestorePushAddS2cResult(builder, commonResult)
	fbsCommon.GameRestorePushAddGameplayState(builder, gameplayState)
	orc := fbsCommon.GameRestorePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameRestorePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// GameRestoreSectionPush 片断数据获取
func GameRestoreSectionPush(userId int, room *Room, seqOperationList []*SeqOperation) *protocal.ImPacket {
	var operationList flatbuffers.UOffsetT
	// 已同意解用户列表
	var dismissUsers flatbuffers.UOffsetT
	// 解散房间剩余同意时间
	var dismissRemainTime uint32
	// 托管用户列表
	var hostingUsers flatbuffers.UOffsetT
	// 在线用户列表
	var onlineUsers flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)

	// 组织数据
	// 同意解散用户列表、解散剩余时间
	if room.DismissTime > 0 {
		dismissRemainTime = uint32(room.DismissTime + int64(config.ROOM_DISMISS_AOTO_ALLOW_INTERVAL) - util.GetTime())
		dismissUserSlice := room.GetDismissUsers()
		length := len(dismissUserSlice)
		fbsCommon.GameRestoreSectionPushStartDismissUsersVector(builder, length)
		for i := length - 1; i >= 0; i-- {
			builder.PrependUint32(uint32(dismissUserSlice[i]))
		}
		dismissUsers = builder.EndVector(length)
	}
	// 已托管用户列表
	hostingUserCnt := len(room.HostingUsers)
	if hostingUserCnt > 0 {
		fbsCommon.GameRestoreSectionPushStartHostingUsersVector(builder, hostingUserCnt)
		for _, v := range room.HostingUsers {
			builder.PrependUint32(uint32(v))
		}
		hostingUsers = builder.EndVector(hostingUserCnt)
	}
	// 操作队列
	if length := len(seqOperationList); length > 0 {
		var operationBinarys []flatbuffers.UOffsetT
		builder, operationBinarys = genSeqOperationList(builder, seqOperationList)

		fbsCommon.GameRestoreSectionPushStartOperationListVector(builder, length)
		for i := len(seqOperationList) - 1; i >= 0; i-- {
			builder.PrependUOffsetT(operationBinarys[i])
		}
		operationList = builder.EndVector(length)
	}
	// 找出在线用户列表
	onlineUserList := room.GetOnlineUsers()
	if onlineCnt := len(onlineUserList); onlineCnt > 0 {
		fbsCommon.GameRestoreSectionPushStartOnlineUsersVector(builder, onlineCnt)
		for _, v := range onlineUserList {
			builder.PrependUint32(uint32(v))
		}
		onlineUsers = builder.EndVector(onlineCnt)
	}

	seq := room.MI.getUser(userId).MSC.GetSeq()

	fbsCommon.GameRestoreSectionPushStart(builder)
	fbsCommon.GameRestoreSectionPushAddDismissRemainTime(builder, dismissRemainTime)
	fbsCommon.GameRestoreSectionPushAddDismissUsers(builder, dismissUsers)
	fbsCommon.GameRestoreSectionPushAddHostingUsers(builder, hostingUsers)
	fbsCommon.GameRestoreSectionPushAddOnlineUsers(builder, onlineUsers)
	fbsCommon.GameRestoreSectionPushAddOperationList(builder, operationList)
	fbsCommon.GameRestoreSectionPushAddStep(builder, uint16(seq))
	orc := fbsCommon.GameRestoreSectionPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameRestoreSectionPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

func genSeqOperationList(builder *flatbuffers.Builder, seqOperationList []*SeqOperation) (*flatbuffers.Builder, []flatbuffers.UOffsetT) {
	var operationBinary flatbuffers.UOffsetT

	length := len(seqOperationList)
	operationBinarys := make([]flatbuffers.UOffsetT, 0, length)
	for _, seqOperation := range seqOperationList {
		builder, operationBinary = genSeqOperation(builder, seqOperation)
		operationBinarys = append(operationBinarys, operationBinary)
	}
	return builder, operationBinarys
}

// 构建单个有序消息
func genSeqOperation(builder *flatbuffers.Builder, seqOperation *SeqOperation) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var operationPush, userOperationPush, clientOperationPush flatbuffers.UOffsetT

	if seqOperation.wOperation != nil {
		builder, operationPush = genOperationPush(builder, seqOperation.wOperation)
	}
	if seqOperation.uOperation != nil {
		builder, userOperationPush = genUserOperation(builder, seqOperation.uOperation)
	}
	if seqOperation.cOperation != nil {
		builder, clientOperationPush = genClientOperation(builder, seqOperation.cOperation)
	}

	fbsCommon.SeqOperationStart(builder)
	fbsCommon.SeqOperationAddStep(builder, uint16(seqOperation.seq))
	fbsCommon.SeqOperationAddOperationPush(builder, operationPush)
	fbsCommon.SeqOperationAddUserOperationPush(builder, userOperationPush)
	fbsCommon.SeqOperationAddClientOperationPush(builder, clientOperationPush)
	return builder, fbsCommon.SeqOperationEnd(builder)
}

// func genGameSettlementPush(builder *flatbuffers.Builder, hu bool, scoreList map[int]*FrontScoreInfo) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
func genGameSettlementPush(builder *flatbuffers.Builder, room *Room) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	hu := room.MI.getHInfo().Hu
	scoreList := room.MI.getFrontData()
	isGoldBam1 := room.MI.isGoldBam1()

	var scoreListBinary flatbuffers.UOffsetT
	var settlementBinary flatbuffers.UOffsetT
	var chikenBinary flatbuffers.UOffsetT
	var chikenListBinary flatbuffers.UOffsetT
	var scoreListBinaryV230 flatbuffers.UOffsetT

	// 构建结算信息的二进制流
	settlements := make([]flatbuffers.UOffsetT, 0, len(scoreList))
	for _, settlementInfo := range scoreList {
		builder, settlementBinary = genSettlementInfo(builder, settlementInfo)
		settlements = append(settlements, settlementBinary)
	}
	fbsCommon.GameSettlementPushStartSettlementInfoVector(builder, len(scoreList))
	for _, settlementBinary := range settlements {
		builder.PrependUOffsetT(settlementBinary)
	}
	scoreListBinary = builder.EndVector(len(scoreList))

	// 构建结算信息的二进制流
	settlementsV230 := make([]flatbuffers.UOffsetT, 0, len(scoreList))
	for _, settlementInfo := range scoreList {
		builder, settlementBinary = genSettlementInfoV230(builder, settlementInfo)
		settlementsV230 = append(settlementsV230, settlementBinary)
	}
	fbsCommon.GameSettlementPushStartSettlementInfoV230Vector(builder, len(scoreList))
	for _, settlementBinary := range settlementsV230 {
		builder.PrependUOffsetT(settlementBinary)
	}
	scoreListBinaryV230 = builder.EndVector(len(scoreList))

	// 构建鸡牌信息的二进制流
	chikens := []flatbuffers.UOffsetT{}
	for _, settlementInfo := range scoreList {
		if len(settlementInfo.PlayChikens) == 0 && len(settlementInfo.HandChikens) == 0 && len(settlementInfo.ShowCardChikens) == 0 {
			// 没有需要结算的鸡
			continue
		}
		builder, chikenBinary = genSettlementChikenInfo(builder, settlementInfo)
		chikens = append(chikens, chikenBinary)
	}
	if len(chikens) > 0 {
		fbsCommon.GameSettlementPushStartChikenInfoVector(builder, len(chikens))
		for _, chikenBinary := range chikens {
			builder.PrependUOffsetT(chikenBinary)
		}
		chikenListBinary = builder.EndVector(len(chikens))
	}

	fbsCommon.GameSettlementPushStart(builder)
	if !hu {
		fbsCommon.GameSettlementPushAddIsHuangPai(builder, int8(1))
	} else {
		fbsCommon.GameSettlementPushAddIsHuangPai(builder, int8(0))
	}
	fbsCommon.GameSettlementPushAddSettlementInfo(builder, scoreListBinary)
	if isGoldBam1 {
		fbsCommon.GameSettlementPushAddIsGoldBam1(builder, int8(1))
	} else {
		fbsCommon.GameSettlementPushAddIsGoldBam1(builder, int8(0))
	}
	fbsCommon.GameSettlementPushAddChikenInfo(builder, chikenListBinary)
	fbsCommon.GameSettlementPushAddSettlementInfoV230(builder, scoreListBinaryV230)
	orc := fbsCommon.GameSettlementPushEnd(builder)

	return builder, orc
}

// 单局结算推送
// func GameSettlementPush(hu bool, scoreList map[int]*FrontScoreInfo) *protocal.ImPacket {
func GameSettlementPush(room *Room) *protocal.ImPacket {
	var orc flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, orc = genGameSettlementPush(builder, room)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameSettlementPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构造结算信息
func genResultInfo(builder *flatbuffers.Builder, finalInfo *FrontFinalInfo) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	fbsCommon.ResultInfoStart(builder)
	fbsCommon.ResultInfoAddUserId(builder, uint32(finalInfo.UserId))
	// fbsCommon.ResultInfoAddTotalScore(builder, int32(finalInfo.Total))
	// 隐藏总积分
	fbsCommon.ResultInfoAddTotalScore(builder, int32(0))
	fbsCommon.ResultInfoAddHuTimes(builder, uint8(finalInfo.HuCount))
	fbsCommon.ResultInfoAddKitchenTimes(builder, uint8(finalInfo.KitchenCount))
	fbsCommon.ResultInfoAddKongTimes(builder, uint8(finalInfo.KongCount))
	fbsCommon.ResultInfoAddDianPaoTimes(builder, uint8(finalInfo.DianPaoCount))
	fbsCommon.ResultInfoAddScore(builder, int32(finalInfo.Score))

	return builder, fbsCommon.ResultInfoEnd(builder)
}

// 构造结算信息 2.3.0以上版本
func genResultInfoV230(builder *flatbuffers.Builder, finalInfo *FrontFinalInfo, showTotal bool) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	nickname := builder.CreateString(finalInfo.Nickname)
	avatar := builder.CreateString(finalInfo.Avatar)
	fbsCommon.ResultInfo_v_2_3_0Start(builder)
	fbsCommon.ResultInfo_v_2_3_0AddUserId(builder, uint32(finalInfo.UserId))
	fbsCommon.ResultInfo_v_2_3_0AddNickname(builder, nickname)
	fbsCommon.ResultInfo_v_2_3_0AddAvatar(builder, avatar)
	// 隐藏总积分
	if showTotal {
		if finalInfo.FromSLevel > 0 {
			fbsCommon.ResultInfo_v_2_3_0AddFromTotolScore(builder, int32(finalInfo.FromSLevel))
			fbsCommon.ResultInfo_v_2_3_0AddTotalScore(builder, int32(finalInfo.FinalSLevel))
		} else {
			fbsCommon.ResultInfo_v_2_3_0AddFromTotolScore(builder, int32(finalInfo.FromTotal))
			fbsCommon.ResultInfo_v_2_3_0AddTotalScore(builder, int32(finalInfo.Total))
		}
	} else {
		fbsCommon.ResultInfo_v_2_3_0AddFromTotolScore(builder, int32(0))
		fbsCommon.ResultInfo_v_2_3_0AddTotalScore(builder, int32(0))
	}
	fbsCommon.ResultInfo_v_2_3_0AddWinSelfTimes(builder, uint16(finalInfo.WinSelfTimes))
	fbsCommon.ResultInfo_v_2_3_0AddWinTimes(builder, uint16(finalInfo.WinTimes))
	fbsCommon.ResultInfo_v_2_3_0AddKitchenTimes(builder, uint16(finalInfo.KitchenCount))
	fbsCommon.ResultInfo_v_2_3_0AddKongDarkTimes(builder, uint16(finalInfo.KongDarkTimes))
	fbsCommon.ResultInfo_v_2_3_0AddKongTimes(builder, uint16(finalInfo.KongTimes))
	fbsCommon.ResultInfo_v_2_3_0AddKongTurnTimes(builder, uint16(finalInfo.KongTurnTimes))
	fbsCommon.ResultInfo_v_2_3_0AddKongTurnFreeTimes(builder, uint16(finalInfo.KongTurnFreeTimes))
	fbsCommon.ResultInfo_v_2_3_0AddDianPaoTimes(builder, uint16(finalInfo.DianPaoCount))
	fbsCommon.ResultInfo_v_2_3_0AddScore(builder, int32(finalInfo.Score))
	fbsCommon.ResultInfo_v_2_3_0AddStarChange(builder, int32(finalInfo.StarChange))
	fbsCommon.ResultInfo_v_2_3_0AddFromExp(builder, int32(finalInfo.FromTotal))
	fbsCommon.ResultInfo_v_2_3_0AddExp(builder, int32(finalInfo.Total))
	fbsCommon.ResultInfo_v_2_3_0AddExpChange(builder, int32(finalInfo.ExpChange))
	fbsCommon.ResultInfo_v_2_3_0AddWinningStreak(builder, int32(finalInfo.WinningStreak))
	fbsCommon.ResultInfo_v_2_3_0AddSecKill(builder, int32(finalInfo.SecKill))
	fbsCommon.ResultInfo_v_2_3_0AddAvatarBox(builder, int32(finalInfo.AvatarBox))

	return builder, fbsCommon.ResultInfoEnd(builder)
}

// GameResultPush 游戏结果推送
func GameResultPush(infoList map[int]*FrontFinalInfo, host int, isDismiss int64, roomNumber string, cType int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)

	number := builder.CreateString(roomNumber)
	var resultInfoBinary flatbuffers.UOffsetT
	var resultListBinary flatbuffers.UOffsetT
	results := make([]flatbuffers.UOffsetT, 0, len(infoList))

	for _, finalInfo := range infoList {
		builder, resultInfoBinary = genResultInfo(builder, finalInfo)
		results = append(results, resultInfoBinary)
	}
	fbsCommon.GameResultPushStartResultInfoVector(builder, len(infoList))
	for _, resultInfoBinary := range results {
		builder.PrependUOffsetT(resultInfoBinary)
	}
	resultListBinary = builder.EndVector(len(infoList))

	var resultInfoV230Binary flatbuffers.UOffsetT
	var resultListV230Binary flatbuffers.UOffsetT
	resultsV230 := make([]flatbuffers.UOffsetT, 0, len(infoList))

	// 是否显示总积分
	showTotal := configService.IsRankRoom(cType)
	for _, finalInfo := range infoList {
		builder, resultInfoV230Binary = genResultInfoV230(builder, finalInfo, showTotal)
		resultsV230 = append(resultsV230, resultInfoV230Binary)
	}
	fbsCommon.GameResultPushStartResultInfoV230Vector(builder, len(infoList))
	for _, resultInfoV230Binary := range resultsV230 {
		builder.PrependUOffsetT(resultInfoV230Binary)
	}
	resultListV230Binary = builder.EndVector(len(infoList))

	fbsCommon.GameResultPushStart(builder)
	fbsCommon.GameResultPushAddResultInfo(builder, resultListBinary)
	fbsCommon.GameResultPushAddHost(builder, uint32(host))
	if isDismiss > 0 {
		fbsCommon.GameResultPushAddIsDismiss(builder, uint8(1))
	} else {
		fbsCommon.GameResultPushAddIsDismiss(builder, uint8(0))
	}
	fbsCommon.GameResultPushAddResultInfoV230(builder, resultListV230Binary)
	fbsCommon.GameResultPushAddNumber(builder, number)
	fbsCommon.GameResultPushAddRandomRoom(builder, byte(cType))
	orc := fbsCommon.GameResultPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameResultPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 系统获取房间信息
func SystemResponse(data map[string]interface{}) *protocal.ImPacket {
	js := simplejson.New()
	js.Set("code", 0)
	js.Set("message", "")
	js.Set("data", data)

	return response.GenJson(protocal.PACKAGE_TYPE_SYSTEM, js)
}

// GameHostingPush 用户托管状态切换
func GameHostingPush(userId, hostingStatus int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.GameHostingPushStart(builder)
	fbsCommon.GameHostingPushAddUserId(builder, uint32(userId))
	fbsCommon.GameHostingPushAddHostingStatus(builder, uint8(hostingStatus))
	orc := fbsCommon.GameHostingPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameHostingPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 生成回放操作
func genPlaybackOperation(builder *flatbuffers.Builder, operation *playbackOperation) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 用户操作
	var userOperationPush flatbuffers.UOffsetT
	// 系统操作
	var clientOperationPush flatbuffers.UOffsetT

	if operation.userOperation != nil {
		builder, userOperationPush = genUserOperation(builder, operation.userOperation)
	}
	if operation.clientOpreation != nil {
		builder, clientOperationPush = genClientOperation(builder, operation.clientOpreation)
	}

	fbsCommon.GamePlaybackOperationStart(builder)
	fbsCommon.GamePlaybackOperationAddT(builder, operation.t)
	fbsCommon.GamePlaybackOperationAddUserOperationPush(builder, userOperationPush)
	fbsCommon.GamePlaybackOperationAddClientOperationPush(builder, clientOperationPush)
	return builder, fbsCommon.GamePlaybackOperationEnd(builder)
}

// 生成回放操作-完整版
func genPlaybackOperationIntact(builder *flatbuffers.Builder, operation *playbackOperation) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 用户操作
	var userOperationPush flatbuffers.UOffsetT
	// 系统操作
	var clientOperationPush flatbuffers.UOffsetT
	var operationPush flatbuffers.UOffsetT

	if operation.userOperation != nil {
		builder, userOperationPush = genUserOperation(builder, operation.userOperation)
	}
	if operation.clientOpreation != nil {
		builder, clientOperationPush = genClientOperation(builder, operation.clientOpreation)
	}
	if operation.userId > 0 {
		builder, operationPush = genOperationPush(builder, operation.opList)
	}

	fbsCommon.GamePlaybackIntactOperationStart(builder)
	fbsCommon.GamePlaybackIntactOperationAddT(builder, operation.t)
	fbsCommon.GamePlaybackIntactOperationAddUserOperationPush(builder, userOperationPush)
	fbsCommon.GamePlaybackIntactOperationAddClientOperationPush(builder, clientOperationPush)
	fbsCommon.GamePlaybackIntactOperationAddUserId(builder, uint32(operation.userId))
	fbsCommon.GamePlaybackIntactOperationAddOperationPush(builder, operationPush)
	return builder, fbsCommon.GamePlaybackOperationEnd(builder)
}

// 生成回放数据
func genPlayback(room *Room, gameScores map[int]int) []byte {
	core.Logger.Debug("[genPlayback]roomId:%v,round:%v", room.RoomId, room.Round)

	// 房间信息
	var roomInfo flatbuffers.UOffsetT
	// 房间用户信息
	var roomUserList flatbuffers.UOffsetT
	// 初始化信息
	var initPush flatbuffers.UOffsetT
	var initList flatbuffers.UOffsetT
	// 操作列表
	var operationList flatbuffers.UOffsetT
	var operationBinary flatbuffers.UOffsetT
	// 结算信息
	var settlementPush flatbuffers.UOffsetT
	// 牌局中的用户信息
	var mahjongUserList flatbuffers.UOffsetT
	var mahjongUserBinary flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, roomInfo = genRoomInfo(builder, room)
	builder, roomUserList = genRoomUsers(builder, room.GetUsers())
	builder, settlementPush = genGameSettlementPush(builder, room)

	// 构建initList
	userCnt := room.GetIndexLen()
	initBinarys := make([]flatbuffers.UOffsetT, 0, userCnt)
	for i := 0; i < userCnt; i++ {
		mu := room.MI.getUsers()[room.GetIndexUserId(i)]
		builder, initPush = genGameInitPush(builder, room.Dealer, room.DealCount, room.MI.getDiceList(), mu.HandTileList.GetInitTiles(), room.Round)
		initBinarys = append(initBinarys, initPush)
	}
	fbsCommon.GamePlayBackStartGameInitListVector(builder, userCnt)
	for i := userCnt - 1; i >= 0; i-- {
		builder.PrependUOffsetT(initBinarys[i])
	}
	initList = builder.EndVector(userCnt)

	// 构建操作列表
	operations := room.MI.getPlaybackOperationList()
	operationBinarys := make([]flatbuffers.UOffsetT, 0)
	for _, operation := range operations {
		// 跳过决策
		if operation.userId > 0 {
			continue
		}
		builder, operationBinary = genPlaybackOperation(builder, operation)
		operationBinarys = append(operationBinarys, operationBinary)
	}
	operationLen := len(operationBinarys)
	fbsCommon.GamePlayBackStartOperationListVector(builder, operationLen)
	for i := operationLen - 1; i >= 0; i-- {
		builder.PrependUOffsetT(operationBinarys[i])
	}
	operationList = builder.EndVector(operationLen)

	// 构建局内用户信息
	mahjongUserBinarys := make([]flatbuffers.UOffsetT, 0, userCnt)
	for userId, gameScore := range gameScores {
		builder, mahjongUserBinary = genMahjongUserInfoOnlyGameScore(builder, userId, gameScore)
		mahjongUserBinarys = append(mahjongUserBinarys, mahjongUserBinary)
	}
	fbsCommon.GamePlayBackStartMahjongUserInfoVector(builder, userCnt)
	for i := userCnt - 1; i >= 0; i-- {
		builder.PrependUOffsetT(mahjongUserBinarys[i])
	}
	mahjongUserList = builder.EndVector(userCnt)

	fbsCommon.GamePlayBackStart(builder)
	fbsCommon.GamePlayBackAddRoomInfo(builder, roomInfo)
	fbsCommon.GamePlayBackAddRoomUserList(builder, roomUserList)
	fbsCommon.GamePlayBackAddGameInitList(builder, initList)
	fbsCommon.GamePlayBackAddOperationList(builder, operationList)
	fbsCommon.GamePlayBackAddGameSettlementPush(builder, settlementPush)
	fbsCommon.GamePlayBackAddMahjongUserInfo(builder, mahjongUserList)
	orc := fbsCommon.GamePlayBackEnd(builder)
	builder.Finish(orc)
	return builder.FinishedBytes()
}

// 生成回放数据 - 完整版
func genPlaybackIntact(room *Room, gameScores map[int]int) []byte {
	// 房间信息
	var roomInfo flatbuffers.UOffsetT
	// 房间用户信息
	var roomUserList flatbuffers.UOffsetT
	// 初始化信息
	var initPush flatbuffers.UOffsetT
	var initList flatbuffers.UOffsetT
	// 操作列表
	var operationList flatbuffers.UOffsetT
	var operationBinary flatbuffers.UOffsetT
	// 结算信息
	var settlementPush flatbuffers.UOffsetT
	// 牌局中的用户信息
	var mahjongUserList flatbuffers.UOffsetT
	var mahjongUserBinary flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, roomInfo = genRoomInfo(builder, room)
	builder, roomUserList = genRoomUsers(builder, room.GetUsers())
	builder, settlementPush = genGameSettlementPush(builder, room)

	// 构建initList
	userCnt := room.GetIndexLen()
	initBinarys := make([]flatbuffers.UOffsetT, 0, userCnt)
	for i := 0; i < userCnt; i++ {
		mu := room.MI.getUsers()[room.GetIndexUserId(i)]
		builder, initPush = genGameInitPush(builder, room.Dealer, room.DealCount, room.MI.getDiceList(), mu.HandTileList.GetInitTiles(), room.Round)
		initBinarys = append(initBinarys, initPush)
	}
	fbsCommon.GamePlayBackIntactStartGameInitListVector(builder, userCnt)
	for i := userCnt - 1; i >= 0; i-- {
		builder.PrependUOffsetT(initBinarys[i])
	}
	initList = builder.EndVector(userCnt)

	// 构建操作列表
	operations := room.MI.getPlaybackOperationList()
	operationLen := len(operations)
	operationBinarys := make([]flatbuffers.UOffsetT, 0, operationLen)
	for _, operation := range operations {
		builder, operationBinary = genPlaybackOperationIntact(builder, operation)
		operationBinarys = append(operationBinarys, operationBinary)
	}
	fbsCommon.GamePlayBackIntactStartOperationListVector(builder, operationLen)
	for i := operationLen - 1; i >= 0; i-- {
		builder.PrependUOffsetT(operationBinarys[i])
	}
	operationList = builder.EndVector(operationLen)

	// 构建局内用户信息
	mahjongUserBinarys := make([]flatbuffers.UOffsetT, 0, userCnt)
	for userId, gameScore := range gameScores {
		builder, mahjongUserBinary = genMahjongUserInfoOnlyGameScore(builder, userId, gameScore)
		mahjongUserBinarys = append(mahjongUserBinarys, mahjongUserBinary)
	}
	fbsCommon.GamePlayBackIntactStartMahjongUserInfoVector(builder, userCnt)
	for i := userCnt - 1; i >= 0; i-- {
		builder.PrependUOffsetT(mahjongUserBinarys[i])
	}
	mahjongUserList = builder.EndVector(userCnt)

	fbsCommon.GamePlayBackIntactStart(builder)
	fbsCommon.GamePlayBackIntactAddRoomInfo(builder, roomInfo)
	fbsCommon.GamePlayBackIntactAddRoomUserList(builder, roomUserList)
	fbsCommon.GamePlayBackIntactAddGameInitList(builder, initList)
	fbsCommon.GamePlayBackIntactAddOperationList(builder, operationList)
	fbsCommon.GamePlayBackIntactAddGameSettlementPush(builder, settlementPush)
	fbsCommon.GamePlayBackIntactAddMahjongUserInfo(builder, mahjongUserList)
	orc := fbsCommon.GamePlayBackIntactEnd(builder)
	builder.Finish(orc)
	return builder.FinishedBytes()
}

// GameAntiCheatingPush
func GameAntiCheatingPush(nearUsers []int, noPositionUsers []int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)

	// 构建距离相近的用户数据
	var nearUsersBinary flatbuffers.UOffsetT
	if len(nearUsers) > 0 {
		fbsCommon.GameEnterPushStartNearUsersVector(builder, len(nearUsers))
		for _, userId := range nearUsers {
			builder.PrependUint32(uint32(userId))
		}
		nearUsersBinary = builder.EndVector(len(nearUsers))
	}

	// 构建未分享位置的用户数据
	var noPositionUsersBinary flatbuffers.UOffsetT
	if len(noPositionUsers) > 0 {
		fbsCommon.GameEnterPushStartNoPositionUsersVector(builder, len(noPositionUsers))
		for _, userId := range noPositionUsers {
			builder.PrependUint32(uint32(userId))
		}
		noPositionUsersBinary = builder.EndVector(len(noPositionUsers))
	}

	fbsCommon.GameAntiCheatingPushStart(builder)
	fbsCommon.GameAntiCheatingPushAddNearUsers(builder, nearUsersBinary)
	fbsCommon.GameAntiCheatingPushAddNoPositionUsers(builder, noPositionUsersBinary)
	orc := fbsCommon.GameAntiCheatingPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameAntiCheatingPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// ClubG2CStartRoomPush 游戏服推送房间开始消息到俱乐部服务
func ClubG2CStartRoomPush(clubId int, roomId int64, round int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.ClubG2CStartRoomPushStart(builder)
	fbsCommon.ClubG2CStartRoomPushAddClubId(builder, int32(clubId))
	fbsCommon.ClubG2CStartRoomPushAddRoomId(builder, uint64(roomId))
	fbsCommon.ClubG2CStartRoomPushAddRound(builder, byte(round))
	orc := fbsCommon.ClubG2CStartRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubG2CStartRoomPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// ClubG2CRoomActiveRequest 游戏服向俱乐部服务询问某房间在俱乐部服务上是否还活跃着
// 主要用于俱乐部服务热更后的房间数据恢复
// 这里虽然是request，但是逻辑上并不依赖于messageNumber，所以就不生成了，直接给0
func ClubG2CRoomActiveRequest(clubId int, roomId int64) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.ClubG2CRoomActiveRequestStart(builder)
	fbsCommon.ClubG2CRoomActiveRequestAddClubId(builder, int32(clubId))
	fbsCommon.ClubG2CRoomActiveRequestAddRoomId(builder, uint64(roomId))
	orc := fbsCommon.ClubG2CRoomActiveRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubG2CRoomActiveRequest, protocal.MSG_TYPE_REQUEST, uint16(0), uint16(0), buf)
}

// ClubG2CDismissRoomPush 游戏服推送房间解散消息到俱乐部服务
func ClubG2CDismissRoomPush(clubId int, roomId int64, code int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.ClubG2CDismissRoomPushStart(builder)
	fbsCommon.ClubG2CDismissRoomPushAddClubId(builder, int32(clubId))
	fbsCommon.ClubG2CDismissRoomPushAddRoomId(builder, uint64(roomId))
	fbsCommon.ClubG2CDismissRoomPushAddCode(builder, int32(code))
	orc := fbsCommon.ClubG2CDismissRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubG2CDismissRoomPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// ClubG2CJoinRoomPush 游戏服推送加入房间消息给俱乐部服务
func ClubG2CJoinRoomPush(clubId int, roomId int64, roomUser *RoomUser) *protocal.ImPacket {
	var roomUserInfoBinary flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, roomUserInfoBinary = genRoomUserInfo(builder, roomUser)
	fbsCommon.ClubG2CJoinRoomPushStart(builder)
	fbsCommon.ClubG2CJoinRoomPushAddClubId(builder, int32(clubId))
	fbsCommon.ClubG2CJoinRoomPushAddRoomId(builder, uint64(roomId))
	fbsCommon.ClubG2CJoinRoomPushAddRoomUserInfo(builder, roomUserInfoBinary)
	orc := fbsCommon.ClubG2CJoinRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubG2CJoinRoomPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// ClubG2CQuitRoomPush 游戏服推送退出房间消息
func ClubG2CQuitRoomPush(clubId int, roomId int64, userId int, index int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.ClubG2CQuitRoomPushStart(builder)
	fbsCommon.ClubG2CQuitRoomPushAddClubId(builder, int32(clubId))
	fbsCommon.ClubG2CQuitRoomPushAddRoomId(builder, uint64(roomId))
	fbsCommon.ClubG2CQuitRoomPushAddUserId(builder, uint32(userId))
	fbsCommon.ClubG2CQuitRoomPushAddIndex(builder, uint8(index))
	orc := fbsCommon.ClubG2CQuitRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubG2CQuitRoomPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// ClubG2CReloadRoomPush 游戏服推送整个房间数据给俱乐部服务
func ClubG2CReloadRoomPush(clubId int, room *Room) *protocal.ImPacket {
	var roomBinary flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, roomBinary = genRoom(builder, room)
	fbsCommon.ClubG2CReloadRoomPushStart(builder)
	fbsCommon.ClubG2CReloadRoomPushAddClubId(builder, int32(clubId))
	fbsCommon.ClubG2CReloadRoomPushAddRoom(builder, roomBinary)
	orc := fbsCommon.ClubG2CReloadRoomPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandClubG2CReloadRoomPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构建distance结构
func genDistance(builder *flatbuffers.Builder, userDistance *UserDistance) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	fbsCommon.GameUserDistanceStart(builder)
	fbsCommon.GameUserDistanceAddMaxUserId(builder, uint32(userDistance.MaxUserId))
	fbsCommon.GameUserDistanceAddMinUserId(builder, uint32(userDistance.MinUserId))
	fbsCommon.GameUserDistanceAddDistance(builder, int32(userDistance.Distance))
	orc := fbsCommon.GameUserDistanceEnd(builder)
	return builder, orc
}

// GameUserDistanceResponse 推送用户之间的距离
func GameUserDistanceResponse(distances []*UserDistance, mNumber uint16) *protocal.ImPacket {
	var distanceBinary flatbuffers.UOffsetT
	var distanceListBinary flatbuffers.UOffsetT
	var commonResult flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	// 生成commonResult
	builder, commonResult = genGameResult(builder, nil)
	// 生成距离数据
	if length := len(distances); length > 0 {
		distanceBinarys := make([]flatbuffers.UOffsetT, 0, length)
		for _, userDistance := range distances {
			builder, distanceBinary = genDistance(builder, userDistance)
			distanceBinarys = append(distanceBinarys, distanceBinary)
		}
		fbsCommon.GameUserDistanceResponseStartDistanceListVector(builder, length)
		for _, distanceBinary := range distanceBinarys {
			builder.PrependUOffsetT(distanceBinary)
		}
		distanceListBinary = builder.EndVector(length)
	}

	// 组织数据
	fbsCommon.GameUserDistanceResponseStart(builder)
	fbsCommon.GameUserDistanceResponseAddS2cResult(builder, commonResult)
	fbsCommon.GameUserDistanceResponseAddDistanceList(builder, distanceListBinary)
	orc := fbsCommon.GameUserDistanceResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameUserDistanceResponse, protocal.MSG_TYPE_RESPONSE, uint16(0), mNumber, buf)
}

// GameUserDistanceFailResponse 获取用户距离失败
func GameUserDistanceFailResponse(err *core.Error, mNumber uint16) *protocal.ImPacket {
	var commonResult flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	// 生成commonResult
	builder, commonResult = genGameResult(builder, err)

	// 构建对象
	fbsCommon.GameUserDistanceResponseStart(builder)
	fbsCommon.GameUserDistanceResponseAddS2cResult(builder, commonResult)
	orc := fbsCommon.GameUserDistanceResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameUserDistanceResponse, protocal.MSG_TYPE_RESPONSE, uint16(0), mNumber, buf)
}

// GameSkipOperateNoticePush 提示用户跳过了一些操作
func GameSkipOperateNoticePush(code int, tile int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.GameSkipOperateNoticePushStart(builder)
	fbsCommon.GameSkipOperateNoticePushAddOp(builder, uint8(code))
	fbsCommon.GameSkipOperateNoticePushAddTile(builder, uint8(tile))
	orc := fbsCommon.GameSkipOperateNoticePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGameSkipOperateNoticePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueS2LGameActivePush 推送房间活跃给联赛服务器
func LeagueS2LGameActivePush(raceId, raceRoomId, roomId int64) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueS2LGameActivePushStart(builder)
	fbsCommon.LeagueS2LGameActivePushAddRaceId(builder, raceId)
	fbsCommon.LeagueS2LGameActivePushAddRaceRoomId(builder, raceRoomId)
	fbsCommon.LeagueS2LGameActivePushAddRoomId(builder, roomId)
	orc := fbsCommon.LeagueS2LGameActivePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueS2LGameActivePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}
