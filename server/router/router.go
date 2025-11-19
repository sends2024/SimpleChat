package router

import (
	"server/api"
	"server/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.Static("/uploads", "./uploads")

	r.Use(middlewares.CORS())

	origen := r.Group("/api")

	origen.POST("/users/register", api.Register)
	origen.POST("/users/login", api.Login)

	user := origen.Group("/user")
	user.Use(middlewares.Auth())
	{
		user.PATCH("/password", api.ChangePassword)
		user.PUT("/avatar", api.ChangeAvatar)
	}

	channels := origen.Group("/channels")
	channels.Use(middlewares.Auth())
	{

		channels.POST("/join", api.JoinChannel)
		channels.POST("/:channel_id/leave", api.LeaveChannel)
		channels.GET("/list", api.GetChannels)
		channels.GET("/:channel_id/history", api.GetHistory)
		channels.GET("/:channel_id/members", api.GetMembers)
	}

	channel := origen.Group("/channel")
	channel.Use(middlewares.Auth())
	{
		channel.POST("/create", api.CreateChannel)
		channel.DELETE("/:channel_id", api.DeleteChannel)
		channel.DELETE("/:channel_id/member/:user_id", api.RemoveMember)
		channel.GET("/invite", api.GetInviteCode)
		channel.PATCH("/:channel_id", api.ChangeChannelName)
	}

	return r
}
