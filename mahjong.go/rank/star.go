package rank

import (
	"fmt"
	"strconv"

	"mahjong.go/library/core"
)

// SLevel 排位等级
type SLevel struct {
	GradeId      int // 段位
	GradeLevel   int // 等级
	Star         int // 星星
	Prev         int // 前一个级别, 0 表示没有前一个等级了
	Next         int // 后一个级别
	MinExp       int // 经验最低值
	MaxExp       int // 经验
	Unrelegation int // 是否保级
}

// FormatSLevel 格式化成一个int
func (s *SLevel) FormatSLevel() int {
	return FormatSLevel(s.GradeId, s.GradeLevel, s.Star)
}

// FormatSLevel 格式化一个段位等级
func FormatSLevel(gradeId, gradeLevel, star int) int {
	v, _ := strconv.Atoi(fmt.Sprintf("%d%02d%05d", gradeId, gradeLevel, star))
	return v
}

// ExplainSLevel 将一个int还原成一个SLevel
func ExplainSLevel(v int) (gradeId, gradeLevel, star int) {
	s := strconv.Itoa(v)
	if len(s) != 8 {
		core.Logger.Error("[rank.ExplainSLevel]length error, v:%v", s)
		return
	}
	buf := []byte(s)
	gradeId, _ = strconv.Atoi(string(buf[:1]))
	gradeLevel, _ = strconv.Atoi(string(buf[1:3]))
	star, _ = strconv.Atoi(string(buf[3:]))

	return
}

// SLevelRev 反转SLevel中的gradeId
func SLevelRev(sLevel int) int {
	gradeId, gradeLevel, star := ExplainSLevel(sLevel)
	// 反转gradeLevel
	gradeLevel = 99 - gradeLevel
	return FormatSLevel(gradeId, gradeLevel, star)
}

// GetSLevel 获取SLevel配置
func GetSLevel(key int) *SLevel {
	if v, ok := RankStarList[key]; ok {
		return v
	}
	return nil
}

// GetPrevSLevel 获取key的前一个key
func GetPrevSLevel(key int) int {
	s := GetSLevel(key)
	if s != nil {
		return s.Prev
	}
	// 表示超出普通的级别，进入不限级别
	return key - 1
}

// GetNextSLevel 获取key的后一个key
func GetNextSLevel(key int) int {
	s := GetSLevel(key)
	if s != nil {
		return s.Next
	}
	// 表示超出普通的级别，进入不限级别
	return key + 1
}

// IsUnlimitedGrade 是否无限星星的等级
// func IsUnlimitedGrade(id int) bool {
// 	return id == 6
// }
