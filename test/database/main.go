package main

import (
	"GuGoTik/src/constant/strings"
	"GuGoTik/src/models"
	"GuGoTik/src/storage/database"
	"github.com/bytedance/gopkg/util/logger"
	"time"
)

func FindVideos() {
	logger.Debug("FindVideos")
	var videos []*models.Video
	result := database.Client.Where("created_at <= ?", time.Unix(1691457467541, 0)).
		Order("created_at DESC").
		Limit(strings.VideoCount).
		Find(&videos)
	logger.Debug(videos)
	if result.Error != nil {
		logger.Debug("FindVideos error")
		return
	}
}
func FindVideos2() {
	logger.Debug("FindVideos")
	var video models.Video
	err := database.Client.Where("created_at <= ?", time.Unix(1691457467541, 0)).
		Order("created_at DESC").
		Limit(strings.VideoCount).
		Find(&video).Error
	logger.Debug(video)
	if err != nil {
		logger.Debug("FindVideos error")
		return
	}
}

func FindAll() {
	logger.Debug("FindVideos")
	var videos models.Video
	result := database.Client.Find(&videos)
	logger.Debug(videos)
	if result.Error != nil {
		logger.Debug("FindVideos error")
		return
	}
}
func Auth() {
	var user []*models.User
	result := database.Client.Where("created_at <= ?", time.Unix(1691457467541, 0)).Find(&user)
	if result.Error != nil {
		logger.Error("FindUserByUsername error")
	}
	logger.Debug(result)
	logger.Debug(user)
}

func main() {
	logger.Debug(123)
	logger.Debug(456)
	FindVideos()
	FindVideos2()
	FindAll()
	Auth()
	time.Sleep(11100000001111)

}
