package api

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"server/common/async"
	"server/common/pkg/redislock"
	"server/common/response"
	"server/common/utils"
	"server/service"

	"github.com/gin-gonic/gin"
)

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type ChangeAvatarPayload struct {
	UserID string `json:"user_id"`
	NewURL string `json:"new_url"`
	Token  string `json:"token"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	if err := service.Register(req.Username, req.Password, req.Email); err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	data, err := service.Login(req.Username, req.Password)
	if err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, data)
}

func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, "Invalid request parameters")
		return
	}

	userID := c.GetString("user_id")
	if err := service.ChangePassword(req.OldPassword, req.NewPassword, userID); err != nil {
		response.HandleServiceError(c, err)
		return
	}

	response.Success(c, nil)
}

func ChangeAvatar(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetString("user_id")

	lockKey := utils.BuildLockKey("user", userID, "change_avatar")
	token, ok, err := redislock.AcquireLock(ctx, lockKey, 30*time.Second)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "Failed to acquire lock")
		return
	}
	if !ok {
		response.Fail(c, http.StatusTooManyRequests, "Avatar update in progress")
		return
	}

	file, err := c.FormFile("new_avatar")
	if err != nil {
		redislock.ReleaseLock(context.Background(), lockKey, token)

		response.Fail(c, http.StatusBadRequest, "Missing picture file")
		return
	}

	filename := fmt.Sprintf("%s_%d_%s", userID, time.Now().Unix(), file.Filename)
	dst := filepath.Join("uploads", filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		redislock.ReleaseLock(context.Background(), lockKey, token)
		response.Fail(c, http.StatusInternalServerError, "Failed to save file")
		return
	}

	url := "http://localhost:8080/uploads/" + filename

	err = async.EnqueueTask("change_avatar", ChangeAvatarPayload{
		UserID: userID,
		NewURL: url,
		Token:  token,
	})

	if err != nil {
		redislock.ReleaseLock(context.Background(), lockKey, token)

		response.Fail(c, http.StatusInternalServerError, "Failed to enqueue task")
		return
	}

	response.Success(c, gin.H{"avatar_url": url})

}
