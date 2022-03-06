package configs

import (
	"fmt"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const ENV = "ENV"

// LoadEnv loads env vars from .env at root of repo
func GetProjectRoot() string {
	re := regexp.MustCompile(`^(.*` + PROJECT_DIR + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))
	return string(rootPath)
}

func LoadEnv() {
	env := Environment(os.Getenv(ENV))
	logrus.WithField("env", env).Infof("loading env")
	if !IsLocal(env) {
		logrus.WithField("env", env).Debug("skipping .env file")
		return
	}
	re := regexp.MustCompile(`^(.*` + PROJECT_DIR + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))
	filePath := fmt.Sprintf("%s/.env", string(rootPath))
	err := godotenv.Load(filePath)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"cwd":      cwd,
			"filePath": filePath,
		}).Fatal("problem loading .env file")
		os.Exit(-1)
	}
}
