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
	dio.SetProperty("app.env", "dev")
	dio.Use(dio.Web(true, true), dio.MySQL()).
		ProvideOnProperty(A{}, "app.env", "dev").
		Run(context.Background())
}

//go:embed configs/*.yaml
var configs embed.FS

func TestYamlConfig(t *testing.T) {
	dio.LoadConfig(configs, "configs/dev.yaml")
	dio.Use(dio.Web(true, true), dio.MySQL()).Run(context.Background())
}
