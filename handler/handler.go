package handler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	e "github.com/RomanLorens/logviewer-module/error"
	l "github.com/RomanLorens/logviewer-module/logger"
	"github.com/RomanLorens/logviewer-module/model"
	"github.com/RomanLorens/logviewer-module/search"
	"github.com/RomanLorens/logviewer-module/stat"
)

var logger = l.L

//DownloadLog download log
func DownloadLog(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	var ld model.LogDownload
	err := json.NewDecoder(r.Body).Decode(&ld)
	if err != nil {
		return nil, e.AppError("missing log download body, %w", err)
	}
	defer r.Body.Close()
	b, er := search.DownloadLog(r.Context(), &ld)
	if er != nil {
		return nil, e.AppError("download log,%w", err)
	}

	w.Header().Add("Content-Type", "application/octet-stream")
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=\"%v\"", path.Base(ld.Log)))
	w.Write(b)
	return nil, nil
}

//Stats stats
func Stats(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	app, err := toApp(r)
	if err != nil {
		return nil, err
	}
	return stat.Get(r.Context(), app)
}

//CollectStats collect stats
func CollectStats(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	var s model.CollectStats
	err := json.NewDecoder(r.Body).Decode(&s)
	if err != nil {
		return nil, e.ClientError("Could not parse req body, %v", err)
	}
	return stat.CollectStats(s.LogPath, s.LogStructure, s.Date)
}

//Errors errors
func Errors(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	app, err := toApp(r)
	if err != nil {
		return nil, err
	}
	return stat.GetErrors(r.Context(), app)
}

//TailLog tail log
func TailLog(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	app, err := toApp(r)
	if err != nil {
		return nil, err
	}
	return search.TailLog(r.Context(), app)
}

//ListLogs list logs
func ListLogs(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	var s, err = toSearch(r)
	if err != nil {
		return nil, err
	}
	return search.ListLogs(r.Context(), s)
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
func SearchHandler(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	var s, err = toSearch(r)
	if err != nil {
		return nil, err
	}
	res, er := search.Find(r.Context(), s)
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
