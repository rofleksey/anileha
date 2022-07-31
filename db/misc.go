package db

type TorrentWithProgress struct {
	Torrent
	BytesRead    int64
	BytesMissing int64
	Progress     float64
}

type Progress struct {
	Progress float64
	Elapsed  float64
	Eta      float64
}

type AnyChannel chan interface{}
