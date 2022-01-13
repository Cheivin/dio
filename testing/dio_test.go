package testing

import (
	"context"
	"embed"
	"fmt"
	"github.com/cheivin/dio"
	"github.com/cheivin/dio-core"
	"testing"
)

type A struct {
	Log core.Log `aware:""`
}

func (A) BeanConstruct() {
	fmt.Println("Load A")
}

func (a A) AfterPropertiesSet() {
	a.Log.Info(context.Background(), "加载完成")
}

func TestRun(t *testing.T) {
	dio.SetProperty("app.env", "dev").
		ProvideOnProperty(A{}, "app.env", "dev").
		Run(context.Background())
}

//go:embed configs/*.yaml
var configs embed.FS

func TestYamlConfig(t *testing.T) {
	dio.LoadConfig(configs, "configs/dev.yaml")
	dio.Run(context.Background())
}

func Test_GetByType(t *testing.T) {
	property := dio.GetProperties("log.", core.Property{}).(core.Property)
	log, err := dio.NewZapLogger(property)
	if err != nil {
		t.Fatal(err)
	}
	dio.SetLogger(log)

	var x core.Log
	t.Log(dio.GetByType(&x))
}
