package game

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/fwhappy/util"
	"mahjong.go/config"
	fbsCommon "mahjong.go/fbs/Common"
	"mahjong.go/library/core"
	"mahjong.go/mi/card"
	"mahjong.go/mi/oc"
	"mahjong.go/rank"
	rankService "mahjong.go/service/rank"
)

type UserScore struct {
	Maps *sync.Map
}
type HuInfo struct {
	// 点炮人信息
	Loser []int
	// 胡牌人的信息
	WInfo           []WinInfo
	Hu              bool
	HuOperationCode int // 胡牌时的操作码
}

type WinInfo struct {
	// 胡牌人的ID
	Id int
	// 胡牌方式
	Way int
	// 胡牌类型
	WType int
	// 清一色标志
	QFlag bool
	// 胡的牌
	Tile int
}

// 用户每局的积分容器
type ScoreInfo struct {
	// 麻将类型
	MahjongType int
	// 积分总值
	Total int
	// 积分条目
	Item []ScoreItem
}

// 单条积分明细
type ScoreItem struct {
	// 这个积分相关用户
	UserId []int
	// 积分条目
	SType int
	// 积分, 可以为负分
	Score int
	// 积分数量
	ScoreCount int
	// 积分的牌
	Tile []int
	// 积分所属组
	Group uint16
}

// FrontFinalInfo 前端显示的总结算结构体
type FrontFinalInfo struct {
	// 用户ID， 用户显示头像昵称
	UserId    int
	Nickname  string
	Avatar    string
	AvatarBox int
	// 累积积分(非本局、包括本把)
	// 在排位赛时，这个算包括本局的排位等级
	Total int
	//胡牌次数
	HuCount int
	// 捉鸡次数
	KitchenCount int
	// 杠牌张数
	KongCount int
	// 点炮
	DianPaoCount int
	// 本把累积积分
	Score int
	// 明杠次数
	KongTimes int
	// 暗杠次数
	KongDarkTimes int
	// 转弯杠次数
	KongTurnTimes int
	// 憨包杠次数
	KongTurnFreeTimes int
	// 自摸次数
	WinSelfTimes int
	// 接炮次数
	WinTimes int
	// 累积积分（非本局、不包括本把）
	FromTotal int
	// 排位赛星星变化
	StarChange int
	// 排位赛开始等级
	FromSLevel int
	// 排位赛最终等级
	FinalSLevel int
	// 排位赛经验值变化
	ExpChange int
	// 排位赛全场秒杀分
	SecKill int
	// 排位赛连胜次数
	WinningStreak int
}

// FrontScoreInfo 前端显示的积分结构体(单局)
type FrontScoreInfo struct {
	// 用户ID， 用户显示头像昵称
	UserId int
	// 自摸, 点炮, 胡， 热炮， 抢杠， 被抢杠
	WinWay int
	// 报听状态 1.报听， 0.未报听
	BaoTing int
	// 报听，叫牌， 没叫牌，大对子....
	WinStatus int
	// 本局积分(仅当前局，不包括本把的其他局)
	Total int
	// 积分条目
	Item []*FrontItem
	// 用户本把累积积分(本把所有局累积)
	GameScore int

	// 听牌状态,0:不显示;1:未叫牌;2:叫牌;3:报听
	TingStatus int
	// 胡牌状态,0:不显示;1:胡牌, 如果胡牌状态大于0，则听牌状态必须是0
	HuStatus int
	// 炮的状态,0:不显示;1:点炮;2:热炮
	PaoStatus int

	PlayChikens     []*FrontChikenInfo
	ShowCardChikens []*FrontChikenInfo
	HandChikens     []*FrontChikenInfo
}

// 客户端需要的鸡信息
type FrontChikenInfo struct {
	Tile       int    // 牌面
	IsRecharge int    // 是否冲锋鸡
	IsBao      int    // 是否包鸡
	IsGlod     int    // 是否金鸡
	ChikenType int    // 鸡牌类型,一个位运算的值
	Extra      string // 预留字段，必要时使用，防止协议修改
}

// 庄 的情况desc特殊，暗杠，转弯杠等情况icon特殊
type FrontItem struct {
	// 代表类型, 庄，胡 , 杠 明杠等
	TypeId int
	// 目前只有 显示庄的时候使用
	Count          int
	Tiles          []int
	ScoreCount     int
	Score          int
	Group          uint16   // 积分所属组
	RelationGroups []uint16 // 积分关联组
}

// 帮FrontItem排序
type FrontItemSorted []*FrontItem

func (s FrontItemSorted) Len() int {
	return len(s)
}

// 这里是倒序
func (s FrontItemSorted) Less(i, j int) bool {
	return s[i].TypeId > s[j].TypeId
}
func (s FrontItemSorted) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (us *UserScore) GetUserScore(userId int) *ScoreInfo {
	if v, ok := us.Maps.Load(userId); ok {
		return v.(*ScoreInfo)
	}
	return nil
}

func (this FrontFinalInfo) String() string {
	var s string
	s = s + fmt.Sprintf("[FrontFinalInfo]userId:%v, 总分:%v,", this.UserId, this.Score)
	s = s + fmt.Sprintf("自摸:%v, 接炮:%v,", this.WinSelfTimes, this.WinTimes)
	s = s + fmt.Sprintf("明杠:%v, 暗杠:%v, 转弯杠:%v, 憨包杠:%v.",
		this.KongTimes, this.KongDarkTimes, this.KongTurnTimes, this.KongTurnFreeTimes)
	return s
}

func (h HuInfo) isWinner(userId int) bool {
	for _, winInfo := range h.WInfo {
		if winInfo.Id == userId {
			return true
		}
	}
	return false
}

func (this FrontScoreInfo) String() string {
	var s string
	s = s + fmt.Sprintf("uid:%v, score total:%v\r\n", this.UserId, this.Total)
	for _, j1 := range this.Item {
		s = s + fmt.Sprintf("类型:%v, 庄数:%v, 牌:%v, 积分:%v*%v\r\n", config.FSTT[j1.TypeId], j1.Count, j1.Tiles, j1.ScoreCount, j1.Score)
	}
	return s
}

func (us *UserScore) String() string {
	var s string
	var total int
	us.Maps.Range(func(k, v interface{}) bool {
		i := k.(int)
		j := v.(*ScoreInfo)
		total = total + j.Total
		s = s + fmt.Sprintf("uid:%v, score total:%v\r\n", i, j.Total)
		for _, j1 := range j.Item {
			s = s + fmt.Sprintf("对象:%v, 条目:%v, 牌:%v, 积分:%v*%v\r\n", j1.UserId, config.STT[j1.SType], j1.Tile, j1.ScoreCount, j1.Score)
		}
		return true
	})
	if total != 0 {
		core.Logger.Error("debug error total not zero")
	}
	return s
}

func (this *UserScore) AddItem(uid int, s ScoreItem) {
	if s.SType == 0 || s.Score == 0 {
		core.Logger.Warn("[AddItem]data error,SType:%v,Score:%v", s.SType, s.Score)
		return
	}
	info := this.GetUserScore(uid)
	info.Total = info.Total + s.ScoreCount*s.Score
	info.Item = append(info.Item, s)
}

// SetScoreItem 设置计分明细
// winner 代表积分存在ScoreCount ！= 0 的情况
func (m *Mahjong) SetScoreItem(winner int, loser []int, way, wayCount, tile, tileCount int) {
	if loser == nil || (len(loser) == 1 && loser[0] == 0) {
		return
	}

	flag := 1
	if way < 0 {
		way = -way
		flag = -1
	}
	winScoreCount := len(loser)

	var score int
	if v, exists := m.scoreMap[way]; exists {
		score = flag * v
	} else {
		core.Logger.Error("对应的积分id未配置,majong type:%v,way:%v", m.MType, way)
	}

	if util.IntInSlice(winner, loser) {
		core.Logger.Error("winner in loser majong type %d, way %d tile %d", m.MType, way, tile)
	}

	// 通三加分
	if m.setting.IsSettingTS() {
		if v, exists := config.TS_EXTRA_SCORE[way]; exists {
			score += flag * v
		}
	}

	score = score * wayCount

	// 是否自摸翻倍
	if m.setting.IsSettingDoubleZM() {
		if m.isZimoHu() && util.IntInSlice(way, config.ZM_DOUBLE_SCORE_TYPE) {
			score *= 2
		}
	}
	// 是否大牌翻倍
	if m.setting.IsSettingDoubleDP() {
		if util.IntInSlice(way, config.DP_DOUBLE_SCORE_TYPE) {
			score *= 2
		}
	}
	// 翻倍鸡
	if m.setting.IsSettingDoubleChiken() {
		if tile == card.MAHJONG_BAM1 {
			score *= 2
		}
	}
	// 庄家翻倍
	if m.setting.EnableDoubleDealer && m.getUser(winner).IsDealer {
		if util.IntInSlice(way, config.DEALER_DOUBLE_SCORE_TYPE) {
			score *= 2
		}
	}

	// 积分倍数
	score *= m.setting.Multiple
	// 本局倍数
	score *= m.setting.MultipleRound

	t := []int{}
	for i := 0; i < tileCount; i++ {
		t = append(t, tile)
	}
	m.settlementGroup++
	// 给赢家增加积分
	winItem := ScoreItem{loser, way, score, winScoreCount, t, m.settlementGroup}
	m.SInfo.AddItem(winner, winItem)

	// 给输家减少积分
	loseItem := ScoreItem{[]int{winner}, way, 0 - score, 1, t, m.settlementGroup}
	for _, id := range loser {
		m.SInfo.AddItem(id, loseItem)
	}
}

// 设置胡牌信息（非黄牌时）
func (this *Mahjong) setHuInfo(winner *MahjongUser, opCode int, tile int) {
	winInfo := WinInfo{}
	winInfo.Id = winner.UserId
	winInfo.Way = this.winWay(winner, opCode)
	// 如果系统支持杠上开花作为独立牌型时，要用杠上开花覆盖掉其他胡牌方式
	// 并且忽略掉清一色
	if this.setting.EnableKongAfterDraw && winInfo.Way == config.HU_WAY_KONG_DRAW {
		winInfo.QFlag = false
		winInfo.WType = config.HU_TYPE_KONG_DRAW
	} else {
		winInfo.QFlag = qingCheck(winner.HandTileList.ToSlice(), winner.ShowCardList.GetAll())
		winInfo.WType = this.winType(winner.HandTileList.ToSlice(), winner.ShowCardList.GetAll(), tile)
	}

	winInfo.Tile = tile
	if oc.IsZMOperation(opCode) {
		// 自摸时，输家是其他人
		this.HInfo.Loser = this.getOtherUserId(winner.UserId)
	} else {
		// 胡牌时，输家是放炮的人
		// 防止一炮多响
		if len(this.HInfo.Loser) == 0 {
			this.HInfo.Loser = append(this.HInfo.Loser, this.LastOperator)
		}
	}
	this.HInfo.Hu = true
	this.HInfo.HuOperationCode = opCode
	this.HInfo.WInfo = append(this.HInfo.WInfo, winInfo)
}

