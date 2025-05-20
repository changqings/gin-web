package middleware

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Middle_01() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		slog.Info("start m_01")
		ctx.Next()
		slog.Info("end m_01")
	}
}

func Middle_02() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		slog.Info("start m_02")
		ctx.Next()
		slog.Info("end m_02")
	}
}

func Middle_03() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		slog.Info("start m_03")
		ctx.Next()
		slog.Info("end m_03")
	}
}

func QuerySpendTime() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now().UnixMilli()
		ctx.Next()
		end := time.Now().UnixMilli()
		slog.Info("query spend time", "msg", fmt.Sprintf("%dms", end-start))
	}
}
