package api

import (
	"context"

	l "github.com/RomanLorens/logger/log"
	"github.com/RomanLorens/logviewer-module/model"
	"github.com/RomanLorens/logviewer-module/search"
	"github.com/RomanLorens/logviewer-module/stat"
)

//LocalAPI local api
type LocalAPI struct {
	logger l.Logger
}

//NewLocalAPI api
func NewLocalAPI(logger l.Logger) *LocalAPI {
	return &LocalAPI{logger: logger}
}

//Grep greps log
func (la LocalAPI) Grep(ctx context.Context, req *model.GrepRequest) []model.GrepResponse {
	return search.Grep(ctx, req, la.logger)
}

//ListLogs list logs
func (la LocalAPI) ListLogs(ctx context.Context, req *model.ListLogsRequest) []model.LogDetails {
	la.logger.Info(ctx, "list logs locally...")
	return search.ListLogs(ctx, req.Logs, la.logger)
}

//DownloadLog download log
func (la LocalAPI) DownloadLog(log string) ([]byte, error) {
	return search.DownloadLog(log)
}

//TailLog tail log
func (la LocalAPI) TailLog(ctx context.Context, log string) (*model.TailLogResponse, error) {
	la.logger.Info(ctx, "Tail logs locally")
	return search.Tail(log)
}

//Stats stats
func (la LocalAPI) Stats(ctx context.Context, req *model.StatsRequest) (map[string]*model.Stat, error) {
	return stat.Stats(req.Log, req.LogStructure)
}

//Errors errors
func (la LocalAPI) Errors(ctx context.Context, req *model.ErrorsRequest) (*model.ErrorDetailsPagination, error) {
	return stat.Errors(req)
}

//CollectStats collect stats
func (la LocalAPI) CollectStats(ctx context.Context, req *model.CollectStatsRequest) (*model.CollectStatsRsults, error) {
	return stat.CollectStats(ctx, req, la.logger)
}
