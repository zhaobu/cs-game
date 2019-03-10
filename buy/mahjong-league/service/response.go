package service

import (
	"encoding/json"
	"mahjong-league/config"
	fbsCommon "mahjong-league/fbs/Common"
	"mahjong-league/ierror"
	"mahjong-league/model"
	"mahjong-league/protocal"
	"mahjong-league/response"

	flatbuffers "github.com/google/flatbuffers/go"
)

// LeagueListPush 列表
func LeagueListPush(leagueList map[int]*model.League, raceList map[int]*model.Race) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	var leagueListBinary flatbuffers.UOffsetT

	length := len(leagueList)
	builder, lists := buildLeagueList(builder, leagueList, raceList)

	// 构建leagueList
	fbsCommon.LeagueListPushStartLeagueListVector(builder, length)
	for i := length - 1; i >= 0; i-- {
		builder.PrependUOffsetT(lists[i])
	}
	leagueListBinary = builder.EndVector(length)

	fbsCommon.LeagueListPushStart(builder)
	fbsCommon.LeagueListPushAddLeagueList(builder, leagueListBinary)
	orc := fbsCommon.LeagueListPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueListPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueListResponse 拉取大厅列表的回应
func LeagueListResponse(mNumber uint16, leagueList map[int]*model.League, raceList map[int]*model.Race) *protocal.ImPacket {
	var leagueListBinary flatbuffers.UOffsetT
	var s2cResult flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, s2cResult = response.BuidGameResult(builder, nil)

	length := len(leagueList)
	builder, lists := buildLeagueList(builder, leagueList, raceList)

	// 构建leagueList
	fbsCommon.LeagueListResponseStartLeagueListVector(builder, length)
	for i := length - 1; i >= 0; i-- {
		builder.PrependUOffsetT(lists[i])
	}
	leagueListBinary = builder.EndVector(length)

	fbsCommon.LeagueListResponseStart(builder)
	fbsCommon.LeagueListResponseAddS2cResult(builder, s2cResult)
	fbsCommon.LeagueListResponseAddLeagueList(builder, leagueListBinary)
	orc := fbsCommon.LeagueListResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueListResponse, protocal.MSG_TYPE_RESPONSE, uint16(0), mNumber, buf)
}

func buildLeagueList(builder *flatbuffers.Builder, leagueList map[int]*model.League, raceList map[int]*model.Race) (*flatbuffers.Builder, []flatbuffers.UOffsetT) {
	var leagueBinary flatbuffers.UOffsetT
	lists := make([]flatbuffers.UOffsetT, 0, len(leagueList))
	for _, leagueInfo := range leagueList {
		builder, leagueBinary = buildLeague(builder, leagueInfo, raceList[leagueInfo.Id])
		lists = append(lists, leagueBinary)
	}
	return builder, lists
}

