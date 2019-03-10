package robot

import (
	"time"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// 机器人计时器
func (this *Robot) startTimer() {
	go func() {
		// 捕获异常
		defer util.RecoverPanic()
		for {
			if this.IsQuit {
				this.trace("机器人已退出，退出计时器:%d", this.UserId)
				break
			}

			// 心跳
			this.heartBeatTimer()

			if this.RoomId > 0 {
				// 网络信号
				this.networkTimer()

				// 说话
				this.sayTimer()

				// 催人
				this.urgeTimer()

				// 退出
				this.quitTimer()
			}

			// 是否超时
			// 1000毫秒一次
			time.Sleep(1000 * time.Millisecond)
		}
	}()
}

// 检查心跳
func (this *Robot) heartBeatTimer() {
	t := util.GetTime()
	if t-this.LastHeartBeatTime >= int64(config.HEART_BEAT_SECOND) {
		this.LastHeartBeatTime = t
		this.HeartBeat()
	}
}

// 网络变化的计时器
func (this *Robot) networkTimer() {
	t := util.GetTime()
	if t >= this.NextNetworkTime {
		// 信号强度, 范围 2 ~ 5
		strength := 2 + util.RandIntn(4)
		var chatId int16
		switch strength {
		case 2:
			chatId = config.CHAT_ID_SIGNAL_WEAK
		case 3:
			chatId = config.CHAT_ID_SIGNAL_NORMAL
		case 4:
			chatId = config.CHAT_ID_SIGNAL_STRONGER
		case 5:
			chatId = config.CHAT_ID_SIGNAL_VERY_STRONGER
		default:
			chatId = config.CHAT_ID_SIGNAL_VERY_STRONGER
		}
		this.network(chatId)

		// 计算下一次发送网络变化时间
		this.NextNetworkTime = t + int64(core.RobotCfg.NetworkInterval[0]+util.RandIntn(core.RobotCfg.NetworkInterval[1]-core.RobotCfg.NetworkInterval[0]+1))
	}
}

// 说话
func (this *Robot) sayTimer() {
	t := util.GetTime()
	if this.LastCheckSayTime == 0 {
		this.LastCheckSayTime = t
		return
	}
	// 检查检测间隔
	if t-this.LastCheckSayTime < int64(core.RobotCfg.ChatCheckInterval) {
		return
	}
	// 更新最后检测时间
	this.LastCheckSayTime = t
	// 检查发送间隔
	if t-this.LastSayTime < int64(core.RobotCfg.ChatInterval) {
		return
	}
	// 检查概率
	if !this.checkRate(core.RobotCfg.ChatRate) {
		return
	}
	// 检查随机说话的id
	if len(core.RobotCfg.ChatIds) == 0 {
		return
	}
	this.LastSayTime = t
	rand := util.RandIntn(len(core.RobotCfg.ChatIds))
	this.chat(int16(core.RobotCfg.ChatIds[rand]), "")
}

// 催人
func (this *Robot) urgeTimer() {
	t := util.GetTime()
	if this.LastOtherOperationTime == 0 {
		return
	}
	// 检查检测间隔
	if t-this.LastCheckUrgeTime < int64(core.RobotCfg.UrgeInterval) {
		return
	}
	// 检查概率
	if !this.checkRate(core.RobotCfg.UrgeRate) {
		return
	}
	// 更新最后检测时间
	this.LastCheckUrgeTime = t
	// 判断是否满足催人条件
	if t-this.LastOtherOperationTime < int64(core.RobotCfg.UrgeInterval) {
		return
	}
	this.chat(int16(core.RobotCfg.UrgeChatId), "")
}

// 退出
func (this *Robot) quitTimer() {
	t := util.GetTime()

	// 1分钟没有操作，则退出机器人
	if t-this.CreateTime < int64(60) {
		return
	}
	if t-this.LastOtherOperationTime < int64(60) {
		return
	}
	// 检查概率
	if !this.checkRate(5) {
		return
	}
	this.quit(8)
}
