package utils

type Pagination struct {
	Skip  int `form:"skip" binding:"min=0"`
	Limit int `form:"limit" binding:"min=10,max=50"`
}

type ListPage[T interface{}] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}
