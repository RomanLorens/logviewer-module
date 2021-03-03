package search

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	e "github.com/RomanLorens/logviewer-module/error"
	"github.com/RomanLorens/logviewer-module/model"
)

//RemoteSearch remote search
type RemoteSearch struct{}

//trust all - mostly for requests from localhost
var (
	client = &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
)

//Tail tail log
func (RemoteSearch) Tail(r *http.Request, app *model.Application) (*model.Result, *e.Error) {
	logger.Info(r.Context(), "Tail log remotely")
	var res *model.Result
	url := ApiURL(app.Host, model.TailLogEndpoint)
	body, err := CallAPI(r, url, app)
	if err != nil {
		return nil, err
	}
	if er := json.Unmarshal(body, &res); er != nil {
		return nil, e.Errorf(500, "Could not unmarshal, %v", er)
	}
	return res, nil
}

//DownloadLog download log
func (RemoteSearch) DownloadLog(r *http.Request, ld *model.LogDownload) ([]byte, *e.Error) {
	logger.Info(r.Context(), "Download log remotely")
	url := ApiURL(ld.Host, model.DownloadLogEndpoint)
	return CallAPI(r, url, ld)
}

//Grep grep logs
func (RemoteSearch) Grep(r *http.Request, url string, s *model.Search) ([]*model.Result, *e.Error) {
	logger.Info(r.Context(), "Grep log remotely")
	url = ApiURL(url, model.SearchEndpoint)
	body, err := CallAPI(r, url, s)
	if err != nil {
		return nil, err
	}
	res := make([]*model.Result, len(s.Logs))
	if er := json.Unmarshal(body, &r); er != nil {
		return res, e.Errorf(500, "Could not read unmarshal, %v", er)
	}
	return res, nil
}

//List list logs
func (RemoteSearch) List(r *http.Request, url string, s *model.Search) ([]*model.LogDetails, *e.Error) {
	var logs []*model.LogDetails
	url = ApiURL(url, model.ListLogsEndpoint)
	body, err := CallAPI(r, url, s)
	if err != nil {
		return nil, err
	}
	if er := json.Unmarshal(body, &logs); er != nil {
		return nil, e.Errorf(500, "Could not read unmarshal, %v", er)
	}
	return logs, nil
}

//CallAPI call api
func CallAPI(r *http.Request, url string, post interface{}) ([]byte, *e.Error) {
	logger.Info(r.Context(), "Remote api for %v", url)
	b, err := json.Marshal(post)
	if err != nil {
		return nil, e.Errorf(500, "Could not marshal post %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, e.AppError("Could not create req for %v, %v", url, err)
	}
	req.Header.Add("Content-Type", "application/json")
	for k, vals := range r.Header {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, e.Errorf(500, "Request to %v failed, %v", url, err)
	}
	logger.Info(r.Context(), "Api response %v", resp)
	if resp.StatusCode != 200 {
		return nil, e.Errorf(resp.StatusCode, "Request to %v failed, %v", url, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, e.Errorf(500, "Could not read response, %v", err)
	}
	return body, nil
}

//ApiURL api url
func ApiURL(url string, api string) string {
	if strings.HasSuffix(url, api) {
		return url
	}
	if url[len(url)-1:] != "/" {
		url = url + "/"
	}
	return url + api
}
