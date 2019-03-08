package game

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"
	"mahjong.go/mi/setting"
	"mahjong.go/model"
	"mahjong.go/rank"
	"mahjong.go/robot"

	fbsCommon "mahjong.go/fbs/Common"
	matchModel "mahjong.go/model/match"
	clubService "mahjong.go/service/club"
	coinService "mahjong.go/service/coin"
	configService "mahjong.go/service/config"
	friendService "mahjong.go/service/friend"
	hallService "mahjong.go/service/hall"
	logService "mahjong.go/service/log"
	rankService "mahjong.go/service/rank"
	roomService "mahjong.go/service/room"
	userService "mahjong.go/service/user"
)

// Room 房间信息
type Room struct {
	RoomId   int64  // 房间id
	Number   string // 房间号
	CType    int    // 房间类型（自主创建、随机创建）
	MType    int    // 麻将类型（贵阳、三丁。。。）
	TRound   int    // 局数类型（8、12。。。）
	CoinType int    // 金币场类型
	// 设置
	// 满堂鸡: 打出去的弃牌也算鸡
	// 上下鸡: 翻出来的牌，前一张
	// 乌骨鸡: 8筒
	// 前后鸡: 每次成功杠牌，从牌墙后向前移动一位，翻开新的，关闭旧的
	// 星期鸡: 星期几对应的万、条、筒都是鸡
	// 意外鸡: 7筒、9万
	// 吹风鸡: 5筒，该把牌所有鸡、豆、胡、杠、包全部无效，直接开始下一局，庄家连庄
	// 滚筒鸡: 和前后鸡的区别就是不关闭旧的，这两个设置是互斥的
	// 本鸡: 翻出来的自己也算鸡
	setting *setting.MSetting

	Mux              *sync.Mutex // 房间锁
	UserOperationMux *sync.Mutex // 操作锁

	Status int // 房间状态（组队中、游戏中）

	Index        *sync.Map // 位置用户索引 index => userId
	Users        *sync.Map // 房间用户列表
	Creator      int       // 创建者
	CreateMode   int       // 房间创建模式
	CreateTime   int64     // 创建时间戳
	StartTime    int64     // 房间开始时间
	Dissmisser   int       // 提出解散者
	DismissTime  int64     // 提出解散时间
	DismissOp    *sync.Map // 解散操作情况
	DissmissChan chan int

	FullTime               int64 // 满员时间
	LastRoundCompletedTime int64 // 上一局结束时间

	// 麻将
	Round     int                // 当前牌局数
	MI        MahjongInterface   // 当前进行中的麻将
	Record    []MahjongInterface // 游戏记录
	Ready     int                // 准备标记，0表示准备中，1表示已准备
	ReadyList []int              // 准备中的，需要记录准备列表

	Dealer    int // 庄家id
	DealCount int // 连庄数

	// 用户积分信息
	ScoreInfo *sync.Map

	HostingUsers []int // 托管用户列表

	// 观察者
	Ob *Ob

	// 俱乐部
	ClubId            int // 所属俱乐部id
	ClubCapitorUserId int // 俱乐部馆主的用户id
	ClubScore         int // 淘汰赛积分

	// 付费方式
	PayType int
	// 付费金额
	PayPrice int

	// 房间版本：创建者版本
	RoomVersion string

	// 作弊相关
	NearUsers       []int           // 距离过近的用户列表
	NoPositionUsers []int           // 未开放位置的用户列表e
	UserDistances   []*UserDistance // 用户之间的距离

	// 关联房间id
	RelationRoomId int64
	// 联赛相关
	LeagueInfo *model.League
	RaceInfo   *model.Race
	RaceRoom   *model.RaceRoom

	// 金币场相关
	CoinConfig *coinService.Config

	// 排位赛相关
	SeasonId int
	GradeId  int
}

type UserDistance struct {
	MinUserId int // 较小的用户id
	MaxUserId int // 较大的用户id
	Distance  int // 距离
}

// GetScoreInfo 获取用户的积分信息
func (room *Room) GetScoreInfo(userId int) *FrontFinalInfo {
	if v, ok := room.ScoreInfo.Load(userId); ok {
		return v.(*FrontFinalInfo)
	}
	return nil
}

// GetScoreInfoList 获取用户的积分列表
func (room *Room) GetScoreInfoList() map[int]*FrontFinalInfo {
	scoreInfoList := make(map[int]*FrontFinalInfo)
	room.ScoreInfo.Range(func(k, v interface{}) bool {
		scoreInfoList[k.(int)] = v.(*FrontFinalInfo)
		return true
	})
	return scoreInfoList
}

// 新建一个房间对象
func NewRoom() *Room {
	room := &Room{}
	// 房间的设置
	room.setting = setting.NewMSetting()
	// 房间用户列表
	room.Users = &sync.Map{}
	// 房间操作者
	room.DismissOp = &sync.Map{}
	// 位置索引
	room.Index = &sync.Map{}

	// 牌局日志
	room.Record = []MahjongInterface{}

	// 锁
	room.Mux = &sync.Mutex{}
	// 操作锁，防止并发操作
	room.UserOperationMux = &sync.Mutex{}

	// 托管用户列表
	room.HostingUsers = make([]int, 0)

	return room
}

// 新建房间
func Create(creatorId int, gType int, round int, cType int, setting []int, mode int) *Room {
	room := NewRoom()
	// 房间类型
	room.CType = cType
	// 游戏类型类型
	room.MType = gType
	// 牌局总局数
	room.TRound = round
	// 生成房间id
	room.RoomId = roomService.GenRoomId()
	// 生成房间号
	room.Number = roomService.GetRoomNumber()

	// 创建者、创建时间
	room.Creator = creatorId
	room.CreateTime = util.GetTime()

	room.setting.SetSetting(setting)

	// 观察者
	room.Ob = NewOb()

	// 房间模式
	room.CreateMode = mode

	// 房间版本
	if mode == config.ROOM_CREATE_MODE_USER {
		if u, _ := UserMap.GetUser(creatorId); u != nil {
			room.RoomVersion = u.Info.Version
		}
	}
	if room.RoomVersion == "" {
		room.RoomVersion = GameVersion
	}

	return room
}

// GetIndexUserId 给房间添加用户
func (room *Room) GetIndexUserId(index int) int {
	userId := 0
	if v, ok := room.Index.Load(index); ok {
		userId = v.(int)
	}
	return userId
}

// GetIndexLen 获取房间索引长度
func (room *Room) GetIndexLen() int {
	return util.SMapLen(room.Index)
}

// GetIndexUserIds 获取房间所有索引的用户id
func (room *Room) GetIndexUserIds() []int {
	userIds := []int{}
	room.Index.Range(func(k, v interface{}) bool {
		userIds = append(userIds, v.(int))
		return true
	})
	return userIds
}

// IndexToString 获取房间所有索引的用户id
func (room *Room) IndexToString() string {
	str := ""
	room.Index.Range(func(k, v interface{}) bool {
		str += fmt.Sprintf("index=%v&userId=%v,", k.(int), v.(int))
		return true
	})
	return str
}

