package model

//GrepRequest req
type GrepRequest struct {
	Value string   `json:"value"`
	Logs  []string `json:"logs"`
}

//GrepResponse search result
type GrepResponse struct {
	LogFile string   `json:"logfile"`
	Lines   []string `json:"lines"`
	Host    string   `json:"host"`
	Time    int64    `json:"time"`
}

//ListLogsRequest list logs
type ListLogsRequest struct {
	Logs []string `json:"logs"`
}

//LogRequest log req
type LogRequest struct {
	Log string `json:"log"`
}

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
	Errors   []ReqID        `json:"errors"`
	Warnings []ReqID        `json:"warnings"`
}

//StatsRequest stats req
type StatsRequest struct {
	Log          string        `json:"log"`
	LogStructure *LogStructure `json:"logStructure"`
}

//ErrorsRequest errors req
type ErrorsRequest struct {
	From int `json:"from"`
	Size int `json:"size"`
	*StatsRequest
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

//TailLogResponse res
type TailLogResponse struct {
	LogFile string   `json:"logfile"`
	Lines   []string `json:"lines"`
	Host    string   `json:"host"`
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
	ErrorDetails []ErrorDetails `json:"errors"`
	Pagination   *Pagination    `json:"pagination"`
}

//LogStructure log structure
type LogStructure struct {
	Date           int    `json:"date"`
	User           int    `json:"user"`
	Reqid          int    `json:"reqid"`
	Level          int    `json:"level"`
	Message        int    `json:"message"`
	DateFormat     string `json:"dateFormat"`
	JavaDateFormat string `json:"javaDateFormat"`
}

//CollectStatsRequest collect stats
type CollectStatsRequest struct {
	*StatsRequest
	Date string `json:"date"`
}

//CollectStatsRsults collect stats results
type CollectStatsRsults struct {
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
