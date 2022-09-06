package dao

import (
	"anileha/db"
	"anileha/util"
	"time"
)

type SeriesResponseDao struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Query     *string   `json:"query"`
	Thumb     string    `json:"thumb"`
	UpdatedAt time.Time `json:"updatedAt"`
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
	UpdatedAt           time.Time                `json:"updatedAt"`
}

type TorrentResponseWithoutFilesDao struct {
	ID                  uint             `json:"id"`
	Name                string           `json:"name"`
	Status              db.TorrentStatus `json:"status"`
	Source              *string          `json:"source"`
	TotalLength         uint             `json:"totalLength"`
	TotalDownloadLength uint             `json:"totalDownloadLength"`
	Progress            util.Progress    `json:"progress"`
	BytesRead           uint             `json:"bytesRead"`
	Auto                bool             `json:"auto"`
	UpdatedAt           time.Time        `json:"updatedAt"`
}

type TorrentFileResponseDao struct {
	Path         string               `json:"path"`
	Status       db.TorrentFileStatus `json:"status"`
	Selected     bool                 `json:"selected"`
	Length       uint                 `json:"length"`
	Episode      string               `json:"episode"`
	EpisodeIndex int                  `json:"episodeIndex"`
	Season       string               `json:"season"`
}

type ConversionResponseDao struct {
	ID            uint                `json:"id"`
	SeriesId      uint                `json:"seriesId"`
	TorrentId     uint                `json:"torrentId"`
	TorrentFileId uint                `json:"torrentFileId"`
	EpisodeId     *uint               `json:"episodeId"`
	EpisodeName   string              `json:"episodeName"`
	Name          string              `json:"name"`
	FFmpegCommand string              `json:"ffmpegCommand"`
	Status        db.ConversionStatus `json:"status"`
	Progress      util.Progress       `json:"progress"`
	UpdatedAt     time.Time           `json:"updatedAt"`
}

type EpisodeResponseDao struct {
	ID           uint      `json:"id"`
	SeriesId     uint      `json:"seriesId"`
	ConversionId uint      `json:"conversionId"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"createdAt"`
	Thumb        *string   `json:"thumb"`
	Length       uint64    `json:"length"`
	DurationSec  int       `json:"durationSec"`
	Url          string    `json:"link"`
}
