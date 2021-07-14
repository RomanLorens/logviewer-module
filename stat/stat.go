package stat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	l "github.com/RomanLorens/logger/log"
	"github.com/RomanLorens/logviewer-module/model"
	"github.com/RomanLorens/logviewer-module/search"
)

//Errors errors
func Errors(req *model.ErrorsRequest) (*model.ErrorDetailsPagination, error) {
	file, err := os.Open(req.Log)
	if err != nil {
		return nil, fmt.Errorf("Could not open log file, %v", err)
	}
	defer file.Close()
	res := make([]model.ErrorDetails, 0, 100)
	requests := make(map[string]int, 0)
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	ls := req.LogStructure
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

		res = append(res, model.ErrorDetails{
			ReqID:   model.ReqID{ReqID: tokens[ls.Reqid], Date: tokens[ls.Date]},
			Level:   level,
			Message: tokens[ls.Message],
			User:    tokens[ls.User],
		})

	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error from scanner, %v", err)
	}

	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		res[i], res[j] = res[j], res[i]
	}

	pagination := &model.Pagination{
		From:  req.From,
		Size:  req.Size,
		Total: len(res),
	}
	start := req.From * req.Size
	end := (req.From * req.Size) + req.Size
	if end >= len(res) {
		end = len(res)
	}
	if start >= end {
		return &model.ErrorDetailsPagination{ErrorDetails: []model.ErrorDetails{}, Pagination: pagination}, nil
	}
	return &model.ErrorDetailsPagination{ErrorDetails: res[start:end], Pagination: pagination}, nil
}

func getFilesByPattern(log string) ([]string, error) {
	dir := filepath.Dir(log)
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("Could not open dir %v, %v", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%v is not dir, %v", dir, info)
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
func CollectStats(ctx context.Context, req *model.CollectStatsRequest, logger l.Logger) (*model.CollectStatsRsults, error) {
	paths, err := getFilesByPattern(req.Log)
	if err != nil {
		return nil, err
	}
	logger.Info(ctx, "collect stats for paths %v and mod date %v", paths, req.Date)
	//user -> level -> counter
	m := make(map[string]map[string]int)
	requests := make(map[string]int, 0)
	for _, p := range paths {
		file, err := os.Open(p)
		if err != nil {
			return nil, fmt.Errorf("Could not open log file, %v", err)
		}
		defer file.Close()
		ls := req.LogStructure
		maxTokens := max(ls)
		scanner := bufio.NewScanner(file)
		buf := make([]byte, 0, 64*1024)
		scanner.Buffer(buf, 1024*1024)
		for scanner.Scan() {
			tokens := strings.Split(scanner.Text(), "|")
			if len(tokens) < maxTokens {
				continue
			}
			if !strings.Contains(tokens[ls.Date], req.Date) {
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
			return nil, fmt.Errorf("Error from scanner, %v", err)
		}
	}

	return &model.CollectStatsRsults{Users: m, TotalRequests: int32(len(requests))}, nil
}

//Stats stats
func Stats(log string, ls *model.LogStructure) (map[string]*model.Stat, error) {
	out := make(map[string]*model.Stat)
	requests := make(map[string]int, 0)
	file, err := os.Open(log)
	if err != nil {
		return nil, fmt.Errorf("Could not open log file, %v", err)
	}
	defer file.Close()

	maxTokens := max(ls)
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
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
			u = &model.Stat{
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
			u.Errors = append(u.Errors, model.ReqID{ReqID: tokens[ls.Reqid], Date: tokens[ls.Date]})
		}
		if level == "WARNING" || level == "WARN" {
			u.Warnings = append(u.Warnings, model.ReqID{
				ReqID: tokens[ls.Reqid], Date: tokens[ls.Date],
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error from scanner, %v", err)
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