// GetUser 获取房间用户
func (room *Room) GetUser(userId int) *RoomUser {
	if v, ok := room.Users.Load(userId); ok {
		return v.(*RoomUser)
	}
	return nil
}

// GetUsers 获取房间用户列表
func (room *Room) GetUsers() map[int]*RoomUser {
	users := make(map[int]*RoomUser)
	room.Users.Range(func(k, v interface{}) bool {
		ru := v.(*RoomUser)
		users[ru.UserId] = ru
		return true
	})
	return users
}

// GetUsersLen 获取房间用户数
func (room *Room) GetUsersLen() int {
	return util.SMapLen(room.Users)
}

// AddUser 给房间添加用户
func (room *Room) AddUser(user *User) {
	// 计算用户索引
	var index int
	for i := 0; i < room.setting.GetSettingPlayerCnt(); i++ {
		if _, ok := room.Index.Load(i); !ok {
			index = i
			break
		}
	}
	room.Index.Store(index, user.UserId)

	// 生成用户
	roomUser := NewRoomUser(user.UserId)
	roomUser.Index = index
	roomUser.Info = user.Info
	roomUser.CType = room.CType
	room.Users.Store(user.UserId, roomUser)
	core.Logger.Debug("[room.AddUser]roomId:%v, userId:%v, user index:%v, room index:%#v", room.RoomId, user.UserId, index, room.IndexToString())
}

// UpdateUser 在用户重新登录之后, 更新用户信息
func (room *Room) UpdateUser(user *User) {
	ru := room.GetUser(user.UserId)
	if ru != nil {
		ru.Info = user.Info
	}
}

// 检查房间设置是否正确
func (room *Room) verifySetting() *core.Error {
	// 从cache中读取房间配置
	return nil
}

// 读取房间真实用户数量(即排除机器人)
func (room *Room) getTruePlayerCount() int {
	cnt := 0
	room.Index.Range(func(index, userId interface{}) bool {
		if !configService.IsRobot(userId.(int)) {
			cnt++
		}
		return true
	})
	return cnt
}

// getRandomSetting 读取随机房间的配置
// 测试服务器的测试用户，可以自行配置
func getRandomSetting(userId, gType int) []int {
	setting := configService.GetRandomRoomDefaultSetting(gType)
	// 这里需要支持一下配置
	if core.AppConfig.EnableRandomRoomSetting == 1 {
		// 判断用户是否测试用户
		testSetting, err := util.GetIntSliceFromFile(fmt.Sprintf("conf/%s/setting-%v.txt", core.AppConfig.Env, userId), ",")
		if err == nil && len(testSetting) == len(setting) {
			setting = testSetting
		}
	}

	return setting
}

// 读取房间的最后活动时间
// 即所有用户的最后心跳时间(不包括机器人)
func (room *Room) GetUserLastHeartTime() int64 {
	lastTime := int64(0)
	room.Users.Range(func(k, v interface{}) bool {
		ru := v.(*RoomUser)
		if !configService.IsRobot(ru.UserId) && ru.Info.LastHeartBeatTime > lastTime {
			lastTime = ru.Info.LastHeartBeatTime
		}
		return true
	})
	return lastTime
}

// 读取房间的同意解散人员���表
// 申请者放在第一个位置
func (room *Room) GetDismissUsers() []int {
	dismissUsers := []int{}
	if room.Dissmisser > 0 {
		dismissUsers = append(dismissUsers, room.Dissmisser)
		room.DismissOp.Range(func(k, v interface{}) bool {
			userId := k.(int)
			if userId != room.Dissmisser {
				dismissUsers = append(dismissUsers, userId)
			}
			return true
		})
	}
	return dismissUsers
}

// 读取未准备用户列表
func (room *Room) GetUnReadyUsers() []int {
	users := []int{}
	room.Users.Range(func(k, v interface{}) bool {
		ru := v.(*RoomUser)
		if !util.IntInSlice(ru.UserId, room.ReadyList) {
			users = append(users, ru.UserId)
		}
		return true
	})
	return users
}

// TimeOutSecond 获取房间过期时长
func (room *Room) TimeOutSecond() int64 {
	if room.IsCreate() || room.IsTV() || room.IsClub() || room.IsLeague() {
		return config.ROOM_TIMEOUT_SECOND
	}
	return config.ROOM_RANDOM_TIMEOUT_SECOND
}

// 获取房间用户连接起来的字符串
func (room *Room) userIdJoin() string {
	return util.SliceJoin(room.getUserIds(), ",")
}

// 获取房间用户id列表
func (room *Room) getUserIds() []int {
	s := []int{}
	room.Index.Range(func(index, userId interface{}) bool {
		s = append(s, userId.(int))
		return true
	})
	return s
}

// SetNextDealer 设置下一把的庄家
// 如果庄家与上一把相同，则连庄数+1
func (room *Room) SetNextDealer(dealer int) {
	if room.StartTime == 0 || dealer != room.Dealer {
		room.DealCount = 2
		room.Dealer = dealer
	} else {
		room.DealCount++
	}
	core.Logger.Debug("设置下一把的庄家, roomId:%d, dealId:%d, dealCount:%v", room.RoomId, dealer, room.DealCount)
}

// SetReady 设置房间已准备
func (room *Room) SetReady() {
	room.Ready = config.ROOM_READY_YES
}

// SetNotReady 设置房间未准备
func (room *Room) SetNotReady() {
	room.Ready = config.ROOM_READY_NO
	room.ReadyList = []int{}
	room.Ob.readyUsers = []int{}
}

// 判断房间人数是否已满
func (room *Room) IsFull() bool {
	return room.GetUsersLen() == room.setting.GetSettingPlayerCnt()
}

// IsReadying 返回房间是否处于准备状态
func (room *Room) IsReadying() bool {
	return room.Ready == config.ROOM_READY_NO
}

// IsFirstReadying 返回房间是否处于满员后到所有人准备的这段时间
// 满员 & 未开始
func (room *Room) IsFirstReadying() bool {
	return room.IsFull() && room.StartTime == 0
}

// 房间是否解散中
func (room *Room) IsDismissing() bool {
	return room.Dissmisser > 0
}

// 判断房间是否超时
func (room *Room) IsTimeout() bool {
	currentTime := util.GetTime()

	// 判断房间创建有没有超过时间间隔
	if currentTime-room.CreateTime < room.TimeOutSecond() {
		return false
	}

	// 判断所有成员，是否在‘解散间隔’有活动过
	if memberLastHeartTime := room.GetUserLastHeartTime(); memberLastHeartTime > 0 && currentTime-memberLastHeartTime < room.TimeOutSecond() {
		return false
	}

	// 判断房间是否在‘解散间隔’时间内有过操作
	if !room.CheckStatus(config.ROOM_STATUS_CREATING) && currentTime-room.MI.getReplyInitTime() < room.TimeOutSecond() {
		return false
	}

	return true
}

// IsRandom 房间是否随机房间
func (room *Room) IsRandom() bool {
	return configService.IsRandomRoom(room.CType)
}

// IsCreate 是否自主组建的房间
func (room *Room) IsCreate() bool {
	return configService.IsCreateRoom(room.CType)
}

// IsMatch 是否比赛房间
func (room *Room) IsMatch() bool {
	return configService.IsMatchRoom(room.CType)
}

// IsTV 是否是电视端房间
func (room *Room) IsTV() bool {
	return configService.IsTVRoom(room.CType)
}

// IsClub 是否是俱乐部房间
func (room *Room) IsClub() bool {
	return configService.IsClubRoom(room.CType) || configService.IsClubMatchRoom(room.CType)
}

// IsClubMatch 是否是淘汰赛房间
func (room *Room) IsClubMatch() bool {
	return configService.IsClubMatchRoom(room.CType)
}

// IsCoin 是否是金币场
func (room *Room) IsCoin() bool {
	return configService.IsCoinRoom(room.CType)
}

// IsClubMode 是否是馆主代开的房间
func (room *Room) IsClubMode() bool {
	return room.CreateMode == config.ROOM_CREATE_MODE_CLUB
}

// IsLeague 是否联赛房间
func (room *Room) IsLeague() bool {
	return configService.IsLeagueRoom(room.CType)
}

// IsRank 是否排位赛房间
func (room *Room) IsRank() bool {
	return configService.IsRankRoom(room.CType)
}

// EnableReturnCost 是否支持退还房费
func (room *Room) EnableReturnCost() bool {
	return room.IsCreate() || room.IsClub()
}

// EnableDismissPunishment 房间是否有解散未回应惩罚
func (room *Room) EnableDismissPunishment() bool {
	return room.IsRandom() || room.IsMatch()
}

// EnableJoin 房间是否允许通过输入房间号或h5加入
func (room *Room) EnableJoin() bool {
	return room.IsCreate() || room.IsTV() || room.IsClub()
}

// EnableRobot 是否支持填充机器人
func (room *Room) EnableRobot() bool {
	return room.IsRandom() || room.IsMatch() || room.IsCoin() || room.IsRank()
	// return room.IsRandom() || room.IsMatch() || (room.IsCoin() && !core.IsProduct()) // IsForReview 当 qa用
}

// EnableAutoReady 是否支持自动准备
func (room *Room) EnableAutoReady() bool {
	return room.IsLeague() || room.IsRank() || room.IsCoin()
}

// 读取房间的收费消费类型
func (room *Room) consumeType() int {
	if room.IsCreate() {
		return config.MONEY_CONSUME_TYPE_CREATE
	} else if room.IsMatch() {
		return config.MONEY_CONSUME_TYPE_MATCH
	} else if room.IsRandom() {
		return config.MONEY_CONSUME_TYPE_RANDOM
	} else if room.IsClub() {
		return config.MONEY_CONSUME_TYPE_CLUB
	}
	return 0
}

// 检查房间状态
func (room *Room) CheckStatus(status int) bool {
	return room.Status == status
}

// 检测房间的加入版本需求
func (room *Room) checkJoinVersion(version string) bool {
	// 版本缺失
	if version == "" {
		return false
	}
	// 最新版
	if version == "latest" {
		return true
	}

	if strings.Compare(version, "1.7.0") == -1 {
		return false
	}

	// 1.8以下的版本，不能加入俱乐部房间
	if room.IsClub() && strings.Compare(version, "1.8.0") == -1 {
		return false
	}
	// 1.9 以下的用户不能加入新玩法
	if strings.Compare(version, "1.9.0") == -1 && room.MType >= fbsCommon.GameTypeMAHJONG_XY {
		return false
	}
	// 2.0 以下的用户不能加入新玩法
	if strings.Compare(version, "2.0.0") == -1 && room.MType >= fbsCommon.GameTypeMAHJONG_QX {
		return false
	}
	// 2.2 以下的用户不能加入仁怀麻将
	if strings.Compare(version, "2.2.0") == -1 && room.MType >= fbsCommon.GameTypeMAHJONG_RH {
		return false
	}
	// 3.1 以下的用户不能加入有最后一局翻倍的房间
	if strings.Compare(version, "3.1.0") == -1 && room.setting.IsSettingDLR() {
		return false
	}
	// 3.1 以下的用户不能加入有杠牌翻倍
	if strings.Compare(version, "3.1.0") == -1 && room.MType == fbsCommon.GameTypeMAHJONG_GFT {
		return false
	}
	// 3.5以下版本，不能加入有换牌玩法的房间
	if strings.Compare(version, "3.5.0") == -1 && room.setting.IsEnableExchange() {
		return false
	}

	return true
}

func (room *Room) dismissPunish() {
	room.Index.Range(func(k, v interface{}) bool {
		userId := v.(int)
		if _, exists := room.DismissOp.Load(userId); !exists {
			// 更新用户的被惩罚次数
			userService.UpdatePunishmentTimes(userId)
			// 更新用户的被惩罚标志
			userService.UpdatePunishmentFlag(userId)
			core.Logger.Info("[dismissPunish]roomId:%d, userId:%d", room.RoomId, userId)
		}
		return true
	})
}

func (room *Room) getLeagueId() int {
	if room.LeagueInfo != nil {
		return room.LeagueInfo.Id
	}
	return 0
}

// 组队满员之后，进入游戏
func (room *Room) enter() {
	// 记录满员时间
	room.FullTime = util.GetTime()

	// 设置初始庄家
	room.SetNextDealer(room.GetIndexUserId(0))

	// 设置未准备
	room.SetNotReady()

	// 发送gameEnter协议
	// 检测到作弊，推送作弊消息
	pushPacket := GameEnterPush()
	room.SendMessageToRoomUser(pushPacket, 0)

	// 推送准备消息给观察者
	room.Ob.sendMessage(pushPacket, 0)

	// 初始化用户积分信息
	room.ScoreInfo = &sync.Map{}

	room.Users.Range(func(k, v interface{}) bool {
		id := k.(int)
		ru := v.(*RoomUser)
		total := ru.GetAccumulativeScore()
		frontFinalInfo := &FrontFinalInfo{
			UserId:    id,
			Nickname:  ru.Info.Nickname,
			Avatar:    ru.Info.Avatar,
			AvatarBox: ru.Info.AvatarBox,
			Total:     total,
			FromTotal: total,
		}
		// 记录排位赛等级
		if room.IsRank() {
			frontFinalInfo.FromSLevel = ru.Info.ScoreRank
			frontFinalInfo.FinalSLevel = ru.Info.ScoreRank
			frontFinalInfo.WinningStreak = rankService.GetSeasonUserWinningStreak(room.SeasonId, id)
		}

		// 如果是比赛房间，给一下初始房间
		if room.IsLeague() && room.RaceInfo != nil {
			frontFinalInfo.Score = model.GetRaceUserScore(room.RaceInfo.Id, id)
		}
		room.ScoreInfo.Store(id, frontFinalInfo)
		return true
	})

	// 检查房间用户和房间用户索引是否匹配
	indexError := false
	for i := 0; i < room.setting.GetSettingPlayerCnt(); i++ {
		indexUserId := 0
		if v, ok := room.Index.Load(i); ok {
			indexUserId = v.(int)
		}
		if indexUserId == 0 {
			core.Logger.Error("[checkIndexError]indexUserId=0, roomId:%v, i:%v, index:%v", room.RoomId, i, room.IndexToString())
			indexError = true
			continue
		}
		ru := room.GetUser(indexUserId)
		if ru == nil {
			core.Logger.Error("[checkIndexError]room user not found, roomId:%v, i:%v, indexUserId:%v", room.RoomId, i, indexUserId)
			indexError = true
			continue
		}
		if i != ru.Index {
			core.Logger.Error("[checkIndexError]index not match, roomId:%v, i:%v, indexUserId:%v, ru.Index:%v", room.RoomId, i, indexUserId, ru.Index)
			indexError = true
			continue
		}
	}
	if indexError {
		core.Logger.Info("[checkIndexError]index出错, 需要进行修正, roomId:%v", room.RoomId)
		userIds := []int{}
		room.Users.Range(func(k, v interface{}) bool {
			ru := v.(*RoomUser)
			ru.Index = len(userIds)
			userIds = append(userIds, k.(int))
			return true
		})
		core.Logger.Info("[checkIndexError]index出错, 需要进行修正, roomId:%v, userIds:%+v", room.RoomId, userIds)
		for k, v := range userIds {
			room.Index.Store(k, v)
		}
		core.Logger.Info("[checkIndexError]index出错, 需要进行修正, roomId:%v, index:%+v", room.RoomId, room.IndexToString())
	}

	// 推送人齐了，开始的通知
	room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_FULL, 0)

	if room.EnableAutoReady() {
		go room.ListenHosting()
		/*
			// 自动准备
			room.ReadyList = room.getUserIds()
			room.SetReady()
			core.Logger.Debug("[room.autoready]roomId:%v, number:%v, round:%v", room.RoomId, room.Number, room.Round+1)
			room.nextGame()
		*/
	}
}

// 开启下一局游戏
func (room *Room) nextGame() {
	// 记录回合数
	room.Round++
	// 更改状态为进入中
	room.Status = config.ROOM_STATUS_PALYING

	if room.Round == 1 {
		// 第一局游戏开始时, 游戏才真的开始, 需要扣费记录一些数据
		// 若扣费失败需要解散游戏
		if !room.charge() {
			// 用户房费不足, 解散房间
			core.Logger.Info("用户房费不足，开房失败, roomId:%d", room.RoomId)
			dismissRoom(room, config.DISMISS_ROOM_CODE_MONEY_NOT_ENOUGH)
			return
		}

		// 需要记录游戏开始时间
		room.StartTime = util.GetTime()
		room.LastRoundCompletedTime = room.StartTime

		// 记录房间信息
		room.logGameInfo()

		// 第一局给离线ios用户发送APNs消息
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_NEXT_GAME, 0)

		// 检查作弊
		// room.checkCheat()

		// 开启房间hosting监听
		// 如果房间支持自动准备，需要在enter中准备
		if !room.EnableAutoReady() {
			go room.ListenHosting()
		}
	}

	// 生成麻将对象
	room.MI = MahjongFactory(room)

	// 开始执行
	room.MI.run()

	// 如果是俱乐部房间，推送房间开始的消息到俱乐部服务
	if room.IsClub() {
		CPool.appendMessage(ClubG2CStartRoomPush(room.ClubId, room.RoomId, room.Round))
	}
}

// 完成所有牌局，结束
func (room *Room) finish() {
	// 发送结算信息
	number := room.Number
	if room.IsCreate() || room.IsClub() {
		number = fmt.Sprintf("%v", room.RoomId)
	}

	resultPushPacket := GameResultPush(room.GetScoreInfoList(), room.GetIndexUserId(0), room.DismissTime, number, room.CType)
	room.SendMessageToRoomUser(resultPushPacket, 0)

	// 给观察者发送结算信息
	room.Ob.sendMessage(resultPushPacket, 0)

	// 保存牌局结算信息
	if room.IsCreate() || room.IsClub() || room.IsRank() {
		go func() {
			defer util.RecoverPanic()
			// 保存用户的最后牌局结果的查看标志
			userService.SaveRoomResultUnread(room.getUserIds(), room.RoomId)
			roomService.SaveResult(room.RoomId, resultPushPacket.GetBody())
		}()
	}

	// 排位赛
	if room.IsRank() {
	}

	// 设置用户随机房间的首日付费
	if room.IsRandom() {
		room.Index.Range(func(k, v interface{}) bool {
			userId := v.(int)
			if !configService.IsRobot(userId) && !userService.IsRandomPayed(userId) {
				userService.SetRandomPayed(userId)
			}
			return true
		})
	}

	// 解散房间
	if room.DismissTime > 0 {
		dismissRoom(room, config.DISMISS_ROOM_CODE_APPLY)
	} else {
		dismissRoom(room, config.DISMISS_ROOM_CODE_FINISH)
	}

	// 更改房间状态
	room.Status = config.ROOM_STATUS_COMPLETED

	// 结算积分
	room.settlementScore()

	// 记录俱乐部房间日志
	if room.IsClub() && len(room.Record) > 0 {
		logService.LogClubRoom(room.RoomId, room.CreateTime, room.ClubId)
		go func() {
			defer util.RecoverPanic()
			// 记录俱乐部房间的成长值
			roomService.StatGrowth(room.ClubCapitorUserId)
			core.Logger.Debug("[StatGrowth]roomId:%v, clubId:%v, manager:%v", room.RoomId, room.ClubId, room.ClubCapitorUserId)
		}()
	}
	// 记录金币场的一起游戏记录
	if room.IsCoin() || room.IsRank() {
		room.logTogetherGame()
	}
	// 删除本服的比赛房间信息
	// 在dismissroom中已经进行了删除
	// if room.IsLeague() {
	// 	model.GetRaceRooms(room.RaceInfo.Id).Del(room.RoomId)
	// }

	core.Logger.Info("[room.finish]roomId:%d, number:%s", room.RoomId, room.Number)
}

// 获取房间哪些用户需要付费
// 机器人无需付费
func (room *Room) getChargeUsers() []int {
	users := []int{}
	if room.IsCreate() || room.IsClub() {
		users = append(users, room.Creator)
	} else if room.IsRandom() {
		// 随机房间，每天只收一次费
		room.Index.Range(func(k, v interface{}) bool {
			userId := v.(int)
			if !configService.IsRobot(userId) && !userService.IsRandomPayed(userId) {
				users = append(users, userId)
			}
			return true
		})
	} else {
		room.Index.Range(func(k, v interface{}) bool {
			userId := v.(int)
			if !configService.IsRobot(userId) {
				users = append(users, userId)
			}
			return true
		})
	}
	return users
}

// 收取房费(用户付费)
func (room *Room) charge() bool {
	if room.PayPrice <= 0 {
		// 无需收费
		core.Logger.Info("[charge]房间无需收费, roomId:%v", room.RoomId)
		return true
	}

	if room.IsClub() && room.PayType == config.ROOM_PAY_TYPE_CLUB {
		return room.chargeByClub(room.PayPrice)
	} else if room.IsCoin() {
		return room.chargeByCoin(room.PayPrice)
	} else if room.IsRank() && room.PayPrice > 0 {
		return room.chargeByRankCard()
	}
	return room.chargeByUser(room.PayPrice)
}

