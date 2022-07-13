package db

type TorrentWithProgress struct {
	Torrent
	BytesRead    int64
	BytesMissing int64
	Progress     float64
}

type Progress struct {
	Progress    float64
	TimeElapsed float64
	Eta         float64
}

type FinishChan chan error
type ProgressChan chan Progress
