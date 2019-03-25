package majiang

type (
	EmOperType     uint8
	EmRecordAction uint8
	EmHuScoreType  uint8  //胡牌得分类型
	EmtimerID      uint32 //定时器枚举
	EmHuType       uint8  //胡牌得分类型
	EmHuMode       uint8  //胡牌方式
)

//定时器ID
const (
	TID_DESK_Begin EmtimerID = 0 //桌子内部定时器开始
	TID_Destory                  //销毁桌子

	TID_GAMESINK_Begin EmtimerID = 1000 //游戏逻辑定时器开始
	TID_DealCard                        //发牌
	TID_GameStartBuHua                  //开始补花
)

//杠类型
const (
	SUO_GANG  EmOperType = iota + 1 //补杠
	AN_GANG                         //暗杠
	MING_GANG                       //明杠
	PENG                            //碰
	CHI                             //吃
)

//游戏记录相关
const (
	ACTION_DRAW EmRecordAction = iota + 1 //摸牌
	ACTION_OUT                            //出牌
	ACTION_PENG                           //碰
	ACTION_GANG                           //杠
	ACTION_CHI                            //吃
	ACTION_HU                             //胡
	ACTION_PASS                           //过
)

//胡牌方式
const (
	ZIMO    EmHuMode = iota + 1 //自摸胡
	PAOHU                       //接炮胡
	QIANGHU                     //抢杠胡
)

//胡分类型(客户端显示用)
const (
	//得分显示
	Zi_Mo          EmHuScoreType = iota + 1 //自摸
	Jie_Pao                                 //接炮
	Qiang_GangHu                            //抢杠胡
	Gang_Shang_Hua                          //杠上花
	Men_Qing                                //门清
	Gang_Shang_Pao                          //杠上炮
	Dui_Dui_Hu                              //对对胡
	Qing_Yi_Se                              //清一色
	Xiao_Qi_Dui                             //小七对
	Quan_Qiu_Ren                            //全求人
	Fang_Pao                                //放炮
	Bei_Qiang_Gang                          //被抢杠
)

// 胡牌类型(胡牌番型)
const (
	NORMAL           EmHuType = iota + 1 //普通胡
	THIRTEEN_ORPHANS                     //十三幺
	SMALL_SEVEN                          //小七对
	LUXURY_SEVEN                         //豪华七小对
	DBL_LUXURY_SEVEN                     //双豪华七小对
	TRI_LUXURY_SEVEN                     //三豪华七小对
	PENGPENGHU                           //碰碰胡
	HUNYISE                              //混一色
	QINGYISE                             //清一色
	SMALL_THREE                          //小三元
	BIG_THREE                            //大三元
	SMALL_FOUR                           //小四喜
	BIG_FOUR                             //大四喜
	EIGHTEEN_ARHAT                       //十八罗汉
	YAOJIU_QY                            //清远幺九
	QUANFENG                             //全风
	FOUR_LAIZI                           //四鬼胡牌
	YAOJIU_CZJY                          //潮汕简易幺九
	PURE_YAOJIU                          //清幺九
	ALL_ZI                               //字一色
	YAOJIU_CZMQ                          //潮州门清幺九
	MIX_YAOJIU_CZMQ                      //潮州门清混幺九
	QUANQIUREN                           //全求人
)
