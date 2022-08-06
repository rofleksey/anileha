package dao

import (
	"anileha/db"
	"anileha/util"
)

type SeriesResponseDao struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Query       *string `json:"query"`
	Thumb       string  `json:"thumb"`
}

type TorrentResponseDao struct {
	ID                  uint                     `json:"id"`
	Name                string                   `json:"name"`
	Status              db.TorrentStatus         `json:"status"`
	Source              *string                  `json:"source"`
	TotalLength         uint                     `json:"totalLength"`
	TotalDownloadLength uint                     `json:"totalDownloadLength"`
	Progress            util.Progress            `json:"progress"`
	BytesRead           uint                     `json:"bytesRead"`
	Auto                bool                     `json:"auto"`
	Files               []TorrentFileResponseDao `json:"files"`
}

type TorrentFileResponseDao struct {
	Path         string               `json:"path"`
	Status       db.TorrentFileStatus `json:"status"`
	Selected     bool                 `json:"selected"`
	Length       uint                 `json:"length"`
	Episode      string               `json:"episode"`
	EpisodeIndex uint                 `json:"episodeIndex"`
	Season       string               `json:"season"`
}

type ConversionResponseDao struct {
	ID            uint                `json:"id"`
	SeriesId      uint                `json:"seriesId"`
	TorrentFileId uint                `json:"torrentFileId"`
	EpisodeId     *uint               `json:"episodeId"`
	EpisodeName   string              `json:"episodeName"`
	Name          string              `json:"name"`
	FFmpegCommand string              `json:"ffmpegCommand"`
	Status        db.ConversionStatus `json:"status"`
	Progress      util.Progress       `json:"progress"`
}

type EpisodeResponseDao struct {
	ID           uint    `json:"id"`
	ConversionId uint    `json:"conversionId"`
	Name         string  `json:"name"`
	Thumb        *string `json:"thumb"`
	Length       uint64  `json:"length"`
	DurationSec  uint64  `json:"durationSec"`
	Url          string  `json:"link"`
}