// 收取房费（用户付费）
func (room *Room) chargeByUser(price int) bool {
	// 开启事务
	orm := core.GetWriter()
	orm.Begin()

	// 需扣费用户列表
	users := room.getChargeUsers()
	// 是否所有人扣费成功
	updateMoneySuccess := true
	for _, userId := range users {
		dbUser := userService.GetUser(userId)
		if err := userService.UpdateMoney(orm, dbUser, price*-1, config.MONEY_CHANGE_TYPE_XF, room.consumeType()); err != nil {
			updateMoneySuccess = false
			break
		}
	}

	if updateMoneySuccess {
		// 扣费成功
		// 提交写操作
		orm.Commit()

		// 因为用户每次切回大厅都会重新获取钻石，所以不用发消息通知了
		for _, userId := range users {
			// 更新内存中的钻石余额
			ru := room.GetUser(userId)
			ru.Info.Money -= price
			core.Logger.Info("[chargeByUser]success, roomId:%v, userId:%v", room.RoomId, userId)
		}
	} else {
		// 扣费失败
		// 回滚数据
		orm.Rollback()
		core.Logger.Error("[chargeByUser]failure, roomId:%v", room.RoomId)
	}
	return updateMoneySuccess
}

// 收取房费(俱乐部付费)
func (room *Room) chargeByClub(price int) bool {
	err := clubService.UpdateFund(room.RoomId, room.ClubId, price*-1, config.MONEY_CHANGE_TYPE_XF, room.CreateTime)
	if err != nil {
		return false
	}
	return true
}

// 收取房费(金币场)
func (room *Room) chargeByCoin(price int) bool {
	// 需扣费用户列表
	room.Users.Range(func(k, v interface{}) bool {
		ru := v.(*RoomUser)
		if !configService.IsRobot(ru.UserId) {
			coin := ru.Info.ScoreCoin - price
			if coin < 0 {
				coin = 0
			}
			if err := userService.UpdateScoreCoin(ru.UserId, coin); err != nil {
				core.Logger.Warn("[chargeByCoin]付费失败,userId:%v,price:%v", ru.UserId, price)
			} else {
				// 记录消耗日志
				ru.Info.ScoreCoin = coin
				logService.LogCoinConsumeLog(room.RoomId, ru.UserId, room.CoinType, price, room.CreateTime)
				core.Logger.Info("[chargeByCoin]success,userId:%v,price:%v,剩余:%v", ru.UserId, price, coin)
			}
		}
		return true
	})
	return true
}

// 收取房费(排位赛)
// 房费由用户的排位赛等级来收取
func (r *Room) chargeByRankCard() bool {
	success := true
	r.Users.Range(func(k, v interface{}) bool {
		ru := v.(*RoomUser)
		if configService.IsRobot(ru.UserId) {
			return true
		}
		if ru.Info.RankConsume <= 0 {
			core.Logger.Warn("[chargeByRankCard]无需付费,userId:%v, roomId:%v, price:%v", ru.UserId, r.RoomId, ru.Info.RankConsume)
			return true
		}
		err := userService.UpdateRankCards(ru.UserId, -1*ru.Info.RankConsume)
		if err != nil {
			core.Logger.Warn("[chargeByRankCard]付费失败, userId:%v, roomId:%v, price:%v, err:%v", ru.UserId, r.RoomId, ru.Info.RankConsume, err.Error())
		} else {
			// 记录消耗日志
			gradeId, gradeLevel, star := rank.ExplainSLevel(ru.Info.ScoreRank)
			logService.LogRankConsumeInfo(r.SeasonId, r.RoomId, ru.UserId, ru.Info.RankConsume, gradeId, gradeLevel, star, r.CreateTime)
			core.Logger.Info("[chargeByRankCard]userId:%v, roomId:%v, price:%v", ru.UserId, r.RoomId, ru.Info.RankConsume)
		}
		return true
	})

	return success
}

// 退还房费
func (room *Room) returnCost() {
	if room.PayPrice <= 0 {
		core.Logger.Debug("[returnCost]房间不收费, 无需付费:%v", room.RoomId)
		return
	}
	if room.IsClub() && room.PayType == config.ROOM_PAY_TYPE_CLUB {
		room.returnCostToClub(room.PayPrice)
	} else {
		room.returnCostToUser(room.PayPrice)
	}
}

// 退还房费给用户
func (room *Room) returnCostToUser(price int) {
	// 读取哪些用户需要退费
	users := room.getChargeUsers()
	for _, userId := range users {
		dbUser := userService.GetUser(userId)
		if err := userService.UpdateMoney(core.GetWriter(), dbUser, price, config.MONEY_CHANGE_TYPE_TF, 0); err != nil {
			core.Logger.Error("给用户退费失败，userId: %d, amount: %d.", userId, price)
		} else {
			SendMessageByUserId(userId, UpdateMoneyPush(price))
			// 更新内存中的钻石余额
			ru := room.GetUser(userId)
			ru.Info.Money += price
			core.Logger.Info("[returnCostToUser]roomId:%v, userId:%v", room.RoomId, userId)
		}
	}
}

// 退还房费到俱乐部基金
func (room *Room) returnCostToClub(price int) {
	err := clubService.UpdateFund(room.RoomId, room.ClubId, price, config.MONEY_CHANGE_TYPE_TF, room.CreateTime)
	if err != nil {
		core.Logger.Error("退还俱乐部基金失败,roomId:%v, clubId:%v, amount:%v", room.RoomId, room.ClubId, price)
	}
	core.Logger.Info("[returnCostToClub]roomId:%v, clubId:%v, amount:%v", room.RoomId, room.ClubId, price)
}

// 更新房间内用户的积分
func (room *Room) settlementScore() {
	// 读取惩罚分
	var punishmentScore int
	if room.DismissTime > 0 && room.EnableDismissPunishment() {
		punishmentScore = config.RoomPunishmentScore
	}

	room.ScoreInfo.Range(func(k, v interface{}) bool {
		ffInfo := v.(*FrontFinalInfo)
		// 输赢分
		score := ffInfo.Score
		if _, exists := room.DismissOp.Load(ffInfo.UserId); !exists {
			score -= punishmentScore
		}
		room.updateUserScore(ffInfo.UserId, score, ffInfo.FromTotal, ffInfo.Total, ffInfo.FromSLevel, ffInfo.FinalSLevel, ffInfo.WinningStreak)
		return true
	})
}

