package configs

import (
	"fmt"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const PROJECT_DIR = "keeper-bot"

// LoadEnv loads env vars from .env at root of repo
func GetProjectRoot() string {
	re := regexp.MustCompile(`^(.*` + PROJECT_DIR + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))
	return string(rootPath)
}

func LoadEnv() {
	logrus.WithField("env", Environment(os.Getenv(ENV))).Infof("loading env")
	re := regexp.MustCompile(`^(.*` + PROJECT_DIR + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))
	filePath := fmt.Sprintf("%s/.env", string(rootPath))
	err := godotenv.Load(filePath)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"cwd":      cwd,
			"filePath": filePath,
		}).Warning("problem loading .env file")
	}
}
