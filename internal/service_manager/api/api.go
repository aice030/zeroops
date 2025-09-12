package api

import (
	"github.com/fox-gonic/fox"
	"github.com/qiniu/zeroops/internal/service_manager/database"
	"github.com/qiniu/zeroops/internal/service_manager/service"
)

type Api struct {
	db      *database.Database
	service *service.Service
	router  *fox.Engine
}

func NewApi(db *database.Database, service *service.Service, router *fox.Engine) (*Api, error) {
	api := &Api{
		db:      db,
		service: service,
		router:  router,
	}

	api.setupRouters(router)
	return api, nil
}

func (api *Api) setupRouters(router *fox.Engine) {
	// 服务信息相关路由
	api.setupInfoRouters(router)

	// 部署管理相关路由
	api.setupDeployRouters(router)
}
