package util

import (
	"errors"
)

var ErrFileMapping = errors.New("file mapping failed")
var ErrCancelled = errors.New("cancelled")
var ErrInvalidStreamSize = errors.New("invalid stream size")
var ErrUnknownByteLengthStr = errors.New("unknown byte length string")
var ErrMoreThanOneVideoStream = errors.New("found more than one video stream")
var ErrVideoStreamNotFound = errors.New("video stream not found")
var ErrUnsupportedSubs = errors.New("unsupported subs")
