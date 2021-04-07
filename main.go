package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	l "github.com/RomanLorens/logger/log"
	e "github.com/RomanLorens/logviewer-module/error"
	h "github.com/RomanLorens/logviewer-module/handler"
	"github.com/RomanLorens/logviewer-module/model"
)

func main() {

	/*
		_ls := &model.LogStructure{Date: 0, Level: 3, User: 1, Reqid: 2, Message: 5}
		_res, _ := stat.CollectStats(context.Background(), "logs/out.log", _ls, "2021/03/04")
		fmt.Println(_res)
		return
	*/

	// ls := search.LocalSearch{Logger: l.PrintLogger(false)}
	// search := &model.Search{Logs: []string{"logs/out.log"}, Value: "1-01-CV-PCVXGPXG1A1BEFYWYWLKBAYFJEAEJFG5487457603@2-2897134#61"}
	// res := ls.Grep(context.Background(), "http://localhost", search)
	// fmt.Printf(">>>> RES = %v\n", res[0].Lines)

	http.HandleFunc("/", root)
	handler := h.NewHandler(l.PrintLogger(false))
	register("/lv/health", handler.HealthHandler)
	register("/lv/"+model.SearchEndpoint, handler.SearchHandler)
	register("/lv/"+model.ListLogsEndpoint, handler.ListLogs)
	register("/lv/"+model.StatsEndpoint, handler.Stats)
	register("/lv/"+model.ErrorsEndpoint, handler.Errors)
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

func register(endpoint string, f func(http.ResponseWriter, *http.Request) (interface{}, *e.Error)) {
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
