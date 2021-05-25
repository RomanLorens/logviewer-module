package search

import (
	"context"
	"testing"

	"github.com/RomanLorens/logger/log"
)

func TestListLogs(t *testing.T) {

	ld := ListLogs(context.Background(), []string{"../../test-logs/"}, log.PrintLogger(false))

	if len(ld) == 0 {
		t.Error("should not be empty")
	}
}