// 构建大厅联赛的
func buildLeague(builder *flatbuffers.Builder, leagueInfo *model.League, raceInfo *model.Race) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	var name, img, settingBinary, rewardsBinary flatbuffers.UOffsetT
	var setting []byte
	var rewards []string
	name = builder.CreateString(leagueInfo.Name)
	img = builder.CreateString(leagueInfo.Img)
	json.Unmarshal([]byte(leagueInfo.Setting), &setting)

	rewards = model.GetFrontLeagueRewards(leagueInfo.Id)

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

	length := len(setting)
	fbsCommon.LeagueStartSettingVector(builder, length)
	for i := length - 1; i >= 0; i-- {
		builder.PrependByte(byte(setting[i]))
	}
	settingBinary = builder.EndVector(length)

	fbsCommon.LeagueStart(builder)
	fbsCommon.LeagueAddId(builder, int32(leagueInfo.Id))
	fbsCommon.LeagueAddName(builder, name)
	fbsCommon.LeagueAddImg(builder, img)
	fbsCommon.LeagueAddSetting(builder, settingBinary)
	fbsCommon.LeagueAddIcon(builder, int32(leagueInfo.Icon))
	fbsCommon.LeagueAddLeagueType(builder, int32(leagueInfo.LeagueType))
	fbsCommon.LeagueAddGameType(builder, uint16(leagueInfo.GameType))
	fbsCommon.LeagueAddPriceEntityId(builder, int32(leagueInfo.PriceEntityId))
	fbsCommon.LeagueAddPrice(builder, int32(leagueInfo.Price))
	fbsCommon.LeagueAddRewards(builder, rewardsBinary)
	fbsCommon.LeagueAddWeight(builder, int32(leagueInfo.Weight))
	if raceInfo != nil {
		fbsCommon.LeagueAddRequireUserCount(builder, int32(raceInfo.RequireUserCount))
		fbsCommon.LeagueAddRequireUserMin(builder, int32(raceInfo.RequireUserMin))
		fbsCommon.LeagueAddSignupTime(builder, raceInfo.SignTime)
		fbsCommon.LeagueAddGiveupTime(builder, raceInfo.GiveupTime)
		fbsCommon.LeagueAddStartTime(builder, raceInfo.StartTime)

		// 如果是报名中的支持氛围的比赛，则显示氛围人数
		if raceInfo.Status == config.RACE_STATUS_SIGNUP && leagueInfo.EnableSimulationUserCount() {
			fbsCommon.LeagueAddSignupUserCount(builder, int32(leagueInfo.SimulationUserCount))
		} else {
			fbsCommon.LeagueAddSignupUserCount(builder, int32(raceInfo.SignupUserCount))
		}
	} else {
		fbsCommon.LeagueAddRequireUserCount(builder, int32(leagueInfo.RequireUserCount))
		fbsCommon.LeagueAddRequireUserMin(builder, int32(leagueInfo.RequireUserMin))
		signTime, giveupTime, startTime := leagueInfo.CalcLeagueRaceTime()
		fbsCommon.LeagueAddSignupTime(builder, signTime)
		fbsCommon.LeagueAddGiveupTime(builder, giveupTime)
		fbsCommon.LeagueAddStartTime(builder, startTime)

		if leagueInfo.EnableSimulationUserCount() {
			fbsCommon.LeagueAddSignupUserCount(builder, int32(leagueInfo.SimulationUserCount))
		}
	}
	fbsCommon.LeagueAddCategory(builder, uint8(leagueInfo.Category))
	leagueBinary := fbsCommon.LeagueEnd(builder)

	return builder, leagueBinary
}

// LeagueRaceSignupCountPush 推送比赛报名人数
func LeagueRaceSignupCountPush(leagueId int, raceId int64, cnt int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueRaceSignupCountPushStart(builder)
	fbsCommon.LeagueRaceSignupCountPushAddLeagueId(builder, int32(leagueId))
	fbsCommon.LeagueRaceSignupCountPushAddRaceId(builder, raceId)
	fbsCommon.LeagueRaceSignupCountPushAddCount(builder, int32(cnt))
	orc := fbsCommon.LeagueRaceSignupCountPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueRaceSignupCountPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueApplyResponse 回应用户报名申请
func LeagueApplyResponse(mNumber uint16, raceInfo *model.Race, raceUserInfo *model.RaceUser, leagueInfo *model.League) *protocal.ImPacket {
	var s2cResult flatbuffers.UOffsetT
	var raceInfoBinary flatbuffers.UOffsetT
	var raceUserBinary flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, s2cResult = response.BuidGameResult(builder, nil)
	builder, raceInfoBinary = buildRace(builder, raceInfo, leagueInfo)
	builder, raceUserBinary = buildRaceUser(builder, raceUserInfo)

	fbsCommon.LeagueApplyResponseStart(builder)
	fbsCommon.LeagueApplyResponseAddS2cResult(builder, s2cResult)
	fbsCommon.LeagueApplyResponseAddRaceInfo(builder, raceInfoBinary)
	fbsCommon.LeagueApplyResponseAddRaceUserInfo(builder, raceUserBinary)
	orc := fbsCommon.LeagueApplyResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueApplyResponse, protocal.MSG_TYPE_RESPONSE, uint16(0), mNumber, buf)
}

func LeagueApplyError(mNumber uint16, err *ierror.Error) *protocal.ImPacket {
	var s2cResult flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, s2cResult = response.BuidGameResult(builder, err)

	fbsCommon.LeagueApplyResponseStart(builder)
	fbsCommon.LeagueApplyResponseAddS2cResult(builder, s2cResult)
	orc := fbsCommon.LeagueApplyResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueApplyResponse, protocal.MSG_TYPE_RESPONSE, uint16(0), mNumber, buf)
}

