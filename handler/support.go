package handler

import (
	"fmt"
	"net/http"
	"runtime"

	"github.com/RomanLorens/logviewer-module/proxy"
)

type memory struct {
	Allocated string
	System    string
	CPU       int
	Threads   int
	Raw       *runtime.MemStats
}

//HealthHandler health
func (h Handler) HealthHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return "OK", nil
}

//ProxyHandler proxy
func (h Handler) ProxyHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	err := proxy.Forward(r.FormValue("url"), &w, r, h.logger)
	if err != nil {
		return nil, fmt.Errorf("proxy error, %v", err)
	}
	return nil, nil
}

//MemoryDiagnostics memory diagnostics
func (h Handler) MemoryDiagnostics(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return &memory{Allocated: convertBytes(m.Alloc),
		System:  convertBytes(m.Sys),
		CPU:     runtime.NumCPU(),
		Threads: runtime.NumGoroutine(),
		Raw:     &m}, nil
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
