package dao

type SeriesDao struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Query       *string `json:"query"`
	ThumbnailId uint    `json:"thumbnail_id" binding:"required"`
}

type NameDao struct {
	Name string `json:"name" binding:"required"`
}

type IdDao struct {
	Id string `json:"id" binding:"required"`
}
