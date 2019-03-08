package game

import (
	"bytes"
	"fmt"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"mahjong.go/config"
	"mahjong.go/library/core"
)

// 获取保存的数据文件名称
func (p *Playback) getDataOssFileName(isIntact bool) string {
	if isIntact {
		return fmt.Sprintf(config.OSS_PLAYBACK_INTACT_FILE_NAME, p.roomId, p.round)
	}
	return fmt.Sprintf(config.OSS_PLAYBACK_SIMPLE_FILE_NAME, p.roomId, p.round)
}

// 获取保存的数据文件名称
func (p *Playback) getVersionOssFileName() string {
	return fmt.Sprintf(config.OSS_PLAYBACK_VERSION_FILE_NAME, p.roomId, p.round)
}

// 保存回放数据
// 回放数据会有一些新旧版本兼容的问题，所以在保存回放数据的时候，顺便保留一份版本数据，用于判断版本是否匹配
func (p *Playback) saveToOss(data []byte, isIntact bool) {
	endpoint, bucketName, accessKeyId, accessKeySecret := core.GetOssCfg()
	if endpoint == "" || accessKeyId == "" || accessKeySecret == "" {
		core.Logger.Error("[saveToOss]oss配置丢失, endpoint:%v,accessKeyId:%v, accessKeySecret:%v",
			endpoint, accessKeyId, accessKeySecret)
	}
	client, err := oss.New(endpoint, accessKeyId, accessKeySecret)
	if err != nil {
		core.Logger.Error("[saveToOss]连接oss失败:%v", err.Error())
		return
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		core.Logger.Error("[saveToOss]selct bucket失败:%v", err.Error())
		return
	}
	// 上传版本文件
	err = bucket.PutObject(p.getVersionOssFileName(), bytes.NewReader([]byte(GameVersion)))
	if err != nil {
		core.Logger.Error("[saveToOss]上传版本文件失败:%v", err.Error())
		return
	}
	// 上传数据文件
	err = bucket.PutObject(p.getDataOssFileName(isIntact), bytes.NewReader(data))
	if err != nil {
		core.Logger.Error("[saveToOss]上传数据文件失败:%v", err.Error())
		return
	}
	core.Logger.Info("[savePlayback][oss]roomId:%v,round:%v,isIntact:%v", p.roomId, p.round, isIntact)
}
