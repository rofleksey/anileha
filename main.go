package main

import (
	"anileha/config"
	"anileha/controller"
	"anileha/db"
	"anileha/logger"
	"anileha/service"
	"go.uber.org/fx"
)

// TODO: improve logs

//logger.Info("APNS: Connection error before reading complete response",
//zap.Int("connectionId", conn.id),
//zap.Int("n", n),
//zap.Error(err),
//)

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
		service.TorrentServiceExport,
		service.ConversionServiceExport,

		// rest controllers
		controller.HealthControllerExport,
		controller.SeriesControllerExport,
		controller.ThumbnailControllerExport,
		controller.TorrentControllerExport,
	).Run()
}
