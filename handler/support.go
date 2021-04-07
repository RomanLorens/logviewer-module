package handler

import (
	"net/http"

	e "github.com/RomanLorens/logviewer-module/error"
	"github.com/RomanLorens/logviewer-module/proxy"
)

//HealthHandler health
func (h Handler) HealthHandler(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	return "OK", nil
}

//ProxyHandler proxy
func (h Handler) ProxyHandler(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	err := proxy.Forward(r.FormValue("url"), &w, r, h.logger)
	if err != nil {
		return nil, e.AppError("proxy error, %v", err)
	}
	return nil, nil
}
