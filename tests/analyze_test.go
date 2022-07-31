package tests

import (
	"anileha/analyze"
	"github.com/go-playground/assert/v2"
	"go.uber.org/zap"
	"testing"
)

func TestAnalyzeFile(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	textAnalyzer := analyze.NewTextAnalyzer(logger)
	probeAnalyzer := analyze.NewProbeAnalyzer(textAnalyzer, logger)
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
	textAnalyzer := analyze.NewTextAnalyzer(logger)
	probeAnalyzer := analyze.NewProbeAnalyzer(textAnalyzer, logger)
	text, err := probeAnalyzer.ExtractSubText("input.mkv", 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(text)
}

func TestAnalyzeStream(t *testing.T) {
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatal(err)
	}
	textAnalyzer := analyze.NewTextAnalyzer(logger)
	probeAnalyzer := analyze.NewProbeAnalyzer(textAnalyzer, logger)
	size, err := probeAnalyzer.GetStreamSize("input.mkv", analyze.StreamAudio, 0)
	if err != nil {
		t.Fatal(err)
	}
	assert.NotEqual(t, size, 0)
	t.Log(size)
}
