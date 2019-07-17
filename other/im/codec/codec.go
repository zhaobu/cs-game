package codec

import (
	"bytes"
	"crypto/hmac"
	"encoding/binary"
	"fmt"
	"hash"
	"io"

	aescrc "cy/other/im/common/crypto/aes"
)

const (
	MaxMsgLen = (1024*1024*8 + 32)
)

type Message struct {
	Total   uint64 // len(mac) + len(payload)
	Mac     []byte
	Payload []byte

	ctx *MessageCtx `json:"-"`
}

type MessageCtx struct {
	NeedCrypto bool
	NeedMAC    bool

	AesKey []byte    // 32
	AesIV  []byte    // 16
	H      hash.Hash // 16
}

func NewMessage(ctx *MessageCtx) *Message {
	m := Message{}
	m.ctx = ctx

	if m.ctx.NeedMAC {
		m.Mac = make([]byte, 32)
	}
	return &m
}

func (m *Message) ReadFrom(r io.Reader) (err error) {
	err = binary.Read(r, binary.BigEndian, &m.Total)
	if err != nil {
		return
	}

	// check total
	if m.Total > MaxMsgLen {
		return fmt.Errorf("len out of range %d", m.Total)
	}

	if m.ctx.NeedMAC {
		_, err = io.ReadFull(r, m.Mac)
		if err != nil {
			return
		}
	}

	payloadLen := m.Total - uint64(len(m.Mac))
	m.Payload = make([]byte, payloadLen)
	_, err = io.ReadFull(r, m.Payload)

	return
}

func (m *Message) Decode(data []byte) error {
	r := bytes.NewReader(data)
	return m.ReadFrom(r)
}

func (m Message) WriteTo(w io.Writer) error {
	data, err := m.Encode()
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (m Message) Encode() (data []byte, err error) {
	total := uint64(len(m.Mac) + len(m.Payload))

	if total > MaxMsgLen {
		return nil, fmt.Errorf("len out of range %d", total)
	}

	buf := make([]byte, 8+total)
	offset := 0

	binary.BigEndian.PutUint64(buf, total)
	offset += 8

	if m.ctx.NeedMAC {
		copy(buf[offset:], m.Mac)
		offset += len(m.Mac)
	}

	copy(buf[offset:], m.Payload)
	offset += len(m.Payload)

	return buf, nil
}

func (m *Message) Encrypto() (err error) {
	if m.ctx.NeedCrypto {
		ciphertext, err := aescrc.AesCbcEncrypto(m.Payload, m.ctx.AesKey, m.ctx.AesIV)
		if err != nil {
			return err
		}
		m.Payload = ciphertext
	}

	if m.ctx.NeedMAC {
		m.ctx.H.Reset()
		_, err = m.ctx.H.Write(m.Payload)
		if err != nil {
			return err
		}
		m.Mac = m.ctx.H.Sum(nil)
	}

	return nil
}

func (m *Message) Decrypto() (err error) {
	if m.ctx.NeedMAC {
		m.ctx.H.Reset()
		_, err = m.ctx.H.Write(m.Payload)
		if err != nil {
			return err
		}

		expectedMAC := m.ctx.H.Sum(nil)
		if !hmac.Equal(m.Mac, expectedMAC) {
			return fmt.Errorf("mac err")
		}
	}

	if m.ctx.NeedCrypto {
		plaintext, err := aescrc.AesCbcDecrypto(m.Payload, m.ctx.AesKey, m.ctx.AesIV)
		if err != nil {
			return err
		}
		m.Payload = plaintext
	}

	return nil
}
