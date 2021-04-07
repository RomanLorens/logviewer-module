package scheduler

import (
	"context"
	"time"

	l "github.com/RomanLorens/logger/log"
	"github.com/RomanLorens/logviewer-module/utils"
)

//Scheduler scheduler
type Scheduler struct {
	logger l.Logger
}

//Task task
type Task struct {
	Name string
	Run  func(ctx context.Context)
}

//NewScheduler new scheduler
func NewScheduler(logger l.Logger) *Scheduler {
	return &Scheduler{logger: logger}
}

//Schedule schedule task
func (s Scheduler) Schedule(ctx context.Context, t *Task, interval time.Duration) {
	s.logger.Info(ctx, "Scheduling %v task with interval %v", t.Name, interval)
	go s.run(ctx, t, interval)
}

func (s Scheduler) run(ctx context.Context, t *Task, interval time.Duration) {
	defer utils.CatchError(ctx, s.logger)
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			go func(ctx context.Context, t *Task) {
				defer utils.CatchError(ctx, s.logger)
				s.logger.Info(ctx, "Scheduled task '%v' invoked", t.Name)
				t.Run(ctx)
			}(ctx, t)
		}
	}
}
