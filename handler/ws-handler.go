package handler

import (
	"context"
	"net/http"
	"time"

	e "github.com/RomanLorens/logviewer-module/error"
	"github.com/RomanLorens/logviewer-module/model"
	"github.com/RomanLorens/logviewer-module/search"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func init() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
}

//Health health
type Health struct {
	Host   string
	App    string
	Env    string
	Status int
}

//TailLogWS tail logs
func TailLogWS(w http.ResponseWriter, r *http.Request) *e.Error {
	c, er := upgrader.Upgrade(w, r, nil)
	if er != nil {
		return e.Errorf(500, "Could not create websocket, %v", er)
	}
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		closeWS(r.Context(), c)
	}()
	var app model.Application
	if er := c.ReadJSON(&app); er != nil {
		return e.Errorf(400, "Could not parse incoming request, %v", er)
	}

	done := make(chan bool)
	go func(c *websocket.Conn) {
		res, err := search.TailLog(r.Context(), &app)
		if err != nil {
			logger.Error(r.Context(), "Error from tail %v", err)
			done <- true
			return
		}
		c.WriteJSON(res)
		for {
			select {
			case <-ticker.C:
				logger.Info(r.Context(), "Checking tail logs with ticker")
				res, err := search.TailLog(r.Context(), &app)
				if err != nil {
					logger.Error(r.Context(), "Error from tail %v", err)
					done <- true
					break
				}
				c.WriteJSON(res)
			}
		}
	}(c)

	go func(c *websocket.Conn) {
		_, _, err := c.ReadMessage()
		if err != nil {
			logger.Info(r.Context(), "Closing connection - %v", err)
			done <- true
		}
	}(c)

	<-done
	return nil
}

//AppsHealth apps health
func AppsHealth(w http.ResponseWriter, r *http.Request) *e.Error {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return e.Errorf(500, "Could not create websocket, %v", err)
	}
	defer closeWS(r.Context(), c)

	//TODO pass heath urls
	/*
		healths := make([]*Health, 0)
		for _, cfg := range config.Config.ApplicationsConfig {
			healthPath := findHealthPath(cfg)
			if healthPath == "" {
				continue
			}
			for _, host := range cfg.Hosts {
				healthURL := path.Join(host.AppHost, healthPath)
				healths = append(healths, &Health{Host: healthURL, App: cfg.Application, Env: cfg.Env})
			}
		}

		var wg sync.WaitGroup
		for _, h := range healths {
			wg.Add(1)
			go func(h *Health, c *websocket.Conn) {
				checkHealth(r.Context(), h)
				err := c.WriteJSON(h)
				if err != nil {
					logger.Error(r.Context(), "error when writing json to ws, %v", err)
				}
				wg.Done()
			}(h, c)
		}

		wg.Wait()
	*/
	return nil
}

func closeWS(ctx context.Context, c *websocket.Conn) {
	logger.Info(ctx, "Closing ws connection")
	c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(2 * time.Second)
	c.Close()
}

func checkHealth(ctx context.Context, h *Health) {
	logger.Info(ctx, "Checking health %v", h.Host)
	resp, err := http.Get(h.Host)
	if err != nil {
		logger.Error(ctx, "error from %v - %v", h.Host, err)
		return
	}
	logger.Info(ctx, "Response %v", resp)
	h.Status = resp.StatusCode
}
