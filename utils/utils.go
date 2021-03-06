package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/debug"

	l "github.com/RomanLorens/logger/log"
)

//CatchError recovers from panic
func CatchError(ctx context.Context, logger l.Logger) {
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
func Hostname() (string, error) {
	h, _ := os.Hostname()
	if h == "" {
		h = os.Getenv("hostname")
		if h == "" {
			return "", errors.New("Could not resolve hostname")
		}
	}
	return h, nil
}
