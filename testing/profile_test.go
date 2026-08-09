package testing

import (
	"testing"

	"github.com/cheivin/dio"
)

// TestProfile 验证 profile 的解析优先级：显式 SetProfile > 环境变量 APP_PROFILE > 空。
func TestProfile(t *testing.T) {
	defer dio.Reset()
	if p := dio.Profile(); p != "" {
		t.Fatalf("default Profile should be empty, got %q", p)
	}
	dio.SetProfile("dev")
	if p := dio.Profile(); p != "dev" {
		t.Fatalf("Profile after SetProfile = %q, want dev", p)
	}
	// 显式设置优先于环境变量
	t.Setenv("APP_PROFILE", "prod")
	if p := dio.Profile(); p != "dev" {
		t.Fatalf("SetProfile should win over APP_PROFILE, got %q", p)
	}
	dio.SetProfile("")
	if p := dio.Profile(); p != "prod" {
		t.Fatalf("Profile after clearing SetProfile = %q, want prod (from env)", p)
	}
}

// TestLoadConfigWithProfile 验证 LoadConfig 自动加载 config-{profile}.yaml 覆盖公共配置：
// 公共配置走 SetDefault（低优先级），profile 覆盖走 Set（高优先级）。
func TestLoadConfigWithProfile(t *testing.T) {
	defer dio.Reset()
	dio.SetProfile("dev")
	dio.LoadConfig(configs, "configs/config.yaml")

	if v := dio.GetPropertyString("app.name"); v != "base-app" {
		t.Fatalf("app.name = %q, want base-app (common config)", v)
	}
	if v := dio.GetPropertyString("app.env"); v != "dev" {
		t.Fatalf("app.env = %q, want dev (profile override)", v)
	}
	if v := dio.GetPropertyString("feature.enable"); v != "true" {
		t.Fatalf("feature.enable = %q, want true (profile override)", v)
	}
}

// TestLoadConfigWithoutProfile 验证未设置 profile 时不加载覆盖配置。
func TestLoadConfigWithoutProfile(t *testing.T) {
	defer dio.Reset()
	dio.LoadConfig(configs, "configs/config.yaml")

	if v := dio.GetPropertyString("app.env"); v != "prod" {
		t.Fatalf("app.env = %q, want prod (no profile)", v)
	}
	if v := dio.GetPropertyString("feature.enable"); v != "false" {
		t.Fatalf("feature.enable = %q, want false (no profile)", v)
	}
}

// TestLoadConfigProfileNotExist 验证 profile 覆盖文件不存在时静默忽略。
func TestLoadConfigProfileNotExist(t *testing.T) {
	defer dio.Reset()
	dio.SetProfile("prod")
	dio.LoadConfig(configs, "configs/config.yaml")

	if v := dio.GetPropertyString("app.env"); v != "prod" {
		t.Fatalf("app.env = %q, want prod (profile file not exist)", v)
	}
}
