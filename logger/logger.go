package logger

import "github.com/RomanLorens/logger/log"

//L app logger
var L *log.Logger

func init() {
	L = log.New(log.WithConfig("logs/logviewer.log").Build())
}
