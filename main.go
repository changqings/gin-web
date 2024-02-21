package main

import (
	"log/slog"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// gin config
	gin.SetMode(gin.DebugMode)
	app := gin.Default()

	// middleware
	app.Use(cors.Default(), middle_01(), middle_02(), middle_03())

	// routers
	app.GET("/getname", getName("scq"))
	app.GET("/json", p_list())

	// main run
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func p_list() func(*gin.Context) {
	return func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "成功",
			"list": []string{"p_01", "p_02", "p_03"},
		})

	}
}

func getName(name string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		if ctx.Query("version") != "" {
			ctx.JSON(200, gin.H{
				"version": ctx.Query("version"),
				"name":    name,
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"name": name,
		})

	}
}

func middle_01() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		slog.Info("start m_01")
		ctx.Next()
		slog.Info("end m_01")
	}
}

func middle_02() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		slog.Info("start m_02")
		ctx.Next()
		slog.Info("end m_02")
	}
}

func middle_03() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		slog.Info("start m_03")
		ctx.Next()
		slog.Info("end m_03")
	}
}
