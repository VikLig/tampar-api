package utils

import (
	"fmt"
	"tampar-api/model/constant"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort   string
	Environment  string
	AppMode      string
	LogDirectory string
	DbName       string
	DbUsername   string
	DbPassword   string
	DbUrl        string
	DbPort       string
	DbSid        string
	DPass        string
	KeyPassword  string
	KeyIv        string
	AppName      string
	// Security
	AccessControlAllowOrigin string
}

func NewEnv() Config {
	var config Config

	viper.SetConfigFile("yaml")
	viper.AddConfigPath("./")
	viper.SetConfigName("config")
	//viper.AddConfigPath("/etc/config")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config file, %s", err)
	}

	config.ServerPort = viper.GetString("serverPort")
	config.Environment = viper.GetString("environment")
	if viper.GetString("releaseMode") == "y" || viper.GetString("releaseMode") == "Y" {
		config.AppMode = gin.ReleaseMode
	} else {
		config.AppMode = gin.DebugMode
	}
	config.LogDirectory = viper.GetString("logDirectory."+config.Environment+".path") + constant.APP_NAME + " " + time.Now().Format("02-Jan-2006") + ".log"
	config.DbName = viper.GetString("database." + config.Environment + ".name")
	config.DbUsername = viper.GetString("database." + config.Environment + ".username")
	config.DbPassword = viper.GetString("database." + config.Environment + ".password")
	config.DbUrl = viper.GetString("database." + config.Environment + ".url")
	config.DbPort = viper.GetString("database." + config.Environment + ".port")
	config.DbSid = viper.GetString("database." + config.Environment + ".sid")
	config.AppName = viper.GetString("appName")

	// Security
	config.AccessControlAllowOrigin = viper.GetString("securitySetting." + config.Environment + ".accessControlAllowOrigin")

	return config
}
