package dao

type QueryRequestDao struct {
	Query string `json:"query" binding:"required"`
}

type TorrentWithFileIndicesRequestDao struct {
	TorrentId   uint   `json:"torrentId" binding:"required"`
	FileIndices string `json:"fileIndices" binding:"required"`
}

type TorrentWithFileIndexRequestDao struct {
	TorrentId uint `json:"torrentId" binding:"required"`
	FileIndex uint `json:"fileIndex"`
}

type TorrentIdRequestDao struct {
	TorrentId uint `json:"torrentId" binding:"required"`
}

type ConvertIdRequestDao struct {
	ConversionId uint `json:"conversionId" binding:"required"`
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
