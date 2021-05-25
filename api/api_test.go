package api

import (
	"context"
	"strings"
	"testing"

	l "github.com/RomanLorens/logger/log"
	"github.com/RomanLorens/logviewer-module/model"
)

var (
	la  = NewLocalAPI(l.PrintLogger(false))
	log = "../test-logs/java-app.log"
	ls  = model.LogStructure{Date: 0, User: 4, Reqid: 5, Level: 2, Message: 6, DateFormat: "2006-01-02"}
)

func TestGrep(t *testing.T) {
	res := la.Grep(context.Background(), &model.GrepRequest{Value: "1-01-CV-QCVMW9XPMLMMUMJKKJNTCR1ETMKB9EG133396101@1-248129#6", Logs: []string{log}})

	if len(res) != 1 {
		t.Fatal("empty response")
	}
	if !strings.Contains(res[0].LogFile, log) {
		t.Fatal("should have log name")
	}
}

func TestListLogs(t *testing.T) {
	res := la.ListLogs(context.Background(), &model.ListLogsRequest{Logs: []string{log}})

	if len(res) != 1 {
		t.Fatal("empty response")
	}
}

func TestDownloadLog(t *testing.T) {
	res, err := la.DownloadLog(log)

	if err != nil {
		t.Fatal(err)
	}
	if len(res) == 0 {
		t.Error("empty content")
	}
}

func TestTailLog(t *testing.T) {
	res, err := la.TailLog(context.Background(), log)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Lines) == 0 {
		t.Error("empty res")
	}
}

func TestStats(t *testing.T) {
	stats, err := la.Stats(context.Background(), &model.StatsRequest{Log: log, LogStructure: &ls})

	if err != nil {
		t.Fatal(err)
	}
	if len(stats) == 0 {
		t.Error("empty res")
	}
}

func TestErrors(t *testing.T) {
	r := model.ErrorsRequest{StatsRequest: &model.StatsRequest{Log: log, LogStructure: &ls}, From: 0, Size: 100}
	res, err := la.Errors(context.Background(), &r)

	if err != nil {
		t.Fatal(err)
	}
	if len(res.ErrorDetails) == 0 {
		t.Error("empty res")
	}
}

func TestCollectStats(t *testing.T) {
	req := &model.StatsRequest{Log: log, LogStructure: &ls}
	stats, err := la.CollectStats(context.Background(), &model.CollectStatsRequest{StatsRequest: req, Date: "2021-05-06"})

	if err != nil {
		t.Fatal(err)
	}
	if len(stats.Users) == 0 {
		t.Error("empty res")
	}
}
