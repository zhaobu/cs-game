package game

import (
	"time"

	"mahjong.go/config"
	"mahjong.go/library/core"

	"sync"

	"github.com/fwhappy/util"
	roomService "mahjong.go/service/room"
)

// 监听用户满员之后，用户不进行准备
func listenRoomKickUser(roomId int64) {
	// 捕获异常
	defer util.RecoverPanic()

	var wg sync.WaitGroup

	core.Logger.Debug("[listenRoomKickUser]监听房间准备超时,roomId:%v", roomId)

	for {
		time.Sleep(time.Second)
		room, err := RoomMap.GetRoom(roomId)
		if err != nil {
			core.Logger.Debug("[listenRoomKickUser]房间[%d]已解散，跳出监听", roomId)
			break
		}

		// 房间未满，继续等待
		if !room.IsFull() {
			continue
		}

		// 房间已开始，退出监听
		if room.StartTime > 0 {
			break
		}

		// 房间已结束
		if room.CheckStatus(config.ROOM_STATUS_COMPLETED) {
			break
		}

		// 判断是否已经准备超时
		if util.GetTime()-room.FullTime <= config.ROOM_FIRST_READY_TIMEOUT_SECOND {
			continue
		}

		wg.Add(1)
		go func() {
			defer util.RecoverPanic()
			defer wg.Done()

			room.Mux.Lock()
			defer room.Mux.Unlock()

			// 如果所有人都未准备，则解散房间
			if len(room.ReadyList) == 0 {
				// 房间已超时，解散房间
				dismissRoom(room, config.DISMISS_ROOM_CODE_TIMEOUT)
			} else {
				for _, userId := range room.GetIndexUserIds() {
					// 跳过已准备用户
					if util.IntInSlice(userId, room.ReadyList) {
						continue
					}
					// 将用户从房间移除
					isDismiss, isSuccess := RemoveRoomUser(room, userId, config.DISMISS_ROOM_CODE_TIMEOUT, config.QUIT_ROOM_CODE_TIMEOUT)
					// 若删除失败，表示用户已不在房间了
					if !isSuccess {
						continue
					}
					if !isDismiss {
						// 若用户在线，需要删除用户内存中的房间
						if user, err := UserMap.GetUser(userId); err == nil {
							// 删除用户内存中的房间
							user.RoomId = int64(0)
						}
						// 构建解散房间的消息
						// responsePacket := CloseRoomPush(config.DISMISS_ROOM_CODE_TIMEOUT, core.GetLang(config.DISMISS_ROOM_CODE_TIMEOUT+config.DISMISS_ROOM_CODE_OFFSET))
						// SendMessageByUserId(userId, responsePacket)
					}

					core.Logger.Info("[listenRoomKickUser]roomId:%v,userId:%v,isDismiss:%v", roomId, userId, isDismiss)
					// 如果已退出，则跳出循环
					if isDismiss {
						break
					}
				}
			}
		}()
		wg.Wait()
	}
}

// 监听房间超时
// 房间解散之后，这里还是处于time.sleep状态, 协程并没有及时关闭，而是在下次循环时，检测到房间不在了，才退出
func listenRoomTimeout(roomId int64) {
	// 捕获异常
	defer util.RecoverPanic()

	for {
		time.Sleep(time.Minute) // 1分钟检测一次
		room, err := RoomMap.GetRoom(roomId)
		if err != nil {
			core.Logger.Debug("[listenRoomTimeout]房间[%d]已解散，跳出监听", roomId)
			break
		}
		// 房间已结束
		if room.CheckStatus(config.ROOM_STATUS_COMPLETED) {
			core.Logger.Debug("[listenRoomTimeout]房间[%d]已完成，跳出监听", roomId)
			break
		}
		if room.IsTimeout() {
			// 房间已超时，解散房间
			dismissRoom(room, config.DISMISS_ROOM_CODE_TIMEOUT)
			core.Logger.Info("房间[%d]已超时，解散房间", roomId)
			break
		} else {
			// 房间未超时，更新指令的有效期
			// fixme 这里的更新有点频繁, 需要更好的策略
			roomService.SetRoomExpire(room.RoomId, room.Number, int(config.ROOM_EXPIRE_SECOND))
			core.Logger.Debug("房间[%d]未超时，延长cache的有效期", roomId)

			// 发送活跃的消息
			if room.IsClub() {
				CPool.appendMessage(ClubG2CRoomActiveRequest(room.ClubId, room.RoomId))
				core.Logger.Debug("发送房间活跃检测请求,clubId:%v,roomId:%v", room.ClubId, room.RoomId)
			} else if room.IsLeague() {
				LPool.appendMessage(LeagueS2LGameActivePush(room.RaceInfo.Id, room.RaceRoom.Id, room.RoomId))
				core.Logger.Debug("发送房间活跃检测请求,raceId:%v, raceRoomId:%v, roomId:%v", room.RaceInfo.Id, room.RaceRoom.Id, room.RoomId)
			}

			/*
				// 检测房间的锁定状态
				go func() {
					room.Mux.Lock()
					core.Logger.Debug("测试房间加锁成功,roomId:%v", room.RoomId)
					defer func() {
						room.Mux.Unlock()
						core.Logger.Debug("测试房间解锁成功,roomId:%v", room.RoomId)
					}()
				}()
				go func() {
					room.UserOperationMux.Lock()
					core.Logger.Debug("测试房间操作加锁成功,roomId:%v", room.RoomId)
					defer func() {
						room.UserOperationMux.Unlock()
						core.Logger.Debug("测试房间操作解锁成功,roomId:%v", room.RoomId)
					}()
				}()
			*/
		}
	}
}

// 监听解散房间
func listenRoomDissmiss(room *Room) {
	// 捕获异常
	defer util.RecoverPanic()

	var handler int

	select {
	case handler = <-room.DissmissChan:
		break
	case <-time.After(config.ROOM_DISMISS_AOTO_ALLOW_INTERVAL * time.Second):
		core.Logger.Info("房间解散申请超时未操作，自动执行同意操作, roomId:%v", room.RoomId)
		break
	}

	core.Logger.Debug("listenRoomDissmiss, roomId:%d, handler:%#v", room.RoomId, handler)

	if handler == config.ROOM_DISMISS_ALLOW {
		// 同意解散房间
		core.Logger.Info("所有人同意，解散房间, roomId:%d", room.RoomId)

		// 如果是第一局，需要退还房费， 非自主创建的，不退费
		if room.Round == 1 && len(room.Record) == 0 && room.EnableReturnCost() {
			room.returnCost()
		}

		// 房间解散未回应有相应的惩罚
		if room.EnableDismissPunishment() {
			room.dismissPunish()
		}

		// 解散房间
		room.finish()
	} else {
		// 拒绝解散房间
		// 更新房间状态恢复房间数据
		room.Dissmisser = 0
		room.DismissTime = int64(0)
		room.DismissOp = &sync.Map{}
		close(room.DissmissChan)

		core.Logger.Info("有人不同意，解散不房间, roomId:%d", room.RoomId)
	}
}
