package model

type HelloWorld struct {
	BaseModel
	Title       string `gorm:"type:varchar(255);not null" json:"title"`
	Description string `gorm:"type:text" json:"description"`
}
