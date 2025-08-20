package rest

import (
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(NewCommonRest),
	fx.Provide(NewRest),
)

type ListRest []Rest

type Rest interface {
	Setup()
}

func NewRest(
	commonRest CommonRest,
) ListRest {
	return ListRest{
		commonRest,
	}
}

func (r ListRest) Setup() {
	for _, rest := range r {
		rest.Setup()
	}
}
