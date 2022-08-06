package tests

import (
	"anileha/ffmpeg"
	"anileha/util"
	"go.uber.org/zap"
	"testing"
)

func TestFfmpegSimple(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	command := ffmpeg.NewCommand("input.mkv", 0, "output.mp4")
	t.Log(command.String())
	outputChan, _, err := command.Execute(logger)
	if err != nil {
		t.Fatal("Failed to start command", err)
	}
	for msg := range outputChan {
		switch casted := msg.(type) {
		case string:
			t.Log(casted)
		case util.Progress:
			t.Log(casted)
		case error:
			t.Fatal("Conversion failed", casted)
		}
	}
}

func TestFfmpegInterrupt(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	command := ffmpeg.NewCommand("input.mkv", 0, "output.mp4")
	t.Log(command.String())
	outputChan, cancelFunc, err := command.Execute(logger)
	if err != nil {
		t.Fatal("Failed to start command", err)
	}
	cancelFunc()
	isKilled := false
	for msg := range outputChan {
		switch casted := msg.(type) {
		case string:
			t.Log(casted)
		case util.Progress:
			t.Log(casted)
		case error:
			if casted.Error() == "signal: killed" {
				isKilled = true
			}
			t.Log("Conversion failed", casted)
		}
	}
	if !isKilled {
		t.Fatal("Ffmpeg is not killed properly")
	}
}

func TestFfmpegInvalidFile(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	command := ffmpeg.NewCommand("input.mkv", 0, "output.mp4")
	t.Log(command.String())
	outputChan, _, err := command.Execute(logger)
	if err != nil {
		t.Fatal("Failed to start command", err)
	}
	okay := false
	for msg := range outputChan {
		switch casted := msg.(type) {
		case string:
			t.Log(casted)
			if casted == "lmao.mkv: No such file or directory" {
				okay = true
			}
		case util.Progress:
			t.Log(casted)
		case error:
			t.Log("Conversion failed", casted)
		}
	}
	if !okay {
		t.Fatal("Test failed")
	}
}
