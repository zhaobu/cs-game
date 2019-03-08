package push

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/garyburd/redigo/redis"
	"mahjong.push/library/core"
	"mahjong.push/library/push"
)

// Send 发送消息
func Send(data interface{}) error {
	// 解析数据
	byts, err := redis.Bytes(data, nil)
	if err != nil {
		return fmt.Errorf("从redis中解析出[]byte失败:%v", data)
	}
	var message *push.Push
	err = json.Unmarshal(byts, &message)
	if err != nil {
		return fmt.Errorf("从[]byte中解析出json失败:%v,err:%s", byts, err)
	}
	core.Logger.Debug("data:%v", string(byts))

	// 验证数据有效性
	err = message.Verification()
	if err != nil {
		return fmt.Errorf("push验证失败:%s", err)
	}

	switch strings.ToLower(message.Device) {
	case push.ANDROID:
		return sendMI(message)
	case push.APNS:
		return sendAPNS(message)
	default:
		return fmt.Errorf("未支持的消息类型：%v", message.Device)
	}
}
