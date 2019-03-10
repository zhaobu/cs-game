package main

import (
	. "cy/chat/def"
)

func sendLoginRsp(c *client) {
	rsp := LoginRsp{}
	rsp.Kind = OpLoginRsp
	rsp.Name = c.id

	c.sends(rsp)
	return
}
