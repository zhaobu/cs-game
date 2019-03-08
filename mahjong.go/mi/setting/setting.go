package setting

// MSetting 游戏设置
type MSetting struct {
	lack          bool // 是否支持定缺
	pinghu        bool // 是否支持平胡
	EnableKongTXZ bool // 明杠是否算通行证
	// 积分倍数
	Multiple      int // 积分倍数(房间)
	MultipleRound int // 倍数(局)
	// 牌型支持
	EnableShuangLongQiDui bool // 是否支持双龙七对，default:true
	EnableBianKaDiao      bool // 是否支持边卡吊，default:false
	EnableDaKuanZhang     bool // 是否支持大宽张，default:false
	EnableDi7Dui          bool // 是否支持地七对，default:false
	EnablePinghuZimo      bool // 是否支持平胡自摸，default:false
	EnableDanDiao         bool // 是否支持单吊，default:false
	EnableHePu7Dui        bool // 是否支持合浦小七对，default:false
	// 绑定玩法支持
	EnableDoubleDealer       bool // 是否支持庄家胡牌翻倍，default:false
	EnableKongAfterDraw      bool // 是否支持杠上开花，这里是指把杠上开花独立作为一个牌型，杠上开花时，不计其他牌型，default:false
	BaoKongNeedTing          bool // 是否对方叫牌才包杠，设置为false的话，不管对方有没有叫牌，都要包，default:true
	BaoChikenNeedTing        bool // 包鸡是否需要对方听牌，设置为false的话，不管对方有没有叫牌，都要包，default:true
	EnableSilverChiken       bool // 是否支持银鸡，default:false
	EnableFullChiken         bool // 是否支持满鸡，default:false
	EnableDiamondChiken      bool // 是否支持钻石鸡，default:false
	EnableBaoKongBam1        bool // 是否支持包杠幺鸡，default:false
	EnableBaoChiken          bool // 是否支持包鸡 default:true
	EnableResponsibilityBam1 bool // 是否支持责任幺鸡 default:true
	EnableInitLack           bool // 是否支持原缺 default: false
	EnableCharge             bool // 是否支持冲锋鸡 default: true
	EnableKongHotPao         bool // 明杠是否算热炮 default: true
	EnableBaoTingKong        bool // 处于报听状态，是否支持明杠 default: false
	EnableKongDouble         bool // 是否杠牌加倍

	// 按顺序，0：满堂鸡；1：连庄；2：上下鸡；3：乌骨鸡；4：前后鸡；5：星期鸡；6：意外鸡；7：吹风鸡；8：滚筒鸡；9：麻将人数；10：麻将张数；11：本鸡
	// 12：站鸡；13：翻倍鸡；14：首圈鸡；15：清一色奖励三分；16：自摸翻倍；17：自摸加一分；18：通三；19：大牌翻倍；20：包杠；21：爬坡鸡；22：查缺不查叫
	// 23: 见7挖; 24: 高挖弹; 25: 龙七对奖3分; 26: 最后一局翻倍; 27:换3张;28:换4张
	// 满堂鸡: 打出去的牌也算鸡
	// 上下鸡: 翻出来的牌，上一张也算鸡，eg: 翻出来5条，如果开启了上下鸡，则4条也算鸡
	// 乌骨鸡: 8筒
	// 前后鸡: 开局时倒数第三墩上面翻开的鸡，每杠一次，在前面一墩翻开一张并关闭当前墩
	// 星期鸡: 今天星期几，则几条、几筒、几万都算鸡，周日算7
	// 吹风鸡: 5筒，翻到5筒，所以的鸡、杠、胡都不算，直接开始下一局
	// 滚筒鸡: 开局时倒数第三墩上面翻开的鸡，每杠一次，在前面一墩翻开一张，当前墩的牌不关闭
	setting []int
}

