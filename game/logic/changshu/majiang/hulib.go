package majiang

//封装胡牌算法
import (
	mjhulib "cy/game/common/mjhulib"
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
		11: 1, 12: 2, 13: 3, 14: 4, 15: 5, 16: 6, 17: 7, 18: 8, 19: 9,
		21: 10, 22: 11, 23: 12, 24: 13, 25: 14, 26: 15, 27: 16, 28: 17, 29: 18,
		31: 19, 32: 20, 33: 21, 34: 22, 35: 23, 36: 24, 37: 25, 38: 26, 39: 27,
		41: 28, 42: 29, 43: 30, 44: 31, 45: 32, 46: 33, 47: 34,
	}
}

type HuTypeList []EmHuType

//普通胡
func (self *HuLib) normalHu(cardInfo *PlayerCardInfo) bool {
	checkcard := make([]int, 34)
	for card, num := range cardInfo.StackCards {
		if key, ok := var2key[card]; ok {
			checkcard[key] = int(num)
		} else {
			log.Errorf("牌值%v不存在,请检查!!!", card)
		}
	}

	if MHuLib.GetHuInfo(checkcard, 0) {
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
	for color, _ := range allColor {
		return color < 4
	}
	return false
}

//混一色
func (self *HuLib) hunYiSe(allColor map[int32]bool) bool {
	return len(allColor) == 2
}

//字一色
func (self *HuLib) ziYiSe(allColor map[int32]bool) bool {
	if len(allColor) != 1 {
		return false
	}
	for color, _ := range allColor {
		return color == 4
	}
	return false
}

//门清
func (self *HuLib) menQing(cardInfo *PlayerCardInfo) bool {
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

func HasHuModeTag(huModeTags []EmHuModeTag, tag EmHuModeTag) bool {
	for _, v := range huModeTags {
		if v == tag {
			return true
		}
	}
	return false
}

// 胡牌牌型
func (self *HuLib) CheckHuType(cardInfo *PlayerCardInfo, balanceInfo *PlayserBalanceInfo, huMode EmHuMode, huModeTags []EmHuModeTag) (bool, HuTypeList) {
	huTypeList := HuTypeList{}

	//判断是否有花牌
	if hasHuaCard(cardInfo) {
		log.Errorf("检测胡牌时还有花牌")
		return false, nil
	}
	baseHu := false
	//先检查能基本胡
	if self.normalHu(cardInfo) {
		baseHu = true
	}
	//普通话检查花数
	if self.checkBaseHuHua(cardInfo, balanceInfo, huMode) && baseHu {
		huTypeList = append(huTypeList, HuType_Normal)
	}
	//再判断能否特殊胡
	if baseHu {
		if self.menQing(cardInfo) { //门清
			huTypeList = append(huTypeList, HuType_MenQing)
		}

		allColor := getColorCount(cardInfo)
		if self.qingYiSe(allColor) { //清一色
			huTypeList = append(huTypeList, HuType_QingYiSe)
		} else if self.hunYiSe(allColor) { //混一色
			huTypeList = append(huTypeList, HuType_HunYiSe)
		} else if self.ziYiSe(allColor) { //字一色
			huTypeList = append(huTypeList, HuType_ZiYiSe)
		}

		if self.duiDuiHu(cardInfo) { //对对胡
			huTypeList = append(huTypeList, HuType_DuiDuiHu)
		}
		if HasHuModeTag(huModeTags, HuModeTag_GangShangHua) { //杠上花
			huTypeList = append(huTypeList, HuType_GangShangKaiHua)
		}
		if HasHuModeTag(huModeTags, HuModeTag_QiangGangHu) { //抢杠胡
			huTypeList = append(huTypeList, HuType_QiangGangHu)
		}
		if self.daDiaoChe(cardInfo) { //大吊车
			huTypeList = append(huTypeList, HuType_DaDiaoChe)
		}
		if HasHuModeTag(huModeTags, HuModeTag_HaiDiLaoYue) { //海底捞月
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
