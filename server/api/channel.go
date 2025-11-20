package api

import (
	"net/http"
	"server/common/response"
	"server/service"

	"github.com/gin-gonic/gin"
)

type JoinChannelRequest struct {
	InviteCode string `json:"invite_code" binding:"required"`
}

func JoinChannel(c *gin.Context) {
	var req JoinChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	userID := c.GetString("user_id")
	err := service.JoinChannel(req.InviteCode, userID)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

func LeaveChannel(c *gin.Context) {
	channelID := c.Param("channel_id")
	userID := c.GetString("user_id")

	err := service.LeaveChannel(channelID, userID)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

func GetChannels(c *gin.Context) {
	userID := c.GetString("user_id")

	data := service.GetChannels(userID)
	response.Success(c, gin.H{"channels": data})
}

func GetMembers(c *gin.Context) {
	channelID := c.Param("channel_id")

	data := service.GetMembers(channelID)
	response.Success(c, gin.H{"members": data})
}

func GetHistory(c *gin.Context) {
	channelID := c.Param("channel_id")
	lastTime := c.Query("before")

	data := service.GetHistory(lastTime, channelID)
	response.Success(c, data)
}
