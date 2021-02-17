package search

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	e "cedt-icg-bitbucket.nam.nsroot.net/bitbucket/users/rl78794/repos/logviewer-module/error"
	"cedt-icg-bitbucket.nam.nsroot.net/bitbucket/users/rl78794/repos/logviewer-module/model"
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
func (RemoteSearch) Tail(ctx context.Context, app *model.Application) (*model.Result, *e.Error) {
	logger.Info(ctx, "Tail log remotely")
	var res *model.Result
	url := ApiURL(app.Host, model.TailLogEndpoint)
	body, err := CallAPI(ctx, url, app)
	if err != nil {
		return nil, err
	}
	if er := json.Unmarshal(body, &res); er != nil {
		return nil, e.Errorf(500, "Could not unmarshal, %v", er)
	}
	return res, nil
}

//DownloadLog download log
func (RemoteSearch) DownloadLog(ctx context.Context, ld *model.LogDownload) ([]byte, *e.Error) {
	logger.Info(ctx, "Download log remotely")
	url := ApiURL(ld.Host, model.DownloadLogEndpoint)
	return CallAPI(ctx, url, ld)
}

//Grep grep logs
func (RemoteSearch) Grep(ctx context.Context, url string, s *model.Search) ([]*model.Result, *e.Error) {
	logger.Info(ctx, "Grep log remotely")
	url = ApiURL(url, model.SearchEndpoint)
	body, err := CallAPI(ctx, url, s)
	if err != nil {
		return nil, err
	}
	r := make([]*model.Result, len(s.Logs))
	if er := json.Unmarshal(body, &r); er != nil {
		return r, e.Errorf(500, "Could not read unmarshal, %v", er)
	}
	return r, nil
}

//List list logs
func (RemoteSearch) List(ctx context.Context, url string, s *model.Search) ([]*model.LogDetails, *e.Error) {
	var logs []*model.LogDetails
	url = ApiURL(url, model.ListLogsEndpoint)
	body, err := CallAPI(ctx, url, s)
	if err != nil {
		return nil, err
	}
	if er := json.Unmarshal(body, &logs); er != nil {
		return nil, e.Errorf(500, "Could not read unmarshal, %v", er)
	}
	return logs, nil
}

//CallAPI call api
func CallAPI(ctx context.Context, url string, post interface{}) ([]byte, *e.Error) {
	logger.Info(ctx, "Remote api for %v", url)
	b, err := json.Marshal(post)
	if err != nil {
		return nil, e.Errorf(500, "Could not marshal post %v", err)
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, e.Errorf(500, "Request to %v failed, %v", url, err)
	}
	logger.Info(ctx, "Api response %v", resp)
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