// 设置胡牌信息（黄牌时）
// 赢家是叫牌了的人
// 输家是未叫牌的人
func (this *Mahjong) setHuInfoWhenHuang() {
	for id, user := range this.getUsers() {
		if user.MTC.IsTing() {
			winInfo := WinInfo{}
			winInfo.Id = id
			wType, tile := this.getTingType(id)
			winInfo.WType = wType
			winInfo.Tile = tile
			winInfo.QFlag = qingCheck(append(user.HandTileList.ToSlice(), tile), user.ShowCardList.GetAll())
			winInfo.Way = config.HU_WAY_PAO
			this.HInfo.WInfo = append(this.HInfo.WInfo, winInfo)
		} else {
			this.HInfo.Loser = append(this.HInfo.Loser, id)
		}
	}
}

func (this *Mahjong) getTingType(id int) (wType, tile int) {
	user := this.getUser(id)
	for _, j := range user.MTC.GetMaps() {
		for _, t := range j {
			pai := append(user.HandTileList.ToSlice(), t)
			tmp := this.winType(pai, user.ShowCardList.GetAll(), t)
			if wType == 0 || config.HuTypeSort[tmp] < config.HuTypeSort[wType] {
				wType = tmp
				tile = t
			}
		}
	}
	return wType, tile
}

// 返回胡的方式，天胡，地胡，杠上开花，自摸，抢杠, 热炮， 普通炮
func (this *Mahjong) winWay(user *MahjongUser, opCode int) int {
	var winway int

	switch opCode {
	case fbsCommon.OperationCodeWIN_SELF:
		if user.canTianhu() {
			winway = config.HU_WAY_TIAN
		} else if user.canDihu() {
			winway = config.HU_WAY_DI
		} else {
			winway = config.HU_WAY_DRAW
		}
	case fbsCommon.OperationCodeWIN_AFTER_KONG_DRAW:
		winway = config.HU_WAY_KONG_DRAW
	case fbsCommon.OperationCodeWIN:
		// 普通炮
		winway = config.HU_WAY_PAO
	case fbsCommon.OperationCodeWIN_AFTER_KONG_PLAY:
		// 热炮
		winway = config.HU_WAY_RE_PAO
	case fbsCommon.OperationCodeWIN_AFTER_KONG_TURN:
		// 抢杠
		winway = config.HU_WAY_QIANG_KONG
	default:
		core.Logger.Error("error opcode %d", opCode)
		winway = config.HU_WAY_PAO
	}

	// 记录用户胡牌状态
	user.WinStatus = winway

	return winway
}

func (this *Mahjong) getFrontData() map[int]*FrontScoreInfo {
	frontData := make(map[int]*FrontScoreInfo)
	this.FrontData.Range(func(k, v interface{}) bool {
		frontData[k.(int)] = v.(*FrontScoreInfo)
		return true
	})
	return frontData
}

func (this *Mahjong) getHInfo() HuInfo {
	return this.HInfo
}

func (this *Mahjong) clacScore() {
	if !this.HInfo.Hu {
		this.setHuInfoWhenHuang()
		if this.setting.IsSettingBaoKong() {
			this.clacBaoKong()
		}
	} else {
		this.clacKong()
		this.clacShaBao()
		this.clacRePao()
		if this.setting.IsSettingBaoKong() {
			this.clacBaoKong()
		} else if this.setting.EnableBaoKongBam1 {
			this.clacBaoKongBam1()
		}
		if this.setting.EnableFullChiken {
			this.clacFullChiken()
		}
	}
	this.clacHu()
	// 查缺
	if this.setting.IsEnableLack() {
		this.chaQue()
	}
	// 原缺
	if this.setting.EnableInitLack {
		this.clacInitLack()
	}
	// 龙七对奖3分
	if this.setting.IsSettingLE() {
		this.clacLE()
	}
	// 不区分庄闲
	if this.setting.IsSettingRemainDealer() {
		this.clacDealerLose()
	}
	this.clacKitchen()
	// core.Logger.Debug("[score]roomId:%d, round:%d, hu info:%v", this.RoomId, this.Round, this.HInfo)
	core.Logger.Debug("[score]roomId:%d, round:%d, score info:%v", this.RoomId, this.Round, this.SInfo)
}

func (this *Mahjong) chaQue() {
	var b []int = this.getBeiBaoQueId()
	if len(b) == 0 {
		return
	}

	for _, user := range this.getUsers() {
		var tile = this.getLackTile(user.UserId)
		var count = getLackCount(tile, user.HandTileList.ToSlice())
		if count != 0 {
			this.SetScoreItem(user.UserId, b, 0-(config.SCORE_TYPE_BAO_QUE), count, tile, 1)
		}
	}
}

// 计算原缺的分
func (m *Mahjong) clacInitLack() {
	m.Users.Range(func(k, v interface{}) bool {
		mu := v.(*MahjongUser)
		if mu.InitLack {
			m.SetScoreItem(mu.UserId, m.getOtherUserId(mu.UserId), config.SCORE_TYPE_INIT_LACK, 1, 0, 0)
		}
		return true
	})
}

// 计算龙七对奖3分
func (m *Mahjong) clacLE() {
	// 先看所有胡牌的人，是否是龙七对或者双龙七对
	for _, wInfo := range m.HInfo.WInfo {
		if wInfo.WType == config.HU_TYPE_LONG_7DUI || wInfo.WType == config.HU_TYPE_SHUANG_LONG_7DUI {
			m.SetScoreItem(wInfo.Id, m.getOtherUserId(wInfo.Id), config.SCORE_TYPE_LE, 1, 0, 0)
		}
	}
	m.Users.Range(func(k, v interface{}) bool {
		mu := v.(*MahjongUser)
		if m.HInfo.isWinner(mu.UserId) {
			return true
		}
		wType, _ := m.getTingType(mu.UserId)
		if wType == config.HU_TYPE_LONG_7DUI || wType == config.HU_TYPE_SHUANG_LONG_7DUI {
			m.SetScoreItem(mu.UserId, m.getOtherUserId(mu.UserId), config.SCORE_TYPE_LE, 1, 0, 0)
		}
		return true
	})
}

func (this *Mahjong) updateFinalInfo(info map[int]*FrontScoreInfo) {
	for _, v := range info {
		this.updateFinal(v)
	}
	// 排位赛中，Total表示排位赛等级，不是普通的房间积分
	// 只在最后一局完成时，根据房间人数进行更新
	// 大赢家增加2分, 小赢家加1分, 输家扣1分, 平分不扣
	var r, _ = RoomMap.GetRoom(this.RoomId)
	if r.IsRank() {
		// 房间结束的判断逻辑
		if this.TRound > 0 && this.Round >= this.TRound {
			maxScore := 0
			loseCnt := 0
			scoreList := r.GetScoreInfoList()
			// 第一趟循环，找出最高的分数, 找出负分人数
			for _, v := range scoreList {
				if v.Score < 0 {
					loseCnt++
				}
				if v.Score > maxScore {
					maxScore = v.Score
				}
			}
			// 第二趟循环，开始加分
			if maxScore > 0 {
				for _, v := range scoreList {
					exp := 0
					if v.Score > 0 {
						// 计算会员等级经验加成
						mu := this.getUser(v.UserId)
						var memberLevelAddExp int
						if mu.MemberAddExp > 0 {
							memberLevelAddExp = v.Score * mu.MemberAddExp / 100
							exp += memberLevelAddExp
							core.Logger.Debug("[updateFinalInfo]rank, roomId:%v, userId:%v, 会员等级奖励经验比例:%v, 奖励经验比例值:%v", this.RoomId, v.UserId, mu.MemberAddExp, memberLevelAddExp)
						}
						// 记录连胜次数
						v.WinningStreak += 1

						// 连胜倍数
						rate := rankService.GetWinningStreakRewards(v.WinningStreak)
						if rate != 0 {
							exp += int(float64(v.Score) * rate)
							core.Logger.Debug("[updateFinalInfo]rank, roomId:%v, userId:%v, 连胜奖励倍数:%v", this.RoomId, v.UserId, rate)
						} else {
							exp += v.Score
						}
						if loseCnt >= (len(this.Index)-1) && loseCnt >= 2 && v.Score == maxScore {
							// 全场秒杀加分
							exp += rank.SECKILL_SCORE
							v.SecKill = rank.SECKILL_SCORE
						}

					} else {
						v.WinningStreak = 0
						if v.Score < 0 {
							gradeId, _, _ := rank.ExplainSLevel(v.FromSLevel)
							exp = rank.GetRealDeductScore(v.Score, gradeId)
						}
					}
					// 计算最终的状态
					if exp != 0 {
						gradeId, gradeLevel, star, totalExp, starChange := rank.CalcRealLevel(v.FromSLevel, v.FromTotal, exp)
						v.Total = totalExp
						v.ExpChange = v.Total - v.FromTotal
						v.FinalSLevel = rank.FormatSLevel(gradeId, gradeLevel, star)
						v.StarChange = starChange

						core.Logger.Info("[updateFinalInfo]rank, roomId:%v, userId:%v, score:%v, fromTotal:%v, total:%v, expChange:%v, fromSLevel:%v, finalSLevel:%v, starChange:%v",
							this.RoomId, v.UserId, v.Score, v.FromTotal, v.Total, v.ExpChange, v.FromSLevel, v.FinalSLevel, v.StarChange)
					}
				}
			}
		}
	}
}

func (this *Mahjong) updateFinal(v *FrontScoreInfo) {
	var id = v.UserId
	var room, _ = RoomMap.GetRoom(this.RoomId)
	var info = room.GetScoreInfo(id)

	// 排位赛中，Total表示排位赛等级，不是普通的房间积分, 不进行总分累积
	if !room.IsRank() {
		info.Total = info.Total + v.Total
	}

	info.Score = info.Score + v.Total
	if this.HInfo.Hu && oc.IsHuOperation(this.HInfo.HuOperationCode) && this.HInfo.Loser[0] == id {
		info.DianPaoCount = info.DianPaoCount + 1
	}
	if this.HInfo.Hu && util.IntInSlice(id, this.getWinner()) {
		if oc.IsHuOperation(this.HInfo.HuOperationCode) {
			info.WinTimes++
		} else {
			info.WinSelfTimes++
		}
	}
	// 统计用户的杠牌次数
	if this.HInfo.Hu {
		mu := this.getUser(id)
		for _, sInfo := range mu.ShowCardList.GetAll() {
			if this.canWinKong(mu) {
				switch sInfo.GetOpCode() {
				case fbsCommon.OperationCodeKONG:
					info.KongTimes++
				case fbsCommon.OperationCodeKONG_DARK:
					info.KongDarkTimes++
				case fbsCommon.OperationCodeKONG_TURN:
					info.KongTurnTimes++
				case fbsCommon.OperationCodeKONG_TURN_FREE:
					info.KongTurnFreeTimes++
				default:
					break
				}
			}
		}
	}

	for _, i := range v.Item {
		switch i.TypeId {
		case config.FRONT_SCORE_TYPE_JI:
			fallthrough
		case config.FRONT_SCORE_TYPE_CHARGE_JI:
			fallthrough
		case config.FRONT_SCORE_TYPE_JIN_JI:
			fallthrough
		case config.FRONT_SCORE_TYPE_JIN_CHARGE_JI:
			fallthrough
		case config.FRONT_SCORE_TYPE_WUGU:
			fallthrough
		case config.FRONT_SCORE_TYPE_CHARGE_WUGU:
			fallthrough
		case config.FRONT_SCORE_TYPE_JIN_WUGU:
			fallthrough
		case config.FRONT_SCORE_TYPE_JIN_CHARGE_WUGU:
			fallthrough
		case config.FRONT_SCORE_TYPE_FANPAI:
			fallthrough
		case config.FRONT_SCORE_TYPE_UD:
			fallthrough
		case config.FRONT_SCORE_TYPE_FB:
			fallthrough
		case config.FRONT_SCORE_TYPE_XQ:
			fallthrough
		case config.FRONT_SCORE_TYPE_BEN:
			fallthrough
		case config.FRONT_SCORE_TYPE_TUMBLING:
			fallthrough
		case config.FRONT_SCORE_TYPE_STAND_KITCHEN:
			fallthrough
		case config.FRONT_SCORE_TYPE_STAND_JIN_KITCHEN:
			fallthrough
		case config.FRONT_SCORE_TYPE_FIRST_CYCLE_CHARGE_KITCHEN:
			fallthrough
		case config.FRONT_SCORE_TYPE_FIRST_CYCLE_CHARGE_JIN_KITCHEN:
			fallthrough
		case config.FRONT_SCORE_TYPE_DIAMOND:
			fallthrough
		case config.FRONT_SCORE_TYPE_PP_KITCHEN:
			fallthrough
		case config.FRONT_SCORE_TYPE_FLOWER_RED_KITCHEN:
			fallthrough
		case config.FRONT_SCORE_TYPE_FLOWER_RED_DIAMOND_KITCHEN:
			info.KitchenCount += len(i.Tiles)
		}
	}
}

