package dao

type SeriesRequestDao struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Query       *string `json:"query"`
	ThumbnailId uint    `json:"thumbnail_id" binding:"required"`
}

type TorrentStartRequestDao struct {
	TorrentId   uint   `json:"torrentId" binding:"required"`
	FileIndices string `json:"fileIndices" binding:"required"`
}

type TorrentIdRequestDao struct {
	TorrentId uint `json:"torrentId" binding:"required"`
}

type ConvertStartRequestDao struct {
	SeriesId      uint `json:"seriesId" binding:"required"`
	TorrentFileId uint `json:"torrentFileId" binding:"required"`
}

type ConvertIdRequestDao struct {
	ConversionId uint `json:"conversionId" binding:"required"`
}
