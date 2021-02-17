package search

import (
	"context"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	e "github.com/RomanLorens/logviewer-module/error"
	"github.com/RomanLorens/logviewer-module/model"
)

//LogSearch available log actions
type LogSearch interface {
	Tail(ctx context.Context, app *model.Application) (*model.Result, *e.Error)
	Grep(ctx context.Context, host string, s *model.Search) []*model.Result
	List(ctx context.Context, url string, s *model.Search) ([]*model.LogDetails, *e.Error)
	DownloadLog(ctx context.Context, req *model.LogDownload) ([]byte, *e.Error)
}

var (
	tailSizeKB = 16
	ls         = LocalSearch{}
	rs         = RemoteSearch{}
)

//DownloadLog download log
func DownloadLog(ctx context.Context, lg *model.LogDownload) ([]byte, *e.Error) {
	if lg.Host == "" || lg.Log == "" {
		return nil, e.Errorf(400, "Payload is missing host and/or log values")
	}
	if IsLocal(ctx, lg.Host) {
		return ls.DownloadLog(ctx, lg)
	}
	return rs.DownloadLog(ctx, lg)
}

//TailLog tail log
func TailLog(ctx context.Context, app *model.Application) (*model.Result, *e.Error) {
	if IsLocal(ctx, app.Host) {
		return ls.Tail(ctx, app)
	}
	return rs.Tail(ctx, app)
}

//Find find logs
func Find(ctx context.Context, s *model.Search) ([]*model.Result, *e.Error) {
	if err := validate(s); err != nil {
		return nil, err
	}
	out := make(chan []*model.Result, len(s.Hosts))
	for _, host := range s.Hosts {
		go func(host string) {
			logger.Info(ctx, "starting goroutine for %v", host)
			start := time.Now()
			local := IsLocal(ctx, host)
			var res []*model.Result
			if local {
				res = ls.Grep(ctx, host, s)
			} else {
				r, err := rs.Grep(ctx, host, s)
				if err != nil {
					res = append(res, &model.Result{Error: err, Time: 0})
				} else {
					res = append(res, r...)
				}
			}
			end := time.Now()
			elapsed := end.Sub(start)
			for _, r := range res {
				r.Time = elapsed.Milliseconds()
			}
			logger.Info(ctx, "goroutine for %v finished", host)
			out <- res
		}(host)
	}
	return <-out, nil
}

//ListLogs list logs for app
func ListLogs(ctx context.Context, s *model.Search) ([]*model.LogDetails, *e.Error) {
	logs := make([]*model.LogDetails, 0)
	hc := make(chan string, len(s.Hosts))
	for _, h := range s.Hosts {
		go func(host string) {
			logger.Info(ctx, "host routine for %v started...", host)
			if IsLocal(ctx, host) {
				l, _ := ls.List(ctx, host, s) //no error for local
				logs = append(logs, l...)
			} else {
				l, err := rs.List(ctx, host, s)
				if err == nil {
					logs = append(logs, l...)
				} else {
					logger.Error(ctx, "Error from api, %v", err)
				}
			}
			hc <- host
			logger.Info(ctx, "host routine for %v finsihed", host)
		}(h)
	}
	<-hc //wait for all threads
	sort.Slice(logs, func(i, j int) bool { return logs[i].ModTime > logs[j].ModTime })
	return logs, nil
}

func IsLocal(ctx context.Context, host string) bool {
	hostname, err := os.Hostname()
	if err != nil {
		logger.Error(ctx, "Could not check hostname, %v", err)
		return false
	}
	if strings.Contains(strings.ToLower(host), strings.ToLower(hostname)) ||
		strings.Contains(host, "://localhost") {
		return true
	}
	return false
}

func validate(s *model.Search) *e.Error {
	if strings.TrimSpace(s.Value) == "" {
		return e.Errorf(http.StatusBadRequest, "Missing value")
	}
	if len(s.Hosts) == 0 {
		return e.Errorf(http.StatusBadRequest, "Missing hosts")
	}
	if len(s.Logs) == 0 {
		return e.Errorf(http.StatusBadRequest, "Missing logs")
	}
	return nil
}
