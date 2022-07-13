package analyze

import "gopkg.in/vansante/go-ffprobe.v2"

type SubsType string

const (
	Text    SubsType = "text"
	Picture SubsType = "picture"
)

type StreamType string

const (
	Video StreamType = "video"
	Audio StreamType = "audio"
	Sub   StreamType = "subtitle"
)

type StreamWithScore struct {
	*ffprobe.Stream
	RelativeIndex int
	Score         uint64
}

type ScoreResult struct {
	Ambiguous       bool
	Video           StreamWithScore
	AudioCandidates []StreamWithScore
	SubCandidates   []StreamWithScore
}

type ParsedStream struct {
	Index         int
	RelativeIndex int
	Language      *string
	Codec         string
	CodecFull     string
	Title         *string
}

type ParsedProbe struct {
	Width      float64
	Height     float64
	Fps        float64
	DurationMs uint
	Format     string
	Bitrate    string
	Video      ParsedStream
	Subs       []ParsedStream
	Audio      []ParsedStream
}
