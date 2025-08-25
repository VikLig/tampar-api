package businessconfig

import (
	"context"
	"tampar-api/mapper"
	"tampar-api/rest"
	"tampar-api/service"
	"tampar-api/utils"

	_ "github.com/sijms/go-ora/v2"

	"go.uber.org/fx"
)

var Module = fx.Options(
	utils.Module,
	mapper.Module,
	rest.Module,
	service.Module,
	fx.Invoke(businessConfig),
)

func businessConfig(
	lifecycle fx.Lifecycle,
	config utils.Config,
	handler utils.RequestHandler,
	logger utils.Logger,
	database utils.Database,
	listRest rest.ListRest,
) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Logger.Infof("SERVER STARTED")
			go func() {
				listRest.Setup()
				handler.Gin.Run(":" + config.ServerPort)
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			database.Database.Close()
			logger.Logger.Infof("SERVER STOPPED")
			return nil
		},
	})
}
