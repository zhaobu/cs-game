package main

import (
	"fmt"
	"sync"

	"github.com/smallnest/rpcx/client"
)

var (
	muGameCli sync.RWMutex
	gameCli   = make(map[string]client.XClient)
)

func getGameCli(gameName string) (cli client.XClient, err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	if gameName == "" {
		return nil, fmt.Errorf("empty game name")
	}

	gameName = "game/" + gameName

	muGameCli.Lock()
	defer muGameCli.Unlock()

	v, ok := gameCli[gameName]
	if !ok {
		servicePath := gameName
		d := client.NewConsulDiscovery(*basePath, servicePath, []string{*consulAddr}, nil)
		cli := client.NewXClient(servicePath, client.Failover, client.RandomSelect, d, client.DefaultOption)
		gameCli[gameName] = cli
		return cli, nil
	}
	return v, nil
}
