package api

import (
	"net/http"

	"server/common/response"
	"server/service"

	"github.com/gin-gonic/gin"
)

type CreateChannelRequest struct {
	ChannelName string `json:"channel_name" binding:"required"`
}

type ChangeChannelNameRequest struct {
	NewName string `json:"new_name" binding:"required"`
}

func CreateChannel(c *gin.Context) {
	var req CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}
	userID := c.GetString("user_id")
	err := service.CreateChannel(req.ChannelName, userID)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

func DeleteChannel(c *gin.Context) {
	clannelID := c.Param("clannel_id")

	response.Success(c, nil)

	// TODO
	// 不想再写一个异步了，有那个意思算了
	service.DeleteChannel(clannelID)
}

func RemoveMember(c *gin.Context) {
	clannelID := c.Param("clannel_id")
	memberID := c.Param("member_id")

	service.RemoveMember(clannelID, memberID)
	response.Success(c, nil)
}

func GetInviteCode(c *gin.Context) {
	clannelID := c.Param("clannel_id")

	data, err := service.GenerateInvite(clannelID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "Failed to get invite_code")
	}
	response.Success(c, gin.H{"invite_code": data})
}

func ChangeChannelName(c *gin.Context) {
	var req ChangeChannelNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	clannelID := c.Param("clannel_id")
	service.ChangeChannelName(req.NewName, clannelID)

	response.Success(c, nil)
}
