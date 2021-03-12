package utils

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	e "github.com/RomanLorens/logviewer-module/error"
	l "github.com/RomanLorens/logviewer-module/logger"
)

var logger = l.L

//CatchError recovers from panic
func CatchError(ctx context.Context) {
	if r := recover(); r != nil {
		m := fmt.Sprintf("Panic recovered: '%v'\n%v", r, string(debug.Stack()))
		if ctx != nil {
			logger.Error(ctx, m)
		} else {
			fmt.Println(m)
		}
	}
}

//Hostname gets hostname from unix box or hostname env
func Hostname() (string, *e.Error) {
	h, _ := os.Hostname()
	if h == "" {
		h = os.Getenv("hostname")
		if h == "" {
			return "", e.AppError("Could not resolve hostname")
		}
	}
	return h, nil
}
