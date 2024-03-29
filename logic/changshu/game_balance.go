package main

//game_balance文件写算分方法
import (
	mj "game/logic/changshu/majiang"
	pbgame_logic "game/pb/game/mj/changshu"
	"math/rand"
	"time"
)

type HuScoreInfo struct {
	mj.HuTypeList //胡牌类型
}

type (
	StartDiceType uint8 //开局色子情况
)

var (
	hitCards    []map[int32]bool      //扳杠头
	huTypeScore map[mj.EmHuType]int32 //牌型分
)

//开局色子类型
const (
	StartDice_None StartDiceType = iota //不加倍
	StartDice_One                       //本局加倍
	StartDice_Two                       //本局和下局加倍
)

//结算信息
type GameBalance struct {
	*mj.RoomLog                          //桌子日志
	game_config  *pbgame_logic.CreateArg //游戏参数
	startDice    StartDiceType           //开局色子
	baozi        int32                   //本局豹子倍数
	loseChair    int32                   //丢分玩家
	gangHuaChair int32                   //杠上花玩家
	huCard       int32                   //胡的牌
	huMode       mj.EmHuMode             //胡牌方式
	huChairs     map[int32]*HuScoreInfo  //胡牌玩家信息
	pghuaShu     []int32                 //碰花,杠花
	duLongHua    int32                   //独龙杠花
	allCards     [][]int32               //扳的所有牌
	hitIndex     [][]int32               //扳到的杠头的索引
	canBaozi     bool                    //建房参数是否能豹子翻倍
}

func init() {
	hitCards = []map[int32]bool{
		0: map[int32]bool{ //庄家
			11: true,
			21: true,
			31: true,
			15: true,
			25: true,
			35: true,
			19: true,
			29: true,
			39: true,
			45: true,
			41: true,
			51: true,
		},
		1: map[int32]bool{ //下家
			14: true,
			24: true,
			34: true,
			18: true,
			28: true,
			38: true,
			44: true,
			54: true,
		},
		2: map[int32]bool{ //对家
			13: true,
			23: true,
			33: true,
			17: true,
			27: true,
			37: true,
			47: true,
			43: true,
			53: true,
		},
		3: map[int32]bool{ //上家
			12: true,
			22: true,
			32: true,
			16: true,
			26: true,
			36: true,
			46: true,
			42: true,
			52: true,
		},
	}
	huTypeScore = map[mj.EmHuType]int32{
		mj.HuType_Normal:          1,
		mj.HuType_MenQing:         5,
		mj.HuType_QingYiSe:        10,
		mj.HuType_ZiYiSe:          15,
		mj.HuType_HunYiSe:         5,
		mj.HuType_DuiDuiHu:        5,
		mj.HuType_GangShangKaiHua: 5,
		mj.HuType_DaDiaoChe:       5,
		mj.HuType_HaiDiLaoYue:     5,
	}
}

func (self *GameBalance) Init(config *pbgame_logic.CreateArg) {
	self.game_config = config
	for _, v := range config.Rule {
		if v.Val == 2 {
			self.canBaozi = true
			break
		}
	}
}

func (self *GameBalance) Reset(curInning uint32) {
	//依据上一局结果判断是否豹子翻倍
	if curInning > 1 && self.canBaozi && (self.huCard == 0 || self.startDice == StartDice_Two) {
		self.baozi = 2
	} else {
		self.baozi = 1
	}
	self.startDice = StartDice_None
	self.loseChair = -1
	self.gangHuaChair = -1
	self.huCard = 0
	self.huMode = mj.HuMode_None
	self.huChairs = make(map[int32]*HuScoreInfo, self.game_config.PlayerCount)
	self.pghuaShu = make([]int32, self.game_config.PlayerCount)
	self.duLongHua = 0
	self.allCards = make([][]int32, self.game_config.PlayerCount) //扳的所有牌
	self.hitIndex = make([][]int32, self.game_config.PlayerCount) //扳到的牌的索引
}

//统计次数
func (self *GameBalance) AddScoreTimes(balanceResult *mj.PlayerBalanceResult, op mj.EmScoreTimes) {
	if num, ok := balanceResult.ScoreTimes[op]; ok {
		balanceResult.ScoreTimes[op] = num + 1
	} else {
		balanceResult.ScoreTimes[op] = 1
	}
}

//处理豹子翻倍
func (self *GameBalance) DealStartDice(randRes [2]int32) {
	if self.canBaozi && randRes[0] == randRes[1] {
		self.baozi = 2
		self.startDice = StartDice_One
		if randRes[0] == 1 || randRes[0] == 4 {
			self.startDice = StartDice_Two
		}
	}
}

//计算杠头数
func (self *GameBalance) CalGangTou(leftCards []int32, bankerId int32) { // 杠头  1 扳4个 2 扳8个 3 独龙杠
	if self.game_config.Barhead == 3 && self.huCard != 0 {
		if len(leftCards) > 0 {
			self.duLongHua = 5
			if mj.GetCardColor(leftCards[len(leftCards)-1]) < 4 {
				self.duLongHua = mj.GetCardValue(int32(leftCards[len(leftCards)-1]))
			}
			leftCards = leftCards[:len(leftCards)-1] //如果是独龙杠,把独龙杠的牌从牌堆去掉
		}
	}
	num := 0
	if self.game_config.Barhead == 2 {
		num = 8
	} else if self.game_config.Barhead == 1 {
		num = 4
	}

	getCanHit := func(chairId int32) map[int32]bool { //获取能中的牌
		for i, j := bankerId, int32(0); j < self.game_config.PlayerCount; i, j = mj.GetNextChair(i, self.game_config.PlayerCount), j+1 {
			if chairId == i {
				return hitCards[j]
			}
		}
		return nil
	}
	count := 0          //计数
	chairId := bankerId //从庄家开始算起数杠头
	for _, v := range leftCards {
		self.allCards[chairId] = append(self.allCards[chairId], v)
		if count < num && getCanHit(chairId)[v] {
			self.hitIndex[chairId] = append(self.hitIndex[chairId], int32(len(self.allCards[chairId])-1))
			count++
		}
		chairId = mj.GetNextChair(chairId, self.game_config.PlayerCount)
	}
	self.Log.Debugf("扳杠头结果:self.duLongHua=%d,\nself.allCards=%+v,\nself.hitIndex=%+v", self.duLongHua, self.allCards, self.hitIndex)
}

//算分
func (self *GameBalance) CalGameBalance(players []*mj.PlayerInfo, bankerId int32) {
	getHuTypeScore := func(huInfo *HuScoreInfo) (score int32) {
		for _, v := range huInfo.HuTypeList {
			if v != mj.HuType_Normal {
				score += huTypeScore[v]
			}
		}
		return
	}
	for winChair, v := range self.huChairs {
		balanceInfo := &players[winChair].BalanceInfo
		//胡牌分
		balanceInfo.HuPoint = 1
		winSocre := 1 + balanceInfo.GetPingHuHua() //胡牌1分+补花+杠花+风花
		//奖码花
		if self.game_config.Barhead == 3 { //独龙杠
			balanceInfo.JiangMaPoint = self.duLongHua
		} else {
			balanceInfo.JiangMaPoint = int32(len(self.hitIndex[winChair]))
		}
		winSocre += balanceInfo.JiangMaPoint
		//特殊牌型花
		balanceInfo.SpecialPoint = getHuTypeScore(v)
		winSocre += balanceInfo.SpecialPoint
		//豹子翻倍
		balanceInfo.Baozi = self.baozi
		winSocre *= self.baozi
		//底飘
		balanceInfo.DiPiaoPoint = int32(self.game_config.Dipiao) * 2
		winSocre += balanceInfo.DiPiaoPoint

		//赢
		if self.huMode == mj.HuMode_ZIMO {
			for i := int32(0); i < self.game_config.PlayerCount; i++ {
				if i == winChair {
					continue
				}
				balanceInfo.Point += winSocre            //赢
				players[i].BalanceInfo.Point -= winSocre //输
				//总分
				players[winChair].BalanceResult.Point += winSocre
				players[i].BalanceResult.Point -= winSocre
			}
		} else {
			for _, huType := range v.HuTypeList { //抢杠胡包三家
				if huType == mj.HuType_QiangGangHu {
					winSocre *= self.game_config.PlayerCount - 1
					break
				}
			}
			balanceInfo.Point += winSocre                         //赢
			players[self.loseChair].BalanceInfo.Point -= winSocre //输
			//总分
			players[winChair].BalanceResult.Point += winSocre
			players[self.loseChair].BalanceResult.Point -= winSocre
		}
	}
}