func (this *Mahjong) genFrontData() {
	var room, _ = RoomMap.GetRoom(this.RoomId)
	// 是否有低版本的用户
	// 有低版本用户时，过滤掉一些不支持的条目
	hasLowVersion := false
	for _, ru := range room.GetUsers() {
		if strings.Compare(ru.Info.Version, "1.9.0") == -1 {
			hasLowVersion = true
			break
		}
	}

	this.SInfo.Maps.Range(func(k, v interface{}) bool {
		id := k.(int)
		score := v.(*ScoreInfo)

		var info = &FrontScoreInfo{}
		info.UserId = id
		info.Total = score.Total
		// 天胡 也显示未报停
		if !this.getUser(id).MTC.IsBaoTing() {
			info.BaoTing = 0
		} else {
			info.BaoTing = 1
		}
		core.Logger.Debug("front baoting %d", info.BaoTing)
		info.WinWay = this.getFrontWinWay(id)
		info.WinStatus = this.getFrontWinStatus(id)
		this.getUser(id).WinType = info.WinStatus
		userScoreItem := this.getFrontScoreItem(id)
		if hasLowVersion {
			for _, item := range userScoreItem {
				if !util.IntInSlice(item.TypeId, config.FrontScoreLowVersionExclude) {
					info.Item = append(info.Item, item)
				}
			}
		} else {
			info.Item = userScoreItem
		}

		// 记录报听状态
		if info.WinStatus == config.FRONT_WINSTATUS_BAOTING {
			info.TingStatus = 3
		}
		if info.WinWay == config.FRONT_WINWAY_HU {
			info.HuStatus = 1
		} else if info.WinWay == config.FRONT_WINWAY_ZIMO {
			info.HuStatus = 2
		} else if info.WinWay == config.FRONT_WINWAY_QIANG {
			info.HuStatus = 3
		} else {
			// 没有胡牌，需要看有没有叫牌
			if info.WinStatus == config.FRONT_WINSTATUS_TING {
				info.TingStatus = 2
			} else if info.WinStatus == config.FRONT_WINSTATUS_NO_TING {
				info.TingStatus = 1
			}
			// 判断有没有点炮，有没有热炮
			if this.HInfo.Hu && util.IntInSlice(id, this.HInfo.Loser) {
				if this.HInfo.WInfo[0].Way == config.HU_WAY_RE_PAO {
					// 热炮
					info.PaoStatus = 2
				} else if this.HInfo.WInfo[0].Way == config.HU_WAY_PAO {
					info.PaoStatus = 1
				} else if this.HInfo.WInfo[0].Way == config.HU_WAY_QIANG_KONG {
					info.PaoStatus = 3
				}
			}
		}

		// 记录放炮状态到用户数据中
		if info.PaoStatus > 0 {
			this.getUser(id).PaoStatus = info.PaoStatus
		}

		// 计算用户可参与计算积分的鸡
		for _, item := range info.Item {
			isChiken, isGold, isCharge, isBao := isShowChikenType(item.TypeId)
			if isChiken == 0 {
				continue
			}
			for i := 0; i < len(item.Tiles); i++ {
				// frontChikenInfo := &FrontChikenInfo{item.Tiles[0], isCharge, isBao, isGold, 0, strconv.Itoa(item.Score * item.ScoreCount / len(item.Tiles))}
				frontChikenInfo := &FrontChikenInfo{item.Tiles[0], isCharge, isBao, isGold, 0, strconv.Itoa(item.Score / len(item.Tiles))}
				info.HandChikens = append(info.HandChikens, frontChikenInfo)
			}
		}
		this.FrontData.Store(id, info)

		return true
	})
	this.updateFinalInfo(this.getFrontData())
	// 读取用户局内积分
	core.Logger.Debug("FrontData===%#v", this.getFrontData())
	core.Logger.Debug("room.ScoreInfo===%#v", room.GetScoreInfoList())
	room.ScoreInfo.Range(func(k, v interface{}) bool {
		id := k.(int)
		ffInfo := v.(*FrontFinalInfo)
		core.Logger.Debug("gameScoreInfo===%#v", ffInfo)
		fi, _ := this.FrontData.Load(id)
		frontData := fi.(*FrontScoreInfo)
		frontData.GameScore = ffInfo.Score
		return true
	})
	core.Logger.Debug("[score]roomId:%d, round:%d, FinalFrontInfo", this.RoomId, this.Round)
}

func (this *Mahjong) getFrontScoreItem(id int) []*FrontItem {
	var kongItem []*FrontItem = this.getKongItem(id)
	var baoJi, baoJiIncome []*FrontItem = this.getBaoAndIncomeKitchenItem(id)
	var baoPai, baoPaiIncome []*FrontItem = this.getBaoAndIncomePaiItem(id)
	var que, queIncome []*FrontItem = this.getChaQueItem(id)
	var rePao, shaBao, zhuang, qiangKong, huItem, winKitchenItem, dianPao []*FrontItem

	// 普通的结算明细，和是否胡牌无关的
	var normalItem []*FrontItem
	// 与胡牌相关的结算明细
	var winItem []*FrontItem

	if this.HInfo.Hu {
		rePao = this.getRePaoItem(id)
		shaBao = this.getShaBaoItem(id)
		zhuang = this.getZhuangItem(id)
		qiangKong = this.getQiangKongItem(id)
		huItem = this.getHuPaiItem(id)
		winKitchenItem = this.getWinKitchenItem(id)
		dianPao = this.getDianPaoItem(id)

		// 赢牌的普通item
		winItem = this.getWinItem(id)
	}

	// 获取普通的，通过score_type 推导front_score_type的结算明细
	normalItem = this.getNormalItem(id)

	var result []*FrontItem
	result = append(result, append(huItem, append(kongItem, append(que, append(queIncome, append(winKitchenItem, append(baoJi, append(baoJiIncome, append(baoPai, append(baoPaiIncome, append(dianPao, append(shaBao, append(rePao, append(qiangKong, zhuang...)...)...)...)...)...)...)...)...)...)...)...)...)...)
	result = append(result, normalItem...)
	result = append(result, winItem...)

	// 合并group操作
	for _, scoreItem := range result {
		if mergeGroup, ok := config.ScoreTypeItemMergeMap[scoreItem.TypeId]; ok {
			scoreItem.Group = mergeGroup
		}
	}

	return result
}

func (this *Mahjong) getChaQueItem(id int) (rBao []*FrontItem, rIncome []*FrontItem) {
	var bao = &FrontItem{}
	bao.ScoreCount = 1
	bao.TypeId = config.FRONT_SCORE_TYPE_CHAQUE
	bao.Tiles = []int{}

	var income = &FrontItem{}
	income.ScoreCount = 1
	income.TypeId = config.FRONT_SCORE_TYPE_CHAQUE_SHOURU
	income.Tiles = []int{}

	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_BAO_QUE:
		default:
			continue
		}
		bao.Group = v.Group
		income.Group = v.Group
		if v.Score < 0 {
			bao.Score = bao.Score + v.Score*v.ScoreCount
		} else if v.Score > 0 {
			income.Score = income.Score + v.Score*v.ScoreCount
		} else {
			core.Logger.Error("score error score %v", v)
			continue
		}
	}
	if bao.Score == 0 {
		rBao = nil
	} else {
		rBao = append(rBao, bao)
	}
	if income.Score == 0 {
		rIncome = nil
	} else {
		rIncome = append(rIncome, income)
	}
	if rBao != nil && rIncome != nil {
		core.Logger.Error("rBao, rIncome all exists")
		return nil, nil
	}
	return rBao, rIncome
}

func (this *Mahjong) getZhuangItem(id int) []*FrontItem {
	// 庄(赢钱): 1.庄胡(4庄), 2.黄牌(包牌, 包牌收入中 )
	// 庄(输钱): 1.庄家点炮，被自摸(下庄, 我是庄，我被自摸,或者点炮，我显示下庄, 对方显示庄家下庄 )
	var typeId int
	var item ScoreItem
	isZM := this.isZimoHu()
	for _, item = range this.SInfo.GetUserScore(id).Item {
		switch item.SType {
		case config.SCORE_TYPE_CONTINUE_DEALER_LOSE:
		case config.SCORE_TYPE_CONTINUE_DEALER:
			if item.Score < 0 && !isZM {
				continue
			}
		default:
			continue
		}
		typeId = getFrontScoreType(item.SType, item.Score)
		if typeId != 0 {
			break
		}
	}
	if typeId != 0 {
		var r = &FrontItem{}
		r.TypeId = typeId
		r.Score = item.Score
		r.Count = this.DealerCount
		r.ScoreCount = item.ScoreCount
		r.Tiles = []int{}
		r.Group = item.Group
		return []*FrontItem{r}
	}
	return nil
}

func (this *Mahjong) getQiangKongItem(id int) []*FrontItem {
	if this.HInfo.WInfo[0].Way != config.HU_WAY_QIANG_KONG {
		return nil
	}
	if len(this.HInfo.Loser) != 1 {
		core.Logger.Error("loser error %v", this.HInfo.Loser)
		return nil
	}
	if this.HInfo.Loser[0] != id && !util.IntInSlice(id, this.getWinner()) {
		return nil
	}
	for _, item := range this.SInfo.GetUserScore(id).Item {
		switch item.SType {
		case config.SCORE_TYPE_KONG_QIANG:
		default:
			continue
		}
		if typeId := getFrontScoreType(item.SType, item.Score); typeId != 0 {
			var qiangKong = &FrontItem{}
			qiangKong.Score = item.Score * item.ScoreCount
			qiangKong.ScoreCount = 1
			qiangKong.TypeId = typeId
			qiangKong.Tiles = []int{}
			qiangKong.Group = item.Group
			return []*FrontItem{qiangKong}
		}
	}
	return nil
}

