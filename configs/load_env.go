package configs

import (
	"fmt"
	"os"
	"regexp"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

// LoadEnv loads env vars from .env at root of repo
func GetProjectRoot() string {
	re := regexp.MustCompile(`^(.*` + PROJECT_DIR + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))
	return string(rootPath)
}

func LoadEnv() {
	env := Environment(os.Getenv("KEEPER_BOT_ENV"))
	if env != LocalEnv && env != NilEnv {
		log.WithField("env", env).Debug("skipping .env file")
		return
	}
	re := regexp.MustCompile(`^(.*` + PROJECT_DIR + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))
	filePath := fmt.Sprintf("%s/.env", string(rootPath))
	err := godotenv.Load(filePath)
	if err != nil {
		log.WithFields(log.Fields{
			"cause":    err,
			"cwd":      cwd,
			"filePath": filePath,
		}).Fatal("problem loading .env file")
		os.Exit(-1)
	}
}
