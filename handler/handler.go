package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	l "github.com/RomanLorens/logger/log"
	e "github.com/RomanLorens/logviewer-module/error"
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
func (h Handler) DownloadLog(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	var ld model.LogDownload
	err := json.NewDecoder(r.Body).Decode(&ld)
	if err != nil {
		return nil, e.AppError("missing log download body, %w", err)
	}
	defer r.Body.Close()
	b, er := search.DownloadLog(r, &ld)
	if er != nil {
		return nil, e.AppError("download log,%w", err)
	}

	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", path.Base(ld.Log)))
	w.Write(b)
	return nil, nil
}

//Stats stats
func (h Handler) Stats(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	app, err := toApp(r)
	if err != nil {
		return nil, err
	}
	return stat.Get(r, app, h.logger)
}

//CollectStats collect stats
func (h Handler) CollectStats(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	var s model.CollectStats
	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		return nil, e.ClientError("Could not parse req body, %v", err)
	}
	return stat.CollectStats(r.Context(), s.LogPath, s.LogStructure, s.Date, h.logger)
}

//Errors errors
func (h Handler) Errors(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	app, err := toApp(r)
	if err != nil {
		return nil, err
	}
	return stat.GetErrors(r, app, h.logger)
}

//TailLog tail log
func (h Handler) TailLog(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	app, err := toApp(r)
	if err != nil {
		return nil, err
	}
	return search.TailLog(r, app)
}

//ListLogs list logs
func (h Handler) ListLogs(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	var s, err = toSearch(r)
	if err != nil {
		return nil, err
	}
	return search.ListLogs(r, s, h.logger)
}

func toApp(r *http.Request) (*model.Application, *e.Error) {
	var app model.Application
	bytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, e.Errorf(http.StatusInternalServerError, "Could not reead req body, %v", err)
	}
	err = json.Unmarshal(bytes, &app)
	if err != nil {
		return nil, e.Errorf(http.StatusInternalServerError, "Could not unmarshal data, %v", err)
	}
	return &app, nil
}

//SearchHandler search
func (h Handler) SearchHandler(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	var s, err = toSearch(r)
	if err != nil {
		return nil, err
	}
	res, er := search.Find(r, s, h.logger)
	return res, er
}

func toSearch(r *http.Request) (*model.Search, *e.Error) {
	var s model.Search
	bytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, e.Errorf(http.StatusInternalServerError, "Could not reead req body, %v", err)
	}
	err = json.Unmarshal(bytes, &s)
	if err != nil {
		return nil, e.Errorf(http.StatusInternalServerError, "Could not unmarshal data, %v", err)
	}
	return &s, nil
}
