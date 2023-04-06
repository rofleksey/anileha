package util

import (
	"regexp"
	"strings"
	"unicode"
)

var SpacesRegex = regexp.MustCompile("\\s+")

func RemoveWhitespace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

func RemoveNonAlphaNonSpace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, str)
}

func SubstrStart(input string, start int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	return string(asRunes[start:])
}

func Substr(input string, start int, end int) string {
	asRunes := []rune(input)

	if start >= len(asRunes) {
		return ""
	}

	if end > len(asRunes) {
		end = len(asRunes)
	}

	return string(asRunes[start:end])
}
