package dao

type SeriesRequestDao struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Query       *string `json:"query"`
	ThumbnailId uint    `json:"thumbnail_id" binding:"required"`
}

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
