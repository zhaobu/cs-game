package game

import (
	"github.com/fwhappy/util"
	"mahjong.go/library/core"
)

// Playback 牌局回放
type Playback struct {
	roomId        int64 // 房间id
	round         int   // 当前局
	operationList []*playbackOperation
}

// 定义一个多类型的操作
type playbackOperation struct {
	t               int64
	userOperation   *UserOperation
	clientOpreation *ClientOperation
	userId          int // 决策者用户id
	opList          []*Operation
}

// NewPlayback 新建一个回放容器
func NewPlayback(roomId int64, round int) *Playback {
	playback := &Playback{}
	playback.roomId = roomId
	playback.round = round
	playback.operationList = make([]*playbackOperation, 0)
	return playback
}

// AppendUserOperation 追加回放数据-用户操作
func (p *Playback) appendUserOperation(operation *UserOperation) {
	op := &playbackOperation{
		t:             util.GetTime(),
		userOperation: operation,
	}
	p.operationList = append(p.operationList, op)
}

// AppendClientOperation 添加一个客户端操作
func (p *Playback) appendClientOperation(operation *ClientOperation) {
	op := &playbackOperation{
		t:               util.GetTime(),
		clientOpreation: operation,
	}
	p.operationList = append(p.operationList, op)
}

// 添加一个决策
func (p *Playback) appendOperationPush(userId int, list []*Operation) {
	op := &playbackOperation{
		t:      util.GetTime(),
		userId: userId,
		opList: list,
	}
	p.operationList = append(p.operationList, op)
}

// 保存回放数据
// 回放数据会有一些新旧版本兼容的问题，所以在保存回放数据的时候，顺便保留一份版本数据，用于判断版本是否匹配
func (p *Playback) save(data []byte, isIntact bool) {
	if p.isSaveToRedis() {
		p.saveToRedis(data, isIntact)
	} else {
		// fixme 过渡时期，需要将版本信息也写入redis, 等版本完成后可以删除
		p.saveVersionToRedis()
		// 保存数据至oss
		p.saveToOss(data, isIntact)
	}
}

// 获取回放存储的方式
func (p *Playback) isSaveToRedis() bool {
	if core.AppConfig.PlaybackSaveType == "oss" {
		return false
	}
	return true
}
