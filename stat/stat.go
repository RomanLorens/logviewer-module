package stat

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	e "github.com/RomanLorens/logviewer-module/error"
	l "github.com/RomanLorens/logviewer-module/logger"
	"github.com/RomanLorens/logviewer-module/model"
	"github.com/RomanLorens/logviewer-module/search"
)

//ReqID req id
type ReqID struct {
	ReqID string `json:"reqid"`
	Date  string `json:"date"`
}

//Stat stats
type Stat struct {
	LastTime string         `json:"lastTime"`
	Counter  int            `json:"counter"`
	Levels   map[string]int `json:"levels"`
	Errors   []*ReqID       `json:"errors"`
	Warnings []*ReqID       `json:"warnings"`
}

//ErrorDetails error details
type ErrorDetails struct {
	ReqID
	User    string `json:"user"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

//Pagination pagination
type Pagination struct {
	Total int `json:"total"`
	From  int `json:"from"`
	Size  int `json:"size"`
}

//ErrorDetailsPagination details with pagination
type ErrorDetailsPagination struct {
	ErrorDetails []*ErrorDetails `json:"errors"`
	Pagination   *Pagination     `json:"pagination"`
}

var logger = l.L

//GetErrors get errors
func GetErrors(r *http.Request, app *model.Application) (*ErrorDetailsPagination, *e.Error) {
	if search.IsLocal(r, app.Host) {
		logger.Info(r.Context(), "Getting error locally")
		return getErrorsLocal(app.Log, app)
	}
	return getErrorsRemotely(r, app.Log, app)
}

func getErrorsRemotely(r *http.Request, log string, app *model.Application) (*ErrorDetailsPagination, *e.Error) {
	logger.Info(r.Context(), "Stats log remotely")
	var res *ErrorDetailsPagination
	url := search.ApiURL(app.Host, model.ErrorsEndpoint)
	body, err := search.CallAPI(r.Context(), url, app, r.Header)
	if err != nil {
		return nil, err
	}
	if er := json.Unmarshal(body, &res); er != nil {
		return nil, e.Errorf(500, "Could not read unmarshal, %v", er)
	}
	return res, nil
}

func getErrorsLocal(log string, app *model.Application) (*ErrorDetailsPagination, *e.Error) {
	file, err := os.Open(log)
	if err != nil {
		return nil, e.Errorf(500, "Could not open log file, %v", err)
	}
	defer file.Close()
	res := make([]*ErrorDetails, 0, 100)
	requests := make(map[string]int, 0)
	scanner := bufio.NewScanner(file)
	ls := &app.LogStructure
	maxTokens := max(ls)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), "|")
		if len(tokens) < maxTokens {
			continue
		}
		level := search.NormalizeText(tokens[ls.Level])
		if !(level == "ERROR" || level == "WARNING" || level == "WARN") {
			continue
		}
		requests[tokens[ls.Reqid]+level]++
		if requests[tokens[ls.Reqid]+level] > 1 {
			continue
		}

		res = append(res, &ErrorDetails{
			ReqID:   ReqID{tokens[ls.Reqid], tokens[ls.Date]},
			Level:   level,
			Message: tokens[ls.Message],
			User:    tokens[ls.User],
		})

	}
	if err := scanner.Err(); err != nil {
		return nil, e.Errorf(500, "Error from scanner, %v", err)
	}

	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}

	pagination := &Pagination{
		From:  app.From,
		Size:  app.Size,
		Total: len(res),
	}
	start := app.From * app.Size
	end := (app.From * app.Size) + app.Size
	if end >= len(res) {
		end = len(res)
	}
	if start >= end {
		return &ErrorDetailsPagination{[]*ErrorDetails{}, pagination}, nil
	}
	return &ErrorDetailsPagination{res[start:end], pagination}, nil
}

//Get gets stats
func Get(r *http.Request, app *model.Application) (map[string]*Stat, *e.Error) {
	if search.IsLocal(r, app.Host) {
		logger.Info(r.Context(), "Checking locally stats")
		if app.LogStructure.Date == 0 && app.LogStructure.Level == 0 {
			return nil, e.ClientError("Must pass log structure - was empty %v", app.LogStructure)
		}
		return stats(app.Log, &app.LogStructure)
	}
	return remoteStats(r, app)

}

func remoteStats(r *http.Request, app *model.Application) (map[string]*Stat, *e.Error) {
	logger.Info(r.Context(), "Stats log remotely")
	var res map[string]*Stat
	url := search.ApiURL(app.Host, model.StatsEndpoint)
	body, err := search.CallAPI(r.Context(), url, app, r.Header)
	if err != nil {
		return nil, err
	}
	if er := json.Unmarshal(body, &res); er != nil {
		return nil, e.Errorf(500, "Could not read unmarshal, %v", er)
	}
	return res, nil
}

func getFilesByPattern(log string) ([]string, *e.Error) {
	dir := filepath.Dir(log)
	info, err := os.Stat(dir)
	if err != nil {
		return nil, e.ClientError("Could not open dir %v, %v", dir, err)
	}
	if !info.IsDir() {
		return nil, e.ClientError("%v is not dir, %v", dir, info)
	}
	pattern := strings.Replace(filepath.Base(log), ".log", "", 1)
	paths := make([]string, 0, 1)
	filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() || !strings.Contains(path, pattern) {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	return paths, nil
}

//CollectStats collects stats
func CollectStats(ctx context.Context, log string, ls *model.LogStructure, date string) (*model.CollectStatsMongo, *e.Error) {
	paths, err := getFilesByPattern(log)
	if err != nil {
		return nil, err
	}
	logger.Info(ctx, "collect stats for paths %v and mod date %v", paths, date)
	//user -> level -> counter
	m := make(map[string]map[string]int)
	requests := make(map[string]int, 0)
	for _, p := range paths {
		file, err := os.Open(p)
		if err != nil {
			return nil, e.AppError("Could not open log file, %v", err)
		}
		defer file.Close()

		maxTokens := max(ls)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			tokens := strings.Split(scanner.Text(), "|")
			if len(tokens) < maxTokens {
				continue
			}
			if !strings.Contains(tokens[ls.Date], date) {
				continue
			}
			user := tokens[ls.User]
			if len(strings.TrimSpace(user)) == 0 {
				continue
			}
			level := strings.ToUpper(search.NormalizeText(tokens[ls.Level]))
			key := tokens[ls.Reqid] + level + user
			requests[key]++
			if requests[key] > 1 {
				continue
			}
			u, ok := m[user]
			if !ok {
				u = make(map[string]int, 0)
				m[user] = u
			}
			u[level]++
		}
		if err := scanner.Err(); err != nil {
			return nil, e.AppError("Error from scanner, %v", err)
		}
	}

	return &model.CollectStatsMongo{Users: m, TotalRequests: int32(len(requests))}, nil
}

func stats(log string, ls *model.LogStructure) (map[string]*Stat, *e.Error) {
	out := make(map[string]*Stat, 0)
	requests := make(map[string]int, 0)
	file, err := os.Open(log)
	if err != nil {
		return nil, e.Errorf(500, "Could not open log file, %v", err)
	}
	defer file.Close()

	maxTokens := max(ls)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), "|")
		if len(tokens) < maxTokens {
			continue
		}
		user := tokens[ls.User]
		if len(strings.TrimSpace(user)) == 0 {
			continue
		}
		u, ok := out[user]
		if !ok {
			u = &Stat{
				Levels: make(map[string]int, 0),
			}
			out[user] = u
		}
		level := strings.ToUpper(search.NormalizeText(tokens[ls.Level]))
		key := tokens[ls.Reqid] + level + user
		requests[key]++
		if requests[key] > 1 {
			continue
		}
		u.LastTime = tokens[ls.Date]
		u.Counter++
		u.Levels[level]++
		if level == "ERROR" {
			u.Errors = append(u.Errors, &ReqID{
				tokens[ls.Reqid], tokens[ls.Date],
			})
		}
		if level == "WARNING" || level == "WARN" {
			u.Warnings = append(u.Warnings, &ReqID{
				tokens[ls.Reqid], tokens[ls.Date],
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, e.Errorf(500, "Error from scanner, %v", err)
	}
	for _, v := range out {
		for i, j := 0, len(v.Errors)-1; i < j; i, j = i+1, j-1 {
			v.Errors[i], v.Errors[j] = v.Errors[j], v.Errors[i]
		}
		for i, j := 0, len(v.Warnings)-1; i < j; i, j = i+1, j-1 {
			v.Warnings[i], v.Warnings[j] = v.Warnings[j], v.Warnings[i]
		}
	}
	return out, nil
}

func max(ls *model.LogStructure) int {
	m := ls.Date
	if ls.User > m {
		m = ls.User
	}
	if ls.Reqid > m {
		m = ls.Reqid
	}
	if ls.Level > m {
		m = ls.Level
	}
	return m
}
