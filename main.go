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

// TODO: MAKE ERRORS MORE INFORMATIVE :/
// TODO: rate limit
// TODO: improve logging
// TODO: delete ready torrent directory

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
		service.ThumbServiceExport,
		service.TorrentServiceExport,
		service.ConversionServiceExport,
		service.EpisodeServiceExport,
		service.PipelineServiceExport,
		service.UserServiceExport,

		// rest controllers
		controller.HealthControllerExport,
		controller.SeriesControllerExport,
		controller.ThumbControllerExport,
		controller.TorrentControllerExport,
		controller.ConvertControllerExport,
		controller.ProbeControllerExport,
		controller.EpisodeControllerExport,
		controller.UserControllerExport,

		// misc
		analyze.TextAnalyzerExport,
		analyze.ProbeAnalyzerExport,
	).Run()
}
