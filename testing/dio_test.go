package testing

import (
	"context"
	"embed"
	"fmt"
	"github.com/cheivin/dio"
	"testing"
)

type A struct {
}

func (A) BeanConstruct() {
	fmt.Println("Load A")
}

func TestRun(t *testing.T) {
	dio.SetProperty("mysql", map[string]interface{}{
		"password":      "Changdao#666",
		"database":      "paycenter",
		"parameters":    "charset=UTF8&parseTime=true&loc=Asia%2FShanghai",
		"pool.max-idle": 10,
	})
	dio.SetProperty("app.env", "dev")
	dio.Web(true, true).
		MySQL().
		ProvideOnProperty(A{}, "app.env", "dev").
		Run(context.Background())
}

//go:embed configs/*.yaml
var configs embed.FS

func TestYamlConfig(t *testing.T) {
	dio.LoadConfig(configs, "configs/dev.yaml")
	dio.Web(true, true).Run(context.Background())
}
