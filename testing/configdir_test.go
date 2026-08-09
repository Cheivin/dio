package testing

import (
	"testing"

	"github.com/cheivin/dio"
)

// TestLoadConfigDir 验证目录配置加载：*.yaml 按文件名排序合并，后加载覆盖同名配置项。
func TestLoadConfigDir(t *testing.T) {
	defer dio.Reset()
	dio.LoadConfigDir(configs, "configs/configdir")

	if v := dio.GetPropertyString("app.name"); v != "dir-app" {
		t.Fatalf("app.name = %q, want dir-app (from 01-base)", v)
	}
	if v := dio.GetPropertyString("feature.flag"); v != "true" {
		t.Fatalf("feature.flag = %q, want true (from 02-extra)", v)
	}
	if v := dio.GetPropertyString("app.port"); v != "9090" {
		t.Fatalf("app.port = %q, want 9090 (03-override should win over 01-base)", v)
	}

	// 显式 Set 高于文件（优先级链最顶层）
	dio.SetProperty("app.port", 9999)
	if v := dio.GetPropertyString("app.port"); v != "9999" {
		t.Fatalf("app.port = %q, want 9999 (explicit Set should win)", v)
	}
}
