package vo

import "time"

type CreateReq struct {
	Title       string `json:"title" binding:"required,max=255"`
	Description string `json:"description"`
}

type UpdateReq struct {
	Title       *string `json:"title" binding:"omitempty,max=255"`
	Description *string `json:"description"`
}

type ListReq struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=100"`
}

type HelloWorldResp struct {
	ID          uint      `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListResp struct {
	Total int64             `json:"total"`
	Items []*HelloWorldResp `json:"items"`
}
