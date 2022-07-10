package db

import "github.com/jinzhu/gorm"

// TODO: add cascade operations

// Series Represents one season of something
type Series struct {
	gorm.Model
	Name        string
	Description string
	Query       *string // Query to automatically add torrents to this series
	ThumbnailID uint
	Thumbnail   Thumbnail           `gorm:"references:ID"`
	Torrents    []Torrent           `gorm:"foreignKey:SeriesId"`
	Conversions []EpisodeConversion `gorm:"foreignKey:SeriesId"`
	Episodes    []Episode           `gorm:"foreignKey:SeriesId"`
}

func NewSeries(name string, description string, query *string, thumbnailId uint) *Series {
	return &Series{
		Name:        name,
		Description: description,
		Query:       query,
		ThumbnailID: thumbnailId,
	}
}

// Thumbnail Represents unique thumbnail image
type Thumbnail struct {
	gorm.Model
	Name        string
	Path        string
	DownloadUrl string
}

func NewThumbnail(name string, path string, downloadUrl string) *Thumbnail {
	return &Thumbnail{
		Name:        name,
		Path:        path,
		DownloadUrl: downloadUrl,
	}
}

type TorrentStatus string

const (
	TORRENT_IDLE        TorrentStatus = "idle"
	TORRENT_DOWNLOADING TorrentStatus = "downloading"
	TORRENT_ERROR       TorrentStatus = "error"
	TORRENT_READY       TorrentStatus = "ready"
)

// Torrent Represents info about torrent (e.g. files)
type Torrent struct {
	gorm.Model
	SeriesId uint
	Name     string
	Status   TorrentStatus
	Path     string
	Files    []TorrentFile `gorm:"foreignKey:TorrentId"`
}

type TorrentFileStatus string

// TorrentFile Represents info about a single torrent file
type TorrentFile struct {
	gorm.Model
	TorrentId    uint
	Index        uint    // Index file index according to .torrent file system
	TorrentPath  string  // TorrentPath file path according to .torrent file system
	DownloadPath *string // DownloadPath file location during torrent download
	ReadyPath    *string // ReadyPath file location after successful download
	State        TorrentStatus
	Size         uint  // Size in bytes
	Probe        Probe `gorm:"foreignKey:TorrentFileId"`
}

// Probe contains JSON of file probe
type Probe struct {
	gorm.Model
	TorrentFileId uint
	Content       string
}

type ConversionStatus string

const (
	CONVERSION_CREATED     ConversionStatus = "created"
	CONVERSION_DOWNLOADING ConversionStatus = "processing"
	CONVERSION_ERROR       ConversionStatus = "error"
	CONVERSION_READY       ConversionStatus = "ready"
)

// EpisodeConversion Represents info about a single attempt to convert TorrentFile to Episode
type EpisodeConversion struct {
	gorm.Model
	SeriesId      uint
	TorrentId     uint
	TorrentFileId uint
	EpisodeId     *uint
	Name          string
	LogPath       *string
	OutputPath    *string
	Commands      string
	Status        ConversionStatus
}

// Episode Represents info about a single ready-to-watch episode
type Episode struct {
	gorm.Model
	SeriesId     uint
	ConversionId uint
	Name         string
	ThumbnailID  uint
	Thumbnail    Thumbnail
	Size         uint // Size in bytes
	Duration     uint // Duration in ms
	Path         string
}
