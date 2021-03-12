package search

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	e "github.com/RomanLorens/logviewer-module/error"
	"github.com/RomanLorens/logviewer-module/model"
	"github.com/RomanLorens/logviewer-module/utils"
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
func DownloadLog(r *http.Request, lg *model.LogDownload) ([]byte, *e.Error) {
	if lg.Host == "" || lg.Log == "" {
		return nil, e.Errorf(400, "Payload is missing host and/or log values")
	}
	if IsLocal(r, lg.Host) {
		return ls.DownloadLog(r.Context(), lg)
	}
	return rs.DownloadLog(r, lg)
}

//TailLog tail log
func TailLog(r *http.Request, app *model.Application) (*model.Result, *e.Error) {
	if IsLocal(r, app.Host) {
		return ls.Tail(r.Context(), app)
	}
	return rs.Tail(r, app)
}

//Find find logs
func Find(r *http.Request, s *model.Search) ([]*model.Result, *e.Error) {
	if err := validate(s); err != nil {
		return nil, err
	}
	out := make(chan []*model.Result, len(s.Hosts))
	for _, host := range s.Hosts {
		go func(host string) {
			logger.Info(r.Context(), "starting grep goroutine for %v", host)
			start := time.Now()
			local := IsLocal(r, host)
			var res []*model.Result
			if local {
				res = ls.Grep(r.Context(), host, s)
			} else {
				r, err := rs.Grep(r, host, s)
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
			logger.Info(r.Context(), "goroutine for %v finished", host)
			out <- res
		}(host)
	}
	return <-out, nil
}

//ListLogs list logs for app
func ListLogs(r *http.Request, s *model.Search) ([]*model.LogDetails, *e.Error) {
	logs := make([]*model.LogDetails, 0)
	hc := make(chan string, len(s.Hosts))
	for _, h := range s.Hosts {
		go func(host string) {
			logger.Info(r.Context(), "host routine for %v started...", host)
			if IsLocal(r, host) {
				l, _ := ls.List(r.Context(), host, s) //no error for local
				logs = append(logs, l...)
			} else {
				l, err := rs.List(r, host, s)
				if err == nil {
					logs = append(logs, l...)
				} else {
					logger.Error(r.Context(), "Error from api, %v", err)
				}
			}
			hc <- host
			logger.Info(r.Context(), "host routine for %v finsihed", host)
		}(h)
	}
	<-hc //wait for all threads
	sort.Slice(logs, func(i, j int) bool { return logs[i].ModTime > logs[j].ModTime })
	return logs, nil
}

//IsLocal is local request
func IsLocal(r *http.Request, url string) bool {
	hostname, err := utils.Hostname()
	if err != nil {
		logger.Error(r.Context(), "Could not resolve hostname - defaults to local host, %v", err)
		return true
	}
	return strings.Contains(strings.ToLower(url), strings.ToLower(hostname))
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
