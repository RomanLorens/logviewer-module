package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"

	l "github.com/RomanLorens/logger/log"
	"github.com/RomanLorens/logviewer-module/model"
	"github.com/RomanLorens/logviewer-module/search"
	"github.com/RomanLorens/logviewer-module/stat"
)

//Handler handler
type Handler struct {
	logger l.Logger
}

//NewHandler new handler
func NewHandler(logger l.Logger) *Handler {
	return &Handler{logger: logger}
}

//DownloadLog download log
func (h Handler) DownloadLog(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var lr model.LogRequest
	err := json.NewDecoder(r.Body).Decode(&lr)
	if err != nil {
		return nil, fmt.Errorf("missing log download body, %w", err)
	}
	defer r.Body.Close()
	b, er := search.DownloadLog(lr.Log)
	if er != nil {
		return nil, err
	}

	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", path.Base(lr.Log)))
	w.Write(b)
	return nil, nil
}

//Stats stats
func (h Handler) Stats(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var sr model.StatsRequest
	err := json.NewDecoder(r.Body).Decode(&sr)
	if err != nil {
		return nil, fmt.Errorf("Could not parse req body as stats request, %v", err)
	}
	return stat.Stats(sr.Log, sr.LogStructure)
}

//CollectStats collect stats
func (h Handler) CollectStats(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var s model.CollectStatsRequest
	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		return nil, fmt.Errorf("Could not parse req body, %v", err)
	}
	return stat.CollectStats(r.Context(), &s, h.logger)
}

//Errors errors
func (h Handler) Errors(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var req model.ErrorsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, fmt.Errorf("Could not parse req body as errors req, %v", err)
	}
	return stat.Errors(&req)
}

//TailLog tail log
func (h Handler) TailLog(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var req model.LogRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, fmt.Errorf("Could not parse req body as log req, %v", err)
	}
	return search.Tail(req.Log)
}

//ListLogs list logs
func (h Handler) ListLogs(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	h.logger.Info(r.Context(), "list logs module handler...")
	var lr model.ListLogsRequest
	err := json.NewDecoder(r.Body).Decode(&lr)
	if err != nil {
		return nil, fmt.Errorf("could not serialize list-logs request, %w", err)
	}
	return search.ListLogs(r.Context(), lr.Logs, h.logger), nil
}

//Search search
func (h Handler) Search(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var gr model.GrepRequest
	err := json.NewDecoder(r.Body).Decode(&gr)
	if err != nil {
		return nil, err
	}
	return search.Grep(r.Context(), &gr, h.logger), nil
}
