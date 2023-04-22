package dao

import (
	"anileha/db"
	"anileha/meta"
	"anileha/util"
	"time"
)

type SeriesResponseDao struct {
	ID         uint      `json:"id"`
	Title      string    `json:"title"`
	Thumb      string    `json:"thumb"`
	LastUpdate time.Time `json:"lastUpdate"`
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
	UpdatedAt           time.Time        `json:"updatedAt"`
}

type TorrentFileResponseDao struct {
	Path              string               `json:"path"`
	Status            db.TorrentFileStatus `json:"status"`
	Selected          bool                 `json:"selected"`
	Length            uint                 `json:"length"`
	ClientIndex       int                  `json:"clientIndex"`
	Type              util.FileType        `json:"type"`
	SuggestedMetadata meta.EpisodeMetadata `json:"suggestedMetadata"`
	Analysis          *db.AnalysisResult   `json:"analysis"`
}

type ConversionResponseDao struct {
	ID            uint                `json:"id"`
	SeriesId      *uint               `json:"seriesId"`
	TorrentId     *uint               `json:"torrentId"`
	TorrentFileId *uint               `json:"torrentFileId"`
	EpisodeId     *uint               `json:"episodeId"`
	EpisodeName   string              `json:"episodeName"`
	Name          string              `json:"name"`
	Command       string              `json:"command"`
	Status        db.ConversionStatus `json:"status"`
	Progress      util.Progress       `json:"progress"`
	UpdatedAt     time.Time           `json:"updatedAt"`
}

type EpisodeResponseDao struct {
	ID           uint      `json:"id"`
	SeriesId     *uint     `json:"seriesId"`
	ConversionId uint      `json:"conversionId"`
	Title        string    `json:"title"`
	Episode      string    `json:"episode"`
	Season       string    `json:"season"`
	CreatedAt    time.Time `json:"createdAt"`
	Thumb        string    `json:"thumb"`
	Length       uint64    `json:"length"`
	DurationSec  int       `json:"durationSec"`
	Url          string    `json:"link"`
}

type UserResponseDao struct {
	ID    uint     `json:"id"`
	Login string   `json:"login"`
	Name  string   `json:"name"`
	Email string   `json:"email"`
	Roles []string `json:"roles"`
	Thumb string   `json:"thumb"`
}