// LeagueCancelResponse 取消报名结果
func LeagueCancelResponse(mNumber uint16, err *ierror.Error) *protocal.ImPacket {
	var s2cResult flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, s2cResult = response.BuidGameResult(builder, err)

	fbsCommon.LeagueCancelResponseStart(builder)
	fbsCommon.LeagueCancelResponseAddS2cResult(builder, s2cResult)
	orc := fbsCommon.LeagueCancelResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueCancelResponse, protocal.MSG_TYPE_RESPONSE, uint16(0), mNumber, buf)
}

// 构建比赛信息
func buildRace(builder *flatbuffers.Builder, raceInfo *model.Race, leagueInfo *model.League) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	// 读取剩余房间数
	playingRoomCnt := 0
	if raceInfo.Status == config.RACE_STATUS_PLAY || raceInfo.Status == config.RACE_STATUS_SETTLEMENT {
		playingRoomCnt = model.GetRaceRooms(raceInfo.Id).PlayingCount()
	}
	var name, img, settingBinary, rewardsBinary, rounds flatbuffers.UOffsetT

	rounds = builder.CreateString(leagueInfo.Rounds)
	name = builder.CreateString(leagueInfo.Name)
	img = builder.CreateString(leagueInfo.Img)

	var setting []byte
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
	fbsCommon.RaceAddRequireUserCount(builder, int32(raceInfo.RequireUserCount))
	fbsCommon.RaceAddRequireUserMin(builder, int32(raceInfo.RequireUserMin))
	fbsCommon.RaceAddSignupUserCount(builder, int32(raceInfo.SignupUserCount))
	fbsCommon.RaceAddPriceEntityId(builder, int32(leagueInfo.PriceEntityId))
	fbsCommon.RaceAddPrice(builder, int32(leagueInfo.Price))
	fbsCommon.RaceAddLeagueType(builder, int32(leagueInfo.LeagueType))
	fbsCommon.RaceAddSignupTime(builder, raceInfo.SignTime)
	fbsCommon.RaceAddGiveupTime(builder, raceInfo.GiveupTime)
	fbsCommon.RaceAddStartTime(builder, raceInfo.StartTime)
	fbsCommon.RaceAddStatus(builder, int32(raceInfo.Status))
	fbsCommon.RaceAddRound(builder, int32(raceInfo.Round))
	fbsCommon.RaceAddRewards(builder, rewardsBinary)
	fbsCommon.RaceAddRoomCnt(builder, int32(playingRoomCnt))
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

// LeagueRacePush 推送当前比赛信息
func LeagueRacePush(raceInfo *model.Race, raceUserInfo *model.RaceUser, leagueInfo *model.League) *protocal.ImPacket {
	var raceInfoBinary flatbuffers.UOffsetT
	var raceUserBinary flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, raceInfoBinary = buildRace(builder, raceInfo, leagueInfo)
	builder, raceUserBinary = buildRaceUser(builder, raceUserInfo)

	fbsCommon.LeagueRacePushStart(builder)
	fbsCommon.LeagueRacePushAddRaceInfo(builder, raceInfoBinary)
	fbsCommon.LeagueRacePushAddRaceUserInfo(builder, raceUserBinary)
	orc := fbsCommon.LeagueRacePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueRacePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueRaceNilPush 推送一个空的RacePush
func LeagueRaceNilPush() *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueRacePushStart(builder)
	orc := fbsCommon.LeagueRacePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueRacePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueListReloadPush 刷新比赛
func LeagueListReloadPush(leagueInfo *model.League, raceInfo *model.Race) *protocal.ImPacket {
	var leagueInfoBinary flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, leagueInfoBinary = buildLeague(builder, leagueInfo, raceInfo)

	fbsCommon.LeagueListReloadPushStart(builder)
	fbsCommon.LeagueListReloadPushAddLeague(builder, leagueInfoBinary)
	orc := fbsCommon.LeagueListReloadPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueListReloadPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)

}

