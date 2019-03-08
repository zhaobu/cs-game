package rank

import (
	"fmt"
	"strconv"
)

// FormatRankScore 格式化成排行积分
func FormatRankScore(sLevel, userId int) int64 {
	// 等级的排序是反的，3级比1级的低，所以也需要反转
	sLevel = SLevelRev(sLevel)
	// 反转用户id
	revUserId := 99999999 - userId
	scoreString := fmt.Sprintf("%d%d", sLevel, revUserId)
	score, _ := strconv.ParseInt(scoreString, 10, 64)
	return score
}

// ExplainRankScore 反向解析排行积分
func ExplainRankScore(score int64) {

}

// GetRealDeductScore 计算实际扣分
func GetRealDeductScore(score, gradeId int) int {
	var rate int
	switch gradeId {
	case 6:
		rate = 100
	case 5:
		rate = 50
	case 4:
		rate = 40
	case 3:
		rate = 30
	case 2:
		fallthrough
	case 1:
		rate = 20
	default:
		rate = 20
	}
	return score * rate / 100
}

// CalcRealLevel 计算用户当前经验以及本次是加减分，来判断是否
func CalcRealLevel(currentSLevel, currentExp, exp int) (gradeId, gradeLevel, star, totalExp, starChange int) {
	totalExp = currentExp + exp
	// 防止溢出
	if totalExp < 0 {
		totalExp = 0
	}

	// 当前排位数据
	currentSInfo := RankStarList[currentSLevel]
	// 是否保级
	// unrelegation := currentSInfo.Unrelegation
	// 解开当前的排位等级
	gradeId, gradeLevel, star = ExplainSLevel(currentSLevel)

	if currentSInfo == nil {
		// TOTO 暂时不处理
		fmt.Printf("[CalcRealLevel]数据异常, currentSLevel:%v, currentExp:%v, exp:%v\n", currentSLevel, currentExp, exp)
		return
	}

	if exp > 0 {
		// 加分
		for {
			if totalExp <= currentSInfo.MaxExp {
				break
			}
			nextSlevel := currentSInfo.Next
			currentSInfo = RankStarList[nextSlevel]
			// 没有下一个等级了
			if currentSInfo == nil {
				break
			}
			gradeId, gradeLevel, star = ExplainSLevel(nextSlevel)
			starChange++
		}
	} else if exp < 0 {
		// 扣分
		for {
			if totalExp >= currentSInfo.MinExp {
				break
			}
			prevSlevel := currentSInfo.Prev
			currentSInfo = RankStarList[prevSlevel]
			// 没有上一个等级了
			if currentSInfo == nil {
				break
			}
			gradeId, gradeLevel, star = ExplainSLevel(prevSlevel)
			starChange--
			// // 保级逻辑
			// if unrelegation == 1 {
			// 	// 保级设置为上一级的最高经验
			// 	totalExp = currentSInfo.MaxExp
			// 	break
			// }
			// unrelegation = currentSInfo.Unrelegation
		}
	}
	return
}
