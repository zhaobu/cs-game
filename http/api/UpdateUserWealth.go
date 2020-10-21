package api

import (
	"game/db/mgo"
	pbhall "game/pb/hall"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

// UpdateWealthReq 更新财富
type UpdateUserWealth struct {
	UserID       uint64 `json:"UserId" form:"UserId"`
	SourceType   uint32 `json:"SourceType" form:"SourceType"`
	SourceData   string `json:"SourceData" form:"SourceData"`
	CurrencyType uint32 `json:"CurrencyType" form:"CurrencyType"`
	Change       int64  `json:"Change" form:"Change"`
	Sign         string `json:"sign"  form:"sign"`
}

func UpdateUserWealthReq(c *gin.Context) {
	req := UpdateUserWealth{}
	c.ShouldBindJSON(&req)
	sign := MakeMD5(fmt.Sprintf("%d%d%d%d%s", req.UserID,req.SourceType,req.CurrencyType,req.Change, Key))
	if sign != req.Sign {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: "签名错误!"})
		return
	}
	//写入流水记录
	err := mgo.WriteWealthRecordData(req.UserID, req.SourceType, req.SourceData,req.CurrencyType,req.Change)
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: err.Error()})
		return
	}
	userInfo,err := mgo.UpdateWealthByNet(req.UserID,req.CurrencyType,req.Change)
	if err != nil {
		c.JSON(http.StatusOK, apiRsp{Code: Failed, Msg: err.Error()})
		return
	}
	//通知客户端
	if req.CurrencyType == 2 {
		err = ToGateNormal(&pbhall.UserWealthChange{
			UserID:        userInfo.UserID,
			Gold:          userInfo.Gold,
			GoldChange:    req.Change,
			Masonry:       userInfo.Masonry,
			MasonryChange: 0,
		},req.UserID)
	}else{
		err = ToGateNormal(&pbhall.UserWealthChange{
			UserID:        userInfo.UserID,
			Gold:          userInfo.Gold,
			GoldChange:    0,
			Masonry:       userInfo.Masonry,
			MasonryChange: req.Change,
		},req.UserID)
	}
	c.JSON(http.StatusOK, apiRsp{Code: Succ, Msg: "成功", Data: nil})
}