func (this *Mahjong) getRePaoItem(id int) []*FrontItem {
	if this.HInfo.WInfo[0].Way != config.HU_WAY_RE_PAO {
		return nil
	}
	if len(this.HInfo.Loser) != 1 {
		core.Logger.Error("loser error %v", this.HInfo.Loser)
		return nil
	}
	if this.HInfo.Loser[0] != id && !util.IntInSlice(id, this.getWinner()) {
		return nil
	}

	for _, item := range this.SInfo.GetUserScore(id).Item {
		switch item.SType {
		case config.SCORE_TYPE_REPAO_EXPOSED:
		case config.SCORE_TYPE_REPAO_DARK:
		case config.SCORE_TYPE_REPAO_TURN:
		default:
			continue
		}
		if typeId := getFrontScoreType(item.SType, item.Score); typeId != 0 {
			var rePao = &FrontItem{}
			rePao.Score = item.Score * item.ScoreCount
			rePao.ScoreCount = 1
			rePao.TypeId = typeId
			rePao.Tiles = []int{}
			rePao.Group = item.Group
			return []*FrontItem{rePao}
		}
	}
	return nil
}

func (this *Mahjong) getZerenItem(id int) []*FrontItem {
	var result []*FrontItem
	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_RB_KITCHEN:
		case config.SCORE_TYPE_RB_JIN_KITCHEN:
		case config.SCORE_TYPE_RB_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_RB_FIRST_CYCLE_KITCHEN:
		case config.SCORE_TYPE_RB_FIRST_CYCLE_JIN_KITCHEN:
		case config.SCORE_TYPE_RB_FIRST_CYCLE_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_FIRST_CYCLE_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_FIRST_CYCLE_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_FIRST_CYCLE_JIN_JIN_KITCHEN:
		default:
			continue
		}
		if typeId := getFrontScoreType(v.SType, v.Score); typeId != 0 {
			var item = &FrontItem{}
			item.Score = v.Score
			item.ScoreCount = v.ScoreCount
			item.TypeId = typeId
			item.Tiles = v.Tile
			item.Group = v.Group
			result = append(result, item)
		}
	}
	return result
}

// 赢鸡(对应输的玩家不显示), 责任鸡收入
func (this *Mahjong) getWinKitchenItem(id int) []*FrontItem {
	var result []*FrontItem
	// 未上听，或者黄牌
	if !this.HInfo.Hu {
		return nil
	}
	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_KITCHEN:
		case config.SCORE_TYPE_CHARGE_KITCHEN:
		case config.SCORE_TYPE_JIN_KITCHEN:
		case config.SCORE_TYPE_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_CHARGE_JIN_KITCHEN:
		case config.SCORE_TYPE_CHARGE_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_WUGU_KITCHEN:
		case config.SCORE_TYPE_CHARGE_WUGU_KITCHEN:
		case config.SCORE_TYPE_WUGU_JIN_KITCHEN:
		case config.SCORE_TYPE_WUGU_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_CHARGE_WUGU_JIN_KITCHEN:
		case config.SCORE_TYPE_CHARGE_WUGU_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_UD_KITCHEN:
		case config.SCORE_TYPE_FP_KITCHEN:
		case config.SCORE_TYPE_FB_KITCHEN:
		case config.SCORE_TYPE_B_KITCHEN:
		case config.SCORE_TYPE_XQ_KITCHEN:
		case config.SCORE_TYPE_GT_KITCHEN:
		case config.SCORE_TYPE_QING_EXTRA: // 清一色奖励三只鸡
		case config.SCORE_TYPE_STAND_KITCHEN: // 站鸡
		case config.SCORE_TYPE_STAND_JIN_KITCHEN: // 金站鸡
		case config.SCORE_TYPE_STAND_JIN_JIN_KITCHEN: // 金金站鸡
		case config.SCORE_TYPE_FIRST_CYCLE_CHARGE_KITCHEN: // 首圈冲锋鸡
		case config.SCORE_TYPE_FIRST_CYCLE_CHARGE_JIN_KITCHEN:
		case config.SCORE_TYPE_FIRST_CYCLE_CHARGE_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_SILVER_CHIKEN:
		case config.SCORE_TYPE_SILVER_DOUBLE_CHIKEN:
		case config.SCORE_TYPE_SILVER_TRIPLE_CHIKEN:
		case config.SCORE_TYPE_DIAMOND_KITCHEN:
		case config.SCORE_TYPE_DIAMOND_DOUBLE_KITCHEN:
		case config.SCORE_TYPE_PP_KITCHEN:
		case config.SCORE_TYPE_FLOWER_RED_KITCHEN_TING:
		case config.SCORE_TYPE_FLOWER_RED_KITCHEN_WIN:
		case config.SCORE_TYPE_FLOWER_RED_KITCHEN_TING_DIAMOND:
		case config.SCORE_TYPE_FLOWER_RED_KITCHEN_WIN_DIAMOND:
		case config.SCORE_TYPE_J7W:
		case config.SCORE_TYPE_GWD:
		default:
			continue
		}
		if typeId := getFrontScoreType(v.SType, v.Score); typeId != 0 {
			var item = &FrontItem{}
			item.Score = v.Score
			item.ScoreCount = v.ScoreCount
			item.TypeId = this.frontScoreTypeTrans(typeId)
			item.Tiles = v.Tile
			item.Group = v.Group
			result = append(result, item)
		}
	}
	return result
}

func (this *Mahjong) getShaBaoItem(id int) []*FrontItem {
	if !util.IntInSlice(id, this.HInfo.Loser) && !util.IntInSlice(id, this.getWinner()) {
		return nil
	}

	var typeId, score int
	var group uint16
	for _, item := range this.SInfo.GetUserScore(id).Item {
		switch item.SType {
		case config.SCORE_TYPE_SHABAO:
		case config.SCORE_TYPE_SHABAO_SHUANG_LONG_7DUI:
		case config.SCORE_TYPE_SHABAO_SHUANG_LONG_7DUI_QING:
		case config.SCORE_TYPE_SHABAO_LONG_7DUI:
		case config.SCORE_TYPE_SHABAO_LONG_7DUI_QING:
		case config.SCORE_TYPE_SHABAO_7DUI:
		case config.SCORE_TYPE_SHABAO_7DUI_QING:
		case config.SCORE_TYPE_SHABAO_DIQIDUI:
		case config.SCORE_TYPE_SHABAO_DIQIDUI_QING:
		case config.SCORE_TYPE_SHABAO_DANDIAO:
		case config.SCORE_TYPE_SHABAO_DANDIAO_QING:
		case config.SCORE_TYPE_SHABAO_DADUI:
		case config.SCORE_TYPE_SHABAO_DADUI_QING:
		case config.SCORE_TYPE_SHABAO_BIANKADIAO:
		case config.SCORE_TYPE_SHABAO_BIANKADIAO_QING:
		case config.SCORE_TYPE_SHABAO_DAKUANZHANG:
		case config.SCORE_TYPE_SHABAO_DAKUANZHANG_QING:
		case config.SCORE_TYPE_SHABAO_QING:
		case config.SCORE_TYPE_SHABAO_HEPU_7DUI:
		case config.SCORE_TYPE_SHABAO_HEPU_7DUI_QING:
		default:
			continue
		}
		typeId = getFrontScoreType(item.SType, item.Score)
		score += item.Score * item.ScoreCount
		group = item.Group
	}

	if typeId != 0 {
		var shaBao = &FrontItem{}
		shaBao.Score = score
		shaBao.ScoreCount = 1
		shaBao.Tiles = []int{}
		shaBao.TypeId = typeId
		shaBao.Group = group
		return []*FrontItem{shaBao}
	}
	return nil
}

func (this *Mahjong) getBaoAndIncomePaiItem(id int) (rBao []*FrontItem, rIncome []*FrontItem) {
	if this.HInfo.Hu {
		return nil, nil
	}

	var bao = &FrontItem{}
	bao.ScoreCount = 1
	bao.TypeId = config.FRONT_SCORE_TYPE_BAOPAI
	bao.Tiles = []int{}

	var income = &FrontItem{}
	income.ScoreCount = 1
	income.TypeId = config.FRONT_SCORE_TYPE_BAOPAI_SHOURU
	income.Tiles = []int{}

	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_BAOTING:
		case config.SCORE_TYPE_QING:
		case config.SCORE_TYPE_SHUANGLONG_QIDUI:
		case config.SCORE_TYPE_HEPU_QIDUI:
		case config.SCORE_TYPE_LONGQIDUI:
		case config.SCORE_TYPE_QIDUI:
		case config.SCORE_TYPE_DIQIDUI:
		case config.SCORE_TYPE_DANDIAO:
		case config.SCORE_TYPE_DADUI:
		case config.SCORE_TYPE_DAKUANZHANG:
		case config.SCORE_TYPE_BIANKADIAO:
		case config.SCORE_TYPE_PINGHU:
		case config.SCORE_TYPE_PINGHU_ZIMO:
		case config.SCORE_TYPE_CONTINUE_DEALER:
		case config.SCORE_TYPE_CONTINUE_DEALER_LOSE:
		default:
			continue
		}
		bao.Group = v.Group
		income.Group = v.Group
		if v.Score < 0 {
			bao.Score += v.Score * v.ScoreCount
		} else if v.Score > 0 {
			income.Score += v.Score * v.ScoreCount
		} else {
			core.Logger.Error("score error score %v", v)
			continue
		}
	}
	if bao.Score == 0 {
		rBao = nil
	} else {
		rBao = append(rBao, bao)
	}
	if income.Score == 0 {
		rIncome = nil
	} else {
		rIncome = append(rIncome, income)
	}
	if rBao != nil && rIncome != nil {
		core.Logger.Error("rBao, rIncome all exists")
		return nil, nil
	}
	return rBao, rIncome
}

func (this *Mahjong) getBaoAndIncomeKitchenItem(id int) (rBao []*FrontItem, rIncome []*FrontItem) {
	var bao = &FrontItem{}
	bao.ScoreCount = 1
	bao.TypeId = config.FRONT_SCORE_TYPE_BAOJI
	bao.Tiles = []int{}

	var income = &FrontItem{}
	income.ScoreCount = 1
	income.TypeId = config.FRONT_SCORE_TYPE_BAOJI_SHOURU
	income.Tiles = []int{}

	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_BAO_KITCHEN:
		case config.SCORE_TYPE_BAO_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_CHARGE_KITCHEN:
		case config.SCORE_TYPE_BAO_CHARGE_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_CHARGE_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_WUGU_KITCHEN:
		case config.SCORE_TYPE_BAO_WUGU_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_WUGU_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_CHARGE_WUGU_KITCHEN:
		case config.SCORE_TYPE_BAO_CHARGE_WUGU_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_CHARGE_WUGU_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_STAND_KITCHEN:
		case config.SCORE_TYPE_BAO_STAND_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_STAND_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_FIRST_CYCLE_CHARGE_KITCHEN:
		case config.SCORE_TYPE_BAO_FIRST_CYCLE_CHARGE_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_FIRST_CYCLE_CHARGE_JIN_JIN_KITCHEN:
		default:
			continue
		}
		bao.Group = v.Group
		income.Group = v.Group
		if v.Score > 0 {
			income.Score += v.Score * v.ScoreCount
		} else {
			bao.Score += v.Score * v.ScoreCount
		}
	}
	if bao.Score == 0 {
		rBao = nil
	} else {
		rBao = append(rBao, bao)
	}
	if income.Score == 0 {
		rIncome = nil
	} else {
		rIncome = append(rIncome, income)
	}
	return rBao, rIncome
}

