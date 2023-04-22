package meta

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type EpisodeMetadata struct {
	Season  string `json:"season"`
	Episode string `json:"episode"`
}

var clusterRegex = regexp.MustCompile("\\s{2,}")
var numberWithSpaceRegex = regexp.MustCompile(" (\\d+(?:.\\d+)?)")
var startsWithNumberRegex = regexp.MustCompile("^(\\d+(?:.\\d+)?).*$")
var fullNumberRegex = regexp.MustCompile("^\\d+$")
var delimiterRegex = regexp.MustCompile("[-_]")

//var badWords = []string{"x264", "x265", "opus", "aac", "mp4", "mp3", "mkv", "hevc", "avc", "flac", "dual", "webdl", "dvd", "cd", "rip"}
//var badRegexes = []*regexp.Regexp{regexp.MustCompile("\\d+\\s*p"), regexp.MustCompile("\\dk")}

var sSeRegex = regexp.MustCompile("(?i)s\\s*(\\d+)\\s*e\\s*(\\d+)")
var sEsRegex = regexp.MustCompile("(?i)e\\s*(\\d+)\\s*s\\s*(\\d+)")
var sEpisodeRegex = regexp.MustCompile("(?i)episode\\s*(\\d+)")
var sEpRegex = regexp.MustCompile("(?i)ep\\s*(\\d+)")
var sSeasonRegex = regexp.MustCompile("(?i)season\\s*(\\d+)")
var sSxERegex = regexp.MustCompile("(?i)(\\d+)\\s*x\\s*(\\d+)")
var eDotSpaceRegex = regexp.MustCompile("(?i)(\\d+)\\.\\s")

func GuessEpisodeMetadata(filename string) EpisodeMetadata {
	var resultSeason string
	var resultEpisode string

	// remove path and extension, make lowercase
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	// replace delimiters with spaces
	spaced := delimiterRegex.ReplaceAllLiteralString(base, " ")

	// try to find popular formats
	if test := sSxERegex.FindStringSubmatch(spaced); test != nil {
		season, _ := strconv.Atoi(test[1])
		if season < 100 {
			return EpisodeMetadata{
				Season:  test[1],
				Episode: test[2],
			}
		}
	}

	if test := sEsRegex.FindStringSubmatch(spaced); test != nil {
		return EpisodeMetadata{
			Season:  test[2],
			Episode: test[1],
		}
	}

	if test := sSeRegex.FindStringSubmatch(spaced); test != nil {
		return EpisodeMetadata{
			Season:  test[1],
			Episode: test[2],
		}
	}

	if test := sEpRegex.FindStringSubmatch(spaced); test != nil {
		resultEpisode = test[1]
		spaced = sEpRegex.ReplaceAllLiteralString(spaced, "")
	}

	if test := sEpisodeRegex.FindStringSubmatch(spaced); test != nil {
		resultEpisode = test[1]
		spaced = sEpisodeRegex.ReplaceAllLiteralString(spaced, "")
	}

	if test := eDotSpaceRegex.FindStringSubmatch(spaced); test != nil {
		resultEpisode = test[1]
		spaced = eDotSpaceRegex.ReplaceAllLiteralString(spaced, "")
	}

	if test := sSeasonRegex.FindStringSubmatch(spaced); test != nil {
		resultSeason = test[1]
		spaced = sSeasonRegex.ReplaceAllLiteralString(spaced, "")
	}

	// remove brackets, they are almost always meaningless
	bracketDepth := 0
	startIndex := 0

	for i := 0; i < len(spaced); i++ {
		if spaced[i] == '(' || spaced[i] == '[' || spaced[i] == '{' {
			bracketDepth++

			startIndex = i
		} else if spaced[i] == ')' || spaced[i] == ']' || spaced[i] == '}' {
			bracketDepth--

			if bracketDepth < 0 {
				break
			}

			if bracketDepth == 0 {
				substr := substringStartEnd(spaced, startIndex+1, i)
				// if bracket's only content is a number, it is a good candidate for episode
				if resultEpisode == "" && fullNumberRegex.MatchString(substr) && len(substr) <= 3 {
					resultEpisode = substr
				}

				spaced = substringStartEnd(spaced, 0, startIndex) + substringStart(spaced, i+1)
				i = -1

				continue
			}
		}
	}

	// trim and split into clusters
	spaced = strings.Trim(spaced, " ")
	clusters := clusterRegex.Split(spaced, -1)

	// (episode) find a single cluster starting with a number
	if resultEpisode == "" {
		clustersStartingWithNumber := 0
		lastNumber := ""
		lastCluster := -1

		for i, cluster := range clusters {
			if test := startsWithNumberRegex.FindStringSubmatch(cluster); test != nil {
				clustersStartingWithNumber++
				lastNumber = test[1]
				lastCluster = i
			}
		}

		if clustersStartingWithNumber == 1 {
			resultEpisode = lastNumber
			clusters[lastCluster] = strings.Replace(clusters[lastCluster], lastNumber, "", 1)
		}
	}

	// (episode) find a single cluster with a single number
	if resultEpisode == "" {
		clustersWithNumbers := 0
		lastNumber := ""
		lastCluster := -1

		for i, cluster := range clusters {
			if test := numberWithSpaceRegex.FindAllStringSubmatch(cluster, -1); test != nil && len(test) == 1 {
				clustersWithNumbers++
				lastNumber = test[0][1]
				lastCluster = i
			}
		}

		if clustersWithNumbers == 1 {
			resultEpisode = lastNumber
			clusters[lastCluster] = strings.Replace(clusters[lastCluster], lastNumber, "", 1)
		}
	}

	// last resort
	if len(clusters) > 0 {
		if resultSeason == "" {
			resultSeason = strings.Trim(clusters[0], " ")
		}

		if len(clusters) == 2 {
			if resultEpisode == "" {
				resultEpisode = strings.Trim(clusters[1], " ")
			}
		}
	}

	if resultEpisode == "" && resultSeason != "" {
		resultEpisode = resultSeason
		resultSeason = ""
	}

	if resultEpisode == "" {
		resultEpisode = strings.Trim(base, " ")
	}

	return EpisodeMetadata{
		Season:  resultSeason,
		Episode: resultEpisode,
	}
}
