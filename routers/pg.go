package routers

import (
	"github.com/changqings/gin-web/db_sql"
	"github.com/gin-gonic/gin"
)

func PgRouters(app *gin.Engine) {

	pg := db_sql.PgClient{DB: db_sql.PG_CLIENT}

	pgGroup := app.Group("/pg")
	pgGroup.PUT("/insert", pg.Insert())
	pgGroup.POST("/insert_json", pg.InsertJson())
	pgGroup.POST("/delete", pg.Delete())
	pgGroup.POST("/update", pg.Update())
	pgGroup.GET("/read", pg.Read())

}
