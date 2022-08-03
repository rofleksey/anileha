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
// TODO: replace all model pointers to objects (where possible)
// TODO: replace all db full updates with partial updates
// TODO: properly recover from all errors
// TODO: check there is no & inside loops for FOR loop variables
// TODO: replace uint with int where possible
// TODO: maximum 2-3 decimal places everywhere
// TODO: check there's is no .model.updates with null values in them

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

		// misc
		analyze.TextAnalyzerExport,
		analyze.ProbeAnalyzerExport,
	).Run()
}
