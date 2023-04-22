package main

import (
	"anileha/config"
	"anileha/db"
	"anileha/db/repo"
	"anileha/ffmpeg/analyze"
	"anileha/ffmpeg/command"
	"anileha/rest/controller"
	"anileha/rest/engine"
	"anileha/search/nyaa"
	"anileha/service"
	"anileha/util/logger"
	"go.uber.org/fx"
)

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
		engine.Export,
		db.ServiceExport,

		// repositories
		repo.SeriesExport,
		repo.UserExport,
		repo.TorrentExport,
		repo.ConversionExport,
		repo.EpisodeExport,

		// search
		nyaa.Export,

		// services
		service.FileExport,
		service.HealthExport,
		service.SeriesExport,
		service.ThumbExport,
		service.TorrentExport,
		service.ConversionExport,
		service.EpisodeExport,
		service.UserExport,
		service.RoomExport,

		// rest controllers
		controller.HealthExport,
		controller.SeriesExport,
		controller.TorrentExport,
		controller.ConvertExport,
		controller.ProbeExport,
		controller.EpisodeExport,
		controller.UserExport,
		controller.WebsocketExport,

		// misc
		analyze.ProbeAnalyzerExport,
		command.ProducerExport,
	).Run()
}
