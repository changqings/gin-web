package main

import (
	"log"

	"github.com/changqings/gin-web/handle_func"
	"github.com/changqings/gin-web/routers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {

	// gin config
	gin.SetMode(gin.DebugMode)
	app := gin.Default()

	// middlewares write or find from offical
	// and you can find some offical on `https://github.com/gin-gonic/contrib`
	app.Use(cors.Default())
	// middle.Limiter(1*time.Second),
	// middle.Middle_01(),
	// middle.Middle_02(),
	// middle.Middle_03())
	// middle.QuerySpendTime())

	// simple mothed usage
	app.GET("/getname", handle_func.GetName("scq"))
	app.GET("/json", handle_func.P_list())

	// security usage
	{
		sec_group := app.Group("/sec")
		sec_group.Use(gin.BasicAuth(gin.Accounts{
			"user01": "PasSw0rd!",
		}))
		sec_group.GET("/info", handle_func.Some_sec_info())
	}

	// metrics usage
	// {
	// 	routers.TxMetrics(app)
	// }

	// pgsql usage
	{
		routers.PgRouters(app)
	}

	// main run
	err := app.Run("127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
}
