package game

import (
	"encoding/json"
	"net"
	"strings"
	"sync"
	"time"

	"fmt"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/library/mq"
	"mahjong.go/mi/protocal"
	"mahjong.go/model/ob"

	hallService "mahjong.go/service/hall"
	roomService "mahjong.go/service/room"
	userService "mahjong.go/service/user"
)

// 用户的扩展信息，由大厅用户和房间用户共用
type UserInfo struct {
	Score             int     // 历史积分
	ScoreRandom       int     // 随机组局总积分
	ScoreMatch        int     // 比赛总积分
	ScoreClub         int     // 俱乐部积分，随当前所在的俱乐部房间变化
	ScoreCoin         int     // 累积金币数
	ScoreLeague       int     // 比赛积分
	ScoreRank         int     // 排位赛等级
	RankExp           int     // 排位赛经验值
	Gender            int     // 用户性别，0：未定义，1：男；2：女
	Money             int     // 用户钻石
	Avatar            string  // 用户头像
	AvatarBox         int     // 头像框
	Nickname          string  // 用户昵称
	Ip                string  // 用户ip地址
	Area              string  // 用户所属地区
	Longitude         float64 // 经度
	Latitude          float64 // 纬度
	Device            string  // 设备:android、ios
	DeviceToken       string  // 设备号
	LastHeartBeatTime int64   // 最后心跳时间
	Version           string  // 用户客户端版本
	RankConsume       int     // 排位赛消耗
	RankCity          int     // 城市(排位赛、金币赛可用)
	MemberLevel       int     // 会员等级
	MemberAddExp      int     // 会员等级经验提升比例
}

// 大厅用户信息
type User struct {
	Mux           *sync.Mutex  // 用户锁
	UserId        int          // 用户id
	ConnectTime   int64        // 连接时间
	Conn          *net.TCPConn // 活动连接
	RoomId        int64        // 房间id，默认为0
	Info          *UserInfo    // 用户扩展信息
	messageStatus bool         // 消息队列的状态
	KickOnce      *sync.Once

	// 握手时，客户端给这几个参数，服务器根据这三个参数，来判断重连数据类型
	handshakeRoomId int64
	handshakeRound  int
	handshakeSeq    int

	MQ *mq.MsgQueue // 消息队列

	// 是否不监听心跳
	NoHeartbeat int
	// 用户来源
	From string
}

// UserList 大厅用户列表
type UserList struct {
	Users *sync.Map
	Mux   *sync.Mutex
}

// 创建一个新用户
func NewUser(userId int, conn *net.TCPConn) *User {
	user := &User{}
	user.UserId = userId
	user.Conn = conn
	// user.Mq = make(chan *protocal.ImPacket, 1024)
	user.MQ = mq.NewMsgQueue(userId, conn)
	user.Mux = &sync.Mutex{}
	user.messageStatus = false
	user.KickOnce = &sync.Once{}
	return user
}

// NewUserList 创建一个新的大厅用户列表
func NewUserList() *UserList {
	userMap := &UserList{}
	userMap.Users = &sync.Map{}
	userMap.Mux = &sync.Mutex{}

	return userMap
}

// GetUser 从大厅列表中获取用户信息
func (list *UserList) GetUser(userId int) (*User, *core.Error) {
	list.Mux.Lock()
	defer list.Mux.Unlock()

	if user, ok := list.Users.Load(userId); ok {
		return user.(*User), nil
	}
	return nil, core.NewError(-201, userId)
}

// SetUser 添加用户到大厅列表
func (list *UserList) SetUser(user *User) {
	list.Mux.Lock()
	defer list.Mux.Unlock()

	list.Users.Store(user.UserId, user)
}

// DelUser 从map中移除用户信息
func (list *UserList) DelUser(userId int) *core.Error {
	list.Mux.Lock()
	defer list.Mux.Unlock()

	if _, ok := list.Users.Load(userId); !ok {
		return core.NewError(-201, userId)
	}
	list.Users.Delete(userId)
	return nil
}

// IsUserExists 判断用户是否存在
func (list *UserList) IsUserExists(userId int) bool {
	list.Mux.Lock()
	defer list.Mux.Unlock()

	_, ok := list.Users.Load(userId)
	return ok
}

// Len 返回用户数量
func (list *UserList) Len() int {
	return util.SMapLen(list.Users)
}

// SendMessageByUserId 给某个在线userId发送消息
func SendMessageByUserId(userId int, imPacket *protocal.ImPacket) {
	user, err := UserMap.GetUser(userId)
	if err == nil {
		user.AppendMessage(imPacket)
	} else {
		core.Logger.Tracef("[SendMessageByUserId]跳过不在线的用户:%d", userId)
	}
}

// AppendMessage 往用户的消息队列追加一条消息
func (user *User) AppendMessage(imPacket *protocal.ImPacket) {
	// 需要判断用户是否已经准备好接收消息
	// 服务端在收到ack到构建gameRestore消息这段时间内，不能给用户发消息，否则会导致消息被重复处理
	if user.messageStatus {
		user.Mux.Lock()
		defer user.Mux.Unlock()
		user.MQ.Append(imPacket)
	}
}

