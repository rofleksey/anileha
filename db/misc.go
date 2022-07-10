package db

type TorrentWithProgress struct {
	Torrent
	BytesRead    int64
	BytesMissing int64
	Progress     float64
}
