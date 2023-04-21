package db

import (
	"anileha/util"
	"gorm.io/datatypes"
	"time"
)

// Series Represents one season of something
type Series struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	LastUpdate time.Time
	Title      string
	Thumb      Thumb `gorm:"embedded"`
}

type TorrentStatus string

const (
	TorrentCreating TorrentStatus = "creating"
	TorrentIdle     TorrentStatus = "idle"
	TorrentDownload TorrentStatus = "download"
	TorrentAnalysis TorrentStatus = "analysis"
	TorrentError    TorrentStatus = "error"
	TorrentReady    TorrentStatus = "ready"
)

type Torrent struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	SeriesId *uint
	Series   *Series `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	Auto                datatypes.JSONType[*AutoTorrent]
	FilePath            string // FilePath path to .torrent file
	Name                string
	BytesRead           uint
	TotalLength         uint
	TotalDownloadLength uint
	util.Progress       `gorm:"embedded"`
	Status              TorrentStatus
	Source              *string       // Source link to torrent url in case it was added automatically via query
	Files               []TorrentFile `gorm:"foreignKey:torrent_id;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type TorrentFileStatus string

const (
	TorrentFileIdle     TorrentFileStatus = "idle"
	TorrentFileDownload TorrentFileStatus = "download"
	TorrentFileAnalysis TorrentFileStatus = "analysis"
	TorrentFileError    TorrentFileStatus = "error"
	TorrentFileReady    TorrentFileStatus = "ready"
)

// TorrentFile Represents info about a single torrent file
type TorrentFile struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	TorrentId uint
	Torrent   Torrent `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	TorrentIndex      int     // TorrentIndex file index according to .torrent file system
	TorrentPath       string  // TorrentPath file path according to .torrent file system
	ClientIndex       int     // ClientIndex sorted by name
	ReadyPath         *string // ReadyPath file location after successful download
	Length            uint    // Length in bytes
	Selected          bool
	Status            TorrentFileStatus
	Type              util.FileType
	SuggestedMetadata datatypes.JSONType[EpisodeMetadata]
	Analysis          datatypes.JSONType[*AnalysisResult]
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
	ID            uint `gorm:"primarykey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	util.Progress `gorm:"embedded"`

	SeriesId      *uint
	Series        *Series `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	TorrentId     *uint
	Torrent       *Torrent `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	TorrentFileId *uint
	TorrentFile   *TorrentFile `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	EpisodeId     *uint
	Episode       *Episode `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	Name             string
	EpisodeName      string
	EpisodeString    string
	SeasonString     string
	OutputDir        string
	VideoPath        string
	LogPath          string
	Command          string
	VideoDurationSec int
	Status           ConversionStatus
}

// Episode Represents info about a single ready-to-watch episode
type Episode struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time

	SeriesId *uint
	Series   *Series `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`

	Title       string
	Episode     string
	Season      string
	Thumb       Thumb  `gorm:"embedded"`
	Length      uint64 // Length in bytes
	DurationSec int    // Duration in seconds
	Path        string
	Url         string
}

type User struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Login     string `gorm:"uniqueIndex"`
	Name      string
	Hash      string
	Email     string `gorm:"uniqueIndex"`
	Roles     datatypes.JSONSlice[string]
	Thumb     Thumb `gorm:"embedded"`
}
