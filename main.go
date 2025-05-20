package main

import (
	"flag"
	"log"
	"log/slog"
	"time"

	"github.com/changqings/gin-web/router"

	"github.com/changqings/gin-web/handler/clbmetrics"
	"github.com/changqings/gin-web/handler/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func etcdTask() {
	// // first check as master or wait
	// startCh := make(chan int, 1)

	// ticker := time.NewTicker(time.Second * 10)
	// if db.ShouldRunAsMaster {
	// 	close(startCh)
	// } else {
	// 	slog.Info("waitting to be master...")
	// 	for range ticker.C {
	// 		// slog.Info("main debug", "shouldRunAsMaster", db.ShouldRunAsMaster)
	// 		if db.ShouldRunAsMaster {
	// 			close(startCh)
	// 			// 执行ticker.Stop()并不会关闭通信，只会不继续发送, 要手动退出循环
	// 			ticker.Stop()
	// 			break
	// 		}
	// 	}
	// }

	// <-startCh
	// slog.Info("running as master...")

	//// lock task
	etcd := db.NewEtcd()
	go db.LockTask01(etcd)
	go db.LockTask02(etcd)

	time.Sleep(time.Second * 10)
	slog.Info("run etcd task completed", "task01", "lockTask01", "task01", "lockTask02")
}

var (
	enableEtcd bool
)

func main() {
	flag.BoolVar(&enableEtcd, "etcd", false, "enable etcd")
	flag.Parse()
	if enableEtcd {
		etcdTask()
		// master election
		etcd := db.NewEtcd()
		etcd.Campaign(ginWebServer)
	} else {
		ginWebServer()
	}

}

func ginWebServer() {

	// gin config
	gin.SetMode(gin.DebugMode)
	app := gin.Default()

	// middlewares write or find from offical
	// and you can find some offical on `https://github.com/gin-gonic/contrib`
	app.Use(cors.Default())
	// middleware.Limiter(1*time.Second),
	// middleware.Middle_01(),
	// middleeare.Middle_02(),
	// middleware.Middle_03())
	// middleware.QuerySpendTime())

	// simple mothed usage
	app.GET("/getname", clbmetrics.GetName("scq"))
	app.GET("/json", clbmetrics.P_list())

	// security usage
	{
		sec_group := app.Group("/sec")
		sec_group.Use(gin.BasicAuth(gin.Accounts{
			"user01": "PasSw0rd!",
		}))

		sec_group.GET("/info", clbmetrics.Some_sec_info())
	}

	// metrics usage
	// {
	// 	if err := router.TxMetrics(app); err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// pgsql usage
	{
		router.PgRouters(app)
	}

	// cicd
	{
		router.CICDRouter(app)
	}

	// loadtest
	{
		router.LoadtestRouter(app)
	}

	// main run
	err := app.Run(":8080")
	if err != nil {
		log.Fatal(err)
	}

}