func (this *Mahjong) getKongItem(id int) []*FrontItem {
	var result = []*FrontItem{}
	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_KONG_TURN:
		case config.SCORE_TYPE_KONG_EXPOSED:
		case config.SCORE_TYPE_KONG_DARK:
		case config.SCORE_TYPE_BAO_KONG:
		case config.SCORE_TYPE_BAO_KONG_DARK:
		case config.SCORE_TYPE_BAO_KONG_TURN:
		default:
			continue
		}
		if typeId := getFrontScoreType(v.SType, v.Score); typeId != 0 {
			var item = &FrontItem{}
			// 杠的积分需要合并
			item.Score = v.Score * v.ScoreCount
			item.ScoreCount = 1
			item.TypeId = typeId
			item.Tiles = v.Tile
			item.Group = v.Group
			result = append(result, item)
		}
	}
	return result
}

// 相对于 getHuPaiItem 去掉了 杠上花，报听
func (this *Mahjong) getDianPaoItem(id int) []*FrontItem {
	// 这把牌是否有人胡
	if !oc.IsHuOperation(this.HInfo.HuOperationCode) {
		return nil
	}

	// fixme 下面的这两个逻辑判断，不太严谨
	// 所以在上面重新添加了一个���辑判断，为了防止出错，短期内让这两个逻辑共存
	// 如果在日志中，不再出现loser count error 这样的内容, 就可以将这两段逻辑删除
	// 黄牌或者自摸没有此item
	if len(this.HInfo.Loser) == 3 {
		return nil
	}
	// 安全性判断，前面代码没bug情况下不会出现此情况
	if len(this.HInfo.Loser) != 1 {
		core.Logger.Error("loser count error %v", this.HInfo.Loser)
		return nil
	}

	// 不是点炮者
	if id != this.HInfo.Loser[0] {
		return nil
	}

	var item = &FrontItem{}
	item.ScoreCount = 1
	item.TypeId = config.FRONT_SCORE_TYPE_DIANPAO
	item.Tiles = []int{}

	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_CONTINUE_DEALER:
		case config.SCORE_TYPE_QING:
		case config.SCORE_TYPE_SHUANGLONG_QIDUI:
		case config.SCORE_TYPE_HEPU_QIDUI:
		case config.SCORE_TYPE_LONGQIDUI:
		case config.SCORE_TYPE_QIDUI:
		case config.SCORE_TYPE_DIQIDUI:
		case config.SCORE_TYPE_DANDIAO:
		case config.SCORE_TYPE_DADUI:
		case config.SCORE_TYPE_DAKUANZHANG:
		case config.SCORE_TYPE_BIANKADIAO:
		case config.SCORE_TYPE_PINGHU:
		case config.SCORE_TYPE_PINGHU_ZIMO:
		case config.SCORE_TYPE_BAOTING:
		case config.SCORE_TYPE_DIHU:
		default:
			continue
		}
		item.Group = v.Group
		item.Score += v.Score * v.ScoreCount
	}
	if item.Score >= 0 {
		core.Logger.Error("dianpao item score >= 0 %d", item.Score)
		for _, v := range this.SInfo.GetUserScore(id).Item {
			core.Logger.Error("this.SInfo[id].Item:%v", v)
		}
		return nil
	}
	return []*FrontItem{item}
}

func (this *Mahjong) getTongSanItem(id int) []*FrontItem {
	var result = []*FrontItem{}
	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_TONG_SAN:
		default:
			continue
		}
		if typeId := getFrontScoreType(v.SType, v.Score); typeId != 0 {
			var item = &FrontItem{}
			item.Score = v.Score
			item.ScoreCount = v.ScoreCount
			item.TypeId = typeId
			item.Group = v.Group
			result = append(result, item)
		}
	}
	return result
}

func (this *Mahjong) getFullChikenItem(id int) []*FrontItem {
	var result = []*FrontItem{}
	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_FULL_CHIKEN:
		default:
			continue
		}
		if typeId := getFrontScoreType(v.SType, v.Score); typeId != 0 {
			var item = &FrontItem{}
			item.Score = v.Score
			item.ScoreCount = v.ScoreCount
			item.TypeId = typeId
			item.Group = v.Group
			result = append(result, item)
		}
	}
	return result
}

// 获取普通的计算明细，不管是不是胡牌都需要计算的
// 必须是通过score_type 直接推算出front_score_type类型的
func (this *Mahjong) getNormalItem(id int) []*FrontItem {
	var result = []*FrontItem{}
	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_INIT_LACK: // 原缺
		case config.SCORE_TYPE_LE: // 龙七对+3分
		case config.SCORE_TYPE_RB_KITCHEN: // 责任鸡
		case config.SCORE_TYPE_RB_JIN_KITCHEN:
		case config.SCORE_TYPE_RB_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_RB_FIRST_CYCLE_KITCHEN:
		case config.SCORE_TYPE_RB_FIRST_CYCLE_JIN_KITCHEN:
		case config.SCORE_TYPE_RB_FIRST_CYCLE_JIN_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_FIRST_CYCLE_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_FIRST_CYCLE_JIN_KITCHEN:
		case config.SCORE_TYPE_BAO_RB_FIRST_CYCLE_JIN_JIN_KITCHEN:
		default:
			continue
		}
		if typeId := getFrontScoreType(v.SType, v.Score); typeId != 0 {
			var item = &FrontItem{}
			item.Score = v.Score
			item.ScoreCount = v.ScoreCount
			item.TypeId = typeId
			item.Tiles = v.Tile
			item.Group = v.Group
			result = append(result, item)
		}
	}
	return result
}

// 获取普通的计算明细，必须要胡牌才参与计算的
// 必须是通过score_type 直接推算出front_score_type类型的
func (this *Mahjong) getWinItem(id int) []*FrontItem {
	var result = []*FrontItem{}
	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_FULL_CHIKEN: // 满鸡
		case config.SCORE_TYPE_TONG_SAN: // 通三
		default:
			continue
		}
		if typeId := getFrontScoreType(v.SType, v.Score); typeId != 0 {
			var item = &FrontItem{}
			item.Score = v.Score
			item.ScoreCount = v.ScoreCount
			item.TypeId = typeId
			item.Group = v.Group
			result = append(result, item)
		}
	}
	return result
}

func (this *Mahjong) getHuPaiItem(id int) []*FrontItem {
	// 非自摸时，输家不显示该条目
	if !this.isZimoHu() && !util.IntInSlice(id, this.getWinner()) {
		return nil
	}
	var result = []*FrontItem{}
	for _, v := range this.SInfo.GetUserScore(id).Item {
		switch v.SType {
		case config.SCORE_TYPE_KONG_DRAW:
		case config.SCORE_TYPE_TIANHU:
		case config.SCORE_TYPE_DIHU:
		case config.SCORE_TYPE_BAOTING:
		case config.SCORE_TYPE_QING:
		case config.SCORE_TYPE_SHUANGLONG_QIDUI:
		case config.SCORE_TYPE_HEPU_QIDUI:
		case config.SCORE_TYPE_LONGQIDUI:
		case config.SCORE_TYPE_QIDUI:
		case config.SCORE_TYPE_DIQIDUI:
		case config.SCORE_TYPE_DANDIAO:
		case config.SCORE_TYPE_DADUI:
		case config.SCORE_TYPE_DAKUANZHANG:
		case config.SCORE_TYPE_BIANKADIAO:
		case config.SCORE_TYPE_PINGHU:
		case config.SCORE_TYPE_PINGHU_ZIMO:
		case config.SCORE_TYPE_ZIMO:
		case config.SCORE_TYPE_ZIMO_EXTRA:
		default:
			continue
		}
		if typeId := getFrontScoreType(v.SType, v.Score); typeId != 0 {
			var item = &FrontItem{}
			item.Score = v.Score
			item.ScoreCount = v.ScoreCount
			item.TypeId = typeId
			item.Group = v.Group
			result = append(result, item)
		}
	}
	return result
}

func (this *Mahjong) getFrontWinStatus(id int) int {
	// 未胡牌或者黄牌都显示听的信息
	if !this.HInfo.Hu || !util.IntInSlice(id, this.getWinner()) {
		mu := this.getUser(id)
		if mu.MTC.IsBaoTing() {
			return config.FRONT_WINSTATUS_BAOTING
		}
		// IsTing包括了听和报听两个状态，但是前面已经判断过报听了，所以这里这么用没问题
		if mu.MTC.IsTing() {
			return config.FRONT_WINSTATUS_TING
		}
		if mu.MTC.IsNormal() {
			return config.FRONT_WINSTATUS_NO_TING
		}
	}

	for _, j := range this.HInfo.WInfo {
		if j.Id != id {
			continue
		}
		var qing int
		if j.QFlag {
			qing = 1
		} else {
			qing = 0
		}
		switch j.WType {
		case config.HU_TYPE_KONG_DRAW:
			return config.FRONT_WINSTATUS_KONG_DRAW
		case config.HU_TYPE_SHUANG_LONG_7DUI:
			return config.FRONT_WINSTATUS_SHUANGLONGQIDUI + qing
		case config.HU_TYPE_HEPU_7DUI:
			return config.FRONT_WINSTATUS_HEPU_QIDUI + qing
		case config.HU_TYPE_LONG_7DUI:
			return config.FRONT_WINSTATUS_LONGQIDUI + qing
		case config.HU_TYPE_7DUI:
			return config.FRONT_WINSTATUS_QIDUI + qing
		case config.HU_TYPE_DIQIDUI:
			return config.FRONT_WINSTATUS_DIQIDUI + qing
		case config.HU_TYPE_DANDIAO:
			return config.FRONT_WINSTATUS_DANDIAO + qing
		case config.HU_TYPE_DADUI:
			return config.FRONT_WINSTATUS_DADUI + qing
		case config.HU_TYPE_BIANKADIAO:
			// 边卡吊，如果是清一色，则显示清一色
			if qing == 1 {
				return config.FRONT_WINSTATUS_QING
			}
			return config.FRONT_WINSTATUS_BIANDIAO
		case config.HU_TYPE_DAKUANZHANG:
			// 大宽张，如果是清一色，则显示清一色
			if qing == 1 {
				return config.FRONT_WINSTATUS_QING
			}
			return config.FRONT_WINSTATUS_DAKUAN
		case config.HU_TYPE_PI:
			if qing == 1 {
				return config.FRONT_WINSTATUS_QING
			}
			return config.FRONT_WINSTATUS_HU
		}
	}

	core.Logger.Error("status error %d", id)
	return config.FRONT_WINSTATUS_NO_TING
}

