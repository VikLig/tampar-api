package utils

import (
	"net/http"
	"strings"
	"tampar-api/model/constant"

	"github.com/gin-gonic/gin"
)

type RequestHandler struct {
	Gin *gin.Engine
}

func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.ToLower(NewEnv().Environment) == "prod" {
			c.Header("Content-Type", "application/json")
			c.Header("X-Content-Type-Options", "nosniff")
			c.Header("X-Frame-Options", "SAMEORIGIN")
			c.Header("Access-Control-Allow-Origin", NewEnv().AccessControlAllowOrigin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, X-Auth-Token, content-type")
			c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			c.Header("Cache-Control", "no-store")
			c.Header("Pragma", "no-cache")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "*")
			c.Header("Access-Control-Allow-Methods", "POST,HEAD,PATCH, OPTIONS, GET, PUT")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}

func NewRequestHandler(config Config) RequestHandler {
	gin.SetMode(config.AppMode)
	engine := gin.New()

	engine.Use(CorsMiddleware())
	engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, "Ooops Page Not Found")
	})

	engine.GET("/"+constant.APP_NAME, func(c *gin.Context) {
		c.JSON(http.StatusOK, "Hello, welcome to Web Service "+constant.APP_NAME+"!")
	})

	engine.GET("/"+constant.APP_NAME+"/", func(c *gin.Context) {
		c.JSON(http.StatusOK, "Hello, welcome to Web Service "+constant.APP_NAME+"!")
	})

	// engine.Use(cors.Default())

	engine.SetTrustedProxies(nil)
	return RequestHandler{Gin: engine}
}
