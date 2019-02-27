package desk

import (
	"cy/game/logic/ddz/card"
	"cy/game/pb/game/ddz"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

func (d *desk) enterCall() {

	logrus.WithFields(logrus.Fields{"deskid": d.id, "uids": d.getSdUids(), "gamearg": d.arg}).Info("开始叫地主")

	d.clearCallFlag()
	d.initCard()
	d.randFirst()
	d.initGiveCard()

	d.callNotif()
}

// 重发牌 清空历史标记
func (d *desk) clearCallFlag() {
	d.callUID = 0
	for _, v := range d.sdPlayers {
		v.call = pbgame_ddz.CallCode_CNotUse
	}
}

// 洗牌
func (d *desk) initCard() {
	d.sdPlayers[0].currCard, d.sdPlayers[1].currCard, d.sdPlayers[2].currCard, d.backCard = card.RandDdzCard()
	d.backCardCt = d.backCard.CalcBackCardType()
}

// 随机首家
func (d *desk) randFirst() {
	d.currPlayer = rand.Intn(seatNumber)
	d.currPlayerID = d.sdPlayers[d.currPlayer].uid
}

// 开局发牌
func (d *desk) initGiveCard() {
	d.toSiteDown(&pbgame_ddz.GameStartNotif{CurrLoopCnt: d.currLoopCnt})
	for _, v := range d.sdPlayers {
		giveCard := &pbgame_ddz.GiveCard{}
		giveCard.Cards = v.currCard.Dump()
		d.toOne(giveCard, v.uid)
	}
}

// 叫地主通知
func (d *desk) callNotif() {
	d.toSiteDown(&pbgame_ddz.CallNotif{UserID: d.currPlayerID, Time: timeOutCallrob})

	d.reqTime = time.Now().UTC()
	seq := d.seq
	d.timer = tw.AfterFunc(timeOutCallrob*time.Second, func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		if d.seq == seq {
			logrus.WithFields(logrus.Fields{"deskid": d.id, "uid": d.currPlayerID}).Info("叫地主超时")

			d.sdPlayers[d.currPlayer].call = pbgame_ddz.CallCode_CNotCall // 默认不叫地主
			d.seq++
			d.callBroadcast(d.currPlayerID, pbgame_ddz.CallCode_CNotCall)

			d.checkCall()
		}
	})
}

// 叫地主广播
func (d *desk) callBroadcast(uid uint64, cc pbgame_ddz.CallCode) {
	d.toSiteDown(&pbgame_ddz.CallBroadcast{UserID: uid, Code: cc})
}

// 玩家动作-叫地主
func (d *desk) handleUserCall(uid uint64, req *pbgame_ddz.UserCall) {
	logrus.WithFields(logrus.Fields{"deskid": d.id, "status": d.f.Current(), "curruid": d.currPlayerID, "requid": uid, "code": req.Code}).Info("handleUserCall")

	if !d.f.Is("SCall") {
		return
	}

	if d.currPlayerID != uid {
		return
	}

	cc := req.Code

	d.sdPlayers[d.currPlayer].call = cc
	d.seq++
	d.callBroadcast(uid, cc)

	if cc == pbgame_ddz.CallCode_CCall {
		d.callUID = uid
	}

	d.checkCall()
}

func (d *desk) checkCall() {
	if d.callUID != 0 {
		if err := d.f.Event("call_end"); err != nil {
			logrus.WithFields(logrus.Fields{"deskid": d.id, "err": err.Error()}).Error()
		}
		d.enterRob()
		return
	}

	notCallCnt := 0 // 不叫地主的数量
	for _, v := range d.sdPlayers {
		if v.call == pbgame_ddz.CallCode_CNotCall {
			notCallCnt++
		}
	}

	// 都不叫 重新发牌
	if notCallCnt == seatNumber {
		logrus.WithFields(logrus.Fields{"deskid": d.id}).Info("重新发牌")

		d.clearCallFlag()
		d.initCard()
		d.randFirst()
		d.initGiveCard()
	} else {
		d.turnNext()
	}
	d.callNotif()
}

// 轮到下个玩家
func (d *desk) turnNext() {
	d.currPlayer = (d.currPlayer + 1) % seatNumber
	d.currPlayerID = d.sdPlayers[d.currPlayer].uid
}
