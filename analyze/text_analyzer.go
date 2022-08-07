package analyze

import (
	"anileha/config"
	"anileha/util"
	"bufio"
	_ "embed"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"os"
	"strings"
)

type TextAnalyzer struct {
	wordsSet map[string]struct{}
	log      *zap.Logger
}

func NewTextAnalyzer(config *config.Config, log *zap.Logger) (*TextAnalyzer, error) {
	file, err := os.Open(config.Conversion.WordsPath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	scanner := bufio.NewScanner(reader)
	words := make(map[string]struct{}, 10000)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		words[word] = struct{}{}
	}
	log.Info("loaded text analyzer", zap.Int("wordCount", len(words)))
	return &TextAnalyzer{
		wordsSet: words,
	}, nil
}

func (a *TextAnalyzer) CountWords(text string) uint64 {
	stripped := util.RemoveNonAlphaNonSpace(text)
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