// NewMSetting 新建一个MSetting的引用
func NewMSetting() *MSetting {
	mSetting := &MSetting{
		setting:       []int{},
		lack:          false,
		pinghu:        false,
		EnableKongTXZ: true,
		// 积分倍数
		Multiple:      1,
		MultipleRound: 1,
		// 牌型支持
		EnableShuangLongQiDui: true,
		EnableBianKaDiao:      false,
		EnableDaKuanZhang:     false,
		EnableDi7Dui:          false,
		EnablePinghuZimo:      false,
		EnableDanDiao:         true,
		EnableDoubleDealer:    false,
		EnableHePu7Dui:        false,
		// 玩法支持
		EnableKongAfterDraw:      false,
		BaoKongNeedTing:          true,
		BaoChikenNeedTing:        true,
		EnableSilverChiken:       false,
		EnableFullChiken:         false,
		EnableDiamondChiken:      false,
		EnableBaoKongBam1:        false,
		EnableBaoChiken:          true,
		EnableResponsibilityBam1: true,
		EnableInitLack:           false,
		EnableCharge:             true,
		EnableKongHotPao:         true,
		EnableBaoTingKong:        false,
	}
	return mSetting
}

// SetSetting 设置需要定缺
func (ms *MSetting) SetSetting(setting []int) {
	ms.setting = setting
}

// SetPositionValue 修改setting指定位置的值
// 若setting长度比给的postion小，则填充0至position长度
func (ms *MSetting) SetPositionValue(position int, value int) {
	sLen := len(ms.setting)
	if sLen > position {
		ms.setting[position] = value
	} else if sLen == position {
		ms.setting = append(ms.setting, value)
	} else {
		fillSlice := make([]int, position-sLen, position-sLen+1)
		fillSlice = append(fillSlice, value)
		ms.setting = append(ms.setting, fillSlice...)
	}
}

// GetSetting 获取设置内容
func (ms *MSetting) GetSetting() []int {
	return ms.setting
}

// SetEnableLack 设置需要定缺
func (ms *MSetting) SetEnableLack() {
	ms.lack = true
}

// IsEnableLack 是否支持定缺
func (ms *MSetting) IsEnableLack() bool {
	return ms.lack
}

// IsEnableExchange 是否支持定缺
func (ms *MSetting) IsEnableExchange() bool {
	return ms.IsSettingExchange3() || ms.IsSettingExchange4()
}

// SetEnablePinghu 设置需要定缺
func (ms *MSetting) SetEnablePinghu() {
	ms.pinghu = true
}

// IsEnablePinghu 是否支持定缺
func (ms *MSetting) IsEnablePinghu() bool {
	return ms.pinghu
}

// GetSettingPlayerCnt 获取设置的玩家人数
func (ms *MSetting) GetSettingPlayerCnt() int {
	return ms.setting[9]
}

// GetSettingTileCnt 获取设置的麻将张数
func (ms *MSetting) GetSettingTileCnt() int {
	return ms.setting[10]
}

// IsSettingAllChikenDraw 是否设置了满堂鸡
func (ms *MSetting) IsSettingAllChikenDraw() bool {
	return ms.setting[0] == 1
}

// IsSettingRemainDealer 是否设置了连庄
func (ms *MSetting) IsSettingRemainDealer() bool {
	return ms.setting[1] == 1
}

// IsSettingChikenUD 是否设置了上下鸡
func (ms *MSetting) IsSettingChikenUD() bool {
	return ms.setting[2] == 1
}

// IsSettingChikenDot8 是否设置了乌骨鸡
func (ms *MSetting) IsSettingChikenDot8() bool {
	return ms.setting[3] == 1
}

// IsSettingChikenFB 是否设置了前后鸡
func (ms *MSetting) IsSettingChikenFB() bool {
	return ms.setting[4] == 1
}

// IsSettingChikenWeekday 是否设置了星期鸡
func (ms *MSetting) IsSettingChikenWeekday() bool {
	return ms.setting[5] == 1
}

// IsSettingChikenUnexpect 是否设置了意外鸡
func (ms *MSetting) IsSettingChikenUnexpect() bool {
	return ms.setting[6] == 1
}

// IsSettingChikenWind 是否支持吹风鸡
func (ms *MSetting) IsSettingChikenWind() bool {
	return ms.setting[7] == 1
}

