package main

import (
	"fmt"
	"sync"
	"time"

	metrics "github.com/rcrowley/go-metrics"
)

var (
	mu       sync.RWMutex
	groupMgr = make(map[string]*Group) // key: group name
)

func init() {
	names := []string{"group_1"}
	for _, name := range names {
		groupMgr[name] = newGroup(name)
	}
}

func listGroup() (r []string) {
	mu.RLock()
	for k := range groupMgr {
		r = append(r, k)
	}
	mu.RUnlock()
	return
}

type Group struct {
	name string
	msgs chan []byte

	muMember sync.RWMutex
	member   map[string]*client // key: clientID
}

func newGroup(name string) *Group {
	g := Group{}
	g.name = name
	g.msgs = make(chan []byte, 1000)
	g.member = make(map[string]*client)

	go g.loopSend()
	go g.statusReport()
	return &g
}

func joinGroup(cli *client, groupName string) error {
	g, ok := groupMgr[groupName]
	if !ok || g == nil {
		return fmt.Errorf("not find group name %s", groupName)
	}
	return g.join(cli)
}

func exitGroup(cli *client, groupName string) error {
	g, ok := groupMgr[groupName]
	if !ok || g == nil {
		return fmt.Errorf("not find group name %s", groupName)
	}
	return g.exit(cli)
}

func cliDrop(cliID string) {
	// 通知所有group 玩家掉线
	for _, g := range groupMgr {
		g.drop(cliID)
	}
}

func inGroup(cliID string, groupName string) (bool, error) {
	g, ok := groupMgr[groupName]
	if !ok || g == nil {
		return false, fmt.Errorf("not find group name %s", groupName)
	}
	return g.isExist(cliID), nil
}

func sendMsgToGroup(msg []byte, groupName string) error {
	g, ok := groupMgr[groupName]
	if !ok || g == nil {
		return fmt.Errorf("not find group name %s", groupName)
	}
	return g.sendMsg(msg)
}

func (g *Group) join(cli *client) error {
	if cli != nil {
		g.muMember.Lock()
		g.member[cli.id] = cli
		g.muMember.Unlock()
	}
	return nil
}

func (g *Group) exit(cli *client) error {
	if cli != nil {
		g.muMember.Lock()
		delete(g.member, cli.id)
		g.muMember.Unlock()
	}
	return nil
}

func (g *Group) drop(cliID string) error {
	g.muMember.Lock()
	delete(g.member, cliID)
	g.muMember.Unlock()
	return nil
}

func (g *Group) isExist(cliID string) (exist bool) {
	g.muMember.RLock()
	_, exist = g.member[cliID]
	g.muMember.RUnlock()
	return
}

func (g *Group) sendMsg(msg []byte) error {
	select {
	case g.msgs <- msg:
	default:
		return fmt.Errorf("group %s full", g.name)
	}
	return nil
}

func (g *Group) loopSend() {
	tik := time.NewTicker(time.Second * 5)
	tik.Stop()

	//snap := make(map[string]*client)
	for {
		select {
		//case <-tik.C:
		//case stop
		case msg := <-g.msgs:
			g.muMember.RLock()
			for _, c := range g.member {
				c.send(msg)
			}
			g.muMember.RUnlock()
		}
	}
}

func (g *Group) statusReport() {
	tik := time.NewTicker(time.Second * 5)
	defer tik.Stop()

	for {
		select {
		case <-tik.C:
			gauge := metrics.GetOrRegisterGauge(fmt.Sprintf("g_%s_Lmsg", g.name), metrics.DefaultRegistry)
			gauge.Update(int64(len(g.msgs)))
		}
	}
}
