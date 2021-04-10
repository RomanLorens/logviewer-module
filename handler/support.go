package handler

import (
	"fmt"
	"net/http"
	"runtime"

	e "github.com/RomanLorens/logviewer-module/error"
	"github.com/RomanLorens/logviewer-module/proxy"
)

type memory struct {
	Allocated      string
	TotalAllocated string
	System         string
	CPU            int
	Threads        int
	Raw            *runtime.MemStats
}

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

//MemoryDiagnostics memory diagnostics
func (h Handler) MemoryDiagnostics(w http.ResponseWriter, r *http.Request) (interface{}, *e.Error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &memory{Allocated: convertBytes(m.Alloc),
		TotalAllocated: convertBytes(m.TotalAlloc),
		System:         convertBytes(m.Sys),
		CPU:            runtime.NumCPU(),
		Threads:        runtime.NumGoroutine(),
		Raw:            &m}, nil
}

func convertBytes(b uint64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
