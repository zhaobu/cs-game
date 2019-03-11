package codec

import (
	"bytes"
	"cy/util"
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

type SerializeType byte

type CompressType byte

const (
	StPb SerializeType = iota
)

const (
	CtNone CompressType = iota
	CtGzip
)

//[7][6][5][4][3][2][1][0]
//[0] SerializeType
//[1] CompressType
//[2] 0x01 Oneway
//[2] 0x02 Heartbeat
//[3] 0x01 BroadCast
//[3] 0x02 MultiCast

type Flag [8]byte

func (f Flag) GetSerializeType() SerializeType {
	return SerializeType(f[0])
}

func (f *Flag) SetSerializeType(t SerializeType) {
	f[0] = byte(t)
}

func (f Flag) GetCompressType() CompressType {
	return CompressType(f[1])
}

func (f *Flag) SetCompressType(t CompressType) {
	f[1] = byte(t)
}

func (f Flag) IsOneway() bool {
	return f[2]&0x01 == 0x01
}

func (f *Flag) SetOneway(oneway bool) {
	if oneway {
		f[2] = f[2] | 0x01
	} else {
		f[2] = f[2] &^ 0x01
	}
}

func (f Flag) IsHeartbeat() bool {
	return f[2]&0x02 == 0x02
}

func (f *Flag) SetHeartbeat(hb bool) {
	if hb {
		f[2] = f[2] | 0x02
	} else {
		f[2] = f[2] &^ 0x02
	}
}

func (f Flag) IsBroadCast() bool {
	return f[3]&0x01 == 0x01
}

func (f *Flag) SetBroadCast(y bool) {
	if y {
		f[3] = f[3] | 0x01
	} else {
		f[3] = f[3] &^ 0x01
	}
}

func (f Flag) IsMultiCast() bool {
	return f[3]&0x02 == 0x02
}

func (f *Flag) SetMultiCast(y bool) {
	if y {
		f[3] = f[3] | 0x02
	} else {
		f[3] = f[3] &^ 0x02
	}
}

// MsgPayload 消息
type MsgPayload struct {
	Seq     uint64 // 防重放攻击 client 从0开始 +1递增
	FromUID uint64
	ToUID   uint64
	// login-token
	Flag
	PayloadName string
	Payload     []byte
}

func NewMsgPayload() *MsgPayload {
	p := MsgPayload{}
	return &p
}

func (p *MsgPayload) ReadFrom(r io.Reader) (err error) {
	err = binary.Read(r, binary.BigEndian, &p.Seq)
	if err != nil {
		err = errors.Wrap(err, "decode Seq")
		return
	}

	err = binary.Read(r, binary.BigEndian, &p.FromUID)
	if err != nil {
		err = errors.Wrap(err, "decode FromUID")
		return
	}

	err = binary.Read(r, binary.BigEndian, &p.ToUID)
	if err != nil {
		err = errors.Wrap(err, "decode ToUID")
		return
	}

	err = binary.Read(r, binary.BigEndian, &p.Flag)
	if err != nil {
		err = errors.Wrap(err, "decode Flag")
		return
	}

	var lenPayloadName uint32
	err = binary.Read(r, binary.BigEndian, &lenPayloadName)
	if err != nil {
		err = errors.Wrap(err, "decode lenPayloadName")
		return
	}
	if lenPayloadName == 0 || lenPayloadName > 100 {
		err = errors.Errorf("PayloadName len(%d) bad", lenPayloadName)
		return
	}

	payloadName := make([]byte, lenPayloadName)
	_, err = io.ReadFull(r, payloadName)
	if err != nil {
		err = errors.Wrap(err, "decode PayloadName")
		return
	}
	p.PayloadName = string(payloadName)

	var lenPayload uint32
	err = binary.Read(r, binary.BigEndian, &lenPayload)
	if err != nil {
		err = errors.Wrap(err, "decode lenPayload")
		return
	}

	// if lenPayload == 0 {
	// 	return nil
	// }
	p.Payload = make([]byte, lenPayload)
	_, err = io.ReadFull(r, p.Payload)
	if err != nil {
		err = errors.Wrap(err, "decode Payload")
		return
	}

	if p.GetCompressType() == CtGzip {
		var err error
		p.Payload, err = util.Unzip(p.Payload)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *MsgPayload) Decode(data []byte) error {
	r := bytes.NewReader(data)
	return p.ReadFrom(r)
}

func (p MsgPayload) WriteTo(w io.Writer) error {
	data, err := p.Encode()
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (p *MsgPayload) Encode() (buf []byte, err error) {
	if p.GetCompressType() == CtGzip {
		var err error
		p.Payload, err = util.Zip(p.Payload)
		if err != nil {
			return nil, err
		}
	}
	var lenTotal int64

	lenTotal = int64(8 + 8 + 8 + 8 + 4 + len(p.PayloadName) + 4 + len(p.Payload))

	buf = make([]byte, lenTotal)
	offset := 0

	binary.BigEndian.PutUint64(buf[offset:], p.Seq)
	offset += 8

	binary.BigEndian.PutUint64(buf[offset:], p.FromUID)
	offset += 8

	binary.BigEndian.PutUint64(buf[offset:], p.ToUID)
	offset += 8

	copy(buf[offset:], p.Flag[:])
	offset += 8

	// PayloadName
	binary.BigEndian.PutUint32(buf[offset:], uint32(len(p.PayloadName)))
	offset += 4
	copy(buf[offset:], p.PayloadName)
	offset += len(p.PayloadName)

	// Payload
	binary.BigEndian.PutUint32(buf[offset:], uint32(len(p.Payload)))
	offset += 4
	copy(buf[offset:], p.Payload)
	offset += len(p.Payload)

	return buf, nil
}
