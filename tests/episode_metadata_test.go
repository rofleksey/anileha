package tests

import (
	"anileha/util"
	"github.com/go-playground/assert/v2"
	"testing"
)

func TestSingleEpisodeMetadata1(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("Koroshi Ai - 06 (WEBDL 1080p HEVC AAC) Ukr DVO.mkv")
	assert.Equal(t, metadata.Season, "")
	assert.Equal(t, metadata.Episode, "06")
}

func TestSingleEpisodeMetadata2(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[Judas] Hunter x Hunter (2011) - S01E008.mkv ")
	assert.Equal(t, metadata.Season, "01")
	assert.Equal(t, metadata.Episode, "008")
}

func TestSingleEpisodeMetadata3(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[Judas] Hunter x Hunter (2011) - S01E012.mkv ")
	assert.Equal(t, metadata.Season, "01")
	assert.Equal(t, metadata.Episode, "012")
}

func TestSingleEpisodeMetadata4(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[Judas] Hunter x Hunter (2011) - S01E113.mkv ")
	assert.Equal(t, metadata.Season, "01")
	assert.Equal(t, metadata.Episode, "113")
}

func TestSingleEpisodeMetadata5(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("Hellsing - Ep. 05 - Brotherhood (480p DVDRip - DUAL Audio).mkv")
	assert.Equal(t, metadata.Season, "")
	assert.Equal(t, metadata.Episode, "05")
}

func TestSingleEpisodeMetadata6(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[Samir755] Hellsing Ultimate 02.mkv")
	assert.Equal(t, metadata.Season, "Hellsing Ultimate")
	assert.Equal(t, metadata.Episode, "02")
}

func TestSingleEpisodeMetadata7(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[CBM]_Hellsing_Ultimate_-_06_-_[1080p-AC3]_[1CB8EDB0].mkv")
	assert.Equal(t, metadata.Season, "Hellsing Ultimate")
	assert.Equal(t, metadata.Episode, "06")
}

func TestSingleEpisodeMetadata8(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[DB]Kabukimonogatari_-_NCED01_(10bit_BD1080p_x265).mkv ")
	assert.Equal(t, metadata.Season, "Kabukimonogatari")
	assert.Equal(t, metadata.Episode, "NCED01")
}

func TestSingleEpisodeMetadata9(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[DB]Nekomonogatari (Black)_-_Recap01_(10bit_BD1080p_x265).mkv")
	assert.Equal(t, metadata.Season, "Nekomonogatari")
	assert.Equal(t, metadata.Episode, "Recap01")
}

func TestSingleEpisodeMetadata10(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[AceAres] Suzumiya Haruhi-chan no Yuuutsu - Episode 06 [1080p BD Dual Audio x265].mkv")
	assert.Equal(t, metadata.Season, "Suzumiya Haruhi chan no Yuuutsu")
	assert.Equal(t, metadata.Episode, "06")
}

func TestSingleEpisodeMetadata11(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[VCB-Studio] Suzumiya Haruhi no Gensou [01][Ma10p_1080p][x265_flac].mkv")
	assert.Equal(t, metadata.Season, "Suzumiya Haruhi no Gensou")
	assert.Equal(t, metadata.Episode, "01")
}

func TestSingleEpisodeMetadata12(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[VCB-Studio] Suzumiya Haruhi no Yuuutsu [26][Ma10p_1080p][x265_3flac].mkv")
	assert.Equal(t, metadata.Season, "Suzumiya Haruhi no Yuuutsu")
	assert.Equal(t, metadata.Episode, "26")
}

func TestSingleEpisodeMetadata13(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[Underwater] Panty and Stocking with Garterbelt 07 - Trans-homers - The Stripping (BD 720p) [E9862607].mkv")
	assert.Equal(t, metadata.Season, "")
	assert.Equal(t, metadata.Episode, "07")
}

func TestSingleEpisodeMetadata14(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[Underwater] Panty and Stocking with Garterbelt OVA - In Sanitarybox (BD 720p) [3525A622].mkv")
	assert.Equal(t, metadata.Season, "Panty and Stocking with Garterbelt OVA")
	assert.Equal(t, metadata.Episode, "In Sanitarybox")
}

func TestSingleEpisodeMetadata15(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("Attack On Titan Season 3 - 19.mkv")
	assert.Equal(t, metadata.Season, "3")
	assert.Equal(t, metadata.Episode, "19")
}

func TestSingleEpisodeMetadata16(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("Shingeki No Kyojin Oad 03.mkv")
	assert.Equal(t, metadata.Season, "Shingeki No Kyojin Oad")
	assert.Equal(t, metadata.Episode, "03")
}

func TestSingleEpisodeMetadata17(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[gg]_Trapeze_-_06_[DAA1989B].mkv")
	assert.Equal(t, metadata.Season, "Trapeze")
	assert.Equal(t, metadata.Episode, "06")
}

func TestSingleEpisodeMetadata18(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[gg]_Trapeze_-_07v2_[985067CA].mkv")
	assert.Equal(t, metadata.Season, "Trapeze")
	assert.Equal(t, metadata.Episode, "07")
}

func TestSingleEpisodeMetadata19(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[Fate-Subs]_Kuuchuu_Buranko_(Trapeze)_08_(1280x720_x264_AAC)_Sub_Ita.mp4")
	assert.Equal(t, metadata.Season, "")
	assert.Equal(t, metadata.Episode, "08")
}

func TestSingleEpisodeMetadata20(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[King] Ousama Ranking - 17 [1080p][D2DCB6D0].mkv")
	assert.Equal(t, metadata.Season, "Ousama Ranking")
	assert.Equal(t, metadata.Episode, "17")
}

func TestSingleEpisodeMetadata21(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("Death Note - 01x06 - Unraveling.mkv")
	assert.Equal(t, metadata.Season, "01")
	assert.Equal(t, metadata.Episode, "06")
}

func TestSingleEpisodeMetadata22(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("[Nep_Blanc] Death Note 35 .mkv ")
	assert.Equal(t, metadata.Season, "Death Note")
	assert.Equal(t, metadata.Episode, "35")
}

func TestSingleEpisodeMetadata23(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("Cowboy Bebop - S00E03 - Knockin' On Heaven's Door.mkv")
	assert.Equal(t, metadata.Season, "00")
	assert.Equal(t, metadata.Episode, "03")
}

func TestSingleEpisodeMetadata24(t *testing.T) {
	metadata := util.ParseSingleEpisodeMetadata("Cowboy Bebop - 03 ITBD Remux.mkv")
	assert.Equal(t, metadata.Season, "Cowboy Bebop")
	assert.Equal(t, metadata.Episode, "03")
}