func (this *Mahjong) getFrontWinWay(id int) int {
	if !this.HInfo.Hu || (!util.IntInSlice(id, this.HInfo.Loser) && !util.IntInSlice(id, this.getWinner())) {
		return config.FRONT_WINWAY_BLANK
	}

	var way int = this.HInfo.WInfo[0].Way
	switch way {
	case config.HU_WAY_TIAN:
		fallthrough
	case config.HU_WAY_DI:
		fallthrough
	case config.HU_WAY_KONG_DRAW:
		fallthrough
	case config.HU_WAY_DRAW:
		if util.IntInSlice(id, this.getWinner()) {
			if oc.IsHuOperation(this.HInfo.HuOperationCode) {
				return config.FRONT_WINWAY_HU
			}
			return config.FRONT_WINWAY_ZIMO
		}
		return config.FRONT_WINWAY_BLANK
	case config.HU_WAY_QIANG_KONG:
		if util.IntInSlice(id, this.getWinner()) {
			return config.FRONT_WINWAY_QIANG
		}
		return config.FRONT_WINWAY_BEI_QIANG
	case config.HU_WAY_RE_PAO:
		if util.IntInSlice(id, this.getWinner()) {
			return config.FRONT_WINWAY_HU
		}
		return config.FRONT_WINWAY_DIANREPAO
	case config.HU_WAY_PAO:
		if util.IntInSlice(id, this.getWinner()) {
			return config.FRONT_WINWAY_HU
		}
		return config.FRONT_WINWAY_DIANPAO
	default:
		return config.FRONT_WINWAY_BLANK
	}
}

// 计算鸡的积分
// 记录所有的鸡
// 记录用户的输赢鸡状态
func (this *Mahjong) clacKitchen() {
	// 算出所有的鸡
	// 幺鸡
	this.Chikens[card.MAHJONG_BAM1] = config.CHIKEN_TYPE_BAM1
	// 乌骨鸡
	if this.setting.IsSettingChikenDot8() {
		this.Chikens[card.MAHJONG_DOT8] = config.CHIKEN_TYPE_DOT8
	}
	// 翻牌鸡
	for _, tile := range this.getChikenDraw() {
		this.Chikens[tile] += config.CHIKEN_TYPE_DRAW
	}
	// 前后鸡
	for _, tile := range this.getChikenFB() {
		this.Chikens[tile] += config.CHIKEN_TYPE_FB
	}
	/*
		// 本鸡
		if this.setting.IsSettingChikenSelf() {
			tile := this.getChikenDrawTile()
			if tile > 0 && card.IsSuit(tile) {
				this.Chikens[tile] += config.CHIKEN_TYPE_SELF
			}
		}
		// 钻石鸡（红中）
		if this.setting.EnableDiamondChiken {
			if this.getDiamondChikenTimes(card.MAHJONG_RED) > 0 {
				this.Chikens[card.MAHJONG_RED] += config.CHIKEN_TYPE_DIAMOND
			}
		}
		// 滚筒鸡
		for _, tile := range this.getChikenTumbling() {
			this.Chikens[tile] += config.CHIKEN_TYPE_TUMBLING
		}
		// 星期鸡
		if this.setting.IsSettingChikenWeekday() {
			for _, tile := range this.chiken.GetWeekChikens() {
				this.Chikens[tile] += config.CHIKEN_TYPE_WEEK
			}
		}
		// 银鸡
		if this.setting.EnableSilverChiken {
			chikens := this.getDrawRelationChikens()
			if util.IntInSlice(card.MAHJONG_CRAK1, chikens) {
				this.Chikens[card.MAHJONG_CRAK1] += config.CHIKEN_TYPE_SILVER
			}
			if util.IntInSlice(card.MAHJONG_DOT1, chikens) {
				this.Chikens[card.MAHJONG_DOT1] += config.CHIKEN_TYPE_SILVER
			}
		}
		// 爬坡鸡
		if this.setting.IsSettingPaPoChiken() {
			if tile := this.getPaPoChiken(); tile > 0 {
				this.Chikens[tile] += config.CHIKEN_TYPE_PAPO
			}
		}
		// 见7挖
		if this.setting.IsSettingJ7W() {

		}
		// 高挖弹
		if this.setting.IsSettingGWD() {

		}

		// 补花鸡
		this.Chikens[card.MAHJONG_RED_FLOWER] += config.CHIKEN_TYPE_FLOWER_RED
	*/
	core.Logger.Debug("本局的鸡,roomId:%v,round:%v,chikens:%#v", this.RoomId, this.Round, this.Chikens)

	for _, user := range this.getUsers() {
		var target, way = this.getKitchenAndTarget(user.UserId)

		// 记录用户的输赢鸡状态 0:不包不赢,1:赢鸡,2:包鸡
		user.KitchenStatus = way
		switch {
		case way == 1 && target != nil:
			this.clacWinKitchenScore(user.UserId, target)
		case way == 2 && target != nil:
			this.clacBaoJiScore(user.UserId, target)
		default:
		}
	}
}

func (this *Mahjong) getChargeUserId(pai int) int {
	if pai == card.MAHJONG_BAM1 {
		return this.getChikenChargeBam1()
	}
	if pai == card.MAHJONG_DOT8 {
		return this.getChikenChargeDot8()
	}
	return 0
}

func (this *Mahjong) clacBaoJiScore(userId int, target []int) {
	if !this.setting.EnableBaoChiken {
		return
	}
	this.clacBaoJi(userId, card.MAHJONG_BAM1, config.SCORE_TYPE_BAO_KITCHEN, target)
	if this.setting.IsSettingChikenDot8() {
		this.clacBaoJi(userId, card.MAHJONG_DOT8, config.SCORE_TYPE_BAO_WUGU_KITCHEN, target)
	}
}

func (this *Mahjong) clacBaoJi(userId int, pai, kType int, target []int) {
	var dk, sk, _ = this.findKitchen(userId, pai)
	var times, cTimes = this.calcTimes(pai)
	// 幺鸡需要判断责任鸡(包责任鸡, 多包打冲锋鸡的人一份)
	if pai == card.MAHJONG_BAM1 {
		var pong = this.getChikenResponsibilityTarget()
		var da = this.getChikenResponsibility()
		if pong == userId && util.IntInSlice(da, target) {
			if this.setting.IsSettingFirstCycleChiken() && this.isFirstCycle() {
				// 包首圈责任鸡
				this.SetScoreItem(userId, []int{da}, 0-(config.SCORE_TYPE_BAO_RB_FIRST_CYCLE_KITCHEN+times), 1, pai, 1)
			} else {
				// 包责任鸡
				this.SetScoreItem(userId, []int{da}, 0-(config.SCORE_TYPE_BAO_RB_KITCHEN+times), 1, pai, 1)
			}
		}
	}
	if userId == this.getChargeUserId(pai) {
		if this.setting.IsSettingFirstCycleChiken() && this.isFirstCycle() && pai == card.MAHJONG_BAM1 {
			// 包首圈冲锋鸡
			this.SetScoreItem(userId, target, config.SCORE_TYPE_BAO_FIRST_CYCLE_CHARGE_KITCHEN+cTimes, 1, pai, 1)
		} else {
			this.SetScoreItem(userId, target, 0-(this.getBaoChargeType(kType)+cTimes), 1, pai, 1)
		}
		dk = dk - 1
	}
	if cTimes != times && dk != 0 {
		this.SetScoreItem(userId, target, 0-(kType+cTimes), dk, pai, dk)
		dk = 0
	}
	if dk+sk != 0 {
		this.SetScoreItem(userId, target, 0-(kType+times), dk+sk, pai, dk+sk)
		dk, sk = 0, 0
	}
}

// 计算用户赢的鸡的条目和分数
func (this *Mahjong) clacWinKitchen(userId, pai, kType int, target []int) {
	// 找出用户弃牌、明牌、手牌中的鸡
	// 如果没有开满堂鸡，那么只有1条和8筒才统计弃牌鸡
	var dk, sk, hk = this.findKitchen(userId, pai)

	// // 明牌和手牌处理逻辑一致，放到手牌处理
	// hk = sk + hk
	// 获取翻倍数量(金鸡, 代表的是配置文件中的距离,不是表面翻倍的意思)
	var times, cTimes int = this.calcTimes(pai)
	// 如果是幺鸡，需要获取责任鸡
	if kType == config.SCORE_TYPE_KITCHEN && userId == this.getChikenResponsibilityTarget() {
		// 碰了责任鸡多赢打责任鸡的人一个鸡, 所以没有 dk = dk - 1
		if this.setting.IsSettingFirstCycleChiken() && this.isFirstCycle() {
			// 首圈责任鸡
			this.SetScoreItem(userId, []int{this.getChikenResponsibility()}, config.SCORE_TYPE_RB_FIRST_CYCLE_KITCHEN+times, 1, pai, 1)
		} else {
			// 非首圈责任鸡
			this.SetScoreItem(userId, []int{this.getChikenResponsibility()}, config.SCORE_TYPE_RB_KITCHEN+times, 1, pai, 1)
		}
	}
	// 冲锋鸡
	if userId == this.getChargeUserId(pai) {
		if this.setting.IsSettingFirstCycleChiken() && this.isFirstCycle() && kType == config.SCORE_TYPE_KITCHEN {
			// 首圈冲锋鸡
			this.SetScoreItem(userId, target, config.SCORE_TYPE_FIRST_CYCLE_CHARGE_KITCHEN+cTimes, 1, pai, 1)
		} else {
			// 普通冲锋鸡
			this.SetScoreItem(userId, target, this.getChargeType(kType)+cTimes, 1, pai, 1)
		}
		dk = dk - 1
	}
	// 站鸡
	if kType == config.SCORE_TYPE_KITCHEN && this.setting.IsSettingStandChiken() {
		if hk > 0 {
			this.SetScoreItem(userId, target, config.SCORE_TYPE_STAND_KITCHEN+times, hk, pai, hk)
			hk = 0
		}
	}

	// 没开满堂鸡，但是有弃牌鸡，说明肯定是一条或者8筒
	if cTimes != times && dk != 0 {
		this.SetScoreItem(userId, target, kType+cTimes, dk, pai, dk)
		dk = 0
	}

	if dk+hk+sk != 0 {
		this.SetScoreItem(userId, target, kType+times, dk+hk+sk, pai, dk+hk+sk)
		dk, hk = 0, 0
	}
}

func (this *Mahjong) clacOtherKitchen(userId, pai, kType int, target []int) {
	// 排除非other类型的鸡
	if pai == card.MAHJONG_BAM1 || // 一条
		(this.setting.IsSettingChikenDot8() && pai == card.MAHJONG_DOT8) || // 乌骨
		(this.setting.EnableSilverChiken && // 支持银鸡
			((pai == card.MAHJONG_CRAK1 && util.IntInSlice(pai, this.getDrawRelationChikens())) || // 一万在不在翻牌相关的数组里
				(pai == card.MAHJONG_DOT1 && util.IntInSlice(pai, this.getDrawRelationChikens())))) { // 一筒在不在翻牌相关的数组里
		return
	}
	// 除了乌骨鸡，幺鸡 其余的鸡不看弃牌(非满堂鸡情况下)
	var dk, sk, hk = this.findKitchen(userId, pai)
	if dk+sk+hk != 0 {
		this.SetScoreItem(userId, target, kType, dk+sk+hk, pai, dk+sk+hk)
	}
}

// 计算单独计算的鸡的积分
// 星期鸡、本鸡
func (this *Mahjong) clacIndependentKitchen(userId, pai, kType int, target []int) {
	// 除了乌骨鸡，幺鸡 其余的鸡不看弃牌(非满堂鸡情况下)
	var dk, sk, hk = this.findKitchen(userId, pai)
	if dk+sk+hk != 0 {
		this.SetScoreItem(userId, target, kType, dk+sk+hk, pai, dk+sk+hk)
	}
}

