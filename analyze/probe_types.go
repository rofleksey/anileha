package analyze

import "gopkg.in/vansante/go-ffprobe.v2"

type SubsType string

const (
	SubsText    SubsType = "text"
	SubsPicture SubsType = "picture"
	SubsUnknown SubsType = "unknown"
)

// intermediate types

type StreamType string

const (
	StreamVideo StreamType = "video"
	StreamAudio StreamType = "audio"
	StreamSub   StreamType = "subtitle"
)

type ParsedStream struct {
	Index         int
	RelativeIndex int
	Language      *string
	Codec         string
	CodecFull     string
	Title         *string
}

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

// result types

type ResultStream struct {
	RelativeIndex int
}

type VideoStream struct {
	ResultStream
	Width       int
	Height      int
	DurationSec uint64
}

type SubStream struct {
	ResultStream
	Type SubsType
}

type Result struct {
	Video VideoStream
	Audio *ResultStream
	Sub   *SubStream
}
