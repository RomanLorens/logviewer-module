package handler

import (
	"net/http"

	e "github.com/RomanLorens/logviewer-module/error"
	"github.com/RomanLorens/logviewer-module/proxy"
)

//HealthHandler health
func HealthHandler(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	return "OK", nil
}

//ProxyHandler proxy
func ProxyHandler(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	err := proxy.Forward(r.FormValue("url"), &w, r)
	if err != nil {
		return nil, e.AppError("proxy error, %v", err)
	}
	return nil, nil
}
