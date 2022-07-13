package dao

type SeriesRequestDao struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Query       *string `json:"query"`
	ThumbnailId uint    `json:"thumbnail_id" binding:"required"`
}
