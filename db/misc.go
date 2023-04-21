package db

import "os"

type AuthUser struct {
	ID    uint     `json:"id"`
	Roles []string `json:"roles"`
}

func NewAuthUser(user User) AuthUser {
	return AuthUser{
		ID:    user.ID,
		Roles: user.Roles,
	}
}

type Thumb struct {
	Path string `gorm:"column:thumb_path"`
	Url  string `gorm:"column:thumb_url"`
}

func (t *Thumb) Delete() {
	_ = os.Remove(t.Path)
}

type AutoTorrent struct {
	AudioLang string `json:"audioLang"`
	SubLang   string `json:"subLang"`
}

type SubsType string

const (
	SubsText    SubsType = "text"
	SubsPicture SubsType = "picture"
	SubsUnknown SubsType = "unknown"
)

type BaseStream struct {
	RelativeIndex int    `json:"index"`
	Name          string `json:"name"`
	Size          uint64 `json:"size"`
	Lang          string `json:"lang"`
}

type VideoStream struct {
	BaseStream
	Width       int `json:"width"`
	Height      int `json:"height"`
	DurationSec int `json:"durationSec"`
}

type AudioStream struct {
	BaseStream
}

type SubStream struct {
	BaseStream
	Type       SubsType `json:"type"`
	TextLength int      `json:"textLength"`
}

type AnalysisResult struct {
	Video VideoStream   `json:"video"`
	Audio []AudioStream `json:"audio"`
	Sub   []SubStream   `json:"sub"`
}

type EpisodeMetadata struct {
	Episode string `json:"episode"`
	Season  string `json:"season"`
}
