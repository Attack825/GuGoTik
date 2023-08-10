package main

import (
	"GuGoTik/src/constant/config"
	"GuGoTik/src/constant/strings"
	"GuGoTik/src/extra/tracing"
	"GuGoTik/src/models"
	"GuGoTik/src/rpc/feed"
	"GuGoTik/src/rpc/publish"
	"GuGoTik/src/storage/database"
	"GuGoTik/src/storage/file"
	"GuGoTik/src/utils/consul"
	"GuGoTik/src/utils/logging"
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bakape/thumbnailer/v2"
	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	"image/jpeg"
	"io"
	"net/http"
)

var FeedClient feed.FeedServiceClient

func init() {
	feedErr := consul.RegisterConsul(config.FeedRpcServerName, config.FeedRpcServerPort)
	if feedErr != nil {
		logging.Logger.WithFields(logrus.Fields{
			"err": feedErr,
		}).Errorf("User Service meet trouble.")
	}
}

// getThumbnail Generate JPEG thumbnail from video
func getThumbnail(input io.ReadSeeker) ([]byte, error) {
	_, thumb, err := thumbnailer.Process(input, thumbnailer.Options{})
	if err != nil {
		return nil, errors.New("failed to create thumbnail")
	}
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, thumb, nil)
	if err != nil {
		return nil, errors.New("failed to create buffer")
	}
	return buf.Bytes(), nil
}

// PublishServiceImpl implements the last service interface defined in the IDL.
type PublishServiceImpl struct {
	publish.PublishServiceServer
}

// CreateVideo implements the PublishServiceImpl interface.
func (s *PublishServiceImpl) CreateVideo(ctx context.Context, req *publish.CreateVideoRequest) (resp *publish.CreateVideoResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "PublishServiceImpl.CreateVideo")
	defer span.End()
	logger := logging.LogService("PublishServiceImpl.CreateVideo").WithContext(ctx)

	detectedContentType := http.DetectContentType(req.Data)
	if detectedContentType != "video/mp4" {
		logger.WithFields(logrus.Fields{
			"content_type": detectedContentType,
		}).Debug("invalid content type")
		resp = &publish.CreateVideoResponse{
			StatusCode: strings.PublishServiceInnerErrorCode,
			StatusMsg:  strings.PublishServiceInnerError,
		}
		return
	}
	// byte[] -> reader
	reader := bytes.NewReader(req.Data)

	// V7 based on timestamp
	generatedUUID, err := uuid.NewV7()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"err": err,
		}).Debug("error generating uuid")
		resp = &publish.CreateVideoResponse{
			StatusCode: strings.PublishServiceInnerErrorCode,
			StatusMsg:  strings.PublishServiceInnerError,
		}
		return
	}
	logger = logger.WithFields(logrus.Fields{
		"uuid": generatedUUID,
	})
	logger.Debug("generated uuid")

	// Upload video file
	fileName := fmt.Sprintf("%d/%s.%s", req.ActorId, generatedUUID.String(), "mp4")
	_, err = file.Upload(ctx, fileName, reader)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"file_name": fileName,
			"err":       err,
		}).Debug("failed to upload video")
		resp = &publish.CreateVideoResponse{
			StatusCode: strings.PublishServiceInnerErrorCode,
			StatusMsg:  strings.PublishServiceInnerError,
		}
		return
	}
	logger.WithFields(logrus.Fields{
		"file_name": fileName,
	}).Debug("uploaded video")

	// Generate thumbnail
	coverName := fmt.Sprintf("%d/%s.%s", req.ActorId, generatedUUID.String(), "jpg")
	thumbData, err := getThumbnail(reader)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"file_name":  fileName,
			"cover_name": coverName,
			"err":        err,
		}).Debug("failed to create thumbnail")
		resp = &publish.CreateVideoResponse{
			StatusCode: strings.PublishServiceInnerErrorCode,
			StatusMsg:  strings.PublishServiceInnerError,
		}
		return
	}
	logger.WithFields(logrus.Fields{
		"cover_name": coverName,
		"data_size":  len(thumbData),
	}).Debug("generated thumbnail")

	// Upload thumbnail
	_, err = file.Upload(ctx, coverName, bytes.NewReader(thumbData))
	if err != nil {
		logger.WithFields(logrus.Fields{
			"file_name":  fileName,
			"cover_name": coverName,
			"err":        err,
		}).Debug("failed to upload cover")
		resp = &publish.CreateVideoResponse{
			StatusCode: strings.PublishServiceInnerErrorCode,
			StatusMsg:  strings.PublishServiceInnerError,
		}
		return
	}
	logger.WithFields(logrus.Fields{
		"cover_name": coverName,
		"data_size":  len(thumbData),
	}).Debug("uploaded thumbnail")

	publishModel := models.Video{
		UserId:    int64(req.ActorId),
		FileName:  fileName,
		CoverName: coverName,
		Title:     req.Title,
	}

	err = database.Client.WithContext(ctx).Create(&publishModel).Error
	if err != nil {
		logger.WithFields(logrus.Fields{
			"file_name":  fileName,
			"cover_name": coverName,
			"err":        err,
		}).Debug("failed to create db entry")
		resp = &publish.CreateVideoResponse{
			StatusCode: strings.PublishServiceInnerErrorCode,
			StatusMsg:  strings.PublishServiceInnerError,
		}
		return
	}
	logger.WithFields(logrus.Fields{
		"entry": publishModel,
	}).Debug("saved db entry")

	resp = &publish.CreateVideoResponse{StatusCode: 0, StatusMsg: strings.ServiceOK}
	logger.WithFields(logrus.Fields{
		"response": resp,
	}).Debug("all process done, ready to launch response")
	return
}

