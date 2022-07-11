package util

import (
	"strconv"
	"strings"
	"unicode"
)

func RemoveWhitespace(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, str)
}

// ParseFileIndices parses strings like '1,2-4,5' into a set of uints
func ParseFileIndices(str string) (map[uint]struct{}, error) {
	// This is the closest thing to Set in golang, struct{} is a Unit type
	// map[uint]bool uses unnecessary memory :)
	result := make(map[uint]struct{})
	strNoWhitespace := RemoveWhitespace(str)
	items := strings.Split(strNoWhitespace, ",")
	for _, item := range items {
		dashSplit := strings.Split(item, "-")
		switch len(dashSplit) {
		case 1:
			if len(dashSplit[0]) == 0 {
				return nil, ErrInvalidIndicesString
			}
			num, err := strconv.ParseUint(dashSplit[0], 10, 64)
			if err != nil {
				return nil, ErrInvalidIndicesString
			}
			result[uint(num)] = struct{}{}
		case 2:
			if len(dashSplit[0]) == 0 || len(dashSplit[1]) == 0 {
				return nil, ErrInvalidIndicesString
			}
			start, err := strconv.ParseUint(dashSplit[0], 10, 64)
			if err != nil {
				return nil, ErrInvalidIndicesString
			}
			end, err := strconv.ParseUint(dashSplit[0], 10, 64)
			if err != nil {
				return nil, ErrInvalidIndicesString
			}
			for pointer := start; pointer <= end; pointer++ {
				result[uint(pointer)] = struct{}{}
				pointer++
			}
		default:
			return nil, ErrInvalidIndicesString
		}
	}
	return result, nil
}
