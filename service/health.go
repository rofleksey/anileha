package service

import "go.uber.org/fx"

type HealthService struct{}

func (s *HealthService) GetHealth() bool {
	return true
}

func NewHealthService() *HealthService {
	return &HealthService{}
}

var HealthServiceExport = fx.Options(fx.Provide(NewHealthService))
