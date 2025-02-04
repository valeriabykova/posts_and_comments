package postgres

type Post struct {
	ID            int32   `gorm:"primaryKey"`
	Title         string  `gorm:"not null"`
	Body          string  `gorm:"not null"`
	AllowComments bool    `gorm:"not null"`
	Author        string  `gorm:"not null"`
	Comments      Comment `gorm:"foreignKey:PostID"`
}

type Comment struct {
	ID       int32     `gorm:"primaryKey"`
	PostID   int32     `gorm:"not null;index"`
	ParentID *int32    `gorm:"index"`
	Body     string    `gorm:"not null"`
	Author   string    `gorm:"not null"`
	Replies  []Comment `gorm:"foreignKey:ParentID"`
}
