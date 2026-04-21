package main

import (
	"blog-post-gin/internal/config"
	"blog-post-gin/internal/entity"
	"blog-post-gin/internal/handler"
	"blog-post-gin/internal/route"
	"blog-post-gin/internal/service"

	"github.com/PeterTakahashi/gin-openapi/openapiui"
	"github.com/gin-gonic/gin"
)

// @title Blog Post Kazu API
// @version 1.0
// @description API documentation for Blog Post Kazu service.
// @BasePath /

func main() {
	cfg := config.LoadConfig()
	db, err := config.InitDB(cfg)

	if err != nil {
		panic(err)
	}

	if err := config.RunMigrations(db); err != nil {
		panic(err)
	}

	argonCfg, err := config.ParseArgonConfig(cfg)
	if err != nil {
		panic(err)
	}

	jwtCfg, cookieCfg, err := config.ParseJWTConfig(cfg)
	if err != nil {
		panic(err)
	}

	authService := service.NewAuthService(db, jwtCfg, argonCfg, cookieCfg)

	authHandler := handler.NewAuthHandler(authService, entity.SessionCookieConfig{})
	healthHandler := handler.NewHealthHandler(db)

	router := gin.New()
	router.SetTrustedProxies([]string{"localhost"})

	route.SetupRouter(router, &route.Routes{
		HealthHandler: healthHandler,
		AuthHandler:   authHandler,
	})

	router.GET("/docs/*any", openapiui.WrapHandler(openapiui.Config{
		SpecURL:      "/docs/openapi.json",
		SpecFilePath: "./docs/swagger.json",
		Title:        "Blog Post Kazu API",
		Theme:        "Dark",
	}))

	router.Run()
}
