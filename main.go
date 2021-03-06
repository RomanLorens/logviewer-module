package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	l "github.com/RomanLorens/logger/log"
	h "github.com/RomanLorens/logviewer-module/handler"
	"github.com/RomanLorens/logviewer-module/model"
)

func main() {

	http.HandleFunc("/", root)
	handler := h.NewHandler(l.PrintLogger(false))
	register("/lv/"+model.SearchEndpoint, handler.Search)
	register("/lv/"+model.ListLogsEndpoint, handler.ListLogs)
	register("/lv/"+model.StatsEndpoint, handler.Stats)
	register("/lv/"+model.ErrorsEndpoint, handler.Errors)
	register("/lv/"+model.DownloadLogEndpoint, handler.DownloadLog)
	register("/lv/"+model.CollectStatsEndpoint, handler.CollectStats)
	register("/lv/"+model.TailLogEndpoint, handler.TailLog)

	register("/lv/support/memory", handler.MemoryDiagnostics)
	register("/lv/support/health", handler.HealthHandler)
	register("/lv/support/proxy", handler.ProxyHandler)
	/*
		task := scheduler.Task{Name: "stats collector",
			Callback: func(ctx context.Context) {
				ls := &model.LogStructure{Date: 0, Level: 3, User: 1, Reqid: 2, Message: 5}
				s, err := stat.CollectStats("logs/out.log", ls, "2021/02/18")
				if err != nil {
					fmt.Printf("error %v", err)
				}
				fmt.Println(s)
			},
		}
		scheduler.Schedule(context.Background(), &task, time.Second*20)
	*/

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func root(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func register(endpoint string, f func(http.ResponseWriter, *http.Request) (interface{}, error)) {
	h := func(w http.ResponseWriter, r *http.Request) {
		res, err := f(w, r)
		if err != nil {
			fmt.Printf("error, %v", err)
			w.WriteHeader(500)
			return
		}
		if res == nil {
			return
		}
		if err := json.NewEncoder(w).Encode(res); nil != err {
			fmt.Printf("error, %v", err)
			w.WriteHeader(500)
			return
		}
	}

	http.HandleFunc(endpoint, h)
}
