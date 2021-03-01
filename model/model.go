package model

import (
	e "github.com/RomanLorens/logviewer-module/error"
)

//Application application
type Application struct {
	ApplicationID string       `json:"application"`
	Env           string       `json:"env"`
	Log           string       `json:"log"`
	Host          string       `json:"host"`
	From          int          `json:"from"`
	Size          int          `json:"size"`
	LogStructure  LogStructure `json:"logStructure"`
}

//Search search
type Search struct {
	Value         string   `json:"value"`
	FromTime      int64    `json:"FromTime"`
	ToTime        int64    `json:"ToTime"`
	ApplicationID string   `json:"application"`
	Env           string   `json:"env"`
	Logs          []string `json:"logs"`
	Hosts         []string `json:"hosts"`
}

//Result search result
type Result struct {
	LogFile string   `json:"logfile"`
	Lines   []string `json:"lines"`
	Host    string   `json:"host"`
	Error   *e.Error `json:"error,omitempty"`
	Time    int64    `json:"time"`
}

//LogDetails log details
type LogDetails struct {
	ModTime int64  `json:"modtime"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Host    string `json:"host"`
}

//LogDownload log download
type LogDownload struct {
	Host string `json:"host"`
	Log  string `json:"log"`
}

//LogStructure log structure
type LogStructure struct {
	Date       int    `json:"date"`
	User       int    `json:"user"`
	Reqid      int    `json:"reqid"`
	Level      int    `json:"level"`
	Message    int    `json:"message"`
	DateFormat string `json:"dateFormat"`
}

//CollectStats collect stats
type CollectStats struct {
	LogPath      string        `json:"logPath"`
	LogStructure *LogStructure `json:"logStructure"`
	Date         string        `date:"date"`
}

//CollectStatsMongo collect stats mongo
type CollectStatsMongo struct {
	Users         map[string]map[string]int `json:"users"`
	TotalRequests int32                     `json:"totalRequests"`
}

const (
	//SearchEndpoint search
	SearchEndpoint = "search"
	//ListLogsEndpoint list logs
	ListLogsEndpoint = "list-logs"
	//DownloadLogEndpoint download log
	DownloadLogEndpoint = "download-log"
	//TailLogEndpoint tail log
	TailLogEndpoint = "tail-log"
	//StatsEndpoint stats
	StatsEndpoint = "stats"
	//ErrorsEndpoint errors
	ErrorsEndpoint = "errors"
	//CollectStatsEndpoint collect stats
	CollectStatsEndpoint = "collect-stats"
)
