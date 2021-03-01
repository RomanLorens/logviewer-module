package logger

import (
	"context"
	"os"

	"github.com/RomanLorens/logger/log"
)

//L app logger
var L log.Logger

func init() {
	path := os.Getenv("LOGVIEWER_LOG")
	if path == "" {
		path = "logs/logviewer.log"
	}
	L, err := log.New(log.WithConfig(path).Build())
	if err != nil {
		L.Error(context.Background(), "Could not create file logger, %v", err)
	}
}
