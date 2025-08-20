package main

import (
	"tampar-api/businessconfig"

	"go.uber.org/fx"
)

// Package main is an example REST app
// @title           Swagger Doc API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/
// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @schemes http https
// @BasePath  /tampar-api

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
func main() {
	fx.New(businessconfig.Module).Run()
}
