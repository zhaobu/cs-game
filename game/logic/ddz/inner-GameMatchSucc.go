package main

import (
	"context"
	"cy/game/cache"
	"cy/game/codec"
	"cy/game/logic/ddz/desk"
	"cy/game/pb/common"
	"cy/game/pb/inner"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (p *ddz) GameMatchSucc(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		logrus.Error(err.Error())
		return err
	}

	gameMatchSucc, ok := pb.(*pbinner.GameMatchSucc)
	if !ok {
		err = fmt.Errorf("not *pbinner.GameMatchSucc")
		logrus.Error(err.Error())
		return
	}

	defer func() {
		r := recover()
		stack := ""
		if r != nil {
			stack = string(debug.Stack())
		}

		logrus.WithFields(logrus.Fields{
			"req":   *gameMatchSucc,
			"err":   err,
			"r":     r,
			"stack": stack,
			"name":  args.Name,
		}).Info("recv")
	}()

	// ###############################################
	// 人数满足
	const seatNumber = 3
	if uint32(len(gameMatchSucc.UserIDs)) != seatNumber {
		return errors.Errorf("cnt:%d != seatnumber:%d", len(gameMatchSucc.UserIDs), seatNumber)
	}

	// 房间ID合法
	roomArg := desk.QueryMatchRoomArg(gameMatchSucc.RoomId)
	if roomArg == nil {
		return errors.Errorf("bad roomid %d", gameMatchSucc.RoomId)
	}

	var newDeskID uint64
	defer func() {
		if err != nil {
			logrus.Warn(err)

			if newDeskID != 0 {
				for _, uid := range gameMatchSucc.UserIDs {
					cache.ExitGame(uid, gameName, gameID, newDeskID)
					cache.ExitMatch(uid)
				}

				cache.DeleteClubDeskRelation(newDeskID)
				cache.DelDeskInfo(newDeskID)
				cache.FreeDeskID(newDeskID)
			}

			// 通知匹配失败 TODO
		}
	}()

	newDeskID, err = cache.AllocDeskID()
	if err != nil {
		return errors.Wrap(err, "cache.AllocDeskID")
	}

	_, err = addDeskInfo(newDeskID, gameMatchSucc.UserIDs)
	if err != nil {
		return errors.Wrap(err, "addDeskInfo")
	}

	for _, uid := range gameMatchSucc.UserIDs {
		succ, err := cache.EnterGame(uid, gameName, gameID, newDeskID, true)
		if err != nil || !succ {
			return errors.Errorf("enter game failed uid:%d game name:%s game id:%s desk id:%d err:%v",
				uid, gameName, gameID, newDeskID, err)
		}
	}

	roomArg.Type = 1
	roomArg.RoomId = gameMatchSucc.RoomId
	roomArg.DeskID = newDeskID

	desk.MatchStart(roomArg, gameMatchSucc.UserIDs)

	return nil
}

// 添加桌子信息到cache
func addDeskInfo(id uint64, uids []uint64) (*pbcommon.DeskInfo, error) {
	deskInfo := &pbcommon.DeskInfo{}
	deskInfo.ID = id
	deskInfo.CreateUserID = 0
	deskInfo.CreateTime = time.Now().UTC().Unix()
	deskInfo.GameName = gameName
	deskInfo.GameID = gameID
	deskInfo.UserIDs = uids
	return deskInfo, cache.AddDeskInfo(deskInfo)
}
