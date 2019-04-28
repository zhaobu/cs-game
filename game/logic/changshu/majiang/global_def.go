package majiang

type (
	EmOperType     uint8  //操作类型
	EmRecordAction uint8  //战绩回放
	EmHuScoreType  uint8  //胡牌得分类型
	EmtimerID      uint32 //定时器枚举
	EmHuType       uint8  //胡牌番型
	EmHuMode       uint8  //胡牌方式
	EmHuModeTag    uint8  //胡牌方式标记(抢杠胡,杠上开花,杠上炮等)
	EmScoreTimes   uint8  //结算计分次数统计
)

//定时器ID
const (
	TID_None    EmtimerID = iota
	TID_Common            //通用定时器id,用于不需要取消的定时
	TID_Destory           //解散桌子
	// TID_DealCard                        //发牌
	// TID_GameStartBuHua                  //开始补花
)

//杠类型
const (
	OperType_None      EmOperType = iota
	OperType_BU_GANG              //补杠
	OperType_AN_GANG              //暗杠
	OperType_MING_GANG            //明杠
	OperType_PENG                 //碰
	OperType_LCHI                 //左吃
	OperType_MCHI                 //中吃
	OperType_RCHI                 //右吃
)

//游戏记录相关
const (
	RecordACTION_NONE EmRecordAction = iota
	RecordACTION_DRAW                //摸牌
	RecordACTION_OUT                 //出牌
	RecordACTION_PENG                //碰
	RecordACTION_GANG                //杠
	RecordACTION_CHI                 //吃
	RecordACTION_HU                  //胡
	RecordACTION_PASS                //过
)

//胡牌方式
const (
	HuMode_None  EmHuMode = iota
	HuMode_ZIMO           //自摸胡
	HuMode_PAOHU          //接炮胡
)

//胡牌时特殊标记
const (
	HuModeTag_None         EmHuModeTag = iota
	HuModeTag_QiangGangHu              //抢杠胡
	HuModeTag_GangShangHua             //杠上花
	HuModeTag_GangShangPao             //杠上炮
	HuModeTag_HaiDiLaoYue              //海底捞月
)

//胡分类型(客户端显示用)
const (
	//得分显示
	HuScoreType_None           EmHuType = iota
	HuScoreType_Zi_Mo                   //自摸
	HuScoreType_Jie_Pao                 //接炮
	HuScoreType_Qiang_GangHu            //抢杠胡
	HuScoreType_Gang_Shang_Hua          //杠上花
	HuScoreType_Men_Qing                //门清
	HuScoreType_Gang_Shang_Pao          //杠上炮
	HuScoreType_Dui_Dui_Hu              //对对胡
	HuScoreType_Qing_Yi_Se              //清一色
	HuScoreType_Xiao_Qi_Dui             //小七对
	HuScoreType_Quan_Qiu_Ren            //全求人
	HuScoreType_Fang_Pao                //放炮
	HuScoreType_Bei_Qiang_Gang          //被抢杠
)

// 胡牌类型(胡牌番型)
const (
	HuType_None            EmHuType = iota
	HuType_Normal                   //普通胡
	HuType_ShiSanYao                //十三幺
	HuType_XiaoQiDui                //小七对
	HuType_MenQing                  //门清
	HuType_QingYiSe                 //清一色
	HuType_ZiYiSe                   //字一色
	HuType_HunYiSe                  //混一色
	HuType_DuiDuiHu                 //对对胡
	HuType_GangShangKaiHua          //杠上开花
	HuType_DaDiaoChe                //大吊车
	HuType_HaiDiLaoYue              //海底捞月
	HuType_QiangGangHu              //抢杠胡
	HuType_GangShangPao             //杠上炮
)

//统计次数
const (
	ScoreTimes_None EmScoreTimes = iota
	ScoreTimes_HuPai
	ScoreTimes_JiePao
	ScoreTimes_DianPao
	ScoreTimes_AnGang
	ScoreTimes_MingGang
	ScoreTimes_BuGang
	ScoreTimes_Win
	ScoreTimes_Lose
	ScoreTimes_ZiMo
)