// IsSettingChikenTumbling 是否支持滚筒鸡
func (ms *MSetting) IsSettingChikenTumbling() bool {
	return ms.setting[8] == 1
}

// IsSettingChikenSelf 是否支持本鸡
func (ms *MSetting) IsSettingChikenSelf() bool {
	return len(ms.setting) >= 12 && ms.setting[11] == 1
}

// IsSettingChikenRock 是否设置了滚鸡
func (ms *MSetting) IsSettingChikenRock() bool {
	return ms.IsSettingChikenFB() || ms.IsSettingChikenTumbling()
}

// IsSettingStandChiken 是否支持站鸡
func (ms *MSetting) IsSettingStandChiken() bool {
	return len(ms.setting) >= 13 && ms.setting[12] == 1
}

// IsSettingDoubleChiken 是否支持翻倍鸡
func (ms *MSetting) IsSettingDoubleChiken() bool {
	return len(ms.setting) >= 14 && ms.setting[13] == 1
}

// IsSettingFirstCycleChiken 是否支持首圈鸡
func (ms *MSetting) IsSettingFirstCycleChiken() bool {
	return len(ms.setting) >= 15 && ms.setting[14] == 1
}

// IsSettingQE 是否支持清一色加3分
func (ms *MSetting) IsSettingQE() bool {
	return len(ms.setting) >= 16 && ms.setting[15] == 1
}

// IsSettingDoubleZM 是否支持自摸翻倍
func (ms *MSetting) IsSettingDoubleZM() bool {
	return len(ms.setting) >= 17 && ms.setting[16] == 1
}

// IsSettingZME 是否支持自摸加一分
func (ms *MSetting) IsSettingZME() bool {
	return len(ms.setting) >= 18 && ms.setting[17] == 1
}

// IsSettingTS 是否支持通三
func (ms *MSetting) IsSettingTS() bool {
	return len(ms.setting) >= 19 && ms.setting[18] == 1
}

// IsSettingDoubleDP 是否支持大牌翻倍
func (ms *MSetting) IsSettingDoubleDP() bool {
	return len(ms.setting) >= 20 && ms.setting[19] == 1
}

// IsSettingBaoKong 是否包杠
func (ms *MSetting) IsSettingBaoKong() bool {
	return len(ms.setting) >= 21 && ms.setting[20] == 1
}

// IsSettingPaPoChiken 是否支持爬坡鸡
func (ms *MSetting) IsSettingPaPoChiken() bool {
	return len(ms.setting) >= 22 && ms.setting[21] == 1
}

// IsSettingDanCha 是否支持单查
func (ms *MSetting) IsSettingDanCha() bool {
	return len(ms.setting) >= 23 && ms.setting[22] == 1
}

// IsSettingJ7W 是否设置了见7挖
func (ms *MSetting) IsSettingJ7W() bool {
	return len(ms.setting) >= 24 && ms.setting[23] == 1
}

// IsSettingGWD 是否设置了高挖弹
func (ms *MSetting) IsSettingGWD() bool {
	return len(ms.setting) >= 25 && ms.setting[24] == 1
}

// IsSettingLE 是否设置了龙七对奖3分玩法
func (ms *MSetting) IsSettingLE() bool {
	return len(ms.setting) >= 26 && ms.setting[25] == 1
}

// IsSettingDLR 是否设置了最后一局翻倍 is setting double le
func (ms *MSetting) IsSettingDLR() bool {
	return len(ms.setting) >= 27 && ms.setting[26] == 1
}

// IsSettingExchange3 是否支持换3张
func (ms *MSetting) IsSettingExchange3() bool {
	return len(ms.setting) >= 28 && ms.setting[27] > 0
}

// IsSettingExchange4 是否支持换4张
func (ms *MSetting) IsSettingExchange4() bool {
	return len(ms.setting) >= 29 && ms.setting[28] > 0
}

// GetExchangeCount 获取需要换的牌的数量
func (ms *MSetting) GetExchangeCount() int {
	var cnt int
	if ms.IsSettingExchange3() {
		cnt = 3
	} else if ms.IsSettingExchange4() {
		cnt = 4
	}
	return cnt
}
