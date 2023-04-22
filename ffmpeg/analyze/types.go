package analyze

import (
	"gopkg.in/vansante/go-ffprobe.v2"
	"regexp"
)

var videoRegex = regexp.MustCompile("video:(\\d+)([a-z]+)")
var audioRegex = regexp.MustCompile("audio:(\\d+)([a-z]+)")
var subRegex = regexp.MustCompile("subtitle:(\\d+)([a-z]+)")

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
