package game

import (
	//	"bytes"
	"encoding/json"
	//	"net/http"

	//	flatbuffers "github.com/google/flatbuffers/go"
	"mahjong.go/config"
	//	fbsInfo "mahjong.go/fbs/info"

	"mahjong.go/library/core"
)

// 推送消息给客户端
/*
func httpPostPush(userId int, msg string, deviceToken string) {
	url, ok := core.AppConfig.PushUrl
	if !ok {
		core.Logger.Error("push_url未定义")
		return
	}

	builder := flatbuffers.NewBuilder(0)
	content := builder.CreateString(msg)
	deviceTokenFbs := builder.CreateString(deviceToken)
	fbsInfo.PushRequestStart(builder)
	fbsInfo.PushRequestAddUserId(builder, uint32(userId))
	fbsInfo.PushRequestAddContent(builder, content)
	fbsInfo.PushRequestAddDeviceToken(builder, deviceTokenFbs)
	orc := fbsInfo.PushRequestEnd(builder)
	builder.Finish(orc)
	buf := builder.FinishedBytes()

	//	http.Post(strings.TrimSpace(url.(string)), "application/octet-stream", bytes.NewReader(buf))
	_, error := http.Post(url.(string), "application/octet-stream", bytes.NewReader(buf))
	if error != nil {
		core.Logger.Warn("推送消息失败，userId: %d, msg: %s, deviceToken:%s", userId, msg, deviceToken)
	} else {
		core.Logger.Debug("推送消息成功，userId: %d, msg: %s, deviceToken:%s", userId, msg, deviceToken)
	}
}
*/

// 通过写redis队列的方式，异步发送推送
func redisListPush(data map[string]interface{}) {
	var byt, _ = json.Marshal(data)
	deviceToken := data["deviceToken"].(string)
	core.RedisDo(core.RedisClient4, "lpush", getPushCacheKey(deviceToken), string(byt))
}

// 获取推送的redisKey
func getPushCacheKey(deviceToken string) string {
	// if strings.ToLower(deviceToken) == "android" {
	// 	return config.CACHE_KEY_PUSH_QUEUE_LIST_ANDROID
	// }
	return config.CACHE_KEY_PUSH_QUEUE_LIST
}
