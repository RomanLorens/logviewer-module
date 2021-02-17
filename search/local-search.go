package search

import (
	"bufio"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	e "github.com/RomanLorens/logviewer-module/error"
	l "github.com/RomanLorens/logviewer-module/logger"
	"github.com/RomanLorens/logviewer-module/model"
)

//LocalSearch local search
type LocalSearch struct{}

var logger = l.L

//Grep grep logs
func (LocalSearch) Grep(ctx context.Context, host string, s *model.Search) []*model.Result {
	logger.Info(ctx, "Local grep for %v", host)
	out := make([]*model.Result, 0, len(s.Logs))
	for _, l := range s.Logs {
		r := model.Result{LogFile: l, Host: host}
		lines, err := grepFile(l, s)
		r.Error = err
		r.Lines = lines
		out = append(out, &r)
	}
	return out
}

//DownloadLog read file
func (LocalSearch) DownloadLog(ctx context.Context, req *model.LogDownload) ([]byte, *e.Error) {
	b, err := ioutil.ReadFile(req.Log)
	if err != nil {
		return nil, e.Errorf(500, "Could not read file %v, %v", req.Log, err)
	}
	return b, nil
}

func grepFile(path string, s *model.Search) ([]string, *e.Error) {
	out := make([]string, 0, 20)
	f, err := os.Open(path)
	if err != nil {
		return nil, e.Errorf(http.StatusInternalServerError, "Could not open file %v, %v", path, err)
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	val := strings.ToLower(s.Value)
	for scanner.Scan() {
		if strings.Contains(strings.ToLower(scanner.Text()), val) {
			out = append(out, NormalizeText(scanner.Text()))
		}
	}
	var er *e.Error
	if err = scanner.Err(); err != nil {
		er = e.Errorf(http.StatusInternalServerError, "error when grepping file %v", err)
	}
	return out, er
}

//Tail tail log
func (LocalSearch) Tail(ctx context.Context, app *model.Application) (*model.Result, *e.Error) {
	logger.Info(ctx, "Tail logs locally")
	start := time.Now()
	file, err := os.Open(app.Log)
	if err != nil {
		return nil, e.Errorf(http.StatusInternalServerError, "Could not open file %v", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, e.Errorf(http.StatusInternalServerError, "Could not stat file %v", err)
	}

	offset := info.Size() - int64(tailSizeKB*1024)
	if offset < 0 {
		offset = 0
	}
	bytes := make([]byte, info.Size()-offset)

	_, err = file.ReadAt(bytes, offset)
	if err != nil && err != io.EOF {
		return nil, e.Errorf(http.StatusInternalServerError, "Could not stat file %v", err)
	}

	//start from new line
	for i, b := range bytes {
		if b == '\n' {
			bytes = bytes[i:]
			break
		}
	}

	lines := make([]string, 0, 100)
	for _, l := range strings.Split(string(bytes), "\n") {
		l = NormalizeText(l)
		if strings.TrimSpace(l) != "" {
			lines = append(lines, NormalizeText(l))
		}
	}
	return &model.Result{
		Lines:   lines,
		Time:    time.Now().Sub(start).Milliseconds(),
		Host:    app.Host,
		LogFile: app.Log,
	}, nil
}

//List list logs
func (LocalSearch) List(ctx context.Context, host string, s *model.Search) ([]*model.LogDetails, *e.Error) {
	logger.Info(ctx, "Local list")
	dirs := getDirs(s.Logs)
	logs := make([]*model.LogDetails, 0)
	c := make(chan []*model.LogDetails, len(dirs))
	for _, dir := range dirs {
		go func(dir string) {
			l, err := getStats(dir, host)
			if err != nil {
				logger.Error(ctx, err.Message)
				close(c)
				return
			}
			c <- l
		}(dir)
	}
	logs = append(logs, <-c...)
	return logs, nil
}

func getDirs(paths []string) []string {
	out := make([]string, 0, len(paths))
	m := make(map[string]bool, len(paths))
	for _, p := range paths {
		dir := filepath.Dir(p)
		if _, ok := m[dir]; !ok {
			m[dir] = true
			out = append(out, dir)
		}
	}
	return out
}

func getStats(dir string, host string) ([]*model.LogDetails, *e.Error) {
	logs := make([]*model.LogDetails, 0)
	info, err := os.Stat(dir)
	if err != nil {
		return nil, e.Errorf(http.StatusBadRequest, err.Error())
	}
	if !info.IsDir() {
		return nil, e.Errorf(http.StatusBadRequest, "path %v is not dir", dir)
	}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		info, err = os.Stat(path)
		if err == nil && !info.IsDir() {
			logs = append(logs, &model.LogDetails{
				ModTime: info.ModTime().Unix(),
				Name:    path,
				Size:    info.Size(),
				Host:    host,
			})
		}
		return nil
	})
	return logs, nil
}

//NormalizeText normalize level
func NormalizeText(t string) string {
	t = strings.ReplaceAll(t, "\033[0;31mERROR\033[0m", "ERROR")
	t = strings.ReplaceAll(t, "\033[0;33mWARNING\033[0m", "WARNING")
	t = strings.ReplaceAll(t, "\033[0;32mINFO\033[0m", "INFO")
	return t
}
