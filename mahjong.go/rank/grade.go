package rank

import (
	"fmt"
	"strconv"

	"mahjong.go/library/core"
)

// Grade // 排位等级配置
type Grade struct {
	GradeId       int  // 段位
	GradeLevelMax int  // 最大等级
	StarMax       int  // 最大星星数
	Unlimited     bool // 是否无限星星
	Exp           int  // 每颗星星的经验
}

// FormatGLevel 格式化一个段位等级
func FormatGLevel(gradeId, gradeLevel int) int {
	v, _ := strconv.Atoi(fmt.Sprintf("%d%02d", gradeId, gradeLevel))
	return v
}

// ExplainGLevel 将一个int还原成一个SLevel
func ExplainGLevel(v int) (gradeId, gradeLevel int) {
	s := strconv.Itoa(v)
	if len(s) != 3 {
		core.Logger.Error("[rank.ExplainGLevel]length error, v:%v", s)
		return
	}
	buf := []byte(s)
	gradeId, _ = strconv.Atoi(string(buf[:1]))
	gradeLevel, _ = strconv.Atoi(string(buf[1:3]))

	return
}

// GetPrevGLevel 获取key的前一个段位等级
func GetPrevGLevel(gradeId, gradeLevel int) int {
	// 没有前一个段位等级
	if gradeId == 1 && gradeLevel == 1 {
		return 0
	}
	// 降一个等级
	gradeLevel -= 1
	if gradeLevel == 0 {
		// 如果等级不够降了，则需要开始降段位
		// 并将等级置为下一个段位的最高等级
		gradeId -= 1
		gradeLevel = GradeList[gradeId].GradeLevelMax
	}
	return FormatGLevel(gradeId, gradeLevel)
}
