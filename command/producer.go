package command

import (
	"anileha/dao"
	"anileha/ffmpeg"
	"anileha/util"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"sort"
)

type Producer struct {
	log *zap.Logger
}

func NewProducer(
	log *zap.Logger,
) *Producer {
	return &Producer{
		log: log,
	}
}

func (p *Producer) selectAudio(streams []dao.AudioStream, prefs PreferencesData) *selectedAudioStream {
	if len(streams) == 0 {
		return nil
	}

	if prefs.Disable {
		return nil
	}

	if prefs.ExternalFile != "" {
		return &selectedAudioStream{
			ExternalFile: prefs.ExternalFile,
		}
	}

	if prefs.StreamIndex != nil {
		return &selectedAudioStream{
			StreamIndex: prefs.StreamIndex,
		}
	}

	if prefs.Lang != "" {
		newStreams := pie.Filter(streams, func(stream dao.AudioStream) bool {
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

func (p *Producer) selectSub(streams []dao.SubStream, prefs PreferencesData) *selectedSubStream {
	if len(streams) == 0 {
		return nil
	}

	if prefs.Disable {
		return nil
	}

	if prefs.ExternalFile != "" {
		return &selectedSubStream{
			ExternalFile: prefs.ExternalFile,
			Filter:       subtitlesSubFilter,
		}
	}

	if prefs.StreamIndex != nil {
		index := pie.FindFirstUsing(streams, func(stream dao.SubStream) bool {
			return stream.RelativeIndex == *prefs.StreamIndex
		})
		subsType := streams[index].Type

		var filter subFilter

		if subsType == dao.SubsPicture {
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
		newStreams := pie.Filter(streams, func(stream dao.SubStream) bool {
			return stream.Lang == prefs.Lang
		})
		if len(newStreams) > 0 {
			streams = newStreams
		}
	}

	pictureSubs := pie.Filter(streams, func(stream dao.SubStream) bool {
		return stream.TextLength < 32
	})
	textSubs := pie.Filter(streams, func(stream dao.SubStream) bool {
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

func (p *Producer) GetFFmpegCommand(inputFile string, outputPath string, logsPath string, probe *dao.AnalysisResult,
	prefs Preferences) (*ffmpeg.Command, error) {
	command := ffmpeg.NewCommand(inputFile, probe.Video.DurationSec, outputPath)
	command.AddKeyValue("-acodec", "aac", ffmpeg.OptionOutput)
	command.AddKeyValue("-b:a", "196k", ffmpeg.OptionOutput)
	command.AddKeyValue("-ac", "2", ffmpeg.OptionOutput)
	command.AddKeyValue("-vcodec", "libx264", ffmpeg.OptionOutput)
	command.AddKeyValue("-crf", "18", ffmpeg.OptionOutput)
	command.AddKeyValue("-tune", "animation", ffmpeg.OptionOutput)  // this is bad?
	command.AddKeyValue("-pix_fmt", "yuv420p", ffmpeg.OptionOutput) // yuv420p10le?
	command.AddKeyValue("-preset", "slow", ffmpeg.OptionOutput)
	command.AddKeyValue("-f", "mp4", ffmpeg.OptionOutput)
	command.AddKeyValue("-movflags", "+faststart", ffmpeg.OptionPostOutput)

	audioPick := p.selectAudio(probe.Audio, prefs.Audio)
	subPick := p.selectSub(probe.Sub, prefs.Sub)

	if subPick != nil {
		switch subPick.Filter {
		case overlaySubFilter:
			if subPick.ExternalFile != "" {
				command.AddKeyValue("-filter_complex", fmt.Sprintf("[0:v]subtitles=f='%s'",
					subPick.ExternalFile), ffmpeg.OptionOutput)
			} else {
				command.AddKeyValue("-filter_complex", fmt.Sprintf("[0:v]subtitles=f='%s':si=%d[vo]",
					inputFile, *subPick.StreamIndex), ffmpeg.OptionOutput)
			}
			command.AddKeyValue("-map", "[vo]", ffmpeg.OptionPostOutput)
		case subtitlesSubFilter:
			command.AddKeyValue("-filter_complex", fmt.Sprintf("[0:v][0:s:%d]overlay[vo]",
				*subPick.StreamIndex), ffmpeg.OptionOutput)
			command.AddKeyValue("-map", "[vo]", ffmpeg.OptionPostOutput)
		default:
			return nil, util.ErrUnsupportedSubs
		}
	} else {
		command.AddKeyValue("-map", "0:v", ffmpeg.OptionPostOutput)
	}

	if audioPick != nil {
		command.AddKeyValue("-map", fmt.Sprintf("0:a:%d", *audioPick.StreamIndex), ffmpeg.OptionOutput)
	}
	command.WriteLogsTo(logsPath)
	return command, nil
}

var ProducerExport = fx.Options(fx.Provide(NewProducer))
