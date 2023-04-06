package analyze

import (
	"gopkg.in/vansante/go-ffprobe.v2"
	"regexp"
)

var videoRegex = regexp.MustCompile("video:(\\d+)([a-z]+)")
var audioRegex = regexp.MustCompile("audio:(\\d+)([a-z]+)")
var subRegex = regexp.MustCompile("subtitle:(\\d+)([a-z]+)")

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

type StreamWithIndex struct {
	*ffprobe.Stream
	RelativeIndex int
}

// result types

type BaseStream struct {
	RelativeIndex int
	Size          uint64
	Lang          string
}

type VideoStream struct {
	BaseStream
	Width       int
	Height      int
	DurationSec int
}

type AudioStream struct {
	BaseStream
}

type SubStream struct {
	BaseStream
	Type       SubsType
	TextLength int
}

type Result struct {
	Video VideoStream
	Audio []AudioStream
	Sub   []SubStream
}
