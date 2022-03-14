package conf

import (
	"github.com/spf13/viper"
)

func InitializeViper(path, confName, confType string) error {
	viper.AddConfigPath(path)
	viper.SetConfigName(confName)
	viper.SetConfigType(confType)
	viper.AutomaticEnv()

	return viper.ReadInConfig()
}
