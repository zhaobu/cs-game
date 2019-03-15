package main

import (
	pbgame_logic "cy/game/pb/game/mj/changshu"
	"encoding/json"
	"io/ioutil"
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

	log.Infof("arg tpl: %+v\n", argTpl)
	return nil
}
