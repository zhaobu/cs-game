package main

import (
	"context"
	"cy/other/im/codec"
)

type selectByToID struct {
	servers map[string]string
}

func (s *selectByToID) Select(ctx context.Context, servicePath, serviceMethod string, args interface{}) (wh string) {
	if len(s.servers) == 0 {
		return ""
	}

	req, ok := args.(*codec.MsgPayload)
	if !ok {
		return ""
	}

	return queryPlace(req.ToUID)
}

func (s *selectByToID) UpdateServer(servers map[string]string) {
	s.servers = make(map[string]string)
	for k, v := range servers {
		s.servers[k] = v
	}
}