// 计算单独计算的鸡的积分
// 爬坡鸡
func (this *Mahjong) clacPaPoKitchen(userId, pai, kType int, target []int) {
	// 除了乌骨鸡，幺鸡 其余的鸡不看弃牌(非满堂鸡情况下)
	var dk, sk, hk = this.findKitchen(userId, pai)
	if dk+sk+hk != 0 {
		this.SetScoreItem(userId, target, kType, (dk+sk+hk)*(pai%10), pai, dk+sk+hk)
	}
}

// 计算银鸡的分
func (m *Mahjong) clacWinSilverKitchen(userId, pai, kType int, target []int) {
	// 找出用户弃牌、明牌、手牌中的鸡
	// 如果没有开满堂鸡，那么只有1条和8筒才统计弃牌鸡
	var dk, sk, hk = m.findKitchen(userId, pai)
	if dk+sk+hk != 0 {
		// 获取翻倍数量(金鸡, 代表的是配置文件中的距离,不是表面翻倍的意思)
		var times = m.clacSilverTimes(pai)
		m.SetScoreItem(userId, target, kType+times-1, dk+sk+hk, pai, dk+sk+hk)
	}

}

// 补花鸡
func (m *Mahjong) calcFlowerKitch(userId int, pai int, target []int) {
	// 获取补花鸡数量
	cnt := m.getUser(userId).ShowCardList.GetTileCnt(card.MAHJONG_RED_FLOWER)
	if cnt > 0 {
		scoreType := m.getFlowerChikenScoreType(userId)
		m.SetScoreItem(userId, target, scoreType, cnt, card.MAHJONG_RED_FLOWER, cnt)
	}
}

func (this *Mahjong) getChargeType(kType int) int {
	if kType == config.SCORE_TYPE_KITCHEN {
		return config.SCORE_TYPE_CHARGE_KITCHEN
	}
	if kType == config.SCORE_TYPE_WUGU_KITCHEN {
		return config.SCORE_TYPE_CHARGE_WUGU_KITCHEN
	}
	return 0
}

func (this *Mahjong) getBaoChargeType(kType int) int {
	if kType == config.SCORE_TYPE_BAO_KITCHEN {
		return config.SCORE_TYPE_BAO_CHARGE_KITCHEN
	}
	if kType == config.SCORE_TYPE_BAO_WUGU_KITCHEN {
		return config.SCORE_TYPE_BAO_CHARGE_WUGU_KITCHEN
	}
	return 0
}

func (this *Mahjong) clacWinKitchenScore(userId int, target []int) {
	// 幺鸡
	this.clacWinKitchen(userId, card.MAHJONG_BAM1, config.SCORE_TYPE_KITCHEN, target)
	// 乌骨鸡
	if this.setting.IsSettingChikenDot8() {
		this.clacWinKitchen(userId, card.MAHJONG_DOT8, config.SCORE_TYPE_WUGU_KITCHEN, target)
	}
	// 银鸡
	if this.setting.EnableSilverChiken {
		chikens := this.getDrawRelationChikens()
		if util.IntInSlice(card.MAHJONG_CRAK1, chikens) {
			this.clacWinSilverKitchen(userId, card.MAHJONG_CRAK1, config.SCORE_TYPE_SILVER_CHIKEN, target)
		}
		if util.IntInSlice(card.MAHJONG_DOT1, chikens) {
			this.clacWinSilverKitchen(userId, card.MAHJONG_DOT1, config.SCORE_TYPE_SILVER_CHIKEN, target)
		}
	}
	// 翻牌鸡
	var ji = this.getChikenDraw()
	for _, j := range ji {
		if len(ji) == 2 {
			this.clacOtherKitchen(userId, j, config.SCORE_TYPE_UD_KITCHEN, target)
		} else if len(ji) == 1 {
			this.clacOtherKitchen(userId, j, config.SCORE_TYPE_FP_KITCHEN, target)
		}
	}
	// 前后鸡
	if this.setting.IsSettingChikenFB() {
		var ji = this.getChikenFB()
		for _, j := range ji {
			this.clacOtherKitchen(userId, j, config.SCORE_TYPE_FB_KITCHEN, target)
		}
	}
	// 滚筒鸡
	if this.setting.IsSettingChikenTumbling() {
		for ji := range this.getChikenTumbling() {
			this.clacOtherKitchen(userId, ji, config.SCORE_TYPE_GT_KITCHEN, target)
		}
	}
	// 本鸡
	if this.setting.IsSettingChikenSelf() {
		if ji := this.getChikenDrawTile(); ji > 0 && card.IsSuit(ji) {
			this.clacIndependentKitchen(userId, ji, config.SCORE_TYPE_B_KITCHEN, target)
		}
	}
	// 钻石鸡(红中)
	if this.setting.EnableDiamondChiken {
		if times := this.getDiamondChikenTimes(card.MAHJONG_RED); times > 0 {
			this.clacIndependentKitchen(userId, card.MAHJONG_RED, config.SCORE_TYPE_DIAMOND_KITCHEN+times-1, target)
		}
	}
	// 星期鸡
	if this.setting.IsSettingChikenWeekday() {
		for _, ji := range this.chiken.GetWeekChikens() {
			this.clacIndependentKitchen(userId, ji, config.SCORE_TYPE_XQ_KITCHEN, target)
		}
	}
	// 爬坡鸡
	if this.setting.IsSettingPaPoChiken() {
		if ji := this.getPaPoChiken(); ji > 0 {
			this.clacPaPoKitchen(userId, ji, config.SCORE_TYPE_PP_KITCHEN, target)
		}
	}
	// 补花鸡
	this.calcFlowerKitch(userId, card.MAHJONG_RED_FLOWER, target)
	// 见7挖
	if this.setting.IsSettingJ7W() {
		for _, tile := range this.chiken.GetJ7W() {
			this.clacOtherKitchen(userId, tile, config.SCORE_TYPE_J7W, target)
		}
	}
	// 高挖弹
	if this.setting.IsSettingGWD() {
		for _, tile := range this.chiken.GetGWD() {
			this.clacOtherKitchen(userId, tile, config.SCORE_TYPE_GWD, target)
		}
	}
	// 清一色加3只鸡
	if this.setting.IsSettingQE() {
		mu := this.getUser(userId)
		qing := qingCheck(mu.HandTileList.ToSlice(), mu.ShowCardList.GetAll())
		// 计算是否满足全鸡算清一色规则
		if !qing && this.setting.EnableFullChiken && this.hasFullChiken(mu) {
			qing = true
		}
		if qing {
			this.SetScoreItem(userId, target, config.SCORE_TYPE_QING_EXTRA, 1, 0, 0)
		}
	}
}

func (this *Mahjong) getBeiBaoQueId() []int {
	return this.getTingAndNoshaoJiId()
}

func (m *Mahjong) clacSilverTimes(pai int) int {
	var times int
	if util.IntInSlice(pai, m.getChikenDraw()) {
		times++
	}
	if m.setting.IsSettingChikenFB() {
		if util.IntInSlice(pai, m.getChikenFB()) {
			times++
		}
	} else if m.setting.IsSettingChikenTumbling() {
		if util.IntInSlice(pai, m.getChikenTumbling()) {
			times++
		}
	}
	// 本鸡
	if m.setting.IsSettingChikenSelf() {
		if drawTile := m.getChikenDrawTile(); drawTile == pai {
			times++
		}
		if fbTile := m.getChikenFBTile(); fbTile == pai {
			times++
		}
		// TODO 暂不支持滚筒鸡
	}
	return times
}

func (this *Mahjong) calcTimes(pai int) (times, cTimes int) {
	var f []int = this.getChikenDraw()
	for _, j := range f {
		if j == pai {
			times = times + 1
		}
	}
	if this.setting.IsSettingChikenFB() {
		var g = this.getChikenFB()
		if util.IntInSlice(pai, g) {
			times = times + 1
		}
	} else if this.setting.IsSettingChikenTumbling() {
		var t = this.getChikenTumbling()
		if util.IntInSlice(pai, t) {
			times = times + 1
		}
	}

	if this.setting.IsSettingAllChikenDraw() {
		cTimes = times
	} else {
		cTimes = 0
	}
	return times, cTimes
}

func (this *Mahjong) clacHu() {
	if len(this.HInfo.Loser) == 0 {
		return
	}

	// 是否是自摸
	zimoFlag := this.isZimoHu()

	for _, u := range this.HInfo.WInfo {
		var user *MahjongUser = this.getUser(u.Id)
		var score int = 0
		// 天胡地胡算分
		// 胡牌、自摸、抢杠、热炮的时候，做一下地胡判断
		switch u.Way {
		case config.HU_WAY_TIAN:
			score += 20
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_TIANHU, 1, 0, 0)
		case config.HU_WAY_DI:
			score += 10
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_DIHU, 1, 0, 0)
			// 杠上开花
		case config.HU_WAY_KONG_DRAW:
			// 支持杠上开花作为独立牌型时，杠上开花不再计算牌型分
			if !this.setting.EnableKongAfterDraw {
				this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_KONG_DRAW, 1, 0, 0)
				score++
			}
		case config.HU_WAY_DRAW:
			fallthrough
		case config.HU_WAY_QIANG_KONG:
			fallthrough
		case config.HU_WAY_RE_PAO:
			fallthrough
		case config.HU_WAY_PAO:
			if user.canDihu() {
				score += 10
				this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_DIHU, 1, 0, 0)
			}
		}

		// 报听算分
		if user.MTC.IsBaoTing() {
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_BAOTING, 1, 0, 0)
			score += 10
		}
		// 清一色算分
		if u.QFlag {
			score += 10
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_QING, 1, 0, 0)
		}
		// 胡牌类型算分
		switch u.WType {
		case config.HU_TYPE_KONG_DRAW:
			score += 11
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_KONG_DRAW, 1, 0, 0)
		case config.HU_TYPE_SHUANG_LONG_7DUI:
			score += 30
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_SHUANGLONG_QIDUI, 1, 0, 0)
		case config.HU_TYPE_HEPU_7DUI:
			score += 20
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_HEPU_QIDUI, 1, 0, 0)
		case config.HU_TYPE_LONG_7DUI:
			score += 20
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_LONGQIDUI, 1, 0, 0)
		case config.HU_TYPE_7DUI:
			score += 10
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_QIDUI, 1, 0, 0)
		case config.HU_TYPE_DIQIDUI:
			score += 20
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_DIQIDUI, 1, 0, 0)
		case config.HU_TYPE_DANDIAO:
			score += 10
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_DANDIAO, 1, 0, 0)
		case config.HU_TYPE_DADUI:
			score += 5
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_DADUI, 1, 0, 0)
		case config.HU_TYPE_BIANKADIAO:
			score += 3
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_BIANKADIAO, 1, 0, 0)
		case config.HU_TYPE_DAKUANZHANG:
			score += 4
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_DAKUANZHANG, 1, 0, 0)
		default:
		}

		if score == 0 || score == 1 {
			if zimoFlag && this.setting.EnablePinghuZimo {
				this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_PINGHU_ZIMO, 1, 0, 0)
			} else {
				this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_PINGHU, 1, 0, 0)
			}
		}

		// 胡牌为庄家, 毕节麻将不分庄闲
		if this.setting.IsSettingRemainDealer() {
			if this.Dealer == user.UserId {
				this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_CONTINUE_DEALER, this.DealerCount, 0, 0)
			}
		}

		// 自摸加一分
		if this.setting.IsSettingZME() && zimoFlag {
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_ZIMO_EXTRA, 1, 0, 0)
		}

		// 牌型通三加分
		if this.setting.IsSettingTS() {
			if u.QFlag || util.IntInSlice(u.WType, config.TS_WTYPE_EXTRA_SCORE) {
				this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_TONG_SAN, 1, 0, 0)
			}
		}
	}
}

