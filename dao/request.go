package dao

type QueryRequestDao struct {
	Query string `json:"query" binding:"required"`
}

type StartTorrentRequestDao struct {
	Id          uint  `json:"id" binding:"required"`
	FileIndices []int `json:"fileIndices"`
}

type StartConversionFileChanPrefData struct {
	Disable bool   `json:"disable"`
	Stream  *int   `json:"stream"`
	File    string `json:"file"`
	Lang    string `json:"lang"`
}

type StartConversionFilePrefData struct {
	Index   int                             `json:"index"`
	Episode string                          `json:"episode"`
	Season  string                          `json:"season"`
	Audio   StartConversionFileChanPrefData `json:"audio" binding:"required"`
	Sub     StartConversionFileChanPrefData `json:"sub" binding:"required"`
}

type StartConversionRequestDao struct {
	TorrentId uint                          `json:"torrentId" binding:"required"`
	Files     []StartConversionFilePrefData `json:"files" binding:"required"`
}

type TorrentWithFileIndexRequestDao struct {
	Id        uint `json:"id" binding:"required"`
	FileIndex int  `json:"fileIndex"`
}

type IdRequestDao struct {
	Id uint `json:"id" binding:"required"`
}

type NewUserRequestDao struct {
	User  string `json:"user" binding:"required"`
	Pass  string `json:"pass" binding:"required"`
	Email string `json:"email" binding:"required"`
}

type OwnerCreateUserRequestDao struct {
	Login string   `json:"login" binding:"required"`
	Pass  string   `json:"pass" binding:"required"`
	Email string   `json:"email" binding:"required"`
	Roles []string `json:"roles"`
}

type ModifyUserRequestDao struct {
	Name  string `json:"name"`
	Pass  string `json:"pass"`
	Email string `json:"email"`
}

type AuthRequestDao struct {
	User string `json:"user" binding:"required"`
	Pass string `json:"pass" binding:"required"`
}
