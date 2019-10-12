package main

import (
	"context"
	"game/cache"
	"game/codec"
	"game/logic/ddz/desk"
	"game/pb/common"
	"game/pb/inner"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"
)

func (p *ddz) GameMatchSucc(ctx context.Context, args *codec.Message, reply *codec.Message) (err error) {
	pb, err := codec.Msg2Pb(args)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	req, ok := pb.(*pbinner.GameMatchSucc)
	if !ok {
		err = fmt.Errorf("not *pbinner.GameMatchSucc")
		log.Error(err.Error())
		return
	}

	var newDeskID uint64

	defer func() {
		r := recover()
		stack := ""
		if r != nil {
			stack = string(debug.Stack())
		}

		log.Infof("req:%s %+v err:%s r:%v stack:%s", args.Name, *req, err, r, stack)

		if err != nil {
			log.Warn(err)

			if newDeskID != 0 {
				for _, uid := range req.UserIDs {
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

	// 人数满足
	const seatNumber = 3
	if uint32(len(req.UserIDs)) != seatNumber {
		return errors.Errorf("cnt:%d != seatnumber:%d", len(req.UserIDs), seatNumber)
	}

	// 房间ID合法
	roomArg := desk.QueryMatchRoomArg(req.RoomId)
	if roomArg == nil {
		return errors.Errorf("bad roomid %d", req.RoomId)
	}

	newDeskID, err = cache.AllocDeskID()
	if err != nil {
		return errors.Wrap(err, "cache.AllocDeskID")
	}

	_, err = addDeskInfo(newDeskID, req.UserIDs)
	if err != nil {
		return errors.Wrap(err, "addDeskInfo")
	}

	for _, uid := range req.UserIDs {
		succ, err := cache.EnterGame(uid, gameName, gameID, newDeskID, true)
		if err != nil || !succ {
			return errors.Errorf("enter game failed uid:%d gamename:%s gameid:%s deskid:%d err:%v",
				uid, gameName, gameID, newDeskID, err)
		}
	}

	roomArg.Type = 1
	roomArg.RoomId = req.RoomId
	roomArg.DeskID = newDeskID

	desk.MatchStart(roomArg, req.UserIDs)

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