//小局结算信息
func (self *GameBalance) GetPlayerBalanceInfo(players []*mj.PlayerInfo) (jsonInfo []*pbgame_logic.Json_PlayerBalance_Info) {
	getClientHuType := func(chairId int32) (res []pbgame_logic.HuType) {
		bNormalHu := len(self.huChairs[chairId].HuTypeList) == 1 //有特殊胡时,不算普通胡
		for _, v := range self.huChairs[chairId].HuTypeList {
			if v != mj.HuType_Normal || (v == mj.HuType_Normal && bNormalHu) {
				res = append(res, pbgame_logic.HuType(v))
			}
		}
		return
	}
	getClientScoreType := func(info *mj.PlayserBalanceInfo) map[int32]int32 {
		res := map[int32]int32{
			1: info.Baozi,
			2: 1,
			3: info.BuHuaPoint,
			4: info.SpecialPoint,
			5: info.JiangMaPoint,
			6: info.FengPoint,
			7: info.DiPiaoPoint,
			8: info.GangPoint,
		}
		return res
	}
	for k, v := range players {
		chairId := int32(k)
		info := &pbgame_logic.Json_PlayerBalance_Info{HuMode: pbgame_logic.HuMode_HuModeNone}
		if self.huMode == mj.HuMode_ZIMO {
			if self.huChairs[chairId] != nil {
				info.HuMode = pbgame_logic.HuMode_HuModeZiMo
			}
		} else if self.huMode == mj.HuMode_PAOHU {
			if self.huChairs[chairId] != nil {
				info.HuMode = pbgame_logic.HuMode_HuModeJiePao
			} else if self.loseChair == chairId {
				info.HuMode = pbgame_logic.HuMode_HuModeDianPao
			}
		}
		if self.huChairs[chairId] != nil {
			info.HuType = getClientHuType(chairId)
			info.ScoreType = getClientScoreType(&v.BalanceInfo)
		}
		info.HandCards = v.CardInfo.HandCards
		info.HuaCards = v.CardInfo.HuaCards
		info.Point = v.BalanceInfo.Point
		info.TotalPoint = v.BalanceResult.Point
		info.BanAllCards = self.allCards[chairId]
		info.BanHitIndex = self.hitIndex[chairId]
		jsonInfo = append(jsonInfo, info)
	}
	return
}

//计算下局庄家
func (self *GameBalance) CalNextBankerId(bankerId int32) int32 {
	if self.huChairs[bankerId] != nil { //庄家胡牌可继续连庄
		return bankerId
	} else if len(self.huChairs) == 0 { //荒局后下家坐庄
		return mj.GetNextChair(bankerId, self.game_config.PlayerCount)
	} else { //闲家胡牌后由闲家坐庄
		huId := []int32{}
		for k, _ := range self.huChairs {
			if k != bankerId {
				huId = append(huId, k)
			}
		}
		rand.Seed(int64(time.Now().UnixNano()))
		return huId[rand.Intn(len(huId))]
	}
}
