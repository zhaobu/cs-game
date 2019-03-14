package game_logic

type (
	emOperType     uint8
	emRecordAction uint8
	emHuScoreType  uint8
)

//杠类型
const (
	SUO_GANG  emOperType = iota //补杠
	AN_GANG                     //暗杠
	MING_GANG                   //明杠
	PENG                        //碰
	CHI                         //吃
)

//游戏记录相关
const (
	ACTION_DRAW emRecordAction = iota //摸牌
	ACTION_OUT                 = 2    //出牌
	ACTION_PENG                = 3    //碰
	ACTION_GANG                = 4    //杠
	ACTION_CHI                 = 5    //吃
	ACTION_HU                  = 6    //胡
	ACTION_PASS                = 7    //过
)

//胡分类型(客户端显示用)
const (
	//得分显示
	Zi_Mo          emHuScoreType = iota //自摸
	Jie_Pao                      = 2    //接炮
	Qiang_GangHu                 = 3    //抢杠胡
	Gang_Shang_Hua               = 4    //杠上花
	Men_Qing                     = 5    //门清
	Gang_Shang_Pao               = 6    //杠上炮
	Dui_Dui_Hu                   = 7    //对对胡
	Qing_Yi_Se                   = 8    //清一色
	Xiao_Qi_Dui                  = 9    //小七对
	Quan_Qiu_Ren                 = 10   //全求人
	Fang_Pao                     = 11   //放炮
	Bei_Qiang_Gang               = 12   //被抢杠
)
