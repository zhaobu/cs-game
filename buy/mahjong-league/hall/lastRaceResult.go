package hall

import (
	"mahjong-league/protocal"

	"github.com/fwhappy/util"
)

// LastRaceResult 最后比赛结果
type LastRaceResult struct {
	T int64 // 记录生成时间
	P *protocal.ImPacket
}

// NewLastRaceResult 生成一个新的比赛结果
func NewLastRaceResult(packet *protocal.ImPacket) *LastRaceResult {
	return &LastRaceResult{
		T: util.GetTime(),
		P: packet,
	}
}

// 最后结果是否有效
// 保留一小时内的结果
func (lrr *LastRaceResult) isValid() bool {
	return util.GetTime()-lrr.T < int64(3600)
}

// SetLastRaceResult 保存最后结果
func SetLastRaceResult(userId int, packet *protocal.ImPacket) {
	LastRaceResultSet.Store(userId, NewLastRaceResult(packet))
}

// GetLastRaceResult 获取最后有效的比赛结果
func GetLastRaceResult(userId int) *protocal.ImPacket {
	if v, ok := LastRaceResultSet.Load(userId); ok {
		// LastRaceResultSet.Delete(userId)
		result := v.(*LastRaceResult)
		if result.isValid() {
			return result.P
		}
	}
	return nil
}
