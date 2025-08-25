package utils

import "go.uber.org/fx"

// Depedency
var Module = fx.Options(
	fx.Provide(NewEnv),
	fx.Provide(NewRequestHandler),
	fx.Provide(NewLogger),
	fx.Provide(NewDatabase),
)