// 更新房间用户的积分
// 包括数据库中和内存中的值
func (room *Room) updateUserScore(userId, score, fromTotal, total, fromSLevel, finalSLevel, winningStreak int) {
	core.Logger.Debug("[updateUserScore]userId:%v, score:%v, fromTotal:%v, total:%v", userId, score, fromTotal, total)
	// 跳过用户错误
	ru := room.GetUser(userId)
	if ru == nil {
		core.Logger.Error("updateUserScore, user not found, userId:%v, room.users:%#v", userId, room.GetUsers())
	}
	if room.IsRandom() && score != 0 {
		userService.UpdateScoreRandom(userId, score)
		ru.Info.ScoreRandom += score
	} else if (room.IsCreate() || room.IsTV()) && score != 0 {
		userService.UpdateScore(userId, score)
		ru.Info.Score += score
	} else if room.IsMatch() {
		if score != 0 {
			userService.UpdateScoreMatch(userId, score)
			ru.Info.ScoreMatch += score
		}
		// 更新比赛日积分记录
		// 机器人不记录
		if !configService.IsRobot(userId) {
			matchModel.UpdateDailyScore(userId, score)
		}
	} else if room.IsClubMatch() && score != 0 {
		// 更新俱乐部淘汰赛积分
		ru.Info.ScoreClub += score
		clubService.UpdateUserScore(room.ClubId, userId, score)
	} else if room.IsCoin() && score != 0 {
		ru.Info.ScoreCoin += score
		// 金币有上下限
		if ru.Info.ScoreCoin < 0 {
			ru.Info.ScoreCoin = 0
		}
		if ru.Info.ScoreCoin > config.COIN_UPPER_LIMIT {
			ru.Info.ScoreCoin = config.COIN_UPPER_LIMIT
		}
		if !configService.IsRobot(userId) {
			userService.UpdateScoreCoin(userId, ru.Info.ScoreCoin)
			// 更新省排行
			coinService.UpdateProvinceRank(userId, ru.Info.ScoreCoin)
			// 更新市排行
			if ru.Info.RankCity > 0 {
				coinService.UpdateCityRank(userId, ru.Info.RankCity, ru.Info.ScoreCoin)
			}
			// 更新好友排行
			// 更新我在好友中的排行
			coinService.UpdateFriendRank(userId, ru.Info.ScoreCoin)

			// 最后更新我在好友中的排行
			for _, friendUserId := range friendService.GetFriends(userId) {
				coinService.UpdateUserFriendRank(userId, friendUserId, ru.Info.ScoreCoin)
			}
		}
	} else if room.IsRank() {
		// 更新真实用户的游戏次数游戏次数, 连胜次数
		if !configService.IsRobot(userId) {
			rankService.UpdateSeasonUserTimes(room.SeasonId, userId, fromSLevel, finalSLevel, total)
			if score > 0 {
				rankService.SetSeasonUserWinningStreak(room.SeasonId, userId, winningStreak)
				// 通知发放连胜奖励
				rankService.NotifyWinningStreakRewards(userId, winningStreak)
			} else {
				rankService.DelSeasonUserWinningStreak(room.SeasonId, userId)
			}
		}
		// 用户排位等级有更新
		ru.Info.ScoreRank = finalSLevel
		ru.Info.RankExp = total
		if fromSLevel != finalSLevel {
			if !configService.IsRobot(userId) {
				// 异步处理用户排位改变带来的其他改变，如排行人数等
				go obUserSLevelModify(room.SeasonId, userId, ru.Info.RankCity, fromTotal, total, fromSLevel, finalSLevel)
			}
		}
	}
}

// 房间数据落地, 每一局都记录
// 更新数据库中的用户积分
// 更新内存中的用户总积分
// 记录用户日志
// 存储房间结果
func (room *Room) recordData() {
	// 用户赢的次数记录
	winTimes := make([]int, room.setting.GetSettingPlayerCnt())
	// 用户输的次数记录
	loseTimes := make([]int, room.setting.GetSettingPlayerCnt())
	// 用户总积分
	userScores := make([]map[string]int, 0, room.setting.GetSettingPlayerCnt())

	// 存储游戏结果

	// 存储日志
	// 记录当局日志
	for _, mi := range room.Record {
		room.Index.Range(func(k, v interface{}) bool {
			roomUserId := v.(int)
			roomUserScore := mi.getRoundScore(roomUserId)
			// 记录用户输赢次数
			index := room.GetUser(roomUserId).Index
			if roomUserScore >= 0 {
				winTimes[index] += 1
			} else {
				loseTimes[index] += 1
			}
			return true
		})
	}

	// 记录用户输赢记录
	for userId, _ := range room.MI.getUsers() {
		//  用户输赢总分
		score := room.GetScoreInfo(userId).Score
		userScores = append(userScores, map[string]int{"userId": userId, "score": score})
		// 用户赢了多少局
		index := room.GetUser(userId).Index
		wins := winTimes[index]
		// 用户输了多少局
		loses := loseTimes[index]

		// 记录用户日志
		logService.LogGameUserRecords(userId, score, wins, loses, room.RoomId, room.CreateTime)
	}
	// 存储房间结果
	logService.LogGameResult(room.RoomId, room.Number, room.CType, room.MType, room.TRound, room.userIdJoin(),
		userScores, 0, nil, room.CreateTime, room.StartTime, room.Round)

	// 记录游戏次数
	if room.IsCreate() || room.IsClub() {
		room.Index.Range(func(k, v interface{}) bool {
			userId := v.(int)
			// 机器人不记录
			if !configService.IsRobot(userId) {
				logService.LogUserGameRoundTimes(userId, room.MType, room.CType)
			}
			return true
		})
	}
}

// 给队伍用户发送消息，跳过某用户，跳过不在线用户
func (room *Room) SendMessageToRoomUser(imPacket *protocal.ImPacket, userId int) {
	room.Index.Range(func(k, v interface{}) bool {
		tUserId := v.(int)
		if tUserId != userId { // 跳过不发消息的用户
			SendMessageByUserId(tUserId, imPacket)
		}
		return true
	})
}

// 给队伍不在线用户发送push
func (room *Room) SendPushToOfflineUser(langId, excludeUserId int) {
	var senderUserId int
	var senderNickName string

	if excludeUserId > 0 {
		senderUserId = excludeUserId
		if ru := room.GetUser(senderUserId); ru != nil {
			senderNickName = ru.Info.Nickname
		} else {
			core.Logger.Error("非房间的用户给其他人发送push, userId:%v", excludeUserId)
			return
		}
	}
	room.Users.Range(func(k, v interface{}) bool {
		ru := v.(*RoomUser)
		// 跳过不发的用户
		if ru.UserId != excludeUserId {
			ru.SendPush(langId, senderUserId, senderNickName, "")
		}
		return true
	})
}

