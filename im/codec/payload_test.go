package codec

import (
	"bytes"
	"reflect"
	"testing"
)

func Test_payload(t *testing.T) {
	p := NewMsgPayload()
	p.Seq = 123
	p.FromUID = 1
	p.ToUID = 2
	p.SetOneway(false)
	p.SetHeartbeat(true)
	p.SetCompressType(CtNone)
	p.PayloadName = `123456`
	p.Payload = bytes.Repeat([]byte(`123456789`), 2)

	data, err := p.Encode()
	if err != nil {
		t.Error(err)
		return
	}

	p2 := NewMsgPayload()
	err = p2.Decode(data)
	if err != nil {
		t.Errorf("%v %+v\n", err, *p2)
		return
	}

	if !reflect.DeepEqual(*p, *p2) {
		t.Errorf("p: %+v\n p2: %+v\n", *p, *p2)
		return
	}

	t.Errorf("p: %+v\n p2: %+v\n", *p, *p2)
}
