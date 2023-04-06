package db

import (
	"anileha/util"
	"time"
)

// Series Represents one season of something
type Series struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	LastUpdate time.Time
	Name       string
	Query      *string // Query to automatically add torrents to this series
	ThumbID    *uint   `gorm:"unique"`
	Thumb      *Thumb  `gorm:"foreignKey:ID;references:thumb_id"`
}

// Thumb Represents unique thumbnail image
type Thumb struct {
	ID          uint `gorm:"primarykey"`
	Path        string
	DownloadUrl string
}

type TorrentStatus string

const (
	TorrentCreating    TorrentStatus = "creating"
	TorrentIdle        TorrentStatus = "idle"
	TorrentDownloading TorrentStatus = "download"
	TorrentError       TorrentStatus = "error"
	TorrentReady       TorrentStatus = "ready"
)

type Torrent struct {
	ID                  uint `gorm:"primarykey"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	SeriesId            uint
	FilePath            string // FilePath path to .torrent file
	Name                string
	BytesRead           uint
	TotalLength         uint
	TotalDownloadLength uint
	util.Progress       `gorm:"embedded"`
	Status              TorrentStatus
	Source              *string       // Source link to torrent url in case it was added automatically via query
	Files               []TorrentFile `gorm:"foreignKey:torrent_id"`
}

type TorrentFileStatus string

const (
	TorrentFileIdle        TorrentFileStatus = "idle"
	TorrentFileDownloading TorrentFileStatus = "download"
	TorrentFileError       TorrentFileStatus = "error"
	TorrentFileReady       TorrentFileStatus = "ready"
)

// TorrentFile Represents info about a single torrent file
type TorrentFile struct {
	ID           uint `gorm:"primarykey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	TorrentId    uint
	TorrentIndex int     // TorrentIndex file index according to .torrent file system
	TorrentPath  string  // TorrentPath file path according to .torrent file system
	ClientIndex  int     // ClientIndex sorted by name
	ReadyPath    *string // ReadyPath file location after successful download
	Length       uint    // Length in bytes
	Selected     bool
	Status       TorrentFileStatus
}

type ConversionStatus string

const (
	ConversionCreated    ConversionStatus = "created"
	ConversionProcessing ConversionStatus = "processing"
	ConversionError      ConversionStatus = "error"
	ConversionCancelled  ConversionStatus = "cancelled"
	ConversionReady      ConversionStatus = "ready"
)

// Conversion Represents info about a single attempt to convert TorrentFile to Episode
type Conversion struct {
	ID               uint `gorm:"primarykey"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	util.Progress    `gorm:"embedded"`
	SeriesId         uint
	TorrentId        uint
	TorrentFileId    uint
	EpisodeId        *uint
	Name             string
	EpisodeName      string
	OutputDir        string
	VideoPath        string
	LogPath          string
	Command          string
	VideoDurationSec int
	Status           ConversionStatus
}

// Episode Represents info about a single ready-to-watch episode
type Episode struct {
	ID           uint `gorm:"primarykey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	SeriesId     uint
	ConversionId uint
	Name         string
	ThumbID      *uint
	Thumb        *Thumb `gorm:"foreignKey:ID;references:thumb_id"`
	Length       uint64 // Length in bytes
	DurationSec  int    // Duration in seconds
	Path         string
	Url          string
}

type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Login     string `gorm:"uniqueIndex"`
	Hash      string
	Email     string `gorm:"uniqueIndex"`
	Admin     bool
}
