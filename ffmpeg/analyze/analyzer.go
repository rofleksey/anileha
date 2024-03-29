package analyze

import (
	"anileha/config"
	"anileha/db"
	"anileha/ffmpeg"
	"anileha/util"
	"context"
	"fmt"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gopkg.in/vansante/go-ffprobe.v2"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ProbeAnalyzer Selects streams and generates FFmpeg command for encoding
// For video: Use the last video stream
// For audio: select the only stream, else selects the only japanese stream, else selects the most heavy one (among japanese or all if not present)
// For subs: select the only stream, else selects the only english stream, else selects the stream with most occurrences of english words (among english or all if not present)
type ProbeAnalyzer struct {
	regexMap map[StreamType]*regexp.Regexp
	log      *zap.Logger
	config   *config.Config
}

func NewProbeAnalyzer(
	log *zap.Logger, config *config.Config,
) *ProbeAnalyzer {
	regexMap := make(map[StreamType]*regexp.Regexp)
	regexMap[StreamVideo] = videoRegex
	regexMap[StreamAudio] = audioRegex
	regexMap[StreamSub] = subRegex
	return &ProbeAnalyzer{
		regexMap: regexMap,
		log:      log,
		config:   config,
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
	var resultStr string
	args := p.config.FFMpeg.StreamSizeArgs
	streamLetter := streamType[0:1]
	mapValue := fmt.Sprintf("0:%s:%d", streamLetter, streamIndex)
	sizeCommand := ffmpeg.NewCommand("ffmpeg", args, 0)
	sizeCommand.AddVar("MAX", "2147483647")
	sizeCommand.AddVar("INPUT", inputFile)
	sizeCommand.AddVar("MAP", mapValue)
	p.log.Info("getting stream size",
		zap.String("inputFile", inputFile),
		zap.String("streamType", string(streamType)),
		zap.Int("relativeIndex", streamIndex),
		zap.String("cmd", sizeCommand.String()))
	result, err := sizeCommand.ExecuteSync()

	if result != nil {
		resultStr = string(result)
	}

	if err != nil {
		p.log.Error(fmt.Sprintf("failed to get stream size: %s", resultStr), zap.String("inputFile", inputFile), zap.String("streamType", string(streamType)), zap.Int("relativeIndex", streamIndex), zap.Error(err))
		return 0, err
	}
	size, err := p.parseStreamSize(resultStr, streamType)
	if err != nil {
		p.log.Error("failed to get stream size", zap.String("inputFile", inputFile), zap.String("streamType", string(streamType)), zap.Int("relativeIndex", streamIndex), zap.Error(err))
		return 0, err
	}
	return size, nil
}

func (p *ProbeAnalyzer) GetVideoDurationSec(inputFile string) (int, error) {
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
	return int(number), nil
}

// ExtractSubText gets sub stream text
func (p *ProbeAnalyzer) ExtractSubText(inputFile string, streamIndex int) (string, error) {
	var resultStr string
	srtFileName := inputFile + ".srt"
	defer func() {
		_ = os.Remove(srtFileName)
	}()
	mapValue := fmt.Sprintf("0:s:%d", streamIndex)
	args := p.config.FFMpeg.ExtractSubArgs
	sizeCommand := ffmpeg.NewCommand("ffmpeg", args, 0)
	sizeCommand.AddVar("INPUT", inputFile)
	sizeCommand.AddVar("OUTPUT", srtFileName)
	sizeCommand.AddVar("MAP", mapValue)
	p.log.Info("extracting stream subtitle text",
		zap.String("inputFile", inputFile),
		zap.Int("relativeIndex", streamIndex),
		zap.String("cmd", sizeCommand.String()))
	output, err := sizeCommand.ExecuteSync()
	if output != nil {
		resultStr = string(output)
	}
	if err != nil {
		p.log.Warn(fmt.Sprintf("failed to get sub text: %s", resultStr), zap.String("inputFile", inputFile), zap.Int("streamIndex", streamIndex), zap.Error(err))
		return "", err
	}
	content, err := os.ReadFile(srtFileName)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (p *ProbeAnalyzer) getSubsType(stream *ffprobe.Stream) db.SubsType {
	switch stream.CodecName {
	case "hdmv_pgs_subtitle":
		return db.SubsPicture
	case "dvd_subtitle":
		return db.SubsPicture
	case "ass":
		return db.SubsText
	case "subrip":
		return db.SubsText
	case "srt":
		return db.SubsText
	default:
		return db.SubsUnknown
	}
}

func (p *ProbeAnalyzer) getLang(stream *ffprobe.Stream) string {
	lang, _ := stream.TagList.GetString("language")
	return lang
}

func (p *ProbeAnalyzer) getName(stream *ffprobe.Stream) string {
	title, _ := stream.TagList.GetString("title")
	return title
}

func (p *ProbeAnalyzer) Probe(inputFile string) (*db.AnalysisResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	p.log.Info("probing file", zap.String("file", inputFile))

	probe, err := ffprobe.ProbeURL(ctx, inputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to run ffprobe: %w", err)
	}

	var videoIndex *StreamWithIndex

	audioIndices := make([]StreamWithIndex, 0, 10)
	subIndices := make([]StreamWithIndex, 0, 10)

	for _, stream := range probe.Streams {
		switch stream.CodecType {
		case "video":
			// probably a video cover
			if stream.CodecName == "mjpeg" {
				continue
			}
			if videoIndex != nil {
				return nil, util.ErrMoreThanOneVideoStream
			}
			videoIndex = &StreamWithIndex{
				Stream:        stream,
				RelativeIndex: 0,
			}
		case "audio":
			audioIndices = append(audioIndices, StreamWithIndex{
				Stream:        stream,
				RelativeIndex: len(audioIndices),
			})
		case "subtitle":
			subIndices = append(subIndices, StreamWithIndex{
				Stream:        stream,
				RelativeIndex: len(subIndices),
			})
		}
	}

	if videoIndex == nil {
		return nil, util.ErrVideoStreamNotFound
	}

	audioStreams := make([]db.AudioStream, 0, len(audioIndices))
	subStreams := make([]db.SubStream, 0, len(subIndices))

	for _, audioIndex := range audioIndices {
		size, err := p.GetStreamSize(inputFile, StreamAudio, audioIndex.RelativeIndex)
		if err != nil {
			continue
		}
		audioStreams = append(audioStreams, db.AudioStream{
			BaseStream: db.BaseStream{
				RelativeIndex: audioIndex.RelativeIndex,
				Name:          p.getName(audioIndex.Stream),
				Size:          size,
				Lang:          p.getLang(audioIndex.Stream),
			},
		})
	}

	for _, subIndex := range subIndices {
		text, err := p.ExtractSubText(inputFile, subIndex.RelativeIndex)
		textLength := len(text)

		if err != nil {
			p.log.Warn("failed to get subtitle text", zap.Int("relativeIndex", subIndex.RelativeIndex), zap.Error(err))
		}

		size, err := p.GetStreamSize(inputFile, StreamSub, subIndex.RelativeIndex)
		if err != nil {
			p.log.Warn("failed to get subtitle stream size", zap.Int("relativeIndex", subIndex.RelativeIndex), zap.Error(err))
		}

		subStreams = append(subStreams, db.SubStream{
			BaseStream: db.BaseStream{
				RelativeIndex: subIndex.RelativeIndex,
				Name:          p.getName(subIndex.Stream),
				Size:          size,
				Lang:          p.getLang(subIndex.Stream),
			},
			Type:       p.getSubsType(subIndex.Stream),
			TextLength: textLength,
		})
	}

	durationSec, err := p.GetVideoDurationSec(inputFile)
	if err != nil {
		p.log.Warn("failed to get video duration", zap.String("inputFile", inputFile), zap.Error(err))
	}

	videoStream := db.VideoStream{
		BaseStream: db.BaseStream{
			RelativeIndex: videoIndex.RelativeIndex,
			Name:          p.getName(videoIndex.Stream),
		},
		DurationSec: durationSec,
	}

	return &db.AnalysisResult{
		Video: videoStream,
		Audio: audioStreams,
		Sub:   subStreams,
	}, nil
}

var ProbeAnalyzerExport = fx.Options(fx.Provide(NewProbeAnalyzer))
