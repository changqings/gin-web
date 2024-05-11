package main

import (
	"log"
	"log/slog"
	"time"

	"github.com/changqings/gin-web/db"
	"github.com/changqings/gin-web/handle"
	"github.com/changqings/gin-web/router"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// first check is master or wait
	startCh := make(chan int, 1)

	ticker := time.NewTicker(time.Second * 15)
	if db.ShouldRunAsMaster {
		startCh <- 1
	} else {
		slog.Info("wait to be master...")
		for range ticker.C {
			if db.ShouldRunAsMaster {
				startCh <- 1
			}
		}
	}

	<-startCh
	slog.Info("run as master...")

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
	app.GET("/getname", handle.GetName("scq"))
	app.GET("/json", handle.P_list())

	// security usage
	{
		sec_group := app.Group("/sec")
		sec_group.Use(gin.BasicAuth(gin.Accounts{
			"user01": "PasSw0rd!",
		}))

		sec_group.GET("/info", handle.Some_sec_info())
	}

	// metrics usage
	// {
	// 	routers.TxMetrics(app)
	// }

	// pgsql usage
	{
		router.PgRouters(app)
	}

	// cicd
	{
		router.CICDRouter(app)
	}

	// main run
	err := app.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
