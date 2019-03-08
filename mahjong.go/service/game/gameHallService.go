package game

import (
	"net"
	"sync"
	"time"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	hallService "mahjong.go/service/hall"
)

var (
	UserMap            *UserList               // 大厅用户列表
	RoomMap            *RoomList               // 大厅房间列表
	RandomRoomQueueMap *RandomRoomQueueFactory // 随机队列列表
	MatchRoomQueueMap  *MatchRoomQueueFactory  // 比赛队列列表
	CoinRoomQueueMap   *CoinRoomQueueFactory   // 金币场队列列表
	RankRoomQueueMap   *RankRoomQueueFactory   // 排位赛队列列表
	GameRemoteAddr     string                  // 游戏监听端口:ip
	GameVersion        string                  // 服务器版本
	CPool              *ClubPool               // 俱乐部连接池
	LPool              *LeaguePool             // 联赛连接池
	HeartBeatLogFlag   bool                    // 是否开启心跳日志
	RankLock           *sync.RWMutex           // 排位赛锁
	Cre                bool
	Coe                bool
	Rne                bool
)

// 初始化
func init() {
	UserMap = NewUserList()
	RoomMap = NewRoomList()
	RandomRoomQueueMap = NewRandomRoomQueueFactory()
	MatchRoomQueueMap = NewMatchRoomQueueFactory()
	CoinRoomQueueMap = NewCoinRoomQueueFactory()
	RankRoomQueueMap = NewRankRoomQueueFactory()
	CPool = NewClubPool()
	LPool = NewLeaguePool()
	RankLock = &sync.RWMutex{}
	Cre = true
	Coe = true
	Rne = true
}

// StatLastActionTime 记录大厅的最后活动时间,用于监控此条服务是否有效
// 每秒记录一次
func StatLastActionTime() {
	// 捕获异常
	defer util.RecoverPanic()

	for {
		// 设置服务器在线用户数
		hallService.SetRemoteUserCnt(GetRemoteAddr(), UserMap.Len())
		// 设置服务器最后活动时间
		hallService.SetRemoteActionTime(GetRemoteAddr(), util.GetTime())
		time.Sleep(time.Second)
	}
}

// SetGameVersion 记录服务器版本，只在服务器启动时记录
func SetGameVersion(version string) {
	// versionSetting := core.GetAppConfig("version")
	// if versionSetting != nil {
	// 	Version = versionSetting.(string)
	if version != "" {
		GameVersion = version
	} else {
		GameVersion = config.GAME_VERSION_DEFAULT
	}
	// 保存至redis
	hallService.SetRemoteVersion(GameRemoteAddr, GameVersion)
}

// SetRemoteAddr 设置当前服务器的监听地址+端口
func SetRemoteAddr(remote string) {
	GameRemoteAddr = remote
}

// GetRemoteAddr 返回当前服务器的监听地址+端口
func GetRemoteAddr() string {
	return GameRemoteAddr
}

// ListenHandShakeSuccess 监听用户连接之后，有没有及时handshake
// 如果用户执行了连接之后，30秒内都没有handshake成功，则将用户踢下线
func ListenHandShakeSuccess(conn *net.TCPConn, c chan int) {
	// 捕获异常
	defer util.RecoverPanic()

	select {
	case <-c:
		break
	case <-time.After(30 * time.Second):
		core.Logger.Debugf("用户长时间未handshake成功，被踢下线:%s", conn.RemoteAddr())
		conn.Close()
		break
	}
}
