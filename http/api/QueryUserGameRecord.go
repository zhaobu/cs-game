package api

import (
	"game/db/mgo"
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
	"time"
)

//查询用户游戏记录

// 请求参数
type QueryUserGameRecordParam struct {
	DrawID    uint64 `json:"DrawID" form:"DrawID"`
	Sign      string `json:"sign"  form:"sign"`
}

//返回参数
type QueryUserGameRecordReturn struct {
	DrawID uint64 `json:"DrawID" form:"DrawID"`
	CreateTime string `json:"sign"  form:"sign"`
	List []*QueryUserGameRecordPlayerReturn `json:"List"  form:"List"`
}

type QueryUserGameRecordPlayerReturn struct {
	UserID uint64
	NickName string
	HeadImgUrl string
	GetScore int32
}


//设置用户红名接口
func QueryUserGameRecordReq(c *gin.Context){
	req := QueryUserGameRecordParam{}
	c.ShouldBindJSON(&req)
	sign := MakeMD5(fmt.Sprintf("%d%s", req.DrawID, Key))
	if sign != req.Sign {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: "签名错误!"})
		return
	}
	rdata,err := mgo.QueryRoomRecordByRoom(req.DrawID)
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: err.Error()})
		return
	}

	result := QueryUserGameRecordReturn{
		DrawID:req.DrawID,
		CreateTime:time.Unix(rdata.GameStartTime, 0).Format("2006-01-02 15:04:05"),
		List:make([]*QueryUserGameRecordPlayerReturn,len(rdata.GamePlayers)),
	}
	for i,v := range rdata.GamePlayers{
		udata,err := mgo.QueryUserInfo(v.UserId)
		result.List[i] = &QueryUserGameRecordPlayerReturn{
			UserID:v.UserId,
			NickName: v.Name,
			GetScore: v.PreScore,
		}
		if err == nil{
			result.List[i].HeadImgUrl = udata.Profile
		}
	}
	c.JSON(http.StatusOK, apiRsp{Code: Succ, Msg: "成功", Data: result})
}