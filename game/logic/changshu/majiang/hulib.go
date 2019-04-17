package majiang

//封装胡牌算法
import (
	mjhulib "cy/game/common/mjhulib"
)

var (
	MHuLib *mjhulib.HuLib
)

type HuLib struct {
}

func init() {
	MHuLib = mjhulib.GetSingleton().HuLib
}

type HuTypeList []EmHuType

func (self *HuLib) normal_hu(cardInfo *PlayerCardInfo) (bool, EmHuType) {
	var var2key = map[int32]int{
		11: 1, 12: 2, 13: 3, 14: 4, 15: 5, 16: 6, 17: 7, 18: 8, 19: 9,
		21: 10, 22: 11, 23: 12, 24: 13, 25: 14, 26: 15, 27: 16, 28: 17, 29: 18,
		31: 19, 32: 20, 33: 21, 34: 22, 35: 23, 36: 24, 37: 25, 38: 26, 39: 27,
		41: 28, 42: 29, 43: 30, 44: 31, 45: 32, 46: 33, 47: 34,
	}
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

func hasHuaCard(cardInfo *PlayerCardInfo) bool {
	for k, _ := range cardInfo.StackCards {
		if IsHuaCard(k) {
			return true
		}
	}
	return false
}

// 胡牌牌型
func (self *HuLib) CheckHuType(cardInfo *PlayerCardInfo) (bool, HuTypeList) {
	huTypeList := HuTypeList{}

	//判断是否有花牌
	if hasHuaCard(cardInfo) {
		return false, nil
	}
	if ok, hutype := self.normal_hu(cardInfo); ok {
		huTypeList = append(huTypeList, hutype)
	}
	if len(huTypeList) > 0 {
		return true, huTypeList
	}
	return false, nil
}
