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
	FileIndex uint `json:"fileIndex" binding:"required"`
}

type TorrentIdRequestDao struct {
	TorrentId uint `json:"torrentId" binding:"required"`
}

type ConvertIdRequestDao struct {
	ConversionId uint `json:"conversionId" binding:"required"`
}
