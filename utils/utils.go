package utils

import (
	"context"
	"fmt"
	"runtime/debug"

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
