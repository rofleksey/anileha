package dao

type QueryRequestDao struct {
	Query string `json:"query" binding:"required"`
}

type StartTorrentRequestDao struct {
	TorrentId     uint   `json:"torrentId" binding:"required"`
	FileIndices   []int  `json:"fileIndices"`
	PrefAudioLang string `json:"prefAudioLang"`
	PrefSubLang   string `json:"prefSubLang"`
}

type StartConversionFileChanPrefData struct {
	Disable bool   `json:"disable"`
	Stream  *int   `json:"stream"`
	File    string `json:"file"`
	Lang    string `json:"lang"`
}

type StartConversionFilePrefData struct {
	Index int                             `json:"index" binding:"required"`
	Audio StartConversionFileChanPrefData `json:"audio" binding:"required"`
	Sub   StartConversionFileChanPrefData `json:"sub" binding:"required"`
}

type StartConversionRequestDao struct {
	TorrentId uint                          `json:"torrentId" binding:"required"`
	Files     []StartConversionFilePrefData `json:"files" binding:"required"`
}

type TorrentWithFileIndexRequestDao struct {
	TorrentId uint `json:"torrentId" binding:"required"`
	FileIndex int  `json:"fileIndex" binding:"required"`
}

type IdRequestDao struct {
	Id uint `json:"id" binding:"required"`
}

type NewUserRequestDao struct {
	User  string `json:"user" binding:"required"`
	Pass  string `json:"pass" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type AuthRequestDao struct {
	User string `json:"user" binding:"required"`
	Pass string `json:"pass" binding:"required"`
}
