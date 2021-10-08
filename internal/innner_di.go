package internal

import "github.com/cheivin/di"

var innerDi *di.DI

func SetDi(di *di.DI) {
	innerDi = di
}

func Di() *di.DI {
	return innerDi
}
