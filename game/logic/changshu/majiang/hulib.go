package majiang

//封装胡牌算法
import (
	mjhulib "cy/game/common/mjhulib"
	pbgame_logic "cy/game/pb/game/mj/changshu"
)

var (
	MHuLib  *mjhulib.HuLib
	var2key map[int32]int
)

type HuLib struct {
}

func init() {
	MHuLib = mjhulib.GetSingleton().HuLib
	var2key = map[int32]int{
		11: 0, 12: 1, 13: 2, 14: 3, 15: 4, 16: 5, 17: 6, 18: 7, 19: 8,
		21: 9, 22: 10, 23: 11, 24: 12, 25: 13, 26: 14, 27: 15, 28: 16, 29: 17,
		31: 18, 32: 19, 33: 20, 34: 21, 35: 22, 36: 23, 37: 24, 38: 25, 39: 26,
		41: 27, 42: 28, 43: 29, 44: 30, 45: 31, 46: 32, 47: 33,
	}
}

type HuTypeList []EmHuType

//普通胡
func (self *HuLib) normalHu(cardInfo *PlayerCardInfo, gui_num int) bool {
	checkcard := make([]int, 34)
	for card, num := range cardInfo.StackCards {
		if key, ok := var2key[card]; ok {
			checkcard[key] = int(num)
		} else {
			log.Errorf("牌值%v不存在,请检查!!!", card)
		}
	}

	if MHuLib.GetHuInfo(checkcard, gui_num) {
		return true
	}
	return false
}

//统计颜色种类
func getColorCount(cardInfo *PlayerCardInfo) map[int32]bool {
	allColor := map[int32]bool{}
	//判断吃碰杠
	for _, v := range cardInfo.RiverCards {
		allColor[GetCardColor(v.Card)] = true
	}
	//判断手牌
	for k, _ := range cardInfo.StackCards {
		allColor[GetCardColor(k)] = true
	}
	return allColor
}

//清一色
func (self *HuLib) qingYiSe(allColor map[int32]bool) bool {
	if len(allColor) != 1 {
		return false
	}
	return allColor[1] || allColor[2] || allColor[3]
}

//混一色
func (self *HuLib) hunYiSe(allColor map[int32]bool) bool {
	if len(allColor) != 2 {
		return false
	}
	return allColor[4]
}

//字一色
func (self *HuLib) ziYiSe(allColor map[int32]bool) bool {
	if len(allColor) != 1 {
		return false
	}
	return allColor[4]
}

//门清
func (self *HuLib) menQing(cardInfo *PlayerCardInfo) bool {
	for _, v := range cardInfo.GangCards {
		if v != pbgame_logic.OperType_Oper_AN_GANG { //只有暗杠才算门清
			return false
		}
	}
	return len(cardInfo.ChiCards)+len(cardInfo.PengCards) == 0
}

//对对胡
func (self *HuLib) duiDuiHu(cardInfo *PlayerCardInfo) bool {
	if len(cardInfo.ChiCards) > 0 {
		return false
	}
	pairNum := 0 //一对的数量
	for _, v := range cardInfo.StackCards {
		if pairNum > 1 {
			return false
		}
		if v == 2 {
			pairNum++
		} else if v == 1 || v == 4 {
			return false
		}
	}
	return pairNum == 1
}

//大吊车
func (self *HuLib) daDiaoChe(cardInfo *PlayerCardInfo) bool {
	return len(cardInfo.HandCards) == 2
}

func hasHuaCard(cardInfo *PlayerCardInfo) bool {
	for k, _ := range cardInfo.StackCards {
		if IsHuaCard(k) {
			return true
		}
	}
	return false
}

