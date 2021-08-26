package testing

import (
	"context"
	"github.com/cheivin/dio"
	"testing"
)

func TestRun(t *testing.T) {
	dio.Web(true, true).
		MySQL().
		Run(context.Background())
}
