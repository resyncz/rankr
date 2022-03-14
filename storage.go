package main

import (
	"github.com/resyncz/rankr/internal/cache"
	"github.com/resyncz/rankr/internal/cache/redis"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"strings"
)

func PostgresDb(connString string) *gorm.DB {
	if strings.TrimSpace(connString) == "" {
		logrus.Error("missing connection string [env|config.yml]")
		logrus.Warn("CAUTION! service will be running without database connection!")
		return nil
	}

	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		logrus.Error("failed to connect database: ", err)
		logrus.Warn("CAUTION! service will be running without database connection!")
	}

	return db
}

func NewRedisCacheStorage(host, port, password string, cacheDb int) cache.Storage {
	redisConfig := redis.NewConfig()
	redisConfig.Host = host
	redisConfig.Port = port
	redisConfig.Password = password
	redisConfig.DB = cacheDb

	redisConn := &redis.Storage{
		Config: redisConfig,
	}

	if err := redisConn.Initialize(); err != nil {
		logrus.Fatal("failed to initialize redis connection: ", err)
	}

	return redisConn
}
