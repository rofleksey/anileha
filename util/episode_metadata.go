package util

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var clusterRegex = regexp.MustCompile("\\s{2,}")
var numberWithSpaceRegex = regexp.MustCompile(" (\\d+)")
var startsWithNumberRegex = regexp.MustCompile("^(\\d+).*$")
var fullNumberRegex = regexp.MustCompile("^\\d+$")

//var badWords = []string{"x264", "x265", "opus", "aac", "mp4", "mp3", "mkv", "hevc", "avc", "flac", "dual", "webdl", "dvd", "cd", "rip"}
//var badRegexes = []*regexp.Regexp{regexp.MustCompile("\\d+\\s*p"), regexp.MustCompile("\\dk")}

var sSeRegex = regexp.MustCompile("(?i)s\\s*(\\d+)\\s*e\\s*(\\d+)")
var sEsRegex = regexp.MustCompile("(?i)e\\s*(\\d+)\\s*s\\s*(\\d+)")
var sEpisodeRegex = regexp.MustCompile("(?i)episode\\s*(\\d+)")
var sEpRegex = regexp.MustCompile("(?i)ep\\s*(\\d+)")
var sSeasonRegex = regexp.MustCompile("(?i)season\\s*(\\d+)")
var sSxERegex = regexp.MustCompile("(?i)(\\d+)\\s*x\\s*(\\d+)")

// EpisodeMetadata best attempt to order files according to episode airing chronology
// Should follow these rules:
// * If season(file1) == season(file2) -> metadata(file1).Season == metadata(file2).Season
// * If season(file1) < season(file2) -> metadata(file1).Season < metadata(file2).Season
// * If episode(file1) == episode(file2) -> metadata(file1).Episode == metadata(file2).Episode
// * If episode(file1) < episode(file2) -> metadata(file1).Episode < metadata(file2).Episode
// * Episode field should be displayable to end user
// * Episode field should be as short as possible
type EpisodeMetadata struct {
	Season  string
	Episode string
}

func ParseSingleEpisodeMetadata(filename string) EpisodeMetadata {
	var resultSeason string
	var resultEpisode string

	// remove path and extension, make lowercase
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	// replace _ and - with spaces
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.ReplaceAll(base, "-", " ")

	fmt.Println(base)

	// try to find popular formats
	if test := sSxERegex.FindStringSubmatch(base); test != nil {
		season, _ := strconv.Atoi(test[1])
		if season < 100 {
			return EpisodeMetadata{
				Season:  test[1],
				Episode: test[2],
			}
		}
	}
	if test := sEsRegex.FindStringSubmatch(base); test != nil {
		return EpisodeMetadata{
			Season:  test[2],
			Episode: test[1],
		}
	}
	if test := sSeRegex.FindStringSubmatch(base); test != nil {
		return EpisodeMetadata{
			Season:  test[1],
			Episode: test[2],
		}
	}
	if test := sEpRegex.FindStringSubmatch(base); test != nil {
		resultEpisode = test[1]
	}
	if test := sEpisodeRegex.FindStringSubmatch(base); test != nil {
		resultEpisode = test[1]
	}
	if test := sSeasonRegex.FindStringSubmatch(base); test != nil {
		resultSeason = test[1]
	}

	// remove brackets, they are almost always meaningless
	bracketDepth := 0
	startIndex := 0
	for i := 0; i < len(base); i++ {
		if base[i] == '(' || base[i] == '[' || base[i] == '{' {
			bracketDepth++
			startIndex = i
		} else if base[i] == ')' || base[i] == ']' || base[i] == '}' {
			bracketDepth--
			if bracketDepth < 0 {
				break
			}
			if bracketDepth == 0 {
				substr := Substr(base, startIndex+1, i)
				if resultEpisode == "" && fullNumberRegex.MatchString(substr) {
					resultEpisode = substr
				}
				base = Substr(base, 0, startIndex) + SubstrStart(base, i+1)
				i = -1
				continue
			}
		}
	}

	// trim and split into clusters
	base = strings.Trim(base, " ")
	fmt.Println(base)
	clusters := clusterRegex.Split(base, -1)
	for _, cluster := range clusters {
		fmt.Println(fmt.Sprintf("-%s-", cluster))
	}

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
	if len(clusters) == 1 {
		if resultSeason == "" {
			resultSeason = strings.Trim(clusters[0], " ")
		}
		return EpisodeMetadata{
			Season:  resultSeason,
			Episode: resultEpisode,
		}
	}

	if len(clusters) == 2 {
		if resultSeason == "" {
			resultSeason = strings.Trim(clusters[0], " ")
		}
		if resultEpisode == "" {
			resultEpisode = strings.Trim(clusters[1], " ")
		}
		return EpisodeMetadata{
			Season:  resultSeason,
			Episode: resultEpisode,
		}
	}

	return EpisodeMetadata{
		Season:  resultSeason,
		Episode: resultEpisode,
	}
}
