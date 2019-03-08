package push

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
	"mahjong.push/library/core"
	"mahjong.push/library/push"
	"mahjong.push/library/util"
)

var client *apns2.Client

func getClient() error {
	if client == nil {
		cert, err := certificate.FromP12File(core.GetApnsP12File(), core.GetApnsP12Passwd())
		if err != nil {
			return fmt.Errorf("解析apns p12文件失败:%v", err.Error())
		}
		client = apns2.NewClient(cert).Production()
	}
	return nil
}

func sendAPNS(message *push.Push) error {
	err := getClient()
	if err != nil {
		return err
	}

	// 换算出原始deviceToken
	deviceToken := getDeviceToken(message.DeviceToken)
	core.Logger.Debug("deviceToken:%v", deviceToken)
	// 构建推送对象
	notification := &apns2.Notification{}
	notification.DeviceToken = deviceToken
	notification.Topic = "com.maxpanda.gzmahjong"
	// notification.Payload = []byte(`{"aps":{"alert":"Hello, world!"}}`) // See Payload section below
	notification.Payload = payload.NewPayload().Alert(message.GetMessage())
	// 记录开始发送时间
	sendTime := util.GetTimestamp()
	// 发送
	// res, err := client.Push(notification)
	// 带超时时间的发送
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	res, err := client.PushWithContext(ctx, notification)
	defer cancel()
	if err != nil {
		return fmt.Errorf("推送失败, deviceToken:%v,msg:%v,err:%v", message.DeviceToken, message.GetMessage(), err.Error())
	}

	if res.StatusCode != 200 {
		return fmt.Errorf("推送失败, deviceToken:%v,msg:%v,code:%v,err:%v", message.DeviceToken, message.GetMessage(), res.StatusCode, res.Reason)
	}

    fmt.Printf("%v,%v,%v\n", res.StatusCode, res.Reason, res.Timestamp)
	core.Logger.Info("推送成功,ID:%v, code:%v, createTime:%v, sendTime:%v, completeTime:%v, deviceToken:%v", res.ApnsID, res.Reason, util.FormatUnixTime(message.Time), sendTime, util.GetTimestamp(), message.DeviceToken)
	return nil
}

func getDeviceToken(fromToken string) string {
	bs, _ := base64.StdEncoding.DecodeString(fromToken)
	return hex.EncodeToString(bs)
}
