package protobuf

import (
	"cy/other/im/pb/misc"

	"github.com/golang/protobuf/proto"
)

func GroupPb(name string, m ...proto.Message) proto.Message {
	g := &misc.GroupMsg{}
	g.Name = name
	for _, v := range m {
		sm := &misc.SomeMsg{}
		sm.Name = proto.MessageName(v)
		sm.Payload, _ = proto.Marshal(v)
		g.Msgs = append(g.Msgs, sm)
	}
	return g
}

func GroupAppend(msg proto.Message, m ...proto.Message) proto.Message {
	g, ok := msg.(*misc.GroupMsg)
	if ok {
		for _, v := range m {
			sm := &misc.SomeMsg{}
			sm.Name = proto.MessageName(v)
			sm.Payload, _ = proto.Marshal(v)
			g.Msgs = append(g.Msgs, sm)
		}
	}
	return g
}
