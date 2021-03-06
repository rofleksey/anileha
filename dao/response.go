package dao

import "anileha/db"

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
	TotalLength         int64                    `json:"totalLength"`
	TotalDownloadLength int64                    `json:"totalDownloadLength"`
	Progress            *float64                 `json:"progress"`
	BytesRead           *int64                   `json:"bytesRead"`
	BytesMissing        *int64                   `json:"bytesMissing"`
	Files               []TorrentFileResponseDao `json:"files"`
}

type TorrentFileResponseDao struct {
	Path     string               `json:"path"`
	Status   db.TorrentFileStatus `json:"status"`
	Selected bool                 `json:"selected"`
	Length   uint                 `json:"length"`
	Episode  string               `json:"episode"`
	Season   string               `json:"season"`
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
	Eta           float64             `json:"eta"`
	Progress      float64             `json:"progress"`
	Elapsed       float64             `json:"elapsed"`
}

type EpisodeResponseDao struct {
	ID           uint   `json:"id"`
	ConversionId uint   `json:"conversionId"`
	Name         string `json:"name"`
	ThumbnailId  *uint  `json:"thumbnailId"`
	Length       uint64 `json:"length"`
	DurationSec  uint64 `json:"durationSec"`
	Url          string `json:"link"`
}
