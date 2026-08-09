module github.com/cheivin/dio

go 1.25

require (
	github.com/cheivin/di v0.6.2
	github.com/cheivin/dio-core v0.6.2
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	go.uber.org/zap v1.28.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/lestrrat-go/strftime v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.8.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

// 本地开发：指向本地源码。发布后移除。
replace (
	github.com/cheivin/di => ../di
	github.com/cheivin/dio-core => ../dio-core
)
