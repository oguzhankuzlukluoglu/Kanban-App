package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

var (
	configSet = make(map[string]string)
	runmod    string
	configs   string
)

type Configs struct {
	Debug   Config
	Release Config
}

type Config struct {
	Public        string `json:"public"`
	Domain        string `json:"domain"`
	SessionSecret string `json:"session_secret"`
	Database      DatabaseConfig
	Sentry        Sentry
}

type DatabaseConfig struct {
	Host     string
	Name     string
	User     string
	Password string
	Port     string
}

type Sentry struct {
	Sentry string
}

var config *Config

func SetConfig(key, value string) {
	configSet[key] = value
	switch key {

	case "runmod":
		runmod = value
	case "config":
		configs = value
	}
}

func LoadConfig(runmod string) {
	data, err := ioutil.ReadFile("config/config.json")
	if err != nil {
		panic(err)
	}
	configs := &Configs{}
	err = json.Unmarshal(data, configs)
	if err != nil {
		panic(err)
	}
	if runmod == "dev" {
		gin.SetMode(gin.DebugMode)
		config = &configs.Debug

	} else if runmod == "prod" {
		gin.SetMode(gin.ReleaseMode)
		config = &configs.Release
	}

	if !path.IsAbs(config.Public) {
		workingDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		config.Public = path.Join(workingDir, config.Public)
	}
}

func GetConfig() *Config {
	return config
}

func PublicPath() string {
	return config.Public
}

func Path() string {
	workingDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return workingDir
}

func UploadsPath() string {
	return path.Join(config.Public, "uploads")
}

func GetConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=True&loc=Local",
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name)
}

func Con() {
	er := sentry.Init(sentry.ClientOptions{
		Dsn:   config.Sentry.Sentry,
		Debug: false,
	})
	if er != nil {
		log.Fatalf("sentry.Init: %s", er)
	}
	defer sentry.Flush(2 * time.Second)
}
