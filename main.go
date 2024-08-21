package main

import (
	"io"
	"kanban/config"
	"kanban/controllers"
	"kanban/models"
	"kanban/routers"
	"os"
	"os/user"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-colorable"
	"github.com/sirupsen/logrus"
)

func main() {

	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	runmod := "prod"
	configPath := "/config/config.json"
	if user.Name != "root" {
		configPath = "./config/config.json"
		runmod = "dev"
	} else {
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = io.Discard
	}
	initLogger()
	config.SetConfig("runmod", runmod)
	config.SetConfig("configs", configPath)

	config.LoadConfig(runmod)

	models.SetDB(config.GetConnectionString())

	models.User{}.Migrate()
	models.Issue{}.Migrate()
	models.Project{}.Migrate()
	models.Comment{}.Migrate()
	models.Issue{}.Migrate()
	models.Notifications{}.Migrate()
	models.ProjectUser{}.Migrate()
	models.WorkSpace{}.Migrate()

	controllers.LoadTemplates()
	router := gin.Default()
	routers.Load(router)

	if err := router.Run(":8000"); err != nil {
		logrus.Fatal(err)
		os.Exit(1)
	}

}

func initLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetOutput(colorable.NewColorableStdout())
	if gin.Mode() == gin.DebugMode {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
