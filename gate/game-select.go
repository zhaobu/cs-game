package main

import (
	"context"
	"game/cache"
	"game/codec"
	pbgame "game/pb/game"
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/smallnest/rpcx/client"
)

var (
	muGameCli sync.RWMutex
	gameCli   = make(map[string]client.XClient)
)

type gameSelector struct {
	servers map[string]string
}

func (s *gameSelector) Select(ctx context.Context, servicePath, serviceMethod string, args interface{}) string {
	switch serviceMethod {
	case "MakeDeskReq", "QueryGameConfigReq":
		return s.randomSelect()
	// case "QueryDeskInfoReq", "JoinDeskReq", "DestroyDeskReq", "ExitDeskReq", "GameAction", "SitDownReq", "VoteDestroyDeskReq":
	// 	gameID := ctx.Value("game_id").(string)
	// 	return gameID
	default:
		gameID := ctx.Value("game_id").(string)
		return gameID
	}
}

func (s *gameSelector) randomSelect() string {
	for k := range s.servers {
		return k
	}
	return ""
}

func (s *gameSelector) UpdateServer(servers map[string]string) {
	s.servers = make(map[string]string)
	for k, v := range servers {
		s.servers[k] = v
	}
	return
}

func (s *session) getGameAddr(msg *codec.Message) (gameName, gameID string) {
	defer func() {
		if r := recover(); r != nil {
			gameName = ""
			gameID = ""
		}
	}()

	pb, err := codec.Msg2Pb(msg)
	if err != nil {
		return
	}

	switch msg.Name {
	case proto.MessageName((*pbgame.MakeDeskReq)(nil)):
		makeDeskReq, ok := pb.(*pbgame.MakeDeskReq)
		if !ok {
			return
		}
		return makeDeskReq.GameName, ""
	case proto.MessageName((*pbgame.QueryGameConfigReq)(nil)):
		queryGameConfigReq, ok := pb.(*pbgame.QueryGameConfigReq)
		if !ok {
			return
		}
		return queryGameConfigReq.GameName, ""
	case proto.MessageName((*pbgame.JoinDeskReq)(nil)):
		joinDeskReq, ok := pb.(*pbgame.JoinDeskReq)
		if !ok {
			return
		}
		deskInfo, err := cache.QueryDeskInfo(joinDeskReq.DeskID)
		if err == nil && deskInfo != nil {
			if deskInfo.GameName != "" && deskInfo.GameID != "" {
				return deskInfo.GameName, deskInfo.GameID
			}
		}
		s.sendPb(&pbgame.JoinDeskRsp{Code: pbgame.JoinDeskRspCode_JoinDeskNotExist})
	case proto.MessageName((*pbgame.QueryDeskInfoReq)(nil)):
		queryDeskInfoReq, ok := pb.(*pbgame.QueryDeskInfoReq)
		if !ok {
			return
		}
		deskInfo, err := cache.QueryDeskInfo(queryDeskInfoReq.DeskID)
		if err == nil && deskInfo != nil {
			if deskInfo.GameName != "" && deskInfo.GameID != "" {
				return deskInfo.GameName, deskInfo.GameID
			}
		}
		s.sendPb(&pbgame.QueryDeskInfoRsp{Code: 2})
	case proto.MessageName((*pbgame.DestroyDeskReq)(nil)):
		destroyDeskReq, ok := pb.(*pbgame.DestroyDeskReq)
		if !ok {
			return
		}
		deskInfo, err := cache.QueryDeskInfo(destroyDeskReq.DeskID)
		if err == nil && deskInfo != nil {
			if deskInfo.GameName != "" && deskInfo.GameID != "" {
				return deskInfo.GameName, deskInfo.GameID
			}
		}
		s.sendPb(&pbgame.DestroyDeskRsp{Code: 2})
	default:
		sessInfo, err := cache.QuerySessionInfo(msg.UserID)
		if err == nil && sessInfo != nil {
			return sessInfo.GameName, sessInfo.GameID
		}
	}
	return
}

func getGameCli(gameName string) (cli client.XClient, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if gameName == "" {
		return nil, fmt.Errorf("房间不存在")
	}

	gameName = "game/" + gameName

	muGameCli.Lock()
	defer muGameCli.Unlock()

	_, ok := gameCli[gameName]
	if !ok {
		servicePath := gameName
		d := client.NewConsulDiscovery(*basePath, servicePath, []string{*consulAddr}, nil)
		cli := client.NewXClient(servicePath, client.Failover, client.SelectByUser, d, client.DefaultOption)
		cli.SetSelector(&gameSelector{servers: make(map[string]string)})
		gameCli[gameName] = cli
	}

	return gameCli[gameName], nil
}
