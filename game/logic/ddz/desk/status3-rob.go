package desk

import (
	"cy/game/logic/ddz/card"
	"cy/game/pb/game/ddz"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func (d *desk) enterRob() {
	d.loge.Info("start rob")

	f := logrus.Fields{}
	for _, v := range d.sdPlayers {
		f[fmt.Sprintf("%d", v.uid)] = v.currCard.String()
	}
	f["backcard"] = d.backCard.String()

	d.loge.Infof("cardinfo %v", f)

	d.turnNextNNotCall()
	if d.currPlayerID == d.callUID {
		d.robEnd()
		return
	}
	d.robNotif()
}

func (d *desk) robEnd() {
	d.robComplete()
	if err := d.f.Event("rob_end"); err != nil {
		d.loge.Error(err.Error())
	}
	d.enterDouble()
}

func (d *desk) checkRob() {
	if d.currPlayerID == d.callUID {
		d.robEnd()
		return
	}
	d.turnNextNNotCall()
	if d.currPlayerID == d.callUID && d.robUID == 0 { // 没有人抢，直接定了
		d.robEnd()
		return
	}
	d.robNotif()
}

func (d *desk) turnNextNNotCall() {
	for i := 0; i < seatNumber; i++ {
		d.currPlayer = (d.currPlayer + 1) % seatNumber
		d.currPlayerID = d.sdPlayers[d.currPlayer].uid
		// 不叫的 不能抢了
		if d.sdPlayers[d.currPlayer].call != pbgame_ddz.CallCode_CNotCall {
			break
		}
		// 最多轮到叫地主的人
		if d.currPlayerID == d.callUID {
			break
		}
	}
}

// 抢地主通知
func (d *desk) robNotif() {
	d.toSiteDown(&pbgame_ddz.RobNotif{UserID: d.currPlayerID, Time: timeOutCallrob})

	d.reqTime = time.Now().UTC()
	seq := d.seq
	d.timer = tw.AfterFunc(timeOutCallrob*time.Second, func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		if d.seq == seq {
			d.loge.Infof("rob timeout %d", d.currPlayerID)

			d.sdPlayers[d.currPlayer].rob = pbgame_ddz.RobCode_RNotRob // 默认不抢地主
			d.seq++
			d.robBroadcast(d.currPlayerID, pbgame_ddz.RobCode_RNotRob)

			d.checkRob()
		}
	})
}

func (d *desk) robBroadcast(uid uint64, rc pbgame_ddz.RobCode) {
	rb := &pbgame_ddz.RobBroadcast{UserID: uid, Code: rc}
	rb.Mul = d.mul.get()
	d.toSiteDown(rb)
}

func (d *desk) handleUserRob(uid uint64, req *pbgame_ddz.UserRob) {
	d.loge.Infof("handleUserRob status:%s curruid:%d requid:%d code:%d", d.f.Current(), d.currPlayerID, uid, req.Code)

	if !d.f.Is("SRob") {
		return
	}

	if d.currPlayerID != uid {
		return
	}

	rc := pbgame_ddz.RobCode(req.Code)
	if rc == pbgame_ddz.RobCode_RRob {
		d.robUID = uid
		d.mul.callRob++
	}

	d.sdPlayers[d.currPlayer].rob = rc
	d.seq++
	d.robBroadcast(uid, rc)

	d.checkRob()
}

func (d *desk) robComplete() {
	// 最后一个人叫地主的情况
	if d.robUID == 0 {
		d.robUID = d.callUID
	}

	d.landlord = d.robUID

	// 重新定位当前玩家
	for i := 0; i < seatNumber; i++ {
		if d.sdPlayers[i].uid == d.landlord {
			d.currPlayer = i
			d.currPlayerID = d.landlord
			break
		}
	}

	d.lastGiveCardPlayerID = d.landlord                       // 地主特殊处理
	d.sdPlayers[d.currPlayer].currCard.Add(d.backCard.Dump()) // 给地主底牌
	d.mul.back = card.BackCardType2Mul(d.backCardCt)          // 地主确定后 才能改变底牌倍数

	d.backNotif()
}

// 起底牌
func (d *desk) backNotif() {
	seq := d.backCard.Dump()
	d.toSiteDown(&pbgame_ddz.BackNotif{
		Landlord: d.landlord,
		Cards:    seq,
		BackMul:  d.mul.back,
		Ct:       d.backCard.CalcBackCardType(),
		Mul:      d.mul.get(),
	})
}
