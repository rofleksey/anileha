package db

import (
	"anileha/util"
	"gorm.io/gorm"
)

// Series Represents one season of something
type Series struct {
	gorm.Model
	Name        string
	Description string
	Query       *string // Query to automatically add torrents to this series
	ThumbnailID *uint
	Thumbnail   *Thumbnail `gorm:"references:ID"`
}

func NewSeries(name string, description string, query *string, thumbnailId *uint) Series {
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
	TORRENT_CREATING    TorrentStatus = "creating"
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

// Torrent Represents info about torrent (e.g. name, files)
type Torrent struct {
	gorm.Model
	util.Progress       `gorm:"embedded"`
	SeriesId            uint
	FilePath            string // FilePath path to .torrent file
	Name                string
	TotalLength         uint // TotalLength total size of ALL torrent files in bytes
	TotalDownloadLength uint // TotalDownloadLength total size of SELECTED torrent files in bytes
	BytesRead           uint
	Status              TorrentStatus
	Source              *string       // Source link to torrent url in case it was added automatically via query
	Files               []TorrentFile `gorm:"foreignKey:TorrentId"`
}

func NewTorrent(seriesId uint, filePath string) Torrent {
	return Torrent{
		SeriesId: seriesId,
		Status:   TORRENT_CREATING,
		FilePath: filePath,
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
	TorrentId    uint
	TorrentIndex uint    // TorrentIndex file index according to .torrent file system
	TorrentPath  string  // TorrentPath file path according to .torrent file system
	ReadyPath    *string // ReadyPath file location after successful download
	Length       uint    // Length in bytes
	Season       string
	Episode      string
	EpisodeIndex uint // EpisodeIndex file index according season/episode ordering
	Selected     bool
	Status       TorrentFileStatus
}

func NewTorrentFile(
	torrentId uint,
	torrentIndex uint,
	torrentPath string,
	selected bool,
	len uint,
) TorrentFile {
	return TorrentFile{
		TorrentId:    torrentId,
		TorrentIndex: torrentIndex,
		TorrentPath:  torrentPath,
		Selected:     selected,
		Status:       TORRENT_FILE_IDLE,
		Length:       len,
	}
}

type ConversionStatus string

const (
	CONVERSION_CREATED    ConversionStatus = "created"
	CONVERSION_PROCESSING ConversionStatus = "processing"
	CONVERSION_ERROR      ConversionStatus = "error"
	CONVERSION_CANCELLED  ConversionStatus = "cancelled"
	CONVERSION_READY      ConversionStatus = "ready"
)

// Conversion Represents info about a single attempt to convert TorrentFile to Episode
type Conversion struct {
	gorm.Model
	util.Progress    `gorm:"embedded"`
	SeriesId         uint
	TorrentFileId    uint
	EpisodeId        *uint
	Name             string
	EpisodeName      string
	OutputPath       string
	LogsPath         string
	Command          string
	VideoDurationSec uint64
	Status           ConversionStatus
}

func NewConversion(seriesId uint, torrentFileId uint, name string, episodeName string, outputPath string, logsPath string, command string, videoDurationSec uint64) Conversion {
	return Conversion{
		SeriesId:         seriesId,
		TorrentFileId:    torrentFileId,
		Name:             name,
		EpisodeName:      episodeName,
		OutputPath:       outputPath,
		LogsPath:         logsPath,
		Command:          command,
		VideoDurationSec: videoDurationSec,
		Status:           CONVERSION_CREATED,
	}
}

// Episode Represents info about a single ready-to-watch episode
type Episode struct {
	gorm.Model
	SeriesId     uint
	ConversionId uint
	Name         string
	ThumbnailID  *uint
	Thumbnail    *Thumbnail `gorm:"references:ID"`
	Length       uint64     // Length in bytes
	DurationSec  uint64     // Duration in seconds
	Path         string
	Url          string
}

func NewEpisode(seriesId uint, conversionId uint, name string, thumbnailId *uint, length uint64, durationSec uint64, path string, url string) Episode {
	return Episode{
		SeriesId:     seriesId,
		ConversionId: conversionId,
		Name:         name,
		ThumbnailID:  thumbnailId,
		Length:       length,
		DurationSec:  durationSec,
		Path:         path,
		Url:          url,
	}
}
