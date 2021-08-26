package testing

import (
	"context"
	"github.com/cheivin/dio"
	"testing"
)

func TestRun(t *testing.T) {
	dio.SetProperty("mysql", map[string]interface{}{
		"password":      "Changdao#666",
		"database":      "paycenter",
		"parameters":    "charset=UTF8&parseTime=true&loc=Asia%2FShanghai",
		"pool.max-idle": 10,
	})
	dio.Web(true, true).
		MySQL().
		Run(context.Background())
}
