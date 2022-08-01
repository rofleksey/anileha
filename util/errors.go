package util

import "errors"

var ErrNotFound = errors.New("not found")
var ErrCreationFailed = errors.New("creation failed")
var ErrInvalidIndicesString = errors.New("invalid indices string")
var ErrDeleteStartedTorrent = errors.New("can't delete started torrent, should stop it first")
var ErrFileMapping = errors.New("file mapping failed")
var ErrVideoStreamNotFound = errors.New("video stream not found")
var ErrMoreThanOneVideoStream = errors.New("found more than 1 video stream")
var ErrUnknownByteLengthStr = errors.New("unknown byte length identifier")
var ErrNoValidStreamsFound = errors.New("no valid streams found")
var ErrAmbiguousSelection = errors.New("ambiguous selection")
var ErrFileIsNotReadyToBeConverted = errors.New("file is not ready to be converted")
var ErrQueueParallelismInvalid = errors.New("queue parallelism is invalid")
var ErrCancelled = errors.New("cancelled")
var ErrUnsupportedSubs = errors.New("unsupported subs")
var ErrInvalidStreamSize = errors.New("invalid stream size")
var ErrFileStateIsCorrupted = errors.New("file state is corrupted")
var ErrCTorrentNotFound = errors.New("ctorrent not found")
var ErrCTorrentCorrupted = errors.New("corrupted ctorrent")
var ErrAlreadyStarted = errors.New("already started")
var ErrAlreadyStopped = errors.New("already stopped")
