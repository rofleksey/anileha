package command

type PreferencesData struct {
	Disable      bool
	ExternalFile string
	StreamIndex  *int
	Lang         string
}

type Preferences struct {
	Audio PreferencesData
	Sub   PreferencesData
}

type subFilter string

const (
	overlaySubFilter   subFilter = "overlay"
	subtitlesSubFilter subFilter = "subtitles"
)

type selectedAudioStream struct {
	StreamIndex  *int
	ExternalFile string
}

type selectedSubStream struct {
	StreamIndex  *int
	ExternalFile string
	Filter       subFilter
}
