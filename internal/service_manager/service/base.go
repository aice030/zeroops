package service

import (
	"github.com/qiniu/zeroops/internal/service_manager/database"
	"github.com/rs/zerolog/log"
)

type Service struct {
	db *database.Database
}

func NewService(db *database.Database) *Service {
	service := &Service{
		db: db,
	}

	log.Info().Msg("Service initialized successfully")
	return service
}

func (s *Service) Close() error {
	return nil
}

func (s *Service) GetDatabase() *database.Database {
	return s.db
}
