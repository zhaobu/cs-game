package game

import (
	"mahjong.go/config"
)

// 服务端score和客户端score的对应关系
func getFrontScoreType(scoreType int, score int) int {
	var frontScoreType int
	if score < 0 {
		scoreType *= -1
	}
	if v, exists := config.ScoreTypeToFrontMap[scoreType]; exists {
		frontScoreType = v
	}
	return frontScoreType
}

// 判断前端类型，是否需要显示在结算的鸡的界面
func isShowChikenType(frontScoreType int) (isChiken, isGold, isCharge, isBao int) {
	isChiken = 1
	switch frontScoreType {
	case config.FRONT_SCORE_TYPE_JI:
	case config.FRONT_SCORE_TYPE_SILVER_CHIKEN:
	case config.FRONT_SCORE_TYPE_CHARGE_JI:
		isCharge = 1
	case config.FRONT_SCORE_TYPE_JIN_JI:
		isGold = 1
	case config.FRONT_SCORE_TYPE_JIN_CHARGE_JI:
		isCharge = 1
		isGold = 1
	case config.FRONT_SCORE_TYPE_WUGU:
	case config.FRONT_SCORE_TYPE_CHARGE_WUGU:
		isCharge = 1
	case config.FRONT_SCORE_TYPE_JIN_WUGU:
		isGold = 1
	case config.FRONT_SCORE_TYPE_JIN_CHARGE_WUGU:
		isCharge = 1
		isGold = 1
	case config.FRONT_SCORE_TYPE_FANPAI:
	case config.FRONT_SCORE_TYPE_UD:
	case config.FRONT_SCORE_TYPE_FB:
	case config.FRONT_SCORE_TYPE_YIWAI:
	case config.FRONT_SCORE_TYPE_XQ:
	case config.FRONT_SCORE_TYPE_DIAMOND:
	case config.FRONT_SCORE_TYPE_BEN:
	case config.FRONT_SCORE_TYPE_TUMBLING:
	case config.FRONT_SCORE_TYPE_BAOJI:
	case config.FRONT_SCORE_TYPE_STAND_KITCHEN:
	case config.FRONT_SCORE_TYPE_STAND_JIN_KITCHEN:
	case config.FRONT_SCORE_TYPE_FIRST_CYCLE_CHARGE_KITCHEN:
		isCharge = 1
	case config.FRONT_SCORE_TYPE_FIRST_CYCLE_CHARGE_JIN_KITCHEN:
		isCharge = 1
		isGold = 1
	case config.FRONT_SCORE_TYPE_YB:
	case config.FRONT_SCORE_TYPE_PP_KITCHEN:
	case config.FRONT_SCORE_TYPE_FLOWER_RED_KITCHEN:
	case config.FRONT_SCORE_TYPE_FLOWER_RED_DIAMOND_KITCHEN:
	default:
		isChiken = 0
		break
	}
	return
}
