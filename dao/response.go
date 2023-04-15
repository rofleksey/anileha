package dao

import (
	"anileha/db"
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
	Path        string               `json:"path"`
	Status      db.TorrentFileStatus `json:"status"`
	Selected    bool                 `json:"selected"`
	Length      uint                 `json:"length"`
	ClientIndex int                  `json:"clientIndex"`
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

type SubsType string

const (
	SubsText    SubsType = "text"
	SubsPicture SubsType = "picture"
	SubsUnknown SubsType = "unknown"
)

type BaseStream struct {
	RelativeIndex int    `json:"index"`
	Name          string `json:"name"`
	Size          uint64 `json:"size"`
	Lang          string `json:"lang"`
}

type VideoStream struct {
	BaseStream
	Width       int `json:"width"`
	Height      int `json:"height"`
	DurationSec int `json:"durationSec"`
}

type AudioStream struct {
	BaseStream
}

type SubStream struct {
	BaseStream
	Type       SubsType `json:"type"`
	TextLength int      `json:"textLength"`
}

type AnalysisResult struct {
	Video VideoStream   `json:"video"`
	Audio []AudioStream `json:"audio"`
	Sub   []SubStream   `json:"sub"`
}

type EpisodeMetadata struct {
	Episode string `json:"episode"`
	Season  string `json:"season"`
}

type AnalysisResponseDao struct {
	*AnalysisResult
	EpisodeMetadata
}
