package game

import (
	"mahjong.go/library/core"
	"mahjong.go/library/response"
	"mahjong.go/mi/protocal"

	flatbuffers "github.com/google/flatbuffers/go"
	fbsCommon "mahjong.go/fbs/Common"
)

// LeagueS2LPlanPush 回应排赛结果
func LeagueS2LPlanPush(err *core.Error, roomId, raceRoomId, raceId int64) *protocal.ImPacket {
	var commonResult flatbuffers.UOffsetT

	builder := flatbuffers.NewBuilder(0)
	// 生成commonResult
	builder, commonResult = genGameResult(builder, err)

	// 构建对象
	fbsCommon.LeagueS2LPlanPushStart(builder)
	fbsCommon.LeagueS2LPlanPushAddS2cResult(builder, commonResult)
	fbsCommon.LeagueS2LPlanPushAddRaceRoomId(builder, raceRoomId)
	fbsCommon.LeagueS2LPlanPushAddRaceId(builder, raceId)
	fbsCommon.LeagueS2LPlanPushAddRoomId(builder, uint64(roomId))
	orc := fbsCommon.LeagueS2LPlanPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueS2LPlanPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueS2LRoundFinishPush 单局完成的消息
func LeagueS2LRoundFinishPush(raceRoomId, raceId int64, scores []int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueS2LRoundFinishPushStartScoresVector(builder, len(scores))
	for i := len(scores) - 1; i >= 0; i-- {
		builder.PrependInt32(int32(scores[i]))
	}
	scoresBinary := builder.EndVector(len(scores))

	fbsCommon.LeagueS2LRoundFinishPushStart(builder)
	fbsCommon.LeagueS2LRoundFinishPushAddRaceRoomId(builder, raceRoomId)
	fbsCommon.LeagueS2LRoundFinishPushAddRaceId(builder, raceId)
	fbsCommon.LeagueS2LRoundFinishPushAddScores(builder, scoresBinary)
	orc := fbsCommon.LeagueS2LPlanPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueS2LRoundFinishPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
}

// LeagueS2LGameFinishPush 房间完成的消息
func LeagueS2LGameFinishPush(raceRoomId, raceId int64, code int) *protocal.ImPacket {
	builder := flatbuffers.NewBuilder(0)
	fbsCommon.LeagueS2LGameFinishPushStart(builder)
	fbsCommon.LeagueS2LGameFinishPushAddRaceRoomId(builder, raceRoomId)
	fbsCommon.LeagueS2LGameFinishPushAddRaceId(builder, raceId)
	fbsCommon.LeagueS2LGameFinishPushAddCode(builder, int32(code))
	orc := fbsCommon.LeagueS2LGameFinishPushEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()
	return response.GenFbs(protocal.PACKAGE_TYPE_DATA, fbsCommon.CommandLeagueS2LGameFinishPush, protocal.MSG_TYPE_PUSH, uint16(0), uint16(0), buf)
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