// ListVideo implements the PublishServiceImpl interface.
func (s *PublishServiceImpl) ListVideo(ctx context.Context, req *publish.ListVideoRequest) (resp *publish.ListVideoResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "PublishServiceImpl.ListVideo")
	defer span.End()
	logger := logging.LogService("PublishServiceImpl.ListVideo").WithContext(ctx)

	var videos []models.Video
	result := database.Client.WithContext(ctx).
		Where("user_id = ?", req.UserId).
		Order("created_at DESC").
		Find(&videos).Error
	if result.Error != nil {
		logger.WithFields(logrus.Fields{
			"err": err,
		}).Debug("failed to query video")
		resp = &publish.ListVideoResponse{
			StatusCode: strings.PublishServiceInnerErrorCode,
			StatusMsg:  strings.PublishServiceInnerError,
		}
		return
	}

	videoIds := make([]uint64, 0, len(videos))
	for _, video := range videos {
		videoIds = append(videoIds, uint64(video.ID))
	}

	queryVideoResp, err := FeedClient.QueryVideos(ctx, &feed.QueryVideosRequest{
		ActorId:  req.ActorId,
		VideoIds: videoIds,
	})

	logger.WithFields(logrus.Fields{
		"response": resp,
	}).Debug("all process done, ready to launch response")
	return &publish.ListVideoResponse{
		StatusCode: strings.ServiceOKCode,
		StatusMsg:  strings.ServiceOK,
		VideoList:  queryVideoResp.VideoList,
	}, nil
}

// CountVideo implements the PublishServiceImpl interface.
func (s *PublishServiceImpl) CountVideo(ctx context.Context, req *publish.CountVideoRequest) (resp *publish.CountVideoResponse, err error) {
	ctx, span := tracing.Tracer.Start(ctx, "PublishServiceImpl.CountVideo")
	defer span.End()
	logger := logging.LogService("PublishServiceImpl.CountVideo").WithContext(ctx)
	var count int64
	result := database.Client.WithContext(ctx).Where("user_id = ?", req.UserId).Count(&count).Error
	if result.Error != nil {
		logger.WithFields(logrus.Fields{
			"err": err,
		}).Debug("failed to query video")
		resp = &publish.CountVideoResponse{
			StatusCode: strings.PublishServiceInnerErrorCode,
			StatusMsg:  strings.PublishServiceInnerError,
		}
		return
	}

	resp = &publish.CountVideoResponse{
		StatusCode: strings.ServiceOKCode,
		StatusMsg:  strings.ServiceOK,
		Count:      uint64(count),
	}
	return
}
