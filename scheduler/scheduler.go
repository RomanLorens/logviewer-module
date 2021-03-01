package scheduler

import (
	"context"
	"time"

	l "github.com/RomanLorens/logviewer-module/logger"
	"github.com/RomanLorens/logviewer-module/utils"
)

var logger = l.L

//Task task
type Task struct {
	Name     string
	Callback func(ctx context.Context)
}

//Schedule schedule task
func Schedule(ctx context.Context, t *Task, interval time.Duration) {
	logger.Info(ctx, "Scheduling %v task with interval %v", t.Name, interval)
	go run(ctx, t, interval)
}

func run(ctx context.Context, t *Task, interval time.Duration) {
	defer utils.CatchError(ctx)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			go func(ctx context.Context, t *Task) {
				defer utils.CatchError(ctx)
				logger.Info(ctx, "Scheduled task '%v' invoked", t.Name)
				t.Callback(ctx)
			}(ctx, t)
		}
	}
}
