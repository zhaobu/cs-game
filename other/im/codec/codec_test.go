package codec

import (
	"crypto/aes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"cy/other/im/codec/protobuf"
	"cy/other/im/pb"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"
)

func Test_codecJSON(t *testing.T) {

	pay := NewMsgPayload()
	pay.Seq = 1
	pay.FromUID = 1001
	pay.ToUID = 1002
	//pay.Flag = 0

	sendMsgReq := &impb.SendMsgReq{}
	sendMsgReq.Seq = 999
	sendMsgReq.From = 1001
	sendMsgReq.To = 1002
	sendMsgReq.Content = []byte(`123456`)

	pay.PayloadName, pay.Payload, _ = protobuf.Marshal(sendMsgReq)

	testdata, _ := json.Marshal(pay)
	//testdata, _ := pay.Encode()

	ctx := &MessageCtx{}
	ctx.NeedCrypto = false
	ctx.NeedMAC = false

	ctx.AesKey = make([]byte, 32)
	io.ReadFull(rand.Reader, ctx.AesKey)
	ctx.AesIV = make([]byte, aes.BlockSize)
	io.ReadFull(rand.Reader, ctx.AesIV)
	hmacKey := make([]byte, 16)
	io.ReadFull(rand.Reader, hmacKey)
	ctx.H = hmac.New(sha256.New, hmacKey)

	m := NewMessage(ctx)
	m.Payload = testdata

	m.Encrypto()

	jsonData, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	///////////////////////

	m2 := NewMessage(ctx)
	json.Unmarshal(jsonData, &m2)

	jsonData2, _ := json.Marshal(m2)
	t.Error(fmt.Sprintf("%s", string(jsonData2)))

	//pay2 := NewMsgPayload()
	//pay2.Decode(m2.Payload)
	//t.Error(fmt.Sprintf("%+v", *pay2))

}

func Test_codec(t *testing.T) {
	testdata := []byte(`123`)

	ctx := &MessageCtx{}
	ctx.NeedCrypto = false
	ctx.NeedMAC = false

	ctx.AesKey = make([]byte, 32)
	io.ReadFull(rand.Reader, ctx.AesKey)
	ctx.AesIV = make([]byte, aes.BlockSize)
	io.ReadFull(rand.Reader, ctx.AesIV)
	hmacKey := make([]byte, 16)
	io.ReadFull(rand.Reader, hmacKey)
	ctx.H = hmac.New(sha256.New, hmacKey)

	m := NewMessage(ctx)
	m.Payload = testdata

	m.Encrypto()
	data, err := m.Encode()
	if err != nil {
		t.Error(err)
		return
	}

	///////////////////////

	m2 := NewMessage(ctx)
	err = m2.Decode(data)
	if err != nil {
		t.Error(err)
		return
	}
	m2.Decrypto()

	if !reflect.DeepEqual(*m, *m2) {
		t.Errorf("\nm: %+v\n \nm2: %+v\n", *m, *m2)
	}

}
