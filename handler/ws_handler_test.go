package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	l "github.com/RomanLorens/logger/log"
	"github.com/RomanLorens/logviewer-module/model"
	"github.com/gorilla/websocket"
)

var h = NewHandler(l.PrintLogger(false))

func TestWSTailLog(t *testing.T) {

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.TailLogWS(w, r)
	}))
	defer s.Close()
	u := "ws" + strings.TrimPrefix(s.URL, "http")
	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	lr := model.LogRequest{Log: "../test-logs/java-app.log"}
	ws.WriteJSON(lr)
	if err := ws.WriteJSON(lr); err != nil {
		t.Fatalf("%v", err)
	}

	_, p, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("%v", err)
	}
	t.Logf("ws read %v \n", string(p))

	time.Sleep(10 * time.Second)

	ws.Close()
}
