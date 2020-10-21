package main

import (
	pbgame_logic "game/pb/game/mj/changshu"
	"encoding/json"
	"io/ioutil"

	"go.uber.org/zap"
)

var (
	argTpl pbgame_logic.CreateArgTpl
)

func loadArgTpl(fn string) error {
	data, err := ioutil.ReadFile(fn)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &argTpl)
	if err != nil {
		return err
	}

	tlog.Info("arg tpl", zap.Any("CreateArgTpl", argTpl))
	return nil
}
