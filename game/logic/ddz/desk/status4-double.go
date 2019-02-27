package desk

import (
	"cy/game/pb/game/ddz"
	"time"

	"github.com/sirupsen/logrus"
)

func (d *desk) enterDouble() {
	logrus.WithFields(logrus.Fields{"deskid": d.id, "landlord": d.landlord}).Info("开始加倍")

	enableDouble := (d.arg.Type == gameTypMatch)
	d.toSiteDown(&pbgame_ddz.DoubleNotif{Time: timeOutDouble, Enable: enableDouble})

	time.AfterFunc(time.Second*timeOutDouble, func() {
		d.mu.Lock()
		defer d.mu.Unlock()

		if enableDouble {
			for _, v := range d.sdPlayers {
				if v.double == pbgame_ddz.DoubleCode_DNotUse {
					v.double = pbgame_ddz.DoubleCode_DNotDouble
					d.doubleBroadcast(v.uid, v.double)
				}
			}
		}

		if err := d.f.Event("rouble_end"); err != nil {
			logrus.WithFields(logrus.Fields{"deskid": d.id, "err": err.Error()}).Error()
		}

		d.enterPlay()
	})
}

func (d *desk) handleUserDouble(uid uint64, req *pbgame_ddz.UserDouble) {
	logrus.WithFields(logrus.Fields{"deskid": d.id, "status": d.f.Current(), "curruid": d.currPlayerID, "requid": uid, "code": req.Code}).Info("handleUserDouble")

	if !d.f.Is("SDouble") {
		return
	}

	for _, v := range d.sdPlayers {
		if v.uid == uid {
			if v.double == pbgame_ddz.DoubleCode_DNotUse {
				v.double = req.Code
				if req.Code == pbgame_ddz.DoubleCode_DDouble {
					v.doubleMul++
					d.syncDoubleMul(uid)
				}
				d.doubleBroadcast(uid, req.Code)
			}
			break
		}
	}
}

func (d *desk) syncDoubleMul(uid uint64) {
	if uid == d.landlord {
		for _, v := range d.sdPlayers {
			if v.uid != uid {
				v.doubleMul++
			}
		}
	} else {
		for _, v := range d.sdPlayers {
			if v.uid == d.landlord {
				v.doubleMul++
				break
			}
		}
	}
}

func (d *desk) doubleBroadcast(uid uint64, code pbgame_ddz.DoubleCode) {
	for _, v := range d.sdPlayers {
		d.toOne(&pbgame_ddz.DoubleBroadcast{
			UserID: uid,
			Code:   code,
			Mul:    d.mulUser(v.uid),
		}, v.uid)
	}
}
