package api

import (
	"cy/game/db/mgo"
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
)

//查询用户参与游戏信息


// 请求参数
type QueryUserParGameParam struct {
	UserID    uint64 `json:"UserId" form:"UserId"`
	Sign      string `json:"sign"  form:"sign"`
}

//返回参数
type QueryUserParGameReturn struct {
	UserID uint64
	WinCase uint64
	TotalCase uint64
}

//设置用户红名接口
func QueryUserParGameInfoReq(c *gin.Context){
	req := QueryUserParGameParam{}
	c.ShouldBindJSON(&req)
	sign := MakeMD5(fmt.Sprintf("%d%s", req.UserID, Key))
	if sign != req.Sign {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: "签名错误!"})
		return
	}
	udata,err := mgo.QueryUserInfo(req.UserID)
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: err.Error()})
		return
	}
	c.JSON(http.StatusOK, apiRsp{Code: Succ, Msg: "成功", Data:&QueryUserParGameReturn{UserID: req.UserID, WinCase: udata.WinCaseCase, TotalCase: udata.TotalCase}})
}