---
layout: default
title: 安装
nav_order: 2
parent: 入门
---

# 安装

## 版本要求

- **Go 1.25+**

## 安装

```bash
go get github.com/cheivin/dio@latest
```

dio 依赖以下模块（自动拉取）：

| 模块 | 说明 |
|------|------|
| [di](https://github.com/Cheivin/di) | 底层依赖注入容器 |
| [dio-core](https://github.com/Cheivin/dio-core) | 接口定义层 |
| [zap](https://github.com/uber-go/zap) | 日志实现 |
| [file-rotatelogs](https://github.com/lestrrat-go/file-rotatelogs) | 日志滚动 |

## 可选插件

```bash
go get github.com/cheivin/dio-plugin/gin@latest
go get github.com/cheivin/dio-plugin/gorm@latest
```

> gin / gorm 插件共用 `dio-plugin` 仓库，通过 `gin/` / `gorm/` 子模块与 tag 前缀发布，`@latest` 会自动解析各自最新的 `gin/vX` / `gorm/vX` 版本。

## 验证

```go
package main

import (
	"context"
	"time"

	"github.com/cheivin/dio"
)

func main() {
	dio.SetBanner("")
	dio.SetProperty("app.port", 8080)
	dio.Run(context.Background())
	// 按 Ctrl+C 退出，观察优雅停机日志
}
```
