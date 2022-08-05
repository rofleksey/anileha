package analyze

import (
	"anileha/config"
	"anileha/ffmpeg"
	"anileha/util"
	"context"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/vansante/go-ffprobe.v2"
	"io/ioutil"
	"os"
	"os/exec"
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
	textAnalyzer  *TextAnalyzer
	regexMap      map[StreamType]*regexp.Regexp
	log           *zap.Logger
	prefAudioLang string
	prefSubLang   string
}

func NewProbeAnalyzer(
	textAnalyzer *TextAnalyzer,
	config *config.Config,
	log *zap.Logger,
) *ProbeAnalyzer {
	videoRegex := regexp.MustCompile("video:(\\d+)([a-z]+)")
	audioRegex := regexp.MustCompile("audio:(\\d+)([a-z]+)")
	subRegex := regexp.MustCompile("subtitle:(\\d+)([a-z]+)")
	regexMap := make(map[StreamType]*regexp.Regexp)
	regexMap[StreamVideo] = videoRegex
	regexMap[StreamAudio] = audioRegex
	regexMap[StreamSub] = subRegex
	return &ProbeAnalyzer{
		textAnalyzer:  textAnalyzer,
		regexMap:      regexMap,
		log:           log,
		prefAudioLang: config.Conversion.PrefAudioLang,
		prefSubLang:   config.Conversion.PrefSubLang,
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
	if matchArr == nil {
		return 0, util.ErrInvalidStreamSize
	}
	number, err := strconv.ParseUint(matchArr[1], 10, 64)
	if err != nil {
		return 0, err
	}
	var multiplier uint64
	switch matchArr[2] {
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
	sizeCommand.AddKeyValue("-map", mapValue, ffmpeg.OptionInput)
	sizeCommand.AddKeyValue("-analyzeduration", "2147483647", ffmpeg.OptionBase)
	sizeCommand.AddKeyValue("-probesize", "2147483647", ffmpeg.OptionBase)
	sizeCommand.AddKeyValue("-c", "copy", ffmpeg.OptionOutput)
	sizeCommand.AddKeyValue("-f", "null", ffmpeg.OptionOutput)
	result, err := sizeCommand.ExecuteSync()
	if err != nil {
		p.log.Error(fmt.Sprintf("failed to get stream size: %s", *result), zap.String("inputFile", inputFile), zap.String("streamType", string(streamType)), zap.Int("relativeIndex", streamIndex), zap.Error(err))
		return 0, err
	}
	size, err := p.parseStreamSize(*result, streamType)
	if err != nil {
		p.log.Error("failed to get stream size", zap.String("inputFile", inputFile), zap.String("streamType", string(streamType)), zap.Int("relativeIndex", streamIndex), zap.Error(err))
		return 0, err
	}
	return size, nil
}

func (p *ProbeAnalyzer) GetVideoDurationSec(inputFile string) (uint64, error) {
	p.log.Info("getting video duration in seconds", zap.String("inputFile", inputFile))
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", inputFile)
	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		p.log.Error(fmt.Sprintf("failed to get video duration: %s", string(outputBytes)), zap.String("inputFile", inputFile), zap.Error(err))
		return 0, err
	}
	outputStr := strings.Trim(string(outputBytes), " \n")
	number, err := strconv.ParseFloat(outputStr, 64)
	if err != nil {
		p.log.Error("failed to get video duration", zap.String("inputFile", inputFile), zap.Error(err))
		return 0, err
	}
	return uint64(number), nil
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
	sizeCommand.AddKeyValue("-map", mapValue, ffmpeg.OptionInput)
	sizeCommand.AddKeyValue("-f", "srt", ffmpeg.OptionOutput)
	output, err := sizeCommand.ExecuteSync()
	if err != nil {
		p.log.Warn(fmt.Sprintf("failed to get sub text: %s", *output), zap.String("inputFile", inputFile), zap.Int("streamIndex", streamIndex), zap.Error(err))
		return "", err
	}
	content, err := ioutil.ReadFile(srtFileName)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (p *ProbeAnalyzer) getSubsType(stream *ffprobe.Stream) SubsType {
	switch stream.CodecName {
	case "hdmv_pgs_subtitle":
		return SubsPicture
	case "ass":
		return SubsText
	default:
		return SubsUnknown
	}
}

func (p *ProbeAnalyzer) Analyze(inputFile string, allowAmbiguousResults bool) (*Result, error) {
	scoredResult, err := p.getScoreResult(inputFile)
	if err != nil {
		return nil, err
	}
	if !allowAmbiguousResults && (len(scoredResult.SubCandidates) == 0 || len(scoredResult.AudioCandidates) == 0) {
		return nil, util.ErrNoValidStreamsFound
	}
	if !allowAmbiguousResults && (len(scoredResult.SubCandidates) > 1 || len(scoredResult.AudioCandidates) > 1 || scoredResult.Ambiguous) {
		return nil, util.ErrAmbiguousSelection
	}
	var audioStream *ResultStream = nil
	if len(scoredResult.AudioCandidates) > 0 {
		audioStream = &ResultStream{
			RelativeIndex: scoredResult.AudioCandidates[0].RelativeIndex,
		}
	}
	var subStream *SubStream = nil
	if len(scoredResult.SubCandidates) > 0 {
		subStream = &SubStream{
			ResultStream: ResultStream{
				RelativeIndex: scoredResult.SubCandidates[0].RelativeIndex,
			},
			Type: p.getSubsType(scoredResult.SubCandidates[0].Stream),
		}
	}
	duration, err := p.GetVideoDurationSec(inputFile)
	if err != nil {
		return nil, err
	}
	return &Result{
		Video: VideoStream{
			ResultStream: ResultStream{
				RelativeIndex: scoredResult.Video.RelativeIndex,
			},
			Width:       scoredResult.Video.Width,
			Height:      scoredResult.Video.Height,
			DurationSec: duration,
		},
		Audio: audioStream,
		Sub:   subStream,
	}, nil
}

func (p *ProbeAnalyzer) getScoreResult(inputFile string) (*ScoreResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	probe, err := ffprobe.ProbeURL(ctx, inputFile)
	if err != nil {
		return nil, err
	}
	var videoStream *StreamWithScore
	audioStreams := make([]*StreamWithScore, 0, 10)
	subStreams := make([]*StreamWithScore, 0, 10)
	for _, stream := range probe.Streams {
		switch stream.CodecType {
		case "video":
			// probably a video cover
			if stream.CodecName == "mjpeg" {
				continue
			}
			if videoStream != nil {
				return nil, util.ErrMoreThanOneVideoStream
			}
			videoStream = &StreamWithScore{
				Stream:        stream,
				RelativeIndex: 0,
			}
		case "audio":
			audioStreams = append(audioStreams, &StreamWithScore{
				Stream:        stream,
				RelativeIndex: len(audioStreams),
			})
		case "subtitle":
			subStreams = append(subStreams, &StreamWithScore{
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
			Video:           videoStream,
			AudioCandidates: audioStreams,
			SubCandidates:   subStreams,
		}, nil
	}

	p.log.Info("has multiple audio or sub streams", zap.Int("audio", len(audioStreams)), zap.Int("sub", len(subStreams)), zap.String("inputFile", inputFile))

	prefLangAudio := make([]*StreamWithScore, 0, len(audioStreams))
	for _, stream := range audioStreams {
		lang, _ := stream.Stream.TagList.GetString("language")
		if lang == p.prefAudioLang {
			prefLangAudio = append(prefLangAudio, stream)
		}
	}

	prefLangSubs := make([]*StreamWithScore, 0, len(subStreams))
	for _, stream := range subStreams {
		lang, _ := stream.Stream.TagList.GetString("language")
		if lang == p.prefSubLang {
			prefLangSubs = append(prefLangSubs, stream)
		}
	}

	if len(prefLangSubs) == 1 && len(prefLangAudio) == 1 {
		p.log.Info("has exactly 1 preferred lang sub and audio", zap.String("inputFile", inputFile))
		return &ScoreResult{
			Video:           videoStream,
			AudioCandidates: prefLangAudio,
			SubCandidates:   prefLangSubs,
		}, nil
	}

	p.log.Info("does NOT have EXACTLY 1 preferred lang sub and audio", zap.Int("audio", len(prefLangAudio)), zap.Int("sub", len(prefLangSubs)), zap.String("inputFile", inputFile))

	var remainingAudio []*StreamWithScore
	var remainingSubs []*StreamWithScore

	if len(prefLangAudio) == 0 {
		remainingAudio = audioStreams
	} else {
		remainingAudio = prefLangAudio
	}

	if len(prefLangSubs) == 0 {
		remainingSubs = subStreams
	} else {
		remainingSubs = prefLangSubs
	}

	for _, audioStream := range remainingAudio {
		size, err := p.GetStreamSize(inputFile, StreamAudio, audioStream.RelativeIndex)
		if err != nil {
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
			p.log.Warn("failed to get subtitle text", zap.Int("relativeIndex", subStream.RelativeIndex), zap.Error(err))
		}
		if len(text) <= 32 {
			p.log.Info("subtitle doesn't have text, using scoring based on stream size", zap.Int("relativeIndex", subStream.RelativeIndex), zap.Error(err))
			size, err := p.GetStreamSize(inputFile, StreamSub, subStream.RelativeIndex)
			if err != nil {
				continue
			}
			p.log.Info("got sub stream size", zap.String("inputFile", inputFile), zap.Int("relativeIndex", subStream.RelativeIndex), zap.Uint64("size", size))
			subStream.Score = size
		} else {
			numberOfEngWords := p.textAnalyzer.CountWords(text)
			subStream.Score = numberOfEngWords
			p.log.Info("got number of eng words in sub stream", zap.String("inputFile", inputFile), zap.Int("relativeIndex", subStream.RelativeIndex), zap.Uint64("wordCount", numberOfEngWords))
		}
	}
	sort.SliceStable(remainingSubs, func(i, j int) bool {
		return remainingSubs[i].Score > remainingSubs[j].Score
	})

	p.log.Warn("has ambiguous selection of subs/audio", zap.Int("audio", len(prefLangAudio)), zap.Int("sub", len(prefLangSubs)), zap.String("inputFile", inputFile))
	return &ScoreResult{
		Ambiguous:       true,
		Video:           videoStream,
		AudioCandidates: remainingAudio,
		SubCandidates:   remainingSubs,
	}, nil
}

var ProbeAnalyzerExport = fx.Options(fx.Provide(NewProbeAnalyzer))