// 检查作弊
// 判断房间其他人的作弊关系
func (room *Room) checkCheat() {
	// 非自建房间，不判断
	if !room.IsCreate() && !room.IsClub() {
		return
	}
	// 2人局无需判断
	if room.GetUsersLen() < 3 {
		return
	}

	// 需要计算位置的用户列表
	pUsers := []int{}
	// 统计出需要计算距离的用户和未开放位置的用户
	rusers := room.GetUsers()
	for _, ru := range rusers {
		if (ru.Latitude == float64(-1) && ru.Longitude == float64(-1)) ||
			(ru.Latitude == float64(0) && ru.Longitude == float64(0)) {
			room.NoPositionUsers = append(room.NoPositionUsers, ru.UserId)
		} else {
			pUsers = append(pUsers, ru.UserId)
		}
	}

	// 计算所有用户之间的距离
	if len(pUsers) >= 2 {
		sort.Ints(pUsers)
		for i := 0; i < len(pUsers)-1; i++ {
			userI := rusers[pUsers[i]]
			for j := i + 1; j < len(pUsers); j++ {
				userJ := rusers[pUsers[j]]
				// 计算两个用户的距离
				distance := int(math.Ceil(util.EarthDistance(userI.Latitude, userI.Longitude, userJ.Latitude, userJ.Longitude)))
				room.UserDistances = append(room.UserDistances, &UserDistance{pUsers[i], pUsers[j], distance})

				// 记录进距离过近的列表
				if distance < config.CHEAT_DISTANCE_LIMIT {
					if !util.IntInSlice(pUsers[i], room.NearUsers) {
						room.NearUsers = append(room.NearUsers, pUsers[i])
					}
					if !util.IntInSlice(pUsers[j], room.NearUsers) {
						room.NearUsers = append(room.NearUsers, pUsers[j])
					}
				}
			}
		}
	}

	// 推送消息
	for _, ru := range rusers {
		// 从过近和未开放位置的切片中屏蔽用户自己
		nearUsers := util.SliceDel(room.NearUsers, ru.UserId)
		if len(nearUsers) == 1 {
			nearUsers = []int{}
		}
		noPositionUsers := util.SliceDel(room.NoPositionUsers, ru.UserId)

		core.Logger.Debug("[checkCheat]roomId:%v, round:%v, userId:%v, nearUsers:%v, noPositionUsers:%v", room.RoomId, room.Round, ru.UserId, nearUsers, noPositionUsers)
		SendMessageByUserId(ru.UserId, GameAntiCheatingPush(nearUsers, noPositionUsers))
	}
	core.Logger.Info("[checkCheat]roomId:%v, nearUsers:%v, noPositionUsers:%v", room.NearUsers, room.NoPositionUsers)
}

// 解散房间
func dismissRoom(room *Room, code int) {
	// 给用户发推送
	switch code {
	case config.DISMISS_ROOM_CODE_FINISH: // 房间正常结束,不发
	case config.DISMISS_ROOM_CODE_OB_QUIT: // 观察员退出
	case config.DISMISS_ROOM_CODE_HOST_LEAVE: // 房主退出，给其他人
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_DISMISS, room.GetIndexUserId(0))
	default:
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_DISMISS, 0)
	}

	// 解散的房间也要保存回放
	if room.Round > len(room.Record) {
		room.MI.savePlaybackIntact()
	}

	// 删除进程中的房间
	RoomMap.DelRoom(room.RoomId)

	// 清除cache中的房间信息
	roomService.CleanRoom(room.RoomId, room.Number)

	// 将房间id从大厅房间的cache中删掉
	hallService.DelHallRoom(GetRemoteAddr(), room.RoomId)

	// 构建解散房间的消息
	responsePacket := CloseRoomPush(code, core.GetLang(code+config.DISMISS_ROOM_CODE_OFFSET))

	// 推送消息给用户&清除用户房间数据
	users := room.getUserIds()
	if room.IsTV() && len(room.Ob.users) > 0 {
		users = append(users, room.Ob.users...)
	}

	for _, oUserId := range users {
		// 清除用户已加入的房间
		userService.DelRoomId(oUserId, room.RoomId)
		// 清除用户的分享回放标志
		userService.DelSharePlayFlag(oUserId)

		// 跳过不在线的用户
		if otherUser, _ := UserMap.GetUser(oUserId); otherUser != nil {
			// 清除内存中的房间id
			// 这里做一下房间id判断，如果房间id，不是用户当前所在的房间，则不做处理
			if otherUser.RoomId == room.RoomId {
				otherUser.RoomId = int64(0)

				// 发送消息
				otherUser.AppendMessage(responsePacket)
				core.Logger.Debug("通知用户解散房间, roomId:%v, userId:%v", room.RoomId, oUserId)
			}
		}
	}

	go func() {
		defer util.RecoverPanic()

		// 更新金币场实时在线人数
		if room.IsCoin() {
			coinService.DecrCoinUserCnt(room.MType, room.CoinType, len(users))
		}

		// 向俱乐部推送房间解散
		if room.IsClub() {
			CPool.appendMessage(ClubG2CDismissRoomPush(room.ClubId, room.RoomId, code))
		}

		if room.IsLeague() {
			leagueGameFinish(room, code)
		}

		// 记录房间解散标志
		logService.LogGameDismiss(room.RoomId)
	}()

	core.Logger.Info("[dismissRoom]roomId:%d, number:%s, code:%d", room.RoomId, room.Number, code)
}

// 回应准���
func (room *Room) userOperationReady(userId int) (bool, *core.Error) {
	if !room.IsReadying() || !room.IsFull() {
		// 未处于准备回应状态
		return false, core.NewError(-510)
	}

	if room.IsTV() && room.Ob.hasUser(userId) {
		// 判断是否已经准备过了
		if room.Ob.isReady(userId) {
			return false, core.NewError(-511)
		}
		// 加入准备队列
		room.Ob.readyUsers = append(room.Ob.readyUsers, userId)
	} else {
		// 判断是否已经准备过了
		if util.IntInSlice(userId, room.ReadyList) {
			return false, core.NewError(-511)
		}
		// 加入准备队列
		room.ReadyList = append(room.ReadyList, userId)

		// 通知游戏用户，有人准备好了
		pushPacket := GameReadyPush(userId)
		room.SendMessageToRoomUser(pushPacket, 0)

		// 给观察员发消息
		room.Ob.sendMessage(pushPacket, 0)
	}

	core.Logger.Debug("[userOperationReady]userId:%d, roomId:%d, number:%s", userId, room.RoomId, room.Number)

	// 电视端必须要有观察者，且观察者全部准备才能开始
	if room.IsTV() && (len(room.Ob.users) == 0 || !room.Ob.isAllReady()) {
		core.Logger.Debug("等候观察者准备, roomId:%v, number:%v", room.RoomId, room.Number)
		return false, nil
	}
	// 是否有人尚未准备
	if len(room.ReadyList) < room.GetUsersLen() {
		return false, nil
	}

	// 准备成功、人数已满
	room.SetReady()

	return true, nil
}

