package testing

import (
	"context"
	"embed"
	"fmt"
	"github.com/cheivin/dio"
	web "github.com/cheivin/dio/plugin/gin_web"
	orm "github.com/cheivin/dio/plugin/gorm"
	"testing"
)

type A struct {
}

func (A) BeanConstruct() {
	fmt.Println("Load A")
}

func TestRun(t *testing.T) {
	dio.SetProperty("app.env", "dev")
	dio.Use(web.GinWeb(true, true), orm.Gorm()).
		ProvideOnProperty(A{}, "app.env", "dev").
		Run(context.Background())
}

//go:embed configs/*.yaml
var configs embed.FS

func TestYamlConfig(t *testing.T) {
	dio.LoadConfig(configs, "configs/dev.yaml")
	dio.Use(web.GinWeb(true, true), orm.Gorm()).Run(context.Background())
}
