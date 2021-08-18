package search

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	l "github.com/RomanLorens/logger/log"
	"github.com/RomanLorens/logviewer-module/model"
	"github.com/RomanLorens/logviewer-module/utils"
)

var tailSizeKB = 16

//Grep grep logs
func Grep(ctx context.Context, req *model.GrepRequest, logger l.Logger) []model.GrepResponse {
	out := make([]model.GrepResponse, 0, len(req.Logs))
	for _, l := range req.Logs {
		logger.Info(ctx, "Local grep for %v - '%v'", l, req.Value)
		r := model.GrepResponse{LogFile: l}
		lines, err := grepFile(l, req.Value)
		if err != nil {
			logger.Error(ctx, "Could not grep %v, %v", l, err)
			continue
		}
		r.Lines = lines
		out = append(out, r)
	}
	return out
}

//DownloadLog read file
func DownloadLog(log string) ([]byte, error) {
	b, err := ioutil.ReadFile(log)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func grepFile(path string, value string) ([]string, error) {
	out := make([]string, 0, 20)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	val := strings.ToLower(value)
	for scanner.Scan() {
		if strings.Contains(strings.ToLower(scanner.Text()), val) {
			out = append(out, NormalizeText(scanner.Text()))
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

//Tail tail log
func Tail(log string) (*model.TailLogResponse, error) {
	res, _, err := TailLogIfNewer(log, 0)
	return res, err
}

//TailLogIfNewer tail log
func TailLogIfNewer(log string, modtime int64) (*model.TailLogResponse, bool, error) {
	start := time.Now()
	file, err := os.Open(log)
	if err != nil {
		return nil, true, fmt.Errorf("Could not open file %v", err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, true, fmt.Errorf("Could not stat file %v", err)
	}
	if modtime >= info.ModTime().Unix() {
		return nil, false, nil
	}
	offset := info.Size() - int64(tailSizeKB*1024)
	if offset < 0 {
		offset = 0
	}
	bytes := make([]byte, info.Size()-offset)

	_, err = file.ReadAt(bytes, offset)
	if err != nil && err != io.EOF {
		return nil, true, fmt.Errorf("Could not stat file %v", err)
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
	return &model.TailLogResponse{
		Lines:   lines,
		Time:    time.Now().Sub(start).Milliseconds(),
		LogFile: log,
		ModTime: info.ModTime().Unix(),
	}, true, nil
}

//ListLogs list logs
func ListLogs(ctx context.Context, logs []string, logger l.Logger) []model.LogDetails {
	dirs := getDirs(logs)
	res := make([]model.LogDetails, 0, len(logs))
	c := make(chan []model.LogDetails, len(dirs))
	for _, dir := range dirs {
		go func(dir string) {
			utils.CatchError(ctx, logger)
			l, err := getStats(dir)
			if err != nil {
				logger.Error(ctx, err.Error())
				close(c)
				return
			}
			c <- l
		}(dir)
	}
	for i := 0; i < len(dirs); i++ {
		res = append(res, <-c...)
	}
	sort.SliceStable(res, func(i int, j int) bool {
		return res[i].ModTime > res[j].ModTime
	})
	return res
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

func getStats(dir string) ([]model.LogDetails, error) {
	logs := make([]model.LogDetails, 0)
	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("not dir %v", dir)
	}
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if !file.IsDir() {
			logs = append(logs, model.LogDetails{
				ModTime: file.ModTime().Unix(),
				Name:    filepath.Join(dir, file.Name()),
				Size:    file.Size(),
			})
		}
	}

	return logs, nil
}

//NormalizeText normalize level
func NormalizeText(t string) string {
	t = strings.ReplaceAll(t, "\033[0;31mERROR\033[0m", "ERROR")
	t = strings.ReplaceAll(t, "\033[0;33mWARNING\033[0m", "WARNING")
	t = strings.ReplaceAll(t, "\033[0;32mINFO\033[0m", "INFO")
	return t
}