func RemoveRoomUser(room *Room, userId int, dismisCode int, quitCode int) (bool, bool) {
	// 防止并发，当用户不在房间的时候，就不再退出了
	if _, ok := room.Users.Load(userId); !ok {
		core.Logger.Warning("[RemoveRoomUser]失败, 用户已不在房间中,roomId:%v, userId:%v", room.RoomId, userId)
		return false, false
	}
	// 是否需要解散房间
	needDismiss := false
	if room.IsClubMode() {
		// 馆主创建的房间不会因为退出解散
	} else if room.IsCreate() || room.IsTV() || room.IsClub() {
		// 自主创建的房间，房主退出，意味着解散房间
		if userId == room.GetIndexUserId(0) {
			needDismiss = true
		}
	} else {
		// 非自主创建的房间
		// 所有人退出后，才解散房间
		// 或者剩下的人全是机器人，解散房间
		if room.GetUsersLen() == 1 || room.getTruePlayerCount() == 1 {
			needDismiss = true
		}
	}

	if needDismiss {
		core.Logger.Info("[RemoveRoomUser]房主或者所有人退出，解散房间, roomId:%d, userId:%v", room.RoomId, userId)
		// 解散房间
		dismissRoom(room, dismisCode)
	} else {
		// 更新金币场实时在线人数
		if room.IsCoin() {
			coinService.DecrCoinUserCnt(room.MType, room.CoinType, 1)
		}
		// 通知用户退出房间
		room.SendPushToOfflineUser(config.PUSH_CHAT_ID_ROOM_QUIT, userId)

		// 成员退出的push
		userIndex := room.GetUser(userId).Index
		pushPacket := QuitRoomPush(userId, userIndex, quitCode)
		room.SendMessageToRoomUser(pushPacket, 0)
		// 发送成员退出消息给观察员
		room.Ob.sendMessage(pushPacket, 0)
		// 删除用户索引
		room.Index.Delete(userIndex)

		// 比赛和随机的房间，在满员之后, 若房间有人退出，需要将房间放回队列，等候其他人加入
		if room.IsFull() {
			if room.IsRandom() {
				RandomRoomQueueMap.Get(room.MType).Add(room.RoomId, room.Number)
			} else if room.IsMatch() {
				MatchRoomQueueMap.Get(room.MType).Add(room.RoomId, room.Number)
			} else if room.IsCoin() {
				CoinRoomQueueMap.Get(room.CoinType, room.MType).Rooms.Store(room.RoomId, room.Number)
			} else if room.IsRank() {
				RankRoomQueueMap.Get(room.GradeId).Rooms.Store(room.RoomId, room.Number)
			}
			// 重新安排机器人进入
			if room.EnableRobot() {
				robotGameInfo := robot.NewGameInfo(GetRemoteAddr(), room.RoomId, room.MType, room.CType, room.setting.GetSettingPlayerCnt(), room.TRound)
				robotGameInfo.CoinType = room.CoinType
				robotGameInfo.GradeId = room.GradeId
				hallService.AddHallRobotRoom(GetRemoteAddr(), robotGameInfo.String())
			}
		}

		// 将用户退出
		room.Users.Delete(userId)
		userService.DelRoomId(userId, room.RoomId)
		// 清除用户的分享回放标志
		userService.DelSharePlayFlag(userId)

		core.Logger.Debug("[room.RemoveRoomUser]roomId:%v, userId:%v, user index:%v, roomIndex:%v", room.RoomId, userId, userIndex, room.IndexToString())
	}

	return needDismiss, true
}

// HasTogetherGameLog 判断房间内的用户是否与给定的用户，今天有过游戏记录
func (room *Room) HasTogetherGameLog(speciedUserId int) bool {
	hasFlag := false
	room.Index.Range(func(k, v interface{}) bool {
		roomUserId := v.(int)
		if coinService.HasTogetherGameLog(speciedUserId, roomUserId) {
			hasFlag = true
			return false
		}
		return true
	})
	return hasFlag
}

// 房间用户重连（完整）
func (r *Room) restoreIntact(u *User) {
	r.Mux.Lock()
	defer r.Mux.Unlock()

	if u.MQ.WasStarted() {
		u.MQ.Pause()
	}
	u.MQ.Send(GameRestorePush(u.UserId, r))
	// 如果用户不支持restore done，在推送玩restore push后，就直接打开
	if !u.EnableRestoreDone() && u.MQ.WasPaused() {
		u.MQ.Continue()
	}

	core.Logger.Debug("[restoreIntact]roomId:%v, round:%v, userId:%v", r.RoomId, r.Round, u.UserId)
}

// 房间用户重连(步骤)
func (r *Room) restoreSection(u *User, seq int) {
	r.Mux.Lock()
	defer r.Mux.Unlock()

	// 读取待发送的消息列表
	if u.MQ.WasStarted() {
		u.MQ.Pause()
	}
	mu := r.MI.getUser(u.UserId)
	seqOperationList := mu.MSC.GetList(seq)
	u.MQ.Send(GameRestoreSectionPush(u.UserId, r, seqOperationList))
	// 如果用户不支持restore done，在推送玩restore push后，就直接打开
	if !u.EnableRestoreDone() && u.MQ.WasPaused() {
		u.MQ.Continue()
	}

	core.Logger.Debug("[restoreSection]roomId:%v, round:%v, userId:%v, from seq:%v, to seq:%v", r.RoomId, r.Round, u.UserId, seq, mu.MSC.GetSeq())
}

// 判断用户状态，是否允许进行片断重连
func (r *Room) canRestoreSection(u *User, roomId int64, round int, seq int) bool {
	core.Logger.Debug("[canRestoreSection]userId:%v, roomId:%v, round:%v, seq:%v", u.UserId, roomId, round, seq)
	if r.MI == nil {
		return false
	}
	// 已经在不同房间
	if r.RoomId != roomId {
		return false
	}
	// 不同局
	if r.Round != round {
		return false
	}
	// 未开始
	if seq == 0 {
		return false
	}
	// 准备中
	if r.IsReadying() {
		return false
	}
	return true
}

// 获取在线用户列表
func (r *Room) GetOnlineUsers() []int {
	onlineUsers := []int{}
	r.Index.Range(func(k, v interface{}) bool {
		userId := v.(int)
		if UserMap.IsUserExists(userId) || configService.IsRobot(userId) {
			onlineUsers = append(onlineUsers, userId)
		}
		return true
	})
	return onlineUsers
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 日志相关
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 房间信息
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func (this *Room) logGameInfo() {
	roomMaster, _ := this.Index.Load(0)

	roomInfo := new(config.GameInfo)
	roomInfo.RoomId = this.RoomId
	roomInfo.RoomNum = this.Number
	roomInfo.CreatorUserId = this.Creator
	roomInfo.GameType = this.CType
	roomInfo.MahjongType = this.MType
	roomInfo.Setting, _ = util.InterfaceToJsonString(this.setting.GetSetting())
	roomInfo.TotalRounds = this.TRound
	roomInfo.PlayerCount = this.setting.GetSettingPlayerCnt()
	roomInfo.Players = this.userIdJoin()
	roomInfo.CreateTime = this.CreateTime
	roomInfo.StartTime = util.GetTime()
	roomInfo.ServerRemote = GetRemoteAddr()
	roomInfo.RoomMaster = roomMaster.(int)
	logService.LogGameInfo(roomInfo)
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 记录房间内用户的一起玩过游戏的记录
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func (room *Room) logTogetherGame() {
	go func() {
		defer util.RecoverPanic()
		for i := 0; i < room.setting.GetSettingPlayerCnt()-1; i++ {
			for j := i + 1; j < room.setting.GetSettingPlayerCnt(); j++ {
				if room.IsClub() {
					coinService.SetTogetherGameLog(room.GetIndexUserId(i), room.GetIndexUserId(j), room.CreateTime)
				} else if room.IsRank() {
					rankService.AddTogetherGameTimes(room.GetIndexUserId(i), room.GetIndexUserId(j), room.CreateTime)
				}
			}
		}
	}()
}
