package main

import (
	"flag"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/resyncz/rankr/internal/httpserver"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Path = flag.String("cpath", "config/", "configuration path")

func main() {
	flag.Parse()

	logrus.Info("restaurants scanner")

	InitializeViper()

	router := NewRouter()

	//db := PostgresDb(viper.GetString("database.connstring"))
	//if err := db.AutoMigrate(); err != nil {
	//	logrus.Fatal("failed to migrate *User: ", err)
	//}

	ServeHttp(":8080", router)
}

func InitializeViper() {
	viper.AddConfigPath(*Path)
	viper.SetConfigName("app")
	viper.AutomaticEnv()
	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		logrus.Fatal("viper failed to read config file: ", err)
	}
}

func NewRouter() *gin.Engine {
	router := gin.New()

	router.Use(cors.Default())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	return router
}

func ServeHttp(addr string, router *gin.Engine) {
	httpserver.ServeHttp(addr, router)
}
