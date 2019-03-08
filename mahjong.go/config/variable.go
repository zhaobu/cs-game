package config

import (
	fbsCommon "mahjong.go/fbs/Common"
)

//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
//-* 麻将相关
//-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-*-
var (
	// 局数价格定义, {牌局数: 对应价格, ...}
	MahjongRoundPrice                   = map[int]int{0: 0, 4: 2, 8: 3, 12: 4}
	MahjongLDRoundPrice                 = map[int]int{0: 0, 4: 2, 8: 3, 12: 4}
	MahjongRoundPriceDragonBoatFestival = map[int]int{0: 0, 4: 1, 8: 2, 12: 3}
	// 随机组局价格
	MahjongRandomPrice = 5
	// 随机组局价格-送审
	MahjongRandomPriceForReview = 5
	MahjongMatchPrice           = map[int]int{
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_1: 1,  // 贵阳麻将比赛场: 1倍积分
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_2: 3,  // 贵阳麻将比赛场: 3倍积分
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_3: 5,  // 贵阳麻将比赛场: 5倍积分
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_4: 10, // 贵阳麻将比赛场: 10倍积分
	}
	RoomPunishmentScore = 10 // 比赛未同意解散惩罚分

	// 随机组局支持的类型
	RandomGameTypeList = []int{
		fbsCommon.GameTypeMAHJONG_GY,  // 贵阳麻将
		fbsCommon.GameTypeMAHJONG_BJ,  // 毕节麻将
		fbsCommon.GameTypeMAHJONG_ZY,  // 遵义麻将
		fbsCommon.GameTypeMAHJONG_SD,  // 三丁拐
		fbsCommon.GameTypeMAHJONG_LD,  // 两丁拐
		fbsCommon.GameTypeMAHJONG_STT, // 72张
	}

	// 随机组局的默认设置
	// 按顺序，0：满堂鸡；1：连庄；2：上下鸡；3：乌骨鸡；4：前后鸡；5：星期鸡；6：意外鸡；7：吹风鸡；8：滚筒鸡；9：麻将人数；10：麻将张数；11：本鸡
	RandomRoomDefaultSettingList = map[int][]int{
		fbsCommon.GameTypeMAHJONG_GY: []int{
			1,                    // 0满堂鸡
			1,                    // 1连庄
			1,                    // 2上下鸡
			1,                    // 3乌骨鸡
			1,                    // 4前后鸡
			0,                    // 5星期鸡
			0,                    // 6意外鸡
			0,                    // 7吹分鸡
			0,                    // 8滚筒鸡
			4,                    // 9麻将人数
			MAHJONG_TILE_CNT_108, // 10麻将张数
			0,                    // 11本鸡
			0,                    // 12站鸡
			0,                    // 13翻倍鸡
			0,                    // 14首圈冲锋鸡
			0,                    // 15清一色奖励三分
			0,                    // 16自摸翻倍
			0,                    // 17自摸加1分
			0,                    // 18通三
			0,                    // 19大牌翻倍
			0,                    // 20：包杠
			0,                    // 21：爬坡鸡
			0,                    // 22：查缺不查叫
			0,                    // 23: 见7挖
			0,                    // 24: 高挖弹
			0,                    // 25: 龙七对奖3分
			0,                    // 26: 最后一局翻倍
			3,                    // 27:换3张
			0,                    // 28:换4张
		},
		fbsCommon.GameTypeMAHJONG_ZY:  []int{1, 1, 1, 1, 1, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0},
		fbsCommon.GameTypeMAHJONG_BJ:  []int{1, 1, 1, 1, 1, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0},
		fbsCommon.GameTypeMAHJONG_SD:  []int{1, 1, 1, 1, 1, 0, 0, 0, 0, 3, MAHJONG_TILE_CNT_108, 0},
		fbsCommon.GameTypeMAHJONG_LD:  []int{1, 1, 1, 1, 1, 0, 0, 0, 0, 2, MAHJONG_TILE_CNT_108, 0},
		fbsCommon.GameTypeMAHJONG_STT: []int{1, 1, 1, 1, 1, 0, 0, 0, 0, 3, MAHJONG_TILE_CNT_72, 0},
	}

	// 比赛支持的类型
	MatchGameTypeList = []int{
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_1, // 贵阳麻将比赛场: 1倍积分
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_2, // 贵阳麻将比赛场: 3倍积分
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_3, // 贵阳麻将比赛场: 5倍积分
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_4, // 贵阳麻将比赛场: 10倍积分
	}

	// 比赛游戏的默认设置
	MatchRoomDefaultSettingList = map[int][]int{
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_1: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0},
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_2: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0},
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_3: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0},
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_4: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0},
	}

	// 比赛的积分倍数配置
	MatchScoreMultipleList = map[int]int{
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_1: 1,  // 贵阳麻将比赛场: 1倍积分
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_2: 3,  // 贵阳麻将比赛场: 3倍积分
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_3: 5,  // 贵阳麻将比赛场: 5倍积分
		fbsCommon.GameTypeMAHJONG_MATCH_GZ_4: 10, // 贵阳麻将比赛场: 10倍积分
	}

	// 金币场的默认设置
	CoinRoomDefaultSettingList = map[int][]int{
		// 贵阳麻将，满堂鸡、连庄
		fbsCommon.GameTypeMAHJONG_GY: []int{1, 1, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		// 遵义麻将，满堂鸡
		fbsCommon.GameTypeMAHJONG_ZY: []int{1, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		// 毕节麻将
		fbsCommon.GameTypeMAHJONG_BJ: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		// 两丁拐
		fbsCommon.GameTypeMAHJONG_LD: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 2, MAHJONG_TILE_CNT_108, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		// 三丁拐
		fbsCommon.GameTypeMAHJONG_SD: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 3, MAHJONG_TILE_CNT_108, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		// 两房
		fbsCommon.GameTypeMAHJONG_STT: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 3, MAHJONG_TILE_CNT_72, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		// 安顺
		fbsCommon.GameTypeMAHJONG_AS: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0},
		// 兴义麻将
		fbsCommon.GameTypeMAHJONG_XY: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0, 0},
		// 六盘水
		fbsCommon.GameTypeMAHJONG_LPS: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0},
		// 凯里
		fbsCommon.GameTypeMAHJONG_KL: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
		// 都匀
		fbsCommon.GameTypeMAHJONG_DY: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0, 0, 0, 1, 0, 0, 1, 0, 0, 0, 0},
		// 铜仁
		fbsCommon.GameTypeMAHJONG_TR: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_108, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0},
		// 黔西
		fbsCommon.GameTypeMAHJONG_QX: []int{1, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_112, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		// 金沙
		fbsCommon.GameTypeMAHJONG_JS: []int{1, 0, 0, 0, 0, 0, 0, 0, 0, 4, MAHJONG_TILE_CNT_112, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1},
	}

	// 牌型大小排序(关联胡牌类型)
	// value越小，牌型越大
	HuTypeSort = map[int]int{
		HU_TYPE_KONG_DRAW:        5,
		HU_TYPE_SHUANG_LONG_7DUI: 10,
		HU_TYPE_LONG_7DUI:        20,
		HU_TYPE_HEPU_7DUI:        29,
		HU_TYPE_7DUI:             30,
		HU_TYPE_DIQIDUI:          40,
		HU_TYPE_DANDIAO:          50,
		HU_TYPE_DADUI:            60,
		HU_TYPE_BIANKADIAO:       70,
		HU_TYPE_DAKUANZHANG:      80,
		HU_TYPE_PI:               90,
	}
)
