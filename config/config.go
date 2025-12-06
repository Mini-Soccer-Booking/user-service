package config

import (
	"github.com/sirupsen/logrus"
	_ "github.com/spf13/viper/remote"
	"os"
	"user-service/common/util"
)

var Config AppConfig

type AppConfig struct {
	Port                  int      `json:"port"`
	AppName               string   `json:"appName"`
	AppEnv                string   `json:"appEnv"`
	SignatureKey          string   `json:"signatureKey"`
	Database              Database `json:"database"`
	RateLimiterMaxRequest int      `json:"rateLimiterMaxRequest"`
	RateLimiterTimeSecond int      `json:"rateLimiterTimeSecond"`
	JwtSecretKey          string   `json:"jwtSecretKey"`
	JwtExpirationTime     int      `json:"jwtExpirationTime"`
}

type Database struct {
	Host                   string `json:"host"`
	Port                   int    `json:"port"`
	Name                   string `json:"name"`
	Username               string `json:"username"`
	Password               string `json:"password"`
	MaxOpenConnections     int    `json:"maxOpenConnections"`
	MaxLifeTimeConnections int    `json:"maxLifeTimeConnections"`
	MaxIdleConnections     int    `json:"maxIdleConnections"`
	MaxIdleTime            int    `json:"maxIdleTime"`
}

func Init() {
	err := util.BindFromJson(&Config, "config.json", ".")
	if err != nil {
		logrus.Infof("failed to bind config, %v.", err)
		err = util.BindFromConsul(&Config, os.Getenv("CONSUL_HTTP_URL"), os.Getenv("CONSUL_HTTP_KEY"))
	}
}