// 胡牌牌型
func (self *HuLib) CheckHuType(cardInfo *PlayerCardInfo, balanceInfo *PlayserBalanceInfo, huMode EmHuMode, huModeTags map[EmHuModeTag]bool) (bool, HuTypeList) {
	huTypeList := HuTypeList{}

	//判断是否有花牌
	if hasHuaCard(cardInfo) {
		log.Errorf("检测胡牌时还有花牌")
		return false, nil
	}
	baseHu := false
	//先检查能基本胡
	if self.normalHu(cardInfo, 0) {
		baseHu = true
	}
	//普通胡检查花数
	if self.checkBaseHuHua(cardInfo, balanceInfo, huMode) && baseHu {
		huTypeList = append(huTypeList, HuType_Normal)
	}
	//再判断能否特殊胡
	if baseHu {
		allColor := getColorCount(cardInfo)
		if self.qingYiSe(allColor) { //清一色
			huTypeList = append(huTypeList, HuType_QingYiSe)
		} else if self.hunYiSe(allColor) { //混一色
			huTypeList = append(huTypeList, HuType_HunYiSe)
		} else if self.ziYiSe(allColor) { //字一色
			huTypeList = append(huTypeList, HuType_ZiYiSe)
		}
		if self.menQing(cardInfo) { //门清
			huTypeList = append(huTypeList, HuType_MenQing)
		}
		if self.duiDuiHu(cardInfo) { //对对胡
			huTypeList = append(huTypeList, HuType_DuiDuiHu)
		}
		if huModeTags[HuModeTag_GangShangHua] { //杠上花
			huTypeList = append(huTypeList, HuType_GangShangKaiHua)
		}
		if huModeTags[HuModeTag_QiangGangHu] { //抢杠胡
			huTypeList = append(huTypeList, HuType_QiangGangHu)
		}
		if self.daDiaoChe(cardInfo) { //大吊车
			huTypeList = append(huTypeList, HuType_DaDiaoChe)
		}
		if huModeTags[HuModeTag_HaiDiLaoYue] { //海底捞月
			huTypeList = append(huTypeList, HuType_HaiDiLaoYue)
		}
	}
	return len(huTypeList) > 0, huTypeList
}

//检查基本胡花数够不够
func (self *HuLib) checkBaseHuHua(cardInfo *PlayerCardInfo, balanceInfo *PlayserBalanceInfo, huMode EmHuMode) bool {
	huaShu := balanceInfo.GetPingHuHua()
	if huaShu == 0 && huMode == HuMode_ZIMO { //没花的情况下手牌里有两张一样的风牌，可自摸
		for i := 41; i < 47; i++ {
			if cardInfo.StackCards[int32(i)] > 1 {
				return true
			}
		}
	} else if huaShu == 1 && huMode == HuMode_ZIMO { //有一个花的情况下，自摸风牌仍可胡，但点炮不能胡
		card := cardInfo.HandCards[len(cardInfo.HandCards)-1]
		return card >= 41 && card < 47
	} else if huMode == HuMode_ZIMO { //自摸
		return huaShu >= 2
	} else { //点炮
		if huaShu == 2 { //有两个花的情况下，点炮只能胡风牌
			card := cardInfo.HandCards[len(cardInfo.HandCards)-1]
			return card >= 41 && card < 47
		}
		return huaShu >= 3
	}
	return false
}

//检查打出某张牌后能否听
func (self *HuLib) OneCardCanListen(cardInfo *PlayerCardInfo, balanceInfo *PlayserBalanceInfo, huModeTags map[EmHuModeTag]bool) bool {
	huTypeList := HuTypeList{}

	//判断是否有花牌
	if hasHuaCard(cardInfo) {
		log.Errorf("检测胡牌时还有花牌")
		return false
	}
	baseHu := false
	//先检查能基本胡
	if self.normalHu(cardInfo, 1) {
		baseHu = true
	}
	//普通胡检查花数
	if self.checkBaseHuHua(cardInfo, balanceInfo, HuMode_ZIMO) && baseHu {
		huTypeList = append(huTypeList, HuType_Normal)
	}
	//不能普通胡时再检查能否特殊胡,只要能胡一种就可听牌
	if baseHu && len(huTypeList) <= 0 {
		allColor := getColorCount(cardInfo)
		if self.qingYiSe(allColor) { //清一色
			huTypeList = append(huTypeList, HuType_QingYiSe)
		} else if self.hunYiSe(allColor) { //混一色
			huTypeList = append(huTypeList, HuType_HunYiSe)
		} else if self.ziYiSe(allColor) { //字一色
			huTypeList = append(huTypeList, HuType_ZiYiSe)
		} else if self.menQing(cardInfo) { //门清
			huTypeList = append(huTypeList, HuType_MenQing)
		} else if self.duiDuiHu(cardInfo) { //对对胡
			huTypeList = append(huTypeList, HuType_DuiDuiHu)
		} else if huModeTags[HuModeTag_GangShangHua] { //杠上花
			huTypeList = append(huTypeList, HuType_GangShangKaiHua)
		} else if huModeTags[HuModeTag_QiangGangHu] { //抢杠胡
			huTypeList = append(huTypeList, HuType_QiangGangHu)
		} else if self.daDiaoChe(cardInfo) { //大吊车
			huTypeList = append(huTypeList, HuType_DaDiaoChe)
		} else if huModeTags[HuModeTag_HaiDiLaoYue] { //海底捞月
			huTypeList = append(huTypeList, HuType_HaiDiLaoYue)
		}
	}
	return len(huTypeList) > 0
}
