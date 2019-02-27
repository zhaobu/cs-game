package desk

import (
	"cy/game/cache"
	"cy/game/pb/game/ddz"

	"github.com/sirupsen/logrus"
)

func (d *desk) enterWait() {
	d.clearLast()
	logrus.WithFields(logrus.Fields{"deskid": d.id}).Info("等待玩家准备")
}

// 好友场清空上局信息，不包含人
func (d *desk) clearLast() {
	d.backCard = nil
	d.backCardCt = pbgame_ddz.BackCardType_BcCtUnknown
	d.currPlayer = 0
	d.currPlayerID = 0
	d.callUID = 0
	d.robUID = 0
	d.landlord = 0
	d.lastGiveCardPlayerID = 0
	d.lastGiveCard = nil
	d.mul.clear()
}

// 玩家离开
func (d *desk) userExit(uid uint64) {
	// 删除桌子中玩家
	for idx, v := range d.sdPlayers {
		if v.uid == uid {
			delete(d.sdPlayers, idx)
			break
		}
	}

	// 删除玩家桌子关系
	deleteUser2desk(uid)

	// 修改玩家状态
	cache.ExitGame(uid, gameName, gameID, d.id)
}

func (d *desk) isSitDownUser(uid uint64) bool {
	for _, v := range d.sdPlayers {
		if v.uid == uid {
			return true
		}
	}
	return false
}

// 匹配完坐下
func (d *desk) matchSitDwon(uids ...uint64) {
	for _, uid := range uids {
		d.markUserSitDown(uid, true)
	}

	d.toSiteDown(d.deskInfo(0))

	d.canStart()
	return
}

func (d *desk) handleUserReadyReq(uid uint64, _ *pbgame_ddz.UserReadyReq) {
	if !d.f.Is("SWait") {
		d.toOne(&pbgame_ddz.UserReadyRsp{Code: 2}, uid)
		return
	}

	for _, v := range d.sdPlayers {
		if v.uid == uid {
			v.status = pbgame_ddz.UserGameStatus_UGSReady
			d.toOne(&pbgame_ddz.UserReadyRsp{Code: 1}, uid)
			d.toSiteDown(&pbgame_ddz.UserGameStatusBroadcast{UserID: uid, Status: v.status})
			d.canStart()
			return
		}
	}
}

func (d *desk) canStart() {
	readyCnt := 0
	for _, v := range d.sdPlayers {
		if v.status == pbgame_ddz.UserGameStatus_UGSReady {
			readyCnt++
		}
	}

	if readyCnt == seatNumber {
		if err := d.f.Event("wait_end"); err != nil {
			logrus.WithFields(logrus.Fields{"deskid": d.id, "err": err.Error()}).Error()
		}
		d.isStarted = true
		d.enterCall()
	}
}
