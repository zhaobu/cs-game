package oc

import (
	fbsCommon "mahjong.go/fbs/Common"
)

// IsKongOperation 判断是否“杠”的操作, 包括: 暗杠、明杠、转弯杠、憨包杠
func IsKongOperation(opCode int) bool {
	if opCode == fbsCommon.OperationCodeKONG ||
		opCode == fbsCommon.OperationCodeKONG_DARK ||
		IsKongTurnOperation(opCode) {
		return true
	}
	return false
}

// IsKongTurnOperation 是否是转弯杠、憨包杠操作
func IsKongTurnOperation(opCode int) bool {
	if opCode == fbsCommon.OperationCodeKONG_TURN ||
		opCode == fbsCommon.OperationCodeKONG_TURN_FREE {
		return true
	}
	return false
}

// IsWinOperation 判断是否"胡"或者"自摸"的操作
func IsWinOperation(opCode int) bool {
	return IsHuOperation(opCode) || IsZMOperation(opCode)
}

// IsHuOperation 是否“胡”操作, 包括: 胡、抢杠、热炮
func IsHuOperation(opCode int) bool {
	return opCode == fbsCommon.OperationCodeWIN || opCode == fbsCommon.OperationCodeWIN_AFTER_KONG_TURN || opCode == fbsCommon.OperationCodeWIN_AFTER_KONG_PLAY
}

// IsZMOperation 判断是否“自摸”的操作, 包括：自摸, 杠上开花
func IsZMOperation(opCode int) bool {
	return opCode == fbsCommon.OperationCodeWIN_SELF || opCode == fbsCommon.OperationCodeWIN_AFTER_KONG_DRAW
}

// IsDrawOperation 判断是否“抓牌”的操作，包括：从前抓、从后杠
func IsDrawOperation(opCode int) bool {
	return opCode == fbsCommon.OperationCodeDRAW || opCode == fbsCommon.OperationCodeDRAW_AFTER_KONG
}

// IsPongOperation 判断是否“碰”的操作
func IsPongOperation(opCode int) bool {
	return opCode == fbsCommon.OperationCodePONG
}

// IsFlowerOperation 判断是否“补花”的操作
func IsFlowerOperation(opCode int) bool {
	return opCode == fbsCommon.OperationCodeFLOWER_CHANGE
}

// IsPlayOperation 判断是否属于“出牌”的操作
// 包括出牌和报听
func IsPlayOperation(opCode int) bool {
	return opCode == fbsCommon.OperationCodePLAY || opCode == fbsCommon.OperationCodeBAO_TING
}

// IsPassCancelRemain 所有应该被pass_cancel的操作
func IsPassCancelRemain(opCode int) bool {
	return opCode == fbsCommon.OperationCodePLAY ||
		opCode == fbsCommon.OperationCodeTING ||
		opCode == fbsCommon.OperationCodePLAY_SUGGEST ||
		opCode == fbsCommon.OperationCodeROBOT_PLAY_SUGGEST ||
		opCode == fbsCommon.OperationCodeTING_PLAY_SUGGEST
}

// IsLackOperation 判断是否属于“出牌”的操作
// 包括出牌和报听
func IsLackOperation(opCode int) bool {
	return opCode == fbsCommon.OperationCodeNEED_LACK_TILE
}
