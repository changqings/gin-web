package db_sql

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	pg_dsn    = "host=192.168.1.201 port=5432 user=dev password=devdevdev dbname=dev01 sslmode=disable TimeZone=Asia/Shanghai"
	PG_CLIENT = &gorm.DB{}
)

type PersonRoom struct {
	gorm.Model
	Name   string `gorm:"uniqueIndex,not null" json:"name"`
	Owner  string `json:"owner"`
	Size   int    `json:"size"`
	HasNet bool   `gorm:"default:false" json:"net"`
	// HomeThings []HomeThing `gorm:"embedded;embeddedPrefix:home_thing_"`
}

type HomeThing struct {
	Name  string
	Price float64
}

type PgClient struct {
	DB *gorm.DB
}

func init() {
	PG_CLIENT = GetPgClient()
	if err := PG_CLIENT.AutoMigrate(&PersonRoom{}); err != nil {
		slog.Error("auto migrate &personRoom{}", "msg", err)
	}

}

func GetPgClient() *gorm.DB {

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: pg_dsn,
		// and you can add another config
	}))

	if err != nil {
		log.Fatalf("get pg client error:%v", err)
	}
	return db
}

// insert by json body
//
//	{
//	    "name": "lucy_home",
//	    "owner": "lucy",
//	    "net": true,
//	    "size": 40
//	}
func (p *PgClient) InsertJson() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		pr := &PersonRoom{}

		err := ctx.ShouldBind(pr)
		if err != nil {
			ctx.AbortWithError(478, err)
			return
		}

		r := p.DB.Model(&PersonRoom{}).Where("name = ?", pr.Name).FirstOrCreate(pr)
		fmt.Printf("debug insert: %v\n", r.RowsAffected)

		if r.Error != nil {
			ctx.AbortWithError(478, r.Error)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"msg": "insert db data success."})

	}
}

// create if not have, insert by c.Query()
func (p *PgClient) Insert() gin.HandlerFunc {

	return func(ctx *gin.Context) {

		size_str := ctx.Query("size")
		net_str := ctx.Query("net")
		name := ctx.Query("name")
		size, err := strconv.ParseUint(size_str, 10, 64)
		net, err1 := strconv.ParseBool(net_str)

		if len(name) == 0 || err != nil || err1 != nil {
			ctx.AbortWithError(479, errors.Join(err, err1))
			return
		}

		pr := &PersonRoom{
			Name:   ctx.Query("name"),
			Owner:  ctx.Query("owner"),
			Size:   int(size),
			HasNet: net,
		}

		r := p.DB.Model(&PersonRoom{}).Where("name = ?", pr.Name).FirstOrCreate(pr)
		// fmt.Printf("debug insert: %v\n", r.RowsAffected)

		if r.Error != nil {
			ctx.AbortWithError(479, r.Error)
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"msg": "insert db data success."})

	}
}

// read use uniqKey name
func (p *PgClient) Read() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		name := ctx.Query("name")
		if len(name) == 0 {
			ctx.AbortWithError(404, errors.ErrUnsupported)
			return
		}

		var pr PersonRoom
		r := p.DB.Model(&PersonRoom{}).Where("name = ?", name).First(&pr)
		if r.Error != nil {
			ctx.AbortWithError(404, r.Error)
			return
		}

		ctx.JSON(http.StatusOK, pr)

	}
}

// update with json body
//
//	{
//	    "name": "lucy_home",
//	    "owner": "lucy",
//	    "net": false,
//	    "size": 50
//	}
func (p *PgClient) Update() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		pr := &PersonRoom{}

		if err := ctx.ShouldBind(pr); err != nil {
			ctx.AbortWithError(480, err)
			return
		}

		r := p.DB.Model(&PersonRoom{}).Where("name = ?", pr.Name).Updates(pr)
		if r.Error != nil {
			ctx.AbortWithError(480, r.Error)
			return
		}
		ctx.JSON(http.StatusOK, pr)

	}
}

// delete use uniqKey name
// curl -X POST http://127.0.0.1:8080/pg/delete?name=lucy_home
func (p *PgClient) Delete() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		name := ctx.Query("name")
		if len(name) == 0 {
			ctx.AbortWithStatusJSON(481, gin.H{"errorMsg": "query params.name must have value."})
			return
		}

		r := p.DB.Where("name = ?", name).Delete(&PersonRoom{})
		if r.Error != nil {
			ctx.AbortWithError(481, r.Error)
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"msg": "name = " + name + " deleted."})
	}

}
