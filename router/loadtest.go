package router

import (
	"github.com/changqings/gin-web/handler/loadtest"
	"github.com/gin-gonic/gin"
)

func LoadtestRouter(app *gin.Engine) {

	cpuLoad := loadtest.NewLoadTestCpu()
	memoryLoad := loadtest.NewLoadtestMemory()

	//
	loadGroup := app.Group("/loadtest")

	// cpu query params
	// localhost:8080/loadtest/cpu?duration=120
	loadGroup.GET("/cpu", cpuLoad.Run())

	// memory query params
	// localhost:8080/loadtest/memroy?size=512&duration=120
	loadGroup.GET("/memory", memoryLoad.Run())

}
