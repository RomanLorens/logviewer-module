package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	e "github.com/RomanLorens/logviewer-module/error"
	h "github.com/RomanLorens/logviewer-module/handler"
	"github.com/RomanLorens/logviewer-module/model"
)

func main() {
	http.HandleFunc("/", root)
	register("/lv/health", h.HealthHandler)
	register("/lv/"+model.SearchEndpoint, h.SearchHandler)
	register("/lv/"+model.ListLogsEndpoint, h.ListLogs)
	register("/lv/"+model.StatsEndpoint, h.Stats)
	register("/lv/"+model.ErrorsEndpoint, h.Errors)
	register("/lv/support/proxy", h.ProxyHandler)

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