// 满鸡分
func (this *Mahjong) clacFullChiken() {
	for _, u := range this.HInfo.WInfo {
		var user *MahjongUser = this.getUser(u.Id)
		if this.hasFullChiken(user) {
			this.SetScoreItem(user.UserId, this.HInfo.Loser, config.SCORE_TYPE_FULL_CHIKEN, 1, 0, 0)
			core.Logger.Debug("EnableFullChiken,用户拥有满鸡,roomId:%v,round:%v,userId:%v", this.RoomId, this.Round, user.UserId)
		}
	}
}

func (this *Mahjong) clacShaBao() {
	for _, id := range this.HInfo.Loser {
		if this.getUser(id).MTC.IsBaoTing() {
			// 杀报分
			this.SetScoreItem(id, this.getWinner(), 0-config.SCORE_TYPE_SHABAO, 1, 0, 0)
			// 杀报牌型分
			if shaBaoScoreType := this.clacShaBaoScoreType(id); shaBaoScoreType > 0 {
				this.SetScoreItem(id, this.getWinner(), 0-shaBaoScoreType, 1, 0, 0)
			}
		}
	}
}

// 计算杀报的积分条目
func (this *Mahjong) clacShaBaoScoreType(id int) int {
	wType, tile := this.getTingType(id)
	mu := this.getUser(id)
	shabaoQFlag := qingCheck(append(mu.HandTileList.ToSlice(), tile), mu.ShowCardList.GetAll())
	if wType == config.HU_TYPE_PI && !shabaoQFlag {
		// 平胡不算杀报牌型
		return 0
	}
	scoreType := wType + config.SHABAO_HU_TYPE_OFFSET
	if shabaoQFlag {
		scoreType += config.SHABAO_HU_TYPE_QING_OFFSET
	}
	return scoreType
}

// 黄牌也需要在此计算庄家陪庄的情况
func (this *Mahjong) clacDealerLose() {
	for _, j := range this.HInfo.Loser {
		if this.Dealer == j {
			this.SetScoreItem(this.Dealer, this.getWinner(), 0-config.SCORE_TYPE_CONTINUE_DEALER_LOSE, this.DealerCount, 0, 0)
		}
	}
}

func (this *Mahjong) clacRePao() {
	// 不是热炮
	if this.HInfo.WInfo[0].Way != config.HU_WAY_RE_PAO {
		return
	}
	var winner int = this.HInfo.Loser[0]

	mu := this.getUser(winner)
	muserLen := this.getUsersLen()

	if mu.KongCode == fbsCommon.OperationCodeKONG {
		this.SetScoreItem(winner, this.getWinner(), 0-config.SCORE_TYPE_REPAO_EXPOSED, 1, 0, 0)
	} else if mu.KongCode == fbsCommon.OperationCodeKONG_TURN {
		this.SetScoreItem(winner, this.getWinner(), 0-config.SCORE_TYPE_REPAO_TURN, muserLen-1, 0, 0)
	} else {
		this.SetScoreItem(winner, this.getWinner(), 0-config.SCORE_TYPE_REPAO_DARK, muserLen-1, 0, 0)
	}
}

func (this *Mahjong) getWinner() []int {
	var winner = []int{}
	for _, j := range this.HInfo.WInfo {
		winner = append(winner, j.Id)
	}
	return winner
}

func (this *Mahjong) clacBaoKong() {
	// 杠幺鸡的人，未听牌
	for _, mu := range this.getUsers() {
		if mu.MTC.IsNormal() {
			for _, scard := range mu.ShowCardList.GetAll() {
				if !oc.IsKongOperation(scard.GetOpCode()) {
					continue
				}
				// 跳过憨包杠
				if scard.GetOpCode() == fbsCommon.OperationCodeKONG_TURN_FREE {
					continue
				}
				var targets []int
				var baoTargets []int
				if scard.GetTarget() == mu.UserId {
					// 暗杠或者转弯杠，对象是其他人
					targets = this.getOtherUserId(scard.GetTarget())
				} else {
					// 明杠对象是电杠的人
					targets = append(targets, scard.GetTarget())
				}
				for _, target := range targets {
					// 是否需要对方听牌才包
					if this.setting.BaoKongNeedTing && !this.canWinKongOrKitchen(target) {
						continue
					}
					baoTargets = append(baoTargets, target)
				}
				// 有包杠对象
				if len(baoTargets) > 0 {
					this.SetScoreItem(mu.UserId, baoTargets, -1*getBaoKongScoreType(scard.GetOpCode()), 1, scard.GetTile(), scard.GetTilesLen())
				}
			}
		}
	}
}

// 包杠幺鸡
func (this *Mahjong) clacBaoKongBam1() {
	// 杠幺鸡的人，未听牌
	for _, user := range this.getUsers() {
		if user.MTC.IsNormal() {
			var t, target, tilesLen = this.getBaoKongBam1TypeAndTarget(user.UserId)
			if t != 0 && target != nil {
				this.SetScoreItem(user.UserId, target, -1*getBaoKongScoreType(t), 1, card.MAHJONG_BAM1, tilesLen)
			}
		}
	}
}

func (this *Mahjong) getBaoKongBam1TypeAndTarget(id int) (kongType int, target []int, tilesLen int) {
	for _, v := range this.getUser(id).ShowCardList.GetAll() {
		if !oc.IsKongOperation(v.GetOpCode()) || v.GetTile() != card.MAHJONG_BAM1 {
			continue
		}
		var idSlice []int
		if v.GetTarget() == id {
			idSlice = this.getOtherUserId(v.GetTarget())
		} else {
			idSlice = append(idSlice, v.GetTarget())
		}
		for _, u := range idSlice {
			if this.canWinKongOrKitchen(u) {
				target = append(target, u)
			}
		}
		kongType = v.GetOpCode()
		tilesLen = v.GetTilesLen()
		break
	}
	return
}

func (this *Mahjong) clacKong() {
	for id, user := range this.getUsers() {
		if this.canWinKong(user) {
			for _, j := range user.ShowCardList.GetAll() {
				switch j.GetOpCode() {
				case fbsCommon.OperationCodeKONG: // 明杠
					this.SetScoreItem(id, []int{j.GetTarget()}, config.SCORE_TYPE_KONG_EXPOSED, 1, j.GetTile(), 4)
				case fbsCommon.OperationCodeKONG_DARK: // 暗杠
					this.SetScoreItem(id, this.getOtherUserId(id), config.SCORE_TYPE_KONG_DARK, 1, j.GetTile(), 4)
				case fbsCommon.OperationCodeKONG_TURN: // 转弯杠
					this.SetScoreItem(id, this.getOtherUserId(id), config.SCORE_TYPE_KONG_TURN, 1, j.GetTile(), 4)
				case fbsCommon.OperationCodeKONG_TURN_FREE: // 憨包杠不收
				default:
				}
			}
		} else {
			// 被抢杠陪积分
			if len(this.HInfo.WInfo) > 0 && this.HInfo.WInfo[0].Way == config.HU_WAY_QIANG_KONG && this.HInfo.Loser[0] == id {
				this.SetScoreItem(id, this.getWinner(), 0-config.SCORE_TYPE_KONG_QIANG, this.setting.GetSettingPlayerCnt()-1, 0, 0)
			}
		}
	}
}

// 是否可以赢杠
func (this *Mahjong) canWinKong(mu *MahjongUser) bool {
	// 黄牌杠不算钱
	if !this.HInfo.Hu {
		return false
	}
	// 是否烧鸡烧杠
	if this.shaoJi(mu.UserId) {
		return false
	}
	// 未叫牌
	if mu.MTC.IsNormal() {
		// 如果支持单查、且打完了缺
		if this.setting.IsSettingDanCha() &&
			mu.LackTile > 0 && getLackCount(mu.LackTile, mu.HandTileList.ToSlice()) == 0 {
			// do nothing
		} else {
			return false
		}
	}
	return true
}

func (this *Mahjong) canWinKongOrKitchen(id int) bool {
	// 未叫牌, 黄牌
	if this.getUser(id).MTC.IsNormal() || !this.HInfo.Hu {
		return false
	}
	if this.shaoJi(id) {
		return false
	}
	return true
}

func (this *Mahjong) getTingAndNoshaoJiId() []int {
	var winner = []int{}
	for _, u := range this.getUsers() {
		if u.MTC.IsTing() && !this.shaoJi(u.UserId) {
			winner = append(winner, u.UserId)
		}
	}
	return winner
}

func (this *Mahjong) getNoshaoJiId(excludeUserId int) []int {
	var winner = []int{}
	for _, u := range this.getUsers() {
		if u.UserId != excludeUserId && !this.shaoJi(u.UserId) {
			winner = append(winner, u.UserId)
		}
	}
	return winner
}

// 判断用户是否被烧鸡
// 被抢杠或者放热炮的用户，会被烧鸡
func (this *Mahjong) shaoJi(id int) bool {
	if !this.HInfo.Hu {
		return false
	}
	// 被抢杠
	if this.HInfo.WInfo[0].Way == config.HU_WAY_QIANG_KONG && this.HInfo.Loser[0] == id {
		return true
	}
	// 点了热炮
	if this.HInfo.WInfo[0].Way == config.HU_WAY_RE_PAO && this.HInfo.Loser[0] == id {
		return true
	}
	return false
}

// 同一个积分条目，在不同玩法中显示不同的条目
// 没有转换规则的话，就返回原id
func (this *Mahjong) frontScoreTypeTrans(id int) int {
	if m, exists := config.FrontScoreTypeTrans[this.MType]; exists {
		if v, exists := m[id]; exists {
			id = v
		}
	}
	return id
}

// 用户是否有满鸡
func (m *Mahjong) hasFullChiken(mu *MahjongUser) bool {
	// 统计用户所拥有的牌的数量
	// 包括手牌、明牌、如果支持满堂鸡的话，也统计弃牌
	tiles := mu.getMergedTileMap(m.setting.IsSettingAllChikenDraw())
	core.Logger.Debug("getMergedTileMap,roomId:%v,userId:%v,tileMap:%+v", m.RoomId, mu.UserId, tiles)
	// 检查翻牌鸡
	for _, chiken := range m.getChikenDraw() {
		if card.IsFlower(chiken) {
			continue
		}
		if tiles[chiken] == 4 {
			return true
		}
	}
	// 检查前后鸡
	if m.setting.IsSettingChikenFB() {
		for _, chiken := range m.getChikenFB() {
			if card.IsFlower(chiken) {
				continue
			}
			if tiles[chiken] == 4 {
				return true
			}
		}
	}
	// TODO 检查滚筒鸡
	return false
}

// 获取补花鸡的类型
func (m *Mahjong) getFlowerChikenScoreType(userId int) int {
	scoreType := config.SCORE_TYPE_FLOWER_RED_KITCHEN_TING
	// 判断用户是否胡牌
	if m.HInfo.Hu == true && m.HInfo.isWinner(userId) {
		scoreType += 2
	}
	// 判断是否是钻石鸡
	if m.getChikenDrawTile() == card.MAHJONG_RED_FLOWER {
		scoreType += 4
	}
	return scoreType
}
