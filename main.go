package main

import (
	"anileha/analyze"
	"anileha/config"
	"anileha/controller"
	"anileha/db"
	"anileha/logger"
	"anileha/service"
	"go.uber.org/fx"
)

// TODO: embed structs in return type
// func keklol() (>>orel MyStruct<<)

// TODO: support mp4 torrents (?)

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

		// misc
		analyze.TextAnalyzerExport,
		analyze.ProbeAnalyzerExport,
	).Run()
}
