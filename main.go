package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/resyncz/rankr/internal/conf"
	"github.com/resyncz/rankr/internal/httpserver"
	"github.com/resyncz/rankr/internal/web"
	"github.com/sirupsen/logrus"
	"net/http"
)

const (
	confPath = "conf/"
	confName = "app"
	confType = "yml"
)

func main() {
	flag.Parse()

	if err := conf.InitializeViper(confPath, confName, confType); err != nil {
		logrus.Fatal(err)
	}

	router := web.NewRouter()

	//db := PostgresDb(viper.GetString("database.connstring"))
	//if err := db.AutoMigrate(); err != nil {
	//	logrus.Fatal("failed to migrate *User: ", err)
	//}

	router.GET("", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, ":)")
	})

	httpserver.ServeHttp(":8080", router)
}
