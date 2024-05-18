package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/changqings/gin-web/db"
	"github.com/changqings/gin-web/handle"
	"github.com/changqings/gin-web/router"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func init() {
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

	etcd := db.NewEtcd()
	etcd.Campaign(backgroundTask)

}

func backgroundTask() error {
	stopCh, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	for {
		select {
		case <-stopCh.Done():
			slog.Info("get exit signal, good bye.")
			os.Exit(0)
		default:
			slog.Info("background task running...")
			time.Sleep(time.Second * 10)

		}
	}
}

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
