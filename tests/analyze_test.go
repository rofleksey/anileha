package tests

import (
	"anileha/analyze"
	"anileha/config"
	"github.com/go-playground/assert/v2"
	"go.uber.org/zap"
	"testing"
)

func TestAnalyzeFile(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	c, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	textAnalyzer, err := analyze.NewTextAnalyzer(c, logger)
	if err != nil {
		t.Fatal(err)
	}
	probeAnalyzer := analyze.NewProbeAnalyzer(textAnalyzer, c, logger)
	analysis, err := probeAnalyzer.Analyze("input.mkv", true)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, analysis.Video.Width, 1920)
	assert.Equal(t, analysis.Video.Height, 1080)
}

func TestAnalyzeSubs(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	c, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	textAnalyzer, err := analyze.NewTextAnalyzer(c, logger)
	if err != nil {
		t.Fatal(err)
	}
	probeAnalyzer := analyze.NewProbeAnalyzer(textAnalyzer, c, logger)
	text, err := probeAnalyzer.ExtractSubText("input.mkv", 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(text)
}

func TestVideoDuration(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	c, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	textAnalyzer, err := analyze.NewTextAnalyzer(c, logger)
	if err != nil {
		t.Fatal(err)
	}
	probeAnalyzer := analyze.NewProbeAnalyzer(textAnalyzer, c, logger)
	duration, err := probeAnalyzer.GetVideoDurationSec("input.mkv")
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, duration, 0)
	t.Log(duration)
}

func TestAnalyzeStream(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	c, err := config.LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	textAnalyzer, err := analyze.NewTextAnalyzer(c, logger)
	if err != nil {
		t.Fatal(err)
	}
	probeAnalyzer := analyze.NewProbeAnalyzer(textAnalyzer, c, logger)
	size, err := probeAnalyzer.GetStreamSize("input.mkv", analyze.StreamAudio, 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, size, 0)
	t.Log(size)
}
