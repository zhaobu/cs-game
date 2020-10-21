package protobuf

import (
	"reflect"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

func Unmarshal(name string, data []byte) (pb proto.Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("recover %v", r)
		}
	}()

	rt := proto.MessageType(name)
	if rt == nil {
		err = errors.Errorf("unknown type: [%s]", name)
		return
	}

	rvi := reflect.New(rt.Elem()).Interface()
	pb = rvi.(proto.Message)

	err = proto.Unmarshal(data, pb)
	if err != nil {
		err = errors.Errorf("proto unmarshal %v name[%s] data[%v]", err, name, data)
		return
	}
	return
}

func Marshal(pb proto.Message) (name string, data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Errorf("recover %v", r)
		}
	}()

	name = proto.MessageName(pb)
	data, err = proto.Marshal(pb)
	return
}