// 用户重复登录，需将用户踢下线
func RepeatLogin(loginedUser *User, newConn *net.TCPConn) {
	core.Logger.Warn("[RepeatLogin]userId:%d, new remote:%s", loginedUser.UserId, newConn.RemoteAddr().String())

	// 发送踢下线的协议
	impacket := protocal.NewImPacket(protocal.PACKAGE_TYPE_KICK, nil)
	loginedUser.MQ.Send(impacket)

	// 踢下线
	Kick(loginedUser.UserId)
}

// LoadCurrentRemoteRoomId 获取用户的房间id，仅在本服务器上
// 只有在本服务器的房间，才会返回
func LoadCurrentRemoteRoomId(userId int) int64 {
	roomId := userService.GetRoomId(userId)
	if roomId > 0 {
		if _, err := RoomMap.GetRoom(roomId); err != nil {
			roomId = int64(0)
			core.Logger.Warn("[LoadRoomId]room was dismissed, userId:%d, roomId: %d.", userId, roomId)
		}
	}
	return roomId
}

// LoadRoomId 获取用户的房间id
// 如果用户的房间已经不存在了，则清除用户当前所处的房间id
func LoadRoomId(userId int) int64 {
	roomId := userService.GetRoomId(userId)
	if roomId > 0 {
		// 读取room remote
		if roomRemote := roomService.GetRoomRemote(roomId); roomRemote != "" {
			// 如果游戏服活跃且房间存在，则认为房间真的存在
			if hallService.IsRemoteActive(roomRemote) && hallService.IsRoomExists(roomRemote, roomId) {
				return roomId
			}
		}
		core.Logger.Warn("[LoadRoomId]room was dismissed, userId:%d, roomId: %d.", userId, roomId)
		roomId = int64(0)
		userService.DelRoomId(userId, 0)
	}
	return roomId
}

// 心跳监测
func (this *User) ListenHeartBeat() {
	// 捕获异常
	defer util.RecoverPanic()

	for {
		time.Sleep(3 * config.HEART_BEAT_SECOND * time.Second)

		user, err := UserMap.GetUser(this.UserId)
		if err != nil {
			// 用户已下线
			core.Logger.Debug("用户已下线，停止心跳监测, userId:%d", this.UserId)
			break
		}
		if user.ConnectTime != this.ConnectTime {
			// 用户已经被顶号或者重新登录
			core.Logger.Debug("用户已重新登录，停止心跳监测, userId:%d", this.UserId)
			break
		}
		if util.GetTime()-user.Info.LastHeartBeatTime > int64(2*config.HEART_BEAT_SECOND) {
			core.Logger.Debug("用户心跳停止，踢下线, userId:%d", this.UserId)
			Kick(this.UserId)
			break
		}
	}
}

// 读取用户以及房间信息
func getUserRoom(userId int) (*User, *Room, *core.Error) {
	// 判断用户是否已连接
	user, err := UserMap.GetUser(userId)
	if err != nil {
		return nil, nil, err
	}

	// 判断用户是否在roomId
	if user.RoomId == 0 {
		return user, nil, core.NewError(-203, userId, user.RoomId)
	}

	// 判断房间是否已存在
	room, err := RoomMap.GetRoom(user.RoomId)
	if err != nil {
		return user, nil, err
	}

	return user, room, nil
}

// 记录用户的重连日志
func restoreLog(userId int, roomId int64, number string, round int, logType int) {
	redisConn := core.RedisClient3.Get()
	defer redisConn.Close()

	// 获取当前日期分钟数，小时数
	date := util.GetYMD()
	hour := time.Now().Hour()
	minute := time.Now().Minute() / 10 * 10

	// 个人日志
	userCacheKey := fmt.Sprintf(config.CACHE_KEY_USER_RESOTRE_LIST, userId, date)
	log := make(map[string]interface{})
	log["roomId"] = roomId
	log["number"] = number
	log["round"] = round
	log["type"] = logType
	log["time"] = util.GetTime()
	byt, _ := json.Marshal(log)
	redisConn.Do("lpush", userCacheKey, string(byt))
	redisConn.Do("expire", userCacheKey, config.RESTORE_LOG_EXPIRE_SECOND)

	// 服务器重连日志
	serverCacheKey := fmt.Sprintf(config.CACHE_KEY_HALL_RESOTRE_LIST, date, hour, minute)
	log["userId"] = userId
	byt2, _ := json.Marshal(log)
	redisConn.Do("lpush", serverCacheKey, string(byt2))
	redisConn.Do("expire", serverCacheKey, config.RESTORE_LOG_EXPIRE_SECOND)

	// 用户重连次数
	if logType == config.RESTORE_LOG_TYPE_RECONNECT {
		redisConn.Do("hincrby", config.CACHE_KEY_HALL_RESOTRE_COUNT, fmt.Sprintf("%v:%v:%v", date, hour, minute), 1)
	}
}

// 判断用户是否是观察员
func (u *User) isObservers() bool {
	// 判断ip
	if !IsOBIP(u.Info.Ip) {
		return false
	}

	if !ob.IsObservers(u.UserId) {
		return false
	}

	return true
}

// EnableRestoreDone 用户是否支持enabledone，2.3及以上版本，不再支持restore done
func (u *User) EnableRestoreDone() bool {
	return u.Info.Version != "latest" && strings.Compare(u.Info.Version, "2.3.0") == -1
}
