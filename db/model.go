package db

import (
	"gorm.io/gorm"
)

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

func NewSeries(name string, description string, query *string, thumbnailId uint) Series {
	return Series{
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

func NewThumbnail(name string, path string, downloadUrl string) Thumbnail {
	return Thumbnail{
		Name:        name,
		Path:        path,
		DownloadUrl: downloadUrl,
	}
}

type TorrentStatus string

const (
	TORRENT_CREATING    TorrentStatus = "created"
	TORRENT_IDLE        TorrentStatus = "idle"
	TORRENT_DOWNLOADING TorrentStatus = "download" // torrentLib should only have torrents in this state
	TORRENT_ERROR       TorrentStatus = "error"
	TORRENT_READY       TorrentStatus = "ready"
)

type TorrentInfoType string

const (
	TORRENT_INFO_FILE   TorrentInfoType = "file"
	TORRENT_INFO_MAGNET TorrentInfoType = "magnet"
)

// Torrent Represents info about torrent (e.g. files)
type Torrent struct {
	gorm.Model
	SeriesId            uint
	InfoType            TorrentInfoType
	InfoPath            string // InfoPath see TorrentInfoType
	Name                string
	TotalLength         int64 // TotalLength total size of ALL torrent files in bytes
	TotalDownloadLength int64 // TotalDownloadLength total size of SELECTED torrent files in bytes
	Status              TorrentStatus
	Source              *string       // Source link to torrent url in case it was added automatically via query
	Files               []TorrentFile `gorm:"foreignKey:TorrentId"`
}

func NewTorrent(seriesId uint, infoPath string, infoType TorrentInfoType) Torrent {
	return Torrent{
		SeriesId: seriesId,
		Status:   TORRENT_CREATING,
		InfoPath: infoPath,
		InfoType: infoType,
	}
}

type TorrentFileStatus string

const (
	TORRENT_FILE_IDLE        TorrentFileStatus = "idle"
	TORRENT_FILE_DOWNLOADING TorrentFileStatus = "download"
	TORRENT_FILE_ERROR       TorrentFileStatus = "error"
	TORRENT_FILE_READY       TorrentFileStatus = "ready"
)

// TorrentFile Represents info about a single torrent file
type TorrentFile struct {
	gorm.Model
	TorrentId   uint
	Index       uint    // Index file index according to .torrent file system
	TorrentPath string  // TorrentPath file path according to .torrent file system
	ReadyPath   *string // ReadyPath file location after successful download
	Length      uint    // Length in bytes
	Selected    bool
	Status      TorrentFileStatus
}

func NewTorrentFile(torrentId uint, index uint, torrentPath string, selected bool, len uint) TorrentFile {
	return TorrentFile{
		TorrentId:   torrentId,
		Index:       index,
		TorrentPath: torrentPath,
		Selected:    selected,
		Status:      TORRENT_FILE_IDLE,
		Length:      len,
	}
}

type ConversionStatus string

const (
	CONVERSION_CREATED    ConversionStatus = "created"
	CONVERSION_PROCESSING ConversionStatus = "processing"
	CONVERSION_ERROR      ConversionStatus = "error"
	CONVERSION_READY      ConversionStatus = "ready"
)

// EpisodeConversion Represents info about a single attempt to convert TorrentFile to Episode
type EpisodeConversion struct {
	gorm.Model
	SeriesId      uint
	TorrentFileId uint
	EpisodeId     *uint
	Name          string
	OutputPath    string
	FFmpegCommand string
	Status        ConversionStatus
}

// Episode Represents info about a single ready-to-watch episode
type Episode struct {
	gorm.Model
	SeriesId     uint
	ConversionId *uint
	Name         string
	ThumbnailID  uint
	Thumbnail    Thumbnail `gorm:"references:ID"`
	Size         uint      // Size in bytes
	Duration     uint      // Duration in ms
	Path         string
}
