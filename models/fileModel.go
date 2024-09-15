package models

import "time"

type File struct {
    ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
    Email      string    `json:"email" gorm:"not null"`
    FileName   string    `json:"file_name" gorm:"not null"`
    FileURL    string    `json:"file_url" gorm:"not null"`
    UploadDate time.Time `json:"upload_date" gorm:"autoCreateTime"`
    FileSize   int64     `json:"file_size"`
    FileType  string     `json:"file_type" gorm:"not null"`
}
