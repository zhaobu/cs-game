package log

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	simplejson "github.com/bitly/go-simplejson"
	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/model"
	configService "mahjong.go/service/config"
)

// 插入日志
func saveLog(log interface{}) *core.Error {
	o := core.GetWriter()
	if _, dberr := o.Insert(log); dberr != nil {
		core.Logger.Error("写日志错误: %s,data:%#v", dberr.Error(), log)

		return core.NewError(-4, dberr.Error())
	}

	return nil
}

// 更新日志
func updateLog(log interface{}, a ...string) *core.Error {
	o := core.GetWriter()
	if _, dberr := o.Update(log, a...); dberr != nil {
		core.Logger.Error("写日志错误: %s, data:%#v", dberr.Error(), log)

		return core.NewError(-4, dberr.Error())
	}

	return nil
}

// 保存日志
func insertOrUpdateLog(log interface{}) *core.Error {
	o := core.GetWriter()
	if _, dberr := o.InsertOrUpdate(log); dberr != nil {
		core.Logger.Error("写日志错误: %s,data:%#v", dberr.Error(), log)

		return core.NewError(-4, dberr.Error())
	}

	return nil
}

// 构建sn
// YYMMDDHHIISS + userId(固定8位，左补0) + 4位随机数字
func GenSn(userId int) string {
	arr := []string{time.Now().Format("060102150405"), fmt.Sprintf("%08d", userId), util.GetRandString(4)}
	return strings.Join(arr, "")
}

// 消费日志
func LogMoney(userId, money, giftMoney int, change_type string, sn string) *core.Error {
	// 跳过机器人日志
	if configService.IsRobot(userId) {
		return nil
	}

	moneyLog := new(config.UserAccountLog)
	moneyLog.UserId = userId
	moneyLog.Money = money
	moneyLog.GiftMoney = giftMoney
	moneyLog.CreateTime = util.GetTime()
	moneyLog.Sn = sn
	moneyLog.ChangeType = change_type

	return saveLog(moneyLog)
}

// 记录房间用户输赢日志
func LogGameUserRecords(userId, score, wins, loses int, roomId, createTime int64) *core.Error {
	// 已存在， update ，不存在 insert
	o := core.GetWriter()
	record := new(config.GameUserRecords)
	qs := o.QueryTable(record)
	err := qs.Filter("room_id", roomId).Filter("user_id", userId).One(record)

	if err == orm.ErrNoRows {
		gameUserRecord := new(config.GameUserRecords)
		gameUserRecord.RoomId = roomId
		gameUserRecord.UserId = userId
		gameUserRecord.Score = score
		gameUserRecord.Wins = wins
		gameUserRecord.Loses = loses
		gameUserRecord.CreateTime = createTime
		return saveLog(gameUserRecord)
	} else if err == orm.ErrMissPK {
		// 多条的时候报错
		core.Logger.Error("Returned Multi Rows Not One")
	}

	record.Score = score
	record.Wins = wins
	record.Loses = loses
	return updateLog(record, "Score", "Wins", "Loses")
}

// 记录消费日志
func LogConsumeInfo(userId, money, giftMoney int, sn, note string, ctype int) *core.Error {
	// 跳过机器人日志
	if configService.IsRobot(userId) {
		return nil
	}

	userConsumeInfo := new(config.UserConsumeInfo)
	userConsumeInfo.Sn = sn
	userConsumeInfo.UserId = userId
	userConsumeInfo.Num = money
	userConsumeInfo.GiftNum = giftMoney
	userConsumeInfo.Note = note
	userConsumeInfo.Ctype = ctype
	userConsumeInfo.CreateTime = util.GetTime()

	return saveLog(userConsumeInfo)
}

// 记录收入日志
func LogUserTransInfo(userId, targetUserId, amount int, sn string, transType int, diamondType int) *core.Error {
	// 跳过机器人日志
	if configService.IsRobot(userId) {
		return nil
	}
	userTransInfo := new(config.UserTransInfo)
	userTransInfo.Sn = sn
	userTransInfo.UserId = userId
	userTransInfo.TargetUserId = targetUserId
	userTransInfo.Num = amount
	userTransInfo.CreateTime = util.GetTime()
	userTransInfo.DiamondType = diamondType
	userTransInfo.TransType = transType

	return saveLog(userTransInfo)
}

// 记录房间信息
func LogGameInfo(roomInfo *config.GameInfo) *core.Error {
	return saveLog(roomInfo)
}

// 记录房间结果
func LogGameResult(roomId int64, number string, cType, mType, tRound int, players string, userScores []map[string]int, isDismiss int, dismissUsers map[int]int, createTime, startTime int64, round int) *core.Error {
	gameResult := new(config.GameResult)
	gameResult.RoomId = roomId
	gameResult.RoomNum = number
	gameResult.GameType = cType
	gameResult.MahjongType = mType
	gameResult.TotalRounds = tRound
	gameResult.Players = players
	gameResult.Scores, _ = util.InterfaceToJsonString(userScores)
	gameResult.CreateTime = createTime
	gameResult.StartTime = startTime
	gameResult.CompleteTime = util.GetTime()
	gameResult.IsDismiss = isDismiss
	if isDismiss > 0 && dismissUsers != nil {
		users := make([]int, 0, len(dismissUsers))
		for userId, _ := range dismissUsers {
			users = append(users, userId)
		}
		gameResult.DismissUsers = util.SliceJoin(users, ",")
	}
	gameResult.PlayRounds = round

	return insertOrUpdateLog(gameResult)
}

