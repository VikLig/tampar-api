package rest

import (
	"tampar-api/service"
	"tampar-api/utils"
)

type CommonRest struct {
	handler   utils.RequestHandler
	commonSvc service.CommonSvc
	config    utils.Config
}

func NewCommonRest(
	handler utils.RequestHandler,
	commonSvc service.CommonSvc,
	config utils.Config,
) CommonRest {
	return CommonRest{
		handler:   handler,
		commonSvc: commonSvc,
		config:    config,
	}
}

func (r CommonRest) Setup() {
	route := r.handler.Gin.Group("/" + r.config.AppName)
	{
		profileRoute := route.Group("/common")
		{
			profileRoute.POST("/process", r.commonSvc.Process)
			profileRoute.GET("/downloadTemplate", r.commonSvc.DownloadTemplate)
			profileRoute.POST("/getSchema", r.commonSvc.GetSchema)
			//profileRoute.GET("/downloadTemplateLog/:key", r.commonSvc.DownloadTemplateLog)
		}
	}
}
