package analyze

import (
	"anileha/config"
	"anileha/util"
	"bufio"
	_ "embed"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"strings"
)

const NumberOfWords = 371000

type TextAnalyzer struct {
	wordsSet map[string]struct{}
	log      *zap.Logger
}

func NewTextAnalyzer(config *config.Config, log *zap.Logger) *TextAnalyzer {
	reader := strings.NewReader(config.Conversion.WordsPath)
	scanner := bufio.NewScanner(reader)
	words := make(map[string]struct{}, NumberOfWords)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		words[word] = struct{}{}
	}
	log.Info("loaded text analyzer", zap.Int("wordCount", len(words)))
	return &TextAnalyzer{
		wordsSet: words,
	}
}

func (a *TextAnalyzer) CountEnglishWords(text string) uint64 {
	stripped := util.RemoveNonAlpha(text)
	lower := strings.ToLower(stripped)
	splitArr := util.SpacesRegex.Split(lower, -1)
	count := uint64(0)
	for _, word := range splitArr {
		if _, contains := a.wordsSet[word]; contains {
			count++
		}
	}
	return count
}

var TextAnalyzerExport = fx.Options(fx.Provide(NewTextAnalyzer))
