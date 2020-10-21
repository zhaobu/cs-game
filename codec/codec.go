package codec

import (
	"bytes"
	"game/codec/protobuf"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/golang/protobuf/proto"
)

const (
	// MaxPacketLen 最大包长度
	MaxPacketLen = (1024 * 1024 * 8)
	// MaxMessageCount 包中最多消息个数
	MaxMessageCount = 64
)

// | 4byte | 4byte | 2byte | 4byte | 变长 |   4byte   | 变长
//   total  reserve  count  nameLen  name  payloadLen payload

// Packet 包
type Packet struct {
	total   uint32
	reserve uint32
	Msgs    []*Message
}

// Message 消息
type Message struct {
	Name    string
	Payload []byte
	UserID  uint64 `json:"-"` // 不做序列化处理
}

// NewPacket NewPacket
func NewPacket() *Packet {
	return &Packet{Msgs: []*Message{}}
}

// ReadFrom ReadFrom
func (p *Packet) ReadFrom(r io.Reader) (err error) {
	err = binary.Read(r, binary.BigEndian, &p.total)
	if err != nil {
		return
	}

	if p.total > MaxPacketLen {
		return fmt.Errorf("len out of range %d", p.total)
	}
	// todo 小于判断

	buf := make([]byte, p.total)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return
	}

	rb := bytes.NewReader(buf)

	err = binary.Read(rb, binary.BigEndian, &p.reserve)
	if err != nil {
		return
	}

	var pktCnt uint16
	err = binary.Read(rb, binary.BigEndian, &pktCnt)
	if err != nil {
		return
	}
	if pktCnt == 0 {
		return fmt.Errorf("pktCnt = 0")
	}
	if pktCnt > MaxMessageCount {
		return fmt.Errorf("MessageCount %d too big", pktCnt)
	}

	var idx uint16
	for ; idx < pktCnt; idx++ {
		var nameLen uint32
		err = binary.Read(rb, binary.BigEndian, &nameLen)
		if err != nil {
			return
		}

		nameBuf := make([]byte, nameLen)
		_, err = io.ReadFull(rb, nameBuf)
		if err != nil {
			return
		}

		var payloadLen uint32
		err = binary.Read(rb, binary.BigEndian, &payloadLen)
		if err != nil {
			return
		}

		msg := &Message{}
		msg.Name = string(nameBuf)
		msg.Payload = make([]byte, payloadLen)

		_, err = io.ReadFull(rb, msg.Payload)
		if err != nil {
			return
		}

		p.Msgs = append(p.Msgs, msg)
	}

	// if pktCnt != uint16(len(p.Msgs)) {
	// 	err = fmt.Errorf("pktCnt err")
	// 	return
	// }

	return
}

// Decode Decode
func (p *Packet) Decode(data []byte) error {
	br := bytes.NewReader(data)
	return p.ReadFrom(br)
}

// | 4byte | 4byte | 2byte | 4byte | 变长 |   4byte   | 变长
//   total  reserve  count  nameLen  name  payloadLen payload
// Encode Encode
func (p *Packet) Encode() (data []byte, err error) {
	if len(p.Msgs) == 0 {
		err = fmt.Errorf("no msgs")
		return
	}

	if len(p.Msgs) > MaxMessageCount {
		err = fmt.Errorf("too many msgs %d", len(p.Msgs))
		return
	}

	var vLen uint64
	for _, msg := range p.Msgs {
		//nameLen
		vLen += 4
		if len(msg.Name) > math.MaxUint32 {
			err = fmt.Errorf("%s too long", msg.Name)
			return
		}
		//name
		vLen += uint64(len(msg.Name))
		//payloadLen
		vLen += 4
		if len(msg.Payload) > math.MaxUint32 {
			err = fmt.Errorf("%s payload too long", msg.Name)
			return
		}
		//payload
		vLen += uint64(len(msg.Payload))
	}

	//reserve+count+vlen
	var total = 4 + 2 + vLen
	if total > math.MaxUint32 {
		err = fmt.Errorf("total %d to long", total)
		return
	}

	p.total = uint32(total)

	data = make([]byte, 4+p.total)
	var offset uint32

	//按大端方式构建数据 total  reserve  count  nameLen  name  payloadLen payload
	binary.BigEndian.PutUint32(data[offset:], p.total) //total
	offset += 4

	binary.BigEndian.PutUint32(data[offset:], p.reserve) //reserve
	offset += 4

	binary.BigEndian.PutUint16(data[offset:], uint16(len(p.Msgs))) //count
	offset += 2

	for _, msg := range p.Msgs {
		binary.BigEndian.PutUint32(data[offset:], uint32(len(msg.Name)))
		offset += 4

		copy(data[offset:], msg.Name)
		offset += uint32(len(msg.Name))

		binary.BigEndian.PutUint32(data[offset:], uint32(len(msg.Payload)))
		offset += 4

		copy(data[offset:], msg.Payload)
		offset += uint32(len(msg.Payload))
	}

	return
}

func (p *Packet) WriteTo(w io.Writer) error {
	data, err := p.Encode()
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func Pb2Msg(pb proto.Message, oMsg *Message) (err error) {
	oMsg.Name, oMsg.Payload, err = protobuf.Marshal(pb)
	return
}

func Msg2Pb(msg *Message) (pb proto.Message, err error) {
	pb, err = protobuf.Unmarshal(msg.Name, msg.Payload)
	return
}
