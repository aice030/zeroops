package main

import (
	"github.com/fox-gonic/fox"
	"github.com/qiniu/zeroops/internal/config"
	"github.com/qiniu/zeroops/internal/middleware"
	servicemanager "github.com/qiniu/zeroops/internal/service_manager"

	// releasesystem "github.com/qiniu/zeroops/internal/release_system/api"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Info().Msg("Starting zeroops api server")
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	serviceManagerSrv, err := servicemanager.NewServiceManagerServer(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create release system api")
	}
	defer func() {
		serviceManagerSrv.Close()
	}()

	router := fox.New()
	router.Use(middleware.Authentication)
	if err := serviceManagerSrv.UseApi(router); err != nil {
		log.Fatal().Err(err).Msg("bind serviceManagerApi failed.")
	}
	log.Info().Msgf("Starting server on %s", cfg.Server.BindAddr)
	if err := router.Run(cfg.Server.BindAddr); err != nil {
		log.Fatal().Err(err).Msg("start zeroops api server failed.")
	}
	log.Info().Msg("zeroops api server exit...")
}
