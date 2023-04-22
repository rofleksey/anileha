package meta

func substringStart(input string, start int) string {
	runes := []rune(input)
	if start >= len(runes) {
		return ""
	}
	return string(runes[start:])
}

func substringStartEnd(input string, start int, end int) string {
	runes := []rune(input)
	if start >= len(runes) {
		return ""
	}
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}