// 设置房间解散标志
func LogGameDismiss(roomId int64) *core.Error {
	o := core.GetWriter()
	gameResult := config.GameResult{RoomId: roomId}
	o.Read(&gameResult)
	if gameResult.RoomNum != "" {
		gameResult.IsDismiss = 1
		return updateLog(&gameResult, "is_dismiss")
	}
	return nil
}

// 记录当局游戏数据
func LogGameRoundData(roomId int64, round int, userScores []map[string]int, data map[string]interface{}, huang int, winPlayers []int, goldBam1, goldDot8 int, startTime int64) *core.Error {
	gameRoundData := new(config.GameRoundData)
	gameRoundData.RoomId = roomId
	gameRoundData.Round = round
	gameRoundData.Scores, _ = util.InterfaceToJsonString(userScores)
	gameRoundData.Data, _ = util.InterfaceToJsonString(data)
	gameRoundData.Huang = huang
	gameRoundData.WinPlayers = util.SliceJoin(winPlayers, ",")
	gameRoundData.WinPlayersCnt = len(winPlayers)
	gameRoundData.GoldBam1 = goldBam1
	gameRoundData.GoldDot8 = goldDot8
	gameRoundData.StartTime = startTime
	gameRoundData.CompleteTime = util.GetTime()
	return saveLog(gameRoundData)
}

// 记录用户游戏信息
func LogGameUserRound(gameUserRound *config.GameUserRound) *core.Error {
	// 机器人不记录
	if configService.IsRobot(gameUserRound.UserId) {
		return nil
	}
	return saveLog(gameUserRound)
}

// 记录用户游戏次数
func LogUserGameRoundTimes(userId, gameType, cType int) *core.Error {
	// 是否插入
	isInsert := false
	o := core.GetWriter()
	userTimes := config.UserGameRoundTimes{UserId: userId}
	err := o.Read(&userTimes)
	if err == orm.ErrNoRows || err == orm.ErrMissPK {
		isInsert = true
		userTimes.CreateTimesDetail = "{}"
	} else {
		if userTimes.CreateTimesDetail == "" {
			userTimes.CreateTimesDetail = "{}"
		}
	}

	// gameType
	gameTypeStr := strconv.Itoa(gameType)
	var js *simplejson.Json
	// 自主创建
	userTimes.CreateTimes++
	js, _ = simplejson.NewJson([]byte(userTimes.CreateTimesDetail))
	currentTimes, err := js.Get(gameTypeStr).Int()
	if err == nil {
		currentTimes++
	} else {
		currentTimes = 1
	}
	js.Set(gameTypeStr, currentTimes)
	jsMap, _ := js.Map()
	userTimes.CreateTimesDetail, _ = util.InterfaceToJsonString(jsMap)
	userTimes.CreateUpdateTime = util.GetTime()

	if isInsert {
		return saveLog(&userTimes)
	}
	return updateLog(&userTimes)
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 俱乐部
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// LogClubRoom 记录俱乐部已完成的房间列表
func LogClubRoom(roomId, createTime int64, clubId int) *core.Error {
	clubRoom := new(config.ClubRoom)
	clubRoom.ClubId = clubId
	clubRoom.RoomId = roomId
	clubRoom.CreateTime = createTime
	return saveLog(clubRoom)
}

// 俱乐部消费日志
func LogClubConsumeInfo(roomId int64, clubId, diamonds int, logType string, createTime int64) *core.Error {
	clubConsumeInfo := new(config.ClubConsumeLog)
	clubConsumeInfo.ClubId = clubId
	clubConsumeInfo.RoomId = roomId
	clubConsumeInfo.Diamonds = diamonds * -1
	clubConsumeInfo.LogType = logType
	clubConsumeInfo.CreateTime = createTime
	return saveLog(clubConsumeInfo)
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 金币场
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 金币场消费日志
func LogCoinConsumeLog(roomId int64, userId int, coinType int, coin int, createTime int64) *core.Error {
	defer util.RecoverPanic()

	coinConsumeLog := new(config.CoinConsumeLog)
	coinConsumeLog.MatchType = coinType
	coinConsumeLog.RoomId = roomId
	coinConsumeLog.UserId = userId
	coinConsumeLog.Coin = coin
	coinConsumeLog.CreateTime = createTime
	return saveLog(coinConsumeLog)
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 排位赛
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 排位赛card消费记录
func LogRankConsumeInfo(seasonId int, roomId int64, userId int, num int, gradeId, gradeLevel, star int, createTime int64) *core.Error {
	seasonConsumeInfo := new(model.SeasonConsumeLog)
	seasonConsumeInfo.SeasonId = seasonId
	seasonConsumeInfo.UserId = userId
	seasonConsumeInfo.RoomId = roomId
	seasonConsumeInfo.Num = num
	seasonConsumeInfo.Grade = gradeId
	seasonConsumeInfo.Level = gradeLevel
	seasonConsumeInfo.Stars = star
	seasonConsumeInfo.CreateTime = createTime
	return saveLog(seasonConsumeInfo)
}
