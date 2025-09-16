package api

import (
	"fmt"

	"github.com/fox-gonic/fox"
	adb "github.com/qiniu/zeroops/internal/alerting/database"
	"github.com/qiniu/zeroops/internal/alerting/service/healthcheck"
	receiver "github.com/qiniu/zeroops/internal/alerting/service/receiver"
	"github.com/qiniu/zeroops/internal/config"
)

type Api struct{}

func NewApi(router *fox.Engine) *Api { return NewApiWithConfig(router, nil) }

func NewApiWithConfig(router *fox.Engine, cfg *config.Config) *Api {
	api := &Api{}
	api.setupRouters(router, cfg)
	return api
}

func (api *Api) setupRouters(router *fox.Engine, cfg *config.Config) {
	var h *receiver.Handler
	var alertDB *adb.Database
	if cfg != nil {
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)
		if db, err := adb.New(dsn); err == nil {
			alertDB = db
			h = receiver.NewHandlerWithCache(receiver.NewPgDAO(db), receiver.NewCacheFromEnv())
		} else {
			h = receiver.NewHandler(receiver.NewNoopDAO())
		}
	} else {
		h = receiver.NewHandler(receiver.NewNoopDAO())
	}
	receiver.RegisterReceiverRoutes(router, h)

	// Issues query API (reads from Redis cache and loads comments from DB)
	RegisterIssueRoutes(router, healthcheck.NewRedisClientFromEnv(), alertDB)
}
