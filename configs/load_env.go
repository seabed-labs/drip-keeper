package configs

import (
	"fmt"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

const PROJECT_DIR = "drip-keeper"

// LoadEnv loads env vars from .env at root of repo
func GetProjectRoot() string {
	rootOverride := os.Getenv(string(PROJECT_ROOT_OVERRIDE))
	if rootOverride != "" {
		logrus.WithField("override", rootOverride).Infof("override project root")
		return rootOverride
	}
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
