package db

import (
	"anileha/util"
	"time"
)

// Series Represents one season of something
type Series struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Query     *string // Query to automatically add torrents to this series
	ThumbID   *uint
	Thumb     *Thumb `gorm:"references:ID"`
}

func NewSeries(name string, query *string, thumbId *uint) Series {
	return Series{
		Name:    name,
		Query:   query,
		ThumbID: thumbId,
	}
}

// Thumb Represents unique thumbnail image
type Thumb struct {
	ID          uint `gorm:"primarykey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string
	Path        string
	DownloadUrl string
}

func NewThumb(name string, path string, downloadUrl string) Thumb {
	return Thumb{
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

// Torrent Represents info about torrent (e.g. name, files)
type Torrent struct {
	ID                  uint `gorm:"primarykey"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	util.Progress       `gorm:"embedded"`
	SeriesId            uint
	FilePath            string // FilePath path to .torrent file
	Name                string
	TotalLength         uint // TotalLength total size of ALL torrent files in bytes
	TotalDownloadLength uint // TotalDownloadLength total size of SELECTED torrent files in bytes
	BytesRead           uint
	Status              TorrentStatus
	Auto                bool
	Source              *string       // Source link to torrent url in case it was added automatically via query
	Files               []TorrentFile `gorm:"foreignKey:TorrentId"`
}

func NewTorrent(seriesId uint, filePath string, autoConvert bool) Torrent {
	return Torrent{
		SeriesId: seriesId,
		Status:   TORRENT_CREATING,
		FilePath: filePath,
		Auto:     autoConvert,
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
	ID           uint `gorm:"primarykey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	TorrentId    uint
	TorrentIndex int     // TorrentIndex file index according to .torrent file system
	TorrentPath  string  // TorrentPath file path according to .torrent file system
	ReadyPath    *string // ReadyPath file location after successful download
	Length       uint    // Length in bytes
	Season       string
	Episode      string
	EpisodeIndex int // EpisodeIndex file index according season/episode ordering
	Selected     bool
	Status       TorrentFileStatus
}

func NewTorrentFile(
	torrentId uint,
	torrentIndex int,
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

func NewConversion(
	seriesId uint,
	torrentid uint,
	torrentFileId uint,
	name string,
	episodeName string,
	outputDir string,
	videoPath string,
	logPath string,
	command string,
	videoDurationSec int,
) Conversion {
	return Conversion{
		SeriesId:         seriesId,
		TorrentId:        torrentid,
		TorrentFileId:    torrentFileId,
		Name:             name,
		EpisodeName:      episodeName,
		OutputDir:        outputDir,
		VideoPath:        videoPath,
		LogPath:          logPath,
		Command:          command,
		VideoDurationSec: videoDurationSec,
		Status:           CONVERSION_CREATED,
	}
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
	Thumb        *Thumb `gorm:"references:ID"`
	Length       uint64 // Length in bytes
	DurationSec  int    // Duration in seconds
	Path         string
	Url          string
}

func NewEpisode(seriesId uint, conversionId uint, name string, thumbId *uint, length uint64, durationSec int, path string, url string) Episode {
	return Episode{
		SeriesId:     seriesId,
		ConversionId: conversionId,
		Name:         name,
		ThumbID:      thumbId,
		Length:       length,
		DurationSec:  durationSec,
		Path:         path,
		Url:          url,
	}
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

func NewUser(login string, hash string, email string) User {
	return User{
		Login: login,
		Hash:  hash,
		Email: email,
	}
}
