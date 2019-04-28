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
func (self *HuLib) normal_hu(cardInfo *PlayerCardInfo) (bool, EmHuType) {
	checkcard := make([]int, 34)
	for card, num := range cardInfo.StackCards {
		if key, ok := var2key[card]; ok {
			checkcard[key] = int(num)
		} else {
			log.Errorf("牌值%v不存在,请检查!!!", card)
		}
	}

	if MHuLib.GetHuInfo(checkcard, 0) {
		return true, HuType_NORMAL
	}
	return false, 0
}

//清一色
func (self *HuLib) qingyise(cardInfo *PlayerCardInfo) (bool, EmHuType) {

	return false, 0
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
func (self *HuLib) CheckHuType(cardInfo *PlayerCardInfo, balanceInfo *PlayserBalanceInfo, huMode EmHuMode) (bool, HuTypeList) {
	huTypeList := HuTypeList{}

	//判断是否有花牌
	if hasHuaCard(cardInfo) {
		log.Errorf("检测胡牌时还有花牌")
		return false, nil
	}
	if self.checkBaseHuHua(cardInfo, balanceInfo, huMode) {
		if ok, hutype := self.normal_hu(cardInfo); ok {
			huTypeList = append(huTypeList, hutype)
		}
	}
	if len(huTypeList) > 0 {
		return true, huTypeList
	}
	return false, nil
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
