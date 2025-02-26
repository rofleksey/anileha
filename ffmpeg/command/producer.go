package command

import (
	"anileha/config"
	"anileha/db"
	"anileha/ffmpeg"
	"anileha/util"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
	"path"
	"runtime"
	"sort"
	"strconv"
)

type Producer struct {
	log     *zap.Logger
	config  *config.Config
	fontDir string
}

func NewProducer(
	log *zap.Logger,
	config *config.Config,
) (*Producer, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	fontDir := path.Join(workingDir, config.Data.Dir, util.FontSubDir)
	log.Info("cpu count", zap.Int("count", runtime.NumCPU()))
	return &Producer{
		log:     log,
		config:  config,
		fontDir: fontDir,
	}, nil
}

func (p *Producer) selectAudio(streams []db.AudioStream, prefs PreferencesData) *selectedAudioStream {
	if prefs.Disable {
		return nil
	}

	if prefs.ExternalFile != "" {
		return &selectedAudioStream{
			ExternalFile: prefs.ExternalFile,
		}
	}

	if len(streams) == 0 {
		return nil
	}

	if prefs.StreamIndex != nil {
		return &selectedAudioStream{
			StreamIndex: prefs.StreamIndex,
		}
	}

	if prefs.Lang != "" {
		newStreams := pie.Filter(streams, func(stream db.AudioStream) bool {
			return stream.Lang == prefs.Lang
		})
		if len(newStreams) > 0 {
			streams = newStreams
		}
	}

	sort.Slice(streams, func(i, j int) bool {
		return streams[i].Size < streams[j].Size
	})

	index := streams[len(streams)-1].RelativeIndex

	return &selectedAudioStream{
		StreamIndex: &index,
	}
}

func (p *Producer) selectSub(streams []db.SubStream, prefs PreferencesData) *selectedSubStream {
	if prefs.Disable {
		return nil
	}

	if prefs.ExternalFile != "" {
		return &selectedSubStream{
			ExternalFile: prefs.ExternalFile,
			Filter:       subtitlesSubFilter,
		}
	}

	if len(streams) == 0 {
		return nil
	}

	if prefs.StreamIndex != nil {
		index := pie.FindFirstUsing(streams, func(stream db.SubStream) bool {
			return stream.RelativeIndex == *prefs.StreamIndex
		})
		subsType := streams[index].Type

		var filter subFilter

		if subsType == db.SubsPicture {
			filter = overlaySubFilter
		} else {
			filter = subtitlesSubFilter
		}

		return &selectedSubStream{
			StreamIndex: prefs.StreamIndex,
			Filter:      filter,
		}
	}

	if prefs.Lang != "" {
		newStreams := pie.Filter(streams, func(stream db.SubStream) bool {
			return stream.Lang == prefs.Lang
		})
		if len(newStreams) > 0 {
			streams = newStreams
		}
	}

	pictureSubs := pie.Filter(streams, func(stream db.SubStream) bool {
		return stream.TextLength < 32
	})
	textSubs := pie.Filter(streams, func(stream db.SubStream) bool {
		return stream.TextLength >= 32
	})

	// prefer picture subs, pick one with the largest size

	if len(pictureSubs) > 0 {
		sort.Slice(pictureSubs, func(i, j int) bool {
			return pictureSubs[i].Size < pictureSubs[j].Size
		})

		index := pictureSubs[len(pictureSubs)-1].RelativeIndex

		return &selectedSubStream{
			StreamIndex: &index,
			Filter:      overlaySubFilter,
		}
	}

	// pick sub with the longest text content

	sort.Slice(textSubs, func(i, j int) bool {
		return textSubs[i].TextLength < textSubs[j].TextLength
	})

	index := textSubs[len(textSubs)-1].RelativeIndex

	return &selectedSubStream{
		StreamIndex: &index,
		Filter:      subtitlesSubFilter,
	}
}

func (p *Producer) GetFFmpegCommand(inputFile string, outputPath string, logsPath string, probe *db.AnalysisResult,
	prefs Preferences) (*ffmpeg.Command, error) {
	// free 2 virtual CPUs from ffmpeg workload
	numThreads := runtime.NumCPU() - 2
	if numThreads < 1 {
		numThreads = 1
	}

	// ffmpeg doesn't recommend setting this above 16
	if numThreads > 16 {
		numThreads = 16
	}

	args := p.config.FFMpeg.ConvertArgs
	command := ffmpeg.NewCommand("ffmpeg", args, probe.Video.DurationSec)
	command.AddVar("INPUT", inputFile)
	command.AddVar("OUTPUT", outputPath)
	command.AddVar("THREADS", strconv.Itoa(numThreads))

	audioPick := p.selectAudio(probe.Audio, prefs.Audio)
	subPick := p.selectSub(probe.Sub, prefs.Sub)

	if subPick != nil {
		switch subPick.Filter {
		case subtitlesSubFilter:
			if subPick.ExternalFile != "" {
				command.AddVar("FILTER_SUB", "-filter_complex",
					fmt.Sprintf("[0:v]subtitles=f='%s':fontsdir='%s'[vo]", subPick.ExternalFile, p.fontDir))
			} else {
				command.AddVar("FILTER_SUB", "-filter_complex",
					fmt.Sprintf("[0:v]subtitles=f='%s':si=%d[vo]", inputFile, *subPick.StreamIndex))
			}
		case overlaySubFilter:
			command.AddVar("FILTER_SUB", "-filter_complex",
				fmt.Sprintf("[0:v][0:s:%d]overlay[vo]", *subPick.StreamIndex))

		default:
			return nil, util.ErrUnsupportedSubs
		}

		command.AddVar("MAP_SUB", "-map", "[vo]")
	} else {
		command.AddVar("MAP_SUB", "-map", "0:v")
	}

	if audioPick != nil {
		command.AddVar("MAP_AUDIO", "-map", fmt.Sprintf("0:a:%d", *audioPick.StreamIndex))
	}

	command.WriteLogsTo(logsPath)

	return command, nil
}

var ProducerExport = fx.Options(fx.Provide(NewProducer))
