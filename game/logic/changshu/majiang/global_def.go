package majiang

type (
	EmOperType     uint8  //操作类型
	EmRecordAction uint8  //战绩回放
	EmHuScoreType  uint8  //胡牌得分类型
	EmtimerID      uint32 //定时器枚举
	EmHuType       uint8  //胡牌番型
	EmExtraHuType  uint8  //附属胡牌番型
	EmHuMode       uint8  //胡牌方式
	EmScoreTimes   uint8  //结算计分次数统计
)

//定时器ID
const (
	TID_DESK_Begin EmtimerID = 0 //桌子内部定时器开始
	TID_Destory                  //解散桌子

	TID_GAMESINK_Begin EmtimerID = 1000 //游戏逻辑定时器开始
	TID_DealCard                        //发牌
	TID_GameStartBuHua                  //开始补花
)

//杠类型
const (
	OperType_None      EmOperType = iota
	OperType_BU_GANG              //补杠
	OperType_AN_GANG              //暗杠
	OperType_MING_GANG            //明杠
	OperType_PENG                 //碰
	OperType_CHI                  //吃
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
	HuMode_ZIMO    EmHuMode = iota + 1 //自摸胡
	HuMode_PAOHU                       //接炮胡
	HuMode_QIANGHU                     //抢杠胡
)

//胡分类型(客户端显示用)
const (
	//得分显示
	HuScoreType_Zi_Mo          EmHuScoreType = iota + 1 //自摸
	HuScoreType_Jie_Pao                                 //接炮
	HuScoreType_Qiang_GangHu                            //抢杠胡
	HuScoreType_Gang_Shang_Hua                          //杠上花
	HuScoreType_Men_Qing                                //门清
	HuScoreType_Gang_Shang_Pao                          //杠上炮
	HuScoreType_Dui_Dui_Hu                              //对对胡
	HuScoreType_Qing_Yi_Se                              //清一色
	HuScoreType_Xiao_Qi_Dui                             //小七对
	HuScoreType_Quan_Qiu_Ren                            //全求人
	HuScoreType_Fang_Pao                                //放炮
	HuScoreType_Bei_Qiang_Gang                          //被抢杠
)

//附属胡牌类型
const (
	ExtraHuType_QiangGang    EmExtraHuType = iota + 1 //抢杠胡
	ExtraHuType_GangShangHua                          //杠上花
	ExtraHuType_GangShangPao                          //杠上炮
	ExtraHuType_MenQing                               //门清
)

// 胡牌类型(胡牌番型)
const (
	HuType_NORMAL           EmHuType = iota + 1 //普通胡
	HuType_THIRTEEN_ORPHANS                     //十三幺
	HuType_SMALL_SEVEN                          //小七对
	HuType_LUXURY_SEVEN                         //豪华七小对
	HuType_DBL_LUXURY_SEVEN                     //双豪华七小对
	HuType_TRI_LUXURY_SEVEN                     //三豪华七小对
	HuType_PENGPENGHU                           //碰碰胡
	HuType_HUNYISE                              //混一色
	HuType_QINGYISE                             //清一色
	HuType_SMALL_THREE                          //小三元
	HuType_BIG_THREE                            //大三元
	HuType_SMALL_FOUR                           //小四喜
	HuType_BIG_FOUR                             //大四喜
	HuType_EIGHTEEN_ARHAT                       //十八罗汉
	HuType_YAOJIU_QY                            //清远幺九
	HuType_QUANFENG                             //全风
	HuType_FOUR_LAIZI                           //四鬼胡牌
	HuType_YAOJIU_CZJY                          //潮汕简易幺九
	HuType_PURE_YAOJIU                          //清幺九
	HuType_ALL_ZI                               //字一色
	HuType_YAOJIU_CZMQ                          //潮州门清幺九
	HuType_MIX_YAOJIU_CZMQ                      //潮州门清混幺九
	HuType_QUANQIUREN                           //全求人
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
