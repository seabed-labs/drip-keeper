package util

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func GoID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		logrus.WithError(err).Errorf("failed to get goroutine ID")
		return -1
	}
	return id
}
