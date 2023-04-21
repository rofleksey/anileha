package util

import (
	"github.com/elliotchance/pie/v2"
	"path/filepath"
	"strings"
)

type FileType string

const (
	FileTypeVideo    FileType = "video"
	FileTypeAudio    FileType = "audio"
	FileTypeSubtitle FileType = "subtitle"
	FileTypeFont     FileType = "font"
	FileTypeArchive  FileType = "archive"
	FileTypeUnknown  FileType = "unknown"
)

var videoExtensions = []string{
	"3gp", "3gpp", "3g2", "h261", "h263", "h264",
	"m4s", "jpgv", "jpm", "jpgm", "mj2", "mjp2",
	"ts", "mp4", "mp4v", "mpg4", "mpeg", "mpg",
	"mpe", "m1v", "m2v", "ogv", "qt", "mov",
	"uvh", "uvvh", "uvm", "uvvm", "uvp", "uvvp",
	"uvs", "uvvs", "uvv", "uvvv", "dvb", "fvt",
	"mxu", "m4u", "pyv", "uvu", "uvvu", "viv",
	"webm", "f4v", "fli", "flv", "m4v", "mkv",
	"mk3d", "mks", "mng", "asf", "asx", "vob",
	"wm", "wmv", "wmx", "wvx", "avi", "movie",
	"smv",
}

var audioExtensions = []string{
	"3gpp", "adts", "aac", "adp",
	"amr", "au", "snd", "mid",
	"midi", "kar", "rmi", "mxmf",
	"mp3", "m4a", "mp4a", "mpga",
	"mp2", "mp2a", "m2a", "m3a",
	"oga", "ogg", "spx", "opus",
	"s3m", "sil", "uva", "uvva",
	"eol", "dra", "dts", "dtshd",
	"lvp", "pya", "ecelp4800", "ecelp7470",
	"ecelp9600", "rip", "wav", "weba",
	"aif", "aiff", "aifc", "caf",
	"flac", "mka", "m3u", "wax",
	"wma", "ram", "ra", "rmp",
	"xm",
}

var subtitleExtensions = []string{
	"srt", "ssa", "ass",
}

var fontExtensions = []string{
	"ttc", "otf", "ttf", "woff", "woff2",
}

var archiveExtensions = []string{
	"zip", "rar", "tar", "gz", "7z",
}

func GetFileType(path string) FileType {
	ext := strings.ToLower(strings.TrimLeft(filepath.Ext(path), "."))

	// TODO: optimize

	if pie.Contains(videoExtensions, ext) {
		return FileTypeVideo
	}

	if pie.Contains(audioExtensions, ext) {
		return FileTypeAudio
	}

	if pie.Contains(subtitleExtensions, ext) {
		return FileTypeSubtitle
	}

	if pie.Contains(fontExtensions, ext) {
		return FileTypeFont
	}

	if pie.Contains(archiveExtensions, ext) {
		return FileTypeArchive
	}

	return FileTypeUnknown
}
