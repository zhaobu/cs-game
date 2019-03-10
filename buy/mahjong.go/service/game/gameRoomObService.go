// 游戏观察者
package game

import (
	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
	"mahjong.go/mi/protocal"

	roomService "mahjong.go/service/room"
	userService "mahjong.go/service/user"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 结构定义
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 操作
type Ob struct {
	users      []int // 用户列表
	readyUsers []int // 已准备用户列表
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 结构操作
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// NewOb 生成一个新的观察结构
func NewOb() *Ob {
	ob := &Ob{[]int{}, []int{}}
	return ob
}

// 添加一个观察者
func (ob *Ob) add(userId int) bool {
	if !util.IntInSlice(userId, ob.users) {
		ob.users = append(ob.users, userId)
		return true
	}
	return false
}

// 移除一个观察者
func (ob *Ob) remove(userId int) bool {
	if util.IntInSlice(userId, ob.users) {
		ob.users = util.SliceDel(ob.users, userId)
		return true
	}
	return false
}

// 清空操作者
func (ob *Ob) clean() bool {
	if len(ob.users) > 0 {
		ob.users = []int{}
		return true
	}
	return false
}

// 有没有一个观察者
func (ob *Ob) hasUser(userId int) bool {
	return util.IntInSlice(userId, ob.users)
}

// 观察者是否已准备
func (ob *Ob) isReady(userId int) bool {
	return util.IntInSlice(userId, ob.readyUsers)
}

// 是否已全部准备
func (ob *Ob) isAllReady() bool {
	return len(ob.users) == len(ob.readyUsers)
}

// 清空准备状态
func (ob *Ob) clearReady() {
	ob.readyUsers = []int{}
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 扩展操作
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 给所有观察者发送消息
func (ob *Ob) sendMessage(imPacket *protocal.ImPacket, excludeUserId int) {
	for _, userId := range ob.users {
		if userId == excludeUserId {
			continue
		}
		SendMessageByUserId(userId, imPacket)
	}
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 普通逻辑
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
// 判断用户的ip，是否在电视端百名段内
// 如果未设置，则认为是不限IP
func IsOBIP(ip string) bool {
	whiteIPLen, _ := core.RedisDoInt(core.RedisClient4, "scard", config.CACHE_KEY_OB_IPS)
	if whiteIPLen == 0 {
		return true
	}
	flag, _ := core.RedisDoBool(core.RedisClient4, "sismember", config.CACHE_KEY_OB_IPS, ip)
	return flag
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* action
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 观察房间
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func ObRoom(userId int, number string, mNumber uint16) *core.Error {
	// 判断用户是否已连接
	user, err := UserMap.GetUser(userId)
	if err != nil {
		return err
	}

	// 判断用户是否在roomId
	if user.RoomId > 0 {
		return core.NewError(-202, roomService.GetRoomNumberById(user.RoomId))
	}

	// 读取房间编号对应的房间id
	roomId := roomService.GetRoomIdByNumber(number)
	if roomId == 0 {
		return core.NewError(-302, number)
	}

	// 判断房间是否已存在
	room, err := RoomMap.GetRoom(roomId)
	if err != nil {
		return err
	}

	// 非TV类型不允许观察
	if !room.IsTV() {
		return core.NewError(-328, room.RoomId)
	}

	// 判断用户是否是观察者
	if !user.isObservers() {
		return core.NewError(-214, userId)
	}

	// 这里需要加锁，防止并发加入
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 是否已经是观察者了
	if room.Ob.hasUser(userId) {
		return core.NewError(-329, room.RoomId)
	}

	// 更新观察者的房间id
	user.RoomId = room.RoomId
	// 更新房间观察者
	room.Ob.add(userId)

	// 将用户房间数据存入cache
	userService.SetRoomId(userId, room.RoomId)

	// 回应一个成功的消息
	user.AppendMessage(JoinRoomResponse(room, mNumber))

	core.Logger.Info("[obRoom]userId:%d,roomId:%d,number:%s", userId, room.RoomId, room.Number)

	return nil
}

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 结束房间
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
func EndRoom(userId int) *core.Error {
	user, room, err := getUserRoom(userId)
	if err != nil {
		return err
	}

	// 非TV类型不允许观察
	if !room.IsTV() {
		return core.NewError(-328, room.RoomId)
	}

	// 判断用户是否是观察者
	if !user.isObservers() {
		return core.NewError(-214, userId)
	}

	// 游戏中的房间不允许退出
	if room.StartTime > 0 && !room.IsReadying() {
		return core.NewError(-330, room.RoomId)
	}

	// 这里需要加锁，防止并发
	room.Mux.Lock()
	defer room.Mux.Unlock()

	// 房间没开始的时候，直接结束
	if room.StartTime == 0 {
		dismissRoom(room, config.DISMISS_ROOM_CODE_HOST_LEAVE)
	} else {
		room.finish()
	}

	return nil
}
