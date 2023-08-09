package models

import (
	"gorm.io/gorm"
)

// Video 视频表 /*
type Video struct {
	UserId    int64  `json:"user_id" gorm:"not null;index"`
	Title     string `json:"title"`
	FileName  string `json:"play_name"`
	CoverName string `json:"cover_name"`
	gorm.Model
}
