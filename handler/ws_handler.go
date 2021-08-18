package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/RomanLorens/logviewer-module/model"
	"github.com/RomanLorens/logviewer-module/search"
	"github.com/RomanLorens/logviewer-module/utils"
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
func (h Handler) TailLogWS(w http.ResponseWriter, r *http.Request) error {
	c, er := upgrader.Upgrade(w, r, nil)
	if er != nil {
		return fmt.Errorf("Could not create websocket, %v", er)
	}
	ticker := time.NewTicker(5 * time.Second)
	defer func() {
		ticker.Stop()
		h.closeWS(r.Context(), c)
	}()
	h.logger.Info(r.Context(), "Accepted ws connection from %v", r.RemoteAddr)
	var lr model.LogRequest
	done := make(chan bool)
	if er := c.ReadJSON(&lr); er != nil {
		<-done
		return fmt.Errorf("Could not parse incoming request, %v", er)
	}

	go func(c *websocket.Conn) {
		utils.CatchError(r.Context(), h.logger)
		res, err := search.Tail(lr.Log)
		if err != nil {
			h.logger.Error(r.Context(), "Error from tail %v", err)
			done <- true
			return
		}
		c.WriteJSON(res)
		modtime := res.ModTime
		i := 0
		for {
			select {
			case <-ticker.C:
				h.logger.Info(r.Context(), "tail logs ticker %v - modtime %v", i, modtime)
				i++
				res, isNeeded, err := search.TailLogIfNewer(lr.Log, modtime)
				if err != nil {
					h.logger.Error(r.Context(), "Error from tail %v", err)
					done <- true
					break
				}
				if !isNeeded {
					h.logger.Info(r.Context(), "%v was not modified since %v", lr.Log, modtime)
					continue
				}
				modtime = res.ModTime
				c.WriteJSON(res)
				if i >= 100 {
					h.logger.Info(r.Context(), "timeout limit for ws ticker reached - closing connection")
					done <- true
					break
				}
			}
		}
	}(c)

	go func(c *websocket.Conn) {
		utils.CatchError(r.Context(), h.logger)
		_, _, err := c.ReadMessage()
		if err != nil {
			h.logger.Info(r.Context(), "Closing connection - %v", err)
			done <- true
		}
	}(c)

	<-done
	return nil
}

//AppsHealth apps health
func (h Handler) AppsHealth(w http.ResponseWriter, r *http.Request) error {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("Could not create websocket, %v", err)
	}
	defer h.closeWS(r.Context(), c)

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

func (h Handler) closeWS(ctx context.Context, c *websocket.Conn) {
	h.logger.Info(ctx, "Closing ws connection")
	c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	time.Sleep(2 * time.Second)
	c.Close()
}

func (h Handler) checkHealth(ctx context.Context, health *Health) {
	h.logger.Info(ctx, "Checking health %v", health.Host)
	resp, err := http.Get(health.Host)
	if err != nil {
		h.logger.Error(ctx, "error from %v - %v", health.Host, err)
		return
	}
	h.logger.Info(ctx, "Response %v", resp)
	health.Status = resp.StatusCode
}
