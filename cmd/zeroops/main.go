package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/fox-gonic/fox"
	alertapi "github.com/qiniu/zeroops/internal/alerting/api"
	adb "github.com/qiniu/zeroops/internal/alerting/database"
	"github.com/qiniu/zeroops/internal/alerting/service/healthcheck"
	"github.com/qiniu/zeroops/internal/alerting/service/remediation"
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

	// optional alerting DB for healthcheck and remediation
	var alertDB *adb.Database
	{
		dsn := func() string {
			return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
				cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)
		}()
		if db, derr := adb.New(dsn); derr == nil {
			alertDB = db
		} else {
			log.Error().Err(derr).Msg("healthcheck alerting DB init failed; scheduler/consumer will run without DB")
		}
	}

	// start healthcheck scheduler and remediation consumer
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	interval := parseDuration(os.Getenv("HC_SCAN_INTERVAL"), 10*time.Second)
	batch := parseInt(os.Getenv("HC_SCAN_BATCH"), 200)
	workers := parseInt(os.Getenv("HC_WORKERS"), 1)
	if workers < 1 {
		workers = 1
	}
	alertChSize := parseInt(os.Getenv("REMEDIATION_ALERT_CHAN_SIZE"), 1024)
	alertCh := make(chan healthcheck.AlertMessage, alertChSize)

	for i := 0; i < workers; i++ {
		go healthcheck.StartScheduler(ctx, healthcheck.Deps{
			DB:       alertDB,
			Redis:    healthcheck.NewRedisClientFromEnv(),
			AlertCh:  alertCh,
			Batch:    batch,
			Interval: interval,
		})
	}
	rem := remediation.NewConsumer(alertDB, healthcheck.NewRedisClientFromEnv())
	go rem.Start(ctx, alertCh)

	router := fox.New()
	router.Use(middleware.Authentication)
	alertapi.NewApiWithConfig(router, cfg)
	if err := serviceManagerSrv.UseApi(router); err != nil {
		log.Fatal().Err(err).Msg("bind serviceManagerApi failed.")
	}
	log.Info().Msgf("Starting server on %s", cfg.Server.BindAddr)
	if err := router.Run(cfg.Server.BindAddr); err != nil {
		log.Fatal().Err(err).Msg("start zeroops api server failed.")
	}
	log.Info().Msg("zeroops api server exit...")
}

func parseDuration(s string, d time.Duration) time.Duration {
	if s == "" {
		return d
	}
	if v, err := time.ParseDuration(s); err == nil {
		return v
	}
	return d
}

func parseInt(s string, v int) int {
	if s == "" {
		return v
	}
	if n, err := strconv.Atoi(s); err == nil {
		return n
	}
	return v
}
