package handle

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetName(name string) gin.HandlerFunc {
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

func P_list() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"code": 200,
			"msg":  "成功",
			"list": []string{"p_01", "p_02", "p_03"},
		})

	}
}

func Some_sec_info() func(*gin.Context) {
	return func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"msg": "you should learn rust",
		})
	}
}
