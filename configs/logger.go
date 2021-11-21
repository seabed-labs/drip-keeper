package configs

import "github.com/sirupsen/logrus"

func InitLogrus() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}
