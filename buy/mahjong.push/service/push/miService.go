package push

import (
	"context"
	"fmt"

	xiaomipush "github.com/yilee/xiaomi-push"
	"mahjong.push/library/core"
	"mahjong.push/library/push"
	"mahjong.push/library/util"
)

// 小米push
func sendMI(message *push.Push) error {
	core.Logger.Debug("deviceToken:%v,msg:%v", message.DeviceToken, message.GetMessage())

	var client = xiaomipush.NewClient(core.GetAppConfig("mi_secret_key").(string), []string{core.GetAppConfig("mi_package").(string)})
	var msg = xiaomipush.NewAndroidMessage("地道贵州麻将", message.GetMessage()).SetPayload("")

	// 设置呼起应用
	msg.SetLauncherActivity()
	// 设置当应用开启时, 不接收消息
	// msg.AddExtra("notify_foreground", "1")

	sendTime := util.GetTimestamp()
	result, err := client.Send(context.Background(), msg, message.DeviceToken)
	if err != nil {
		return fmt.Errorf("推送失败, deviceToken:%v,msg:%v,err:%v", message.DeviceToken, message.GetMessage(), err.Error())
	}

	core.Logger.Info("推送成功,ID:%v, createTime:%v, sendTime:%v, completeTime:%v", result.Data.ID, util.FormatUnixTime(message.Time), sendTime, util.GetTimestamp())
	return nil
}
