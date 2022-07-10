package main

import (
	"anileha/config"
	"anileha/controller"
	"anileha/db"
	"anileha/logger"
	"anileha/service"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		// main components
		logger.Export,
		config.Export,
		controller.RestExport,
		db.ServiceExport,

		// services
		service.FileServiceExport,
		service.HealthServiceExport,
		service.SeriesServiceExport,
		service.ThumbnailServiceExport,

		// rest controllers
		controller.HealthControllerExport,
		controller.SeriesControllerExport,
		controller.ThumbnailControllerExport,
	).Run()
}
