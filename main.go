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

// HINT: embed structs in return type (?)
// func keklol() (>>orel MyStruct<<)
// HINT: replace all model pointers to objects (where possible)
// HINT: properly recover from all errors
// HINT: remove rows changed check where possible

// FEATURE: support mp4 torrents (?)

// TODO: properly delete series, torrents, episodes, gc conversions
// TODO: MAKE ERRORS MORE INFORMATIVE :/

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
		service.EpisodeServiceExport,

		// rest controllers
		controller.HealthControllerExport,
		controller.SeriesControllerExport,
		controller.ThumbnailControllerExport,
		controller.TorrentControllerExport,
		controller.ConvertControllerExport,
		controller.ProbeControllerExport,
		controller.EpisodeControllerExport,

		// misc
		analyze.TextAnalyzerExport,
		analyze.ProbeAnalyzerExport,
	).Run()
}
