package analyze

import (
	"anileha/ffmpeg"
	"anileha/util"
	"context"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/vansante/go-ffprobe.v2"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ProbeAnalyzer Selects streams and generates FFmpeg command for encoding
// For video: Use the last video stream
// For audio: select the only stream, else selects the only japanese stream, else selects the most heavy one (among japanese or all if not present)
// For subs: select the only stream, else selects the only english stream, else selects the stream with most occurrences of english words (among english or all if not present)
type ProbeAnalyzer struct {
	textAnalyzer *TextAnalyzer
	regexMap     map[StreamType]*regexp.Regexp
	log          *zap.Logger
}

func NewProbeAnalyzer(textAnalyzer *TextAnalyzer, log *zap.Logger) ProbeAnalyzer {
	videoRegex := regexp.MustCompile("video:(\\d+)([a-z]+)")
	audioRegex := regexp.MustCompile("audio:(\\d+)([a-z]+)")
	subRegex := regexp.MustCompile("subtitle:(\\d+)([a-z]+)")
	regexMap := make(map[StreamType]*regexp.Regexp)
	regexMap[Video] = videoRegex
	regexMap[Audio] = audioRegex
	regexMap[Sub] = subRegex
	return ProbeAnalyzer{
		textAnalyzer: textAnalyzer,
		regexMap:     regexMap,
		log:          log,
	}
}

// parseStreamSize returns stream size in bytes
func (p *ProbeAnalyzer) parseStreamSize(sizeCommandResult string, streamType StreamType) (uint64, error) {
	trimmed := strings.TrimSpace(sizeCommandResult)
	splitArr := strings.Split(trimmed, "\n")
	lastLine := splitArr[len(splitArr)-1]
	lowerCase := strings.ToLower(lastLine)
	reg := p.regexMap[streamType]
	matchArr := reg.FindStringSubmatch(lowerCase)
	number, err := strconv.ParseUint(matchArr[0], 10, 64)
	if err != nil {
		return 0, err
	}
	var multiplier uint64
	switch matchArr[1] {
	case "b":
		multiplier = 1
	case "byte":
		multiplier = 1
	case "bytes":
		multiplier = 1
	case "kb":
		multiplier = 1024
	case "mb":
		multiplier = 1024 * 1024
	case "gb":
		multiplier = 1024 * 1024 * 1024
	default:
		p.log.Error("parsing stream size failed", zap.String("multiplier", matchArr[1]))
		return 0, util.ErrUnknownByteLengthStr
	}
	return multiplier * number, nil
}

// GetStreamSize get size of stream
// Executes: ffmpeg -i <inputFile> -map 0:a:<streamIndex> -c copy -f null -
// Last ffmpeg line: video:0kB audio:18619kB subtitle:0kB other streams:0kB global headers:0kB muxing overhead: unknown
func (p *ProbeAnalyzer) GetStreamSize(inputFile string, streamType StreamType, streamIndex int) (uint64, error) {
	p.log.Info("getting stream size", zap.String("inputFile", inputFile), zap.String("streamType", string(streamType)), zap.Int("relativeIndex", streamIndex))
	sizeCommand := ffmpeg.NewCommand(inputFile, 0, "-")
	streamLetter := streamType[0:1]
	mapValue := fmt.Sprintf("0:%s:%d", streamLetter, streamIndex)
	sizeCommand.AddEscapedKeyValue("-map", mapValue, ffmpeg.OptionInput)
	sizeCommand.AddKeyValue("-c", "copy", ffmpeg.OptionOutput)
	sizeCommand.AddKeyValue("-f", "null", ffmpeg.OptionOutput)
	result, err := sizeCommand.ExecuteSync()
	if err != nil {
		return 0, nil
	}
	size, err := p.parseStreamSize(*result, streamType)
	if err != nil {
		return 0, nil
	}
	return size, nil
}

// ExtractSubText gets sub stream text
func (p *ProbeAnalyzer) ExtractSubText(inputFile string, streamIndex int) (string, error) {
	p.log.Info("extracting stream subtitle text", zap.String("inputFile", inputFile), zap.Int("relativeIndex", streamIndex))
	srtFileName := inputFile + ".srt"
	defer func() {
		_ = os.Remove(srtFileName)
	}()
	sizeCommand := ffmpeg.NewCommand(inputFile, 0, srtFileName)
	mapValue := fmt.Sprintf("0:s:%d", streamIndex)
	sizeCommand.AddEscapedKeyValue("-map", mapValue, ffmpeg.OptionInput)
	sizeCommand.AddKeyValue("-f", "srt", ffmpeg.OptionOutput)
	_, err := sizeCommand.ExecuteSync()
	if err != nil {
		return "", nil
	}
	content, err := ioutil.ReadFile(srtFileName)
	if err != nil {
		return "", nil
	}
	return string(content), nil
}

func (p *ProbeAnalyzer) getScoreResult(inputFile string) (*ScoreResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	probe, err := ffprobe.ProbeURL(ctx, inputFile)
	if err != nil {
		return nil, err
	}
	var videoStream *StreamWithScore
	audioStreams := make([]StreamWithScore, 0, 10)
	subStreams := make([]StreamWithScore, 0, 10)
	for _, stream := range probe.Streams {
		switch stream.CodecType {
		case "video":
			if videoStream != nil {
				return nil, util.ErrMoreThanOneVideoStream
			}
			videoStream = &StreamWithScore{
				Stream:        stream,
				RelativeIndex: 0,
			}
		case "audio":
			audioStreams = append(audioStreams, StreamWithScore{
				Stream:        stream,
				RelativeIndex: len(audioStreams),
			})
		case "subtitle":
			subStreams = append(subStreams, StreamWithScore{
				Stream:        stream,
				RelativeIndex: len(subStreams),
			})
		}
	}

	if videoStream == nil {
		return nil, util.ErrVideoStreamNotFound
	}

	if len(audioStreams) == 1 && len(subStreams) == 1 {
		p.log.Info("has single audio and sub stream", zap.String("inputFile", inputFile))
		return &ScoreResult{
			Video:           *videoStream,
			AudioCandidates: audioStreams,
			SubCandidates:   subStreams,
		}, nil
	}

	japaneseAudio := make([]StreamWithScore, 0, len(audioStreams))
	for _, stream := range audioStreams {
		lang, err := stream.Stream.TagList.GetString("language")
		if err != nil && lang == "jpn" {
			japaneseAudio = append(japaneseAudio, stream)
		}
	}

	englishSubs := make([]StreamWithScore, 0, len(subStreams))
	for _, stream := range subStreams {
		lang, err := stream.Stream.TagList.GetString("language")
		if err != nil && lang == "eng" {
			englishSubs = append(englishSubs, stream)
		}
	}

	if len(englishSubs) == 1 && len(japaneseAudio) == 1 {
		p.log.Info("has exactly 1 eng sub and jpn audio", zap.String("inputFile", inputFile))
		return &ScoreResult{
			Video:           *videoStream,
			AudioCandidates: japaneseAudio,
			SubCandidates:   englishSubs,
		}, nil
	}

	var remainingAudio []StreamWithScore
	var remainingSubs []StreamWithScore

	if len(japaneseAudio) == 0 {
		remainingAudio = audioStreams
	} else {
		remainingAudio = japaneseAudio
	}

	if len(englishSubs) == 0 {
		remainingSubs = subStreams
	} else {
		remainingSubs = englishSubs
	}

	for _, audioStream := range remainingAudio {
		size, err := p.GetStreamSize(inputFile, Audio, audioStream.RelativeIndex)
		if err != nil {
			p.log.Error("failed to get stream size", zap.String("streamType", string(Audio)), zap.Int("relativeIndex", audioStream.RelativeIndex), zap.Error(err))
			continue
		}
		p.log.Info("got audio stream size", zap.String("inputFile", inputFile), zap.Int("relativeIndex", audioStream.RelativeIndex), zap.Uint64("size", size))
		audioStream.Score = size
	}
	sort.SliceStable(remainingAudio, func(i, j int) bool {
		return remainingAudio[i].Score > remainingAudio[j].Score
	})

	for _, subStream := range remainingSubs {
		text, err := p.ExtractSubText(inputFile, subStream.RelativeIndex)
		if err != nil {
			p.log.Error("failed to get subtitle text", zap.Int("relativeIndex", subStream.RelativeIndex), zap.Error(err))
			continue
		}
		numberOfEngWords := p.textAnalyzer.CountEnglishWords(text)
		subStream.Score = numberOfEngWords
		p.log.Info("got number of eng words in sub stream", zap.String("inputFile", inputFile), zap.Int("relativeIndex", subStream.RelativeIndex), zap.Uint64("wordCount", numberOfEngWords))
	}
	sort.SliceStable(remainingSubs, func(i, j int) bool {
		return remainingSubs[i].Score > remainingSubs[j].Score
	})

	p.log.Warn("has ambiguous selection of subs/audio", zap.String("inputFile", inputFile))
	return &ScoreResult{
		Ambiguous:       true,
		Video:           *videoStream,
		AudioCandidates: remainingAudio,
		SubCandidates:   remainingSubs,
	}, nil
}

var ProbeAnalyzerExport = fx.Options(fx.Provide(NewProbeAnalyzer))
