package util

import "errors"

var ErrNotFound = errors.New("not found")
var ErrCreationFailed = errors.New("creation failed")
var ErrInvalidInfoType = errors.New("invalid InfoType")
var ErrInvalidIndicesString = errors.New("invalid indices string")
var ErrDeleteStartedTorrent = errors.New("can't delete started torrent, should stop it first")
var ErrFileMapping = errors.New("file mapping failed")
var ErrVideoStreamNotFound = errors.New("video stream not found")
var ErrMoreThanOneVideoStream = errors.New("found more than 1 video stream")
var ErrUnknownByteLengthStr = errors.New("unknown byte length identifier")
