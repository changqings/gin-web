package main

import (
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/chanqings/gin-web/middle"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// gin config
	gin.SetMode(gin.DebugMode)
	app := gin.Default()

	// middleware
	app.Use(cors.Default(), middle.Limiter(1*time.Second), middle_01(), middle_02(), middle_03())

	// routers
	app.GET("/getname", getName("scq"))
	app.GET("/json", p_list())

	// other middle will work on router below
	sec_group := app.Group("/sec")
	sec_group.Use(gin.BasicAuth(gin.Accounts{
		"user01": "PasSw0rd!",
	}))
	sec_group.GET("/info", some_sec_info())

	// main run
	if err := app.Run("127.0.0.1:8080"); err != nil {
		log.Fatal(err)
	}
}

func some_sec_info() func(*gin.Context) {
	return func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"msg": "you should learn rust",
		})
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
