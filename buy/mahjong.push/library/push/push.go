package push

import (
	"errors"

	"mahjong.push/library/core"
)

const (
	APNS    = "ios"
	MI      = "mi"
	ANDROID = "android"
)

// Push 结构
type Push struct {
	SenderID       int    `json:"senderId"`          // 发送者id, 如果是系统发的, 此值为0
	SenderNickname string `json:"senderNickname"`    // 发送者名称, 如果是系统发的, 此值为空字符串
	LangID         int    `json:"langId"`            // 所属语言包id
	Content        string `json:"content,omitempty"` // 所属语言包内容
	Device         string `json:"device"`            // push类型, ios|android
	DeviceToken    string `json:"deviceToken"`       // 小米:regID, ios:device_token
	Time           int64  `json:"time"`              // push产生time
}

// Verification 验证数据有效性
func (p *Push) Verification() error {
	if p.Device == "" {
		return errors.New("device为空")
	}
	if p.Device == "" || p.DeviceToken == "" {
		return errors.New("deviceToken为空")
	}
	if p.LangID == 0 && p.Content == "" {
		return errors.New("device和content不能同时为空")
	}
	if p.Device != APNS && p.Device != ANDROID {
		return errors.New("未支持的消息类型:" + p.Device)
	}

	return nil
}

// GetMessage 读取消息内容
func (p *Push) GetMessage() string {
	if p.Content != "" {
		return p.Content
	} else if p.LangID > 0 {
		return core.GetLang(p.LangID, p.SenderNickname)
	}
	return ""
}
