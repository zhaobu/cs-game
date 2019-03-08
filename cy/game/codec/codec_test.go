package codec

import (
	"bytes"
	"testing"
)

func Test_codec(t *testing.T) {
	p1 := NewPacket()
	p1.Msgs = append(p1.Msgs, &Message{Name: "zz", Payload: []byte("zhengzhong")})
	p1.Msgs = append(p1.Msgs, &Message{Name: "gg", Payload: []byte("guagua")})

	data, _ := p1.Encode()

	p2 := NewPacket()
	p2.Decode(data)

	if len(p2.Msgs) != 2 {
		t.Errorf("bad len %d", len(p2.Msgs))
	}

	if p2.Msgs[0].Name != p1.Msgs[0].Name {
		t.Errorf("bad name %s %s", p2.Msgs[0].Name, p1.Msgs[0].Name)
	}

	if bytes.Compare(p2.Msgs[0].Payload, p1.Msgs[0].Payload) != 0 {
		t.Errorf("bad payload %s %s", p2.Msgs[0].Payload, p1.Msgs[0].Payload)
	}

	if p2.Msgs[1].Name != p1.Msgs[1].Name {
		t.Errorf("bad name %s %s", p2.Msgs[1].Name, p1.Msgs[1].Name)
	}

	if bytes.Compare(p2.Msgs[1].Payload, p1.Msgs[1].Payload) != 0 {
		t.Errorf("bad payload %s %s", p2.Msgs[1].Payload, p1.Msgs[1].Payload)
	}
}