// LeagueListRemovePush 推送比赛移除通知
func LeagueListRemovePush(leagueId int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueListRemovePushStart(builder)
	fbsCommon.LeagueListRemovePushAddLeagueId(builder, int32(leagueId))
	orc := fbsCommon.LeagueListRemovePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueListRemovePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueL2SPlanPush 构建通知游戏服排赛的消息
func LeagueL2SPlanPush(raceRoomId int64) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueL2SPlanPushStart(builder)
	fbsCommon.LeagueL2SPlanPushAddRaceRoomId(builder, raceRoomId)
	orc := fbsCommon.LeagueL2SPlanPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueL2SPlanPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueL2SRankRefreshPush 通知客户端更新比赛排名
func LeagueL2SRankRefreshPush(raceId int64) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueL2SRankRefreshPushStart(builder)
	fbsCommon.LeagueL2SRankRefreshPushAddRaceId(builder, raceId)
	orc := fbsCommon.LeagueL2SRankRefreshPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueL2SRankRefreshPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueGameStartPush 通知用户游戏开始
func LeagueGameStartPush(raceId int64, roomId uint64) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueGameStartPushStart(builder)
	fbsCommon.LeagueGameStartPushAddRaceId(builder, raceId)
	fbsCommon.LeagueGameStartPushAddRoomId(builder, roomId)
	orc := fbsCommon.LeagueGameStartPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueGameStartPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueGameRankPush 通知用户排名更新
func LeagueGameRankPush(rank, score, roomCnt int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueGameRankPushStart(builder)
	fbsCommon.LeagueGameRankPushAddRank(builder, int32(rank))
	fbsCommon.LeagueGameRankPushAddScore(builder, int32(score))
	fbsCommon.LeagueGameRankPushAddRoomCnt(builder, int32(roomCnt))
	orc := fbsCommon.LeagueGameRankPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueGameRankPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueQuitResponse 退赛结果
func LeagueQuitResponse(mNumber uint16, raceInfo *model.Race, leagueInfo *model.League, err *ierror.Error) *protocal.ImPacket {
	var raceInfoBinary, s2cResult flatbuffers.UOffsetT
	builder := flatbuffers.NewBuilder(0)
	builder, s2cResult = response.BuidGameResult(builder, err)
	if raceInfo != nil {
		builder, raceInfoBinary = buildRace(builder, raceInfo, leagueInfo)
	}

	fbsCommon.LeagueQuitResponseStart(builder)
	fbsCommon.LeagueQuitResponseAddRaceInfo(builder, raceInfoBinary)
	fbsCommon.LeagueQuitResponseAddS2cResult(builder, s2cResult)
	orc := fbsCommon.LeagueQuitResponseEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueQuitResponse, protocal.MSG_TYPE_RESPONSE, uint16(0), mNumber, buf)
}

// LeagueRaceCancelPush 取消比赛
func LeagueRaceCancelPush(raceId int64) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueRaceCancelPushStart(builder)
	fbsCommon.LeagueRaceCancelPushAddRaceId(builder, raceId)
	orc := fbsCommon.LeagueRaceCancelPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueRaceCancelPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueRaceResultPush 游戏结果
func LeagueRaceResultPush(raceInfo *model.Race, leagueInfo *model.League, userId, rank int, rewardDesc string) *protocal.ImPacket {
	var raceBinary, rewards flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, raceBinary = buildRace(builder, raceInfo, leagueInfo)
	if len(rewardDesc) != 0 {
		rewards = builder.CreateString(rewardDesc)
	}
	fbsCommon.LeagueRaceResultPushStart(builder)
	fbsCommon.LeagueRaceResultPushAddRaceInfo(builder, raceBinary)
	fbsCommon.LeagueRaceResultPushAddUserId(builder, uint32(userId))
	fbsCommon.LeagueRaceResultPushAddRank(builder, int32(rank))
	fbsCommon.LeagueRaceResultPushAddRewardsDesc(builder, rewards)
	orc := fbsCommon.LeagueRaceResultPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueRaceResultPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// 构建消息
func buildMessage(builder *flatbuffers.Builder, messageId int, content string) (*flatbuffers.Builder, flatbuffers.UOffsetT) {
	s := builder.CreateString(content)
	fbsCommon.S2cMessageStart(builder)
	fbsCommon.S2cMessageAddMessageId(builder, uint32(messageId))
	fbsCommon.S2cMessageAddContent(builder, s)
	v := fbsCommon.RaceEnd(builder)
	return builder, v
}

// PrivateMessagePush 消息推送
func PrivateMessagePush(userId int, messageId int, content string) *protocal.ImPacket {
	var message flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	builder, message = buildMessage(builder, messageId, content)

	fbsCommon.GatewayS2CPrivateMessagePushStart(builder)
	fbsCommon.GatewayS2CPrivateMessagePushAddUserId(builder, uint32(userId))
	fbsCommon.GatewayS2CPrivateMessagePushAddS2cMessage(builder, message)
	orc := fbsCommon.GatewayS2CPrivateMessagePushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandGatewayS2CPrivateMessagePush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}
