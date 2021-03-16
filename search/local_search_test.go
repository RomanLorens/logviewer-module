package search

import (
	"context"
	"testing"

	"github.com/RomanLorens/logviewer-module/model"
)

func TestGrep(t *testing.T) {
	ls := LocalSearch{}
	search := &model.Search{Logs: []string{"logs/out.log"}, Value: "1-01-CV-PCVXGPXG1A1BEFYWYWLKBAYFJEAEJFG5487457603@2-2897134#61"}
	res := ls.Grep(context.Background(), "http://localhost", search)
	for _, r := range res {
		t.Logf("lines %v", r.Lines)
	}
	if len(res[0].Lines) == 0 {
		t.Fatal("Should be match")
	}
}
