package mapper

import (
	"tampar-api/utils"
)

type CommonMapper struct {
	handler     utils.RequestHandler
	logger      utils.Logger
}

func NewCommonMapper(
	handler utils.RequestHandler,
	logger utils.Logger,
) CommonMapper {
	return CommonMapper{
		handler:     handler,
		logger:      logger,
	}
}
