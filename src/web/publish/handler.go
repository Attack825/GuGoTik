package publish

import (
	"GuGoTik/src/constant/config"
	"GuGoTik/src/constant/strings"
	"GuGoTik/src/extra/tracing"
	"GuGoTik/src/rpc/publish"
	"GuGoTik/src/utils/logging"
	"GuGoTik/src/web/models"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"mime/multipart"
	"net/http"
)

var Client publish.PublishServiceClient

func init() {
	//Dial creates a client connection to the given target.
	conn, err := grpc.Dial(
		fmt.Sprintf("consul://%s/%s?wait=15s", config.EnvCfg.ConsulAddr, config.PublishRpcServerName),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
	)

	if err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Errorf("Build PublishServer Client meet trouble")
	}
	Client = publish.NewPublishServiceClient(conn)
}

func paramValidate(c *gin.Context) (err error) {
	_, span := tracing.Tracer.Start(c.Request.Context(), "Publish-paramValidate")
	defer span.End()
	var wrappedError error
	form, err := c.MultipartForm()
	if err != nil {
		wrappedError = fmt.Errorf("invalid form: %w", err)
	}
	title := form.Value["title"]
	if len(title) <= 0 {
		wrappedError = fmt.Errorf("not title")
	}

	data := form.File["data"]
	if len(data) <= 0 {
		wrappedError = fmt.Errorf("not data")
	}
	if wrappedError != nil {
		return wrappedError
	}
	return nil
}

func ActionHandle(c *gin.Context) {
	_, span := tracing.Tracer.Start(c.Request.Context(), "Publish-ActionHandle")
	defer span.End()
	logger := logging.LogService("GateWay.PublishAction").WithContext(c.Request.Context())

	//检查参数
	if err := paramValidate(c); err != nil {
		logging.Logger.WithFields(logrus.Fields{
			"err": err,
		}).Errorf("Publish paramValidate meet trouble")
		return
	}

	var req models.PublishActionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, models.ListVideoRes{
			StatusCode: strings.GateWayParamsErrorCode,
			StatusMsg:  strings.GateWayParamsError,
			NextTime:   nil,
			VideoList:  nil,
		})
	}

	res, err := Client.CreateVideo(c.Request.Context(), &publish.CreateVideoRequest{})
	title := req.Title
	file := req.Data
	if err != nil {
		logger.WithFields(logrus.Fields{
			"Data":  file,
			"Title": title,
		}).Warnf("Error when trying to connect with PublishService")
		c.JSON(http.StatusOK, res)
		return
	}
	//检查文件
	opened, _ := file.Open()
	defer func(opened multipart.File) {
		err := opened.Close()
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err,
			}).Errorf("opened.Close() failed")
		}
	}(opened)
	var data = make([]byte, file.Size)
	readSize, err := opened.Read(data)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Errorf("Open file failed")
		return
	}
	if readSize != int(file.Size) {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Errorf("Size not match")
		return
	}
	logger.WithFields(logrus.Fields{
		"title":    title,
		"dataSize": len(data),
	}).Debugf("Success create video")
	c.JSON(
		http.StatusOK,
		res,
	)

}

func ListHandle(c *gin.Context) {
	_, span := tracing.Tracer.Start(c.Request.Context(), "Publish-ListHandle")
	defer span.End()
	logger := logging.LogService("GateWay.PublishList").WithContext(c.Request.Context())
	var req models.PublishListReq
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, models.ListVideoRes{
			StatusCode: strings.GateWayParamsErrorCode,
			StatusMsg:  strings.GateWayParamsError,
			NextTime:   nil,
			VideoList:  nil,
		})
	}

	res, err := Client.ListVideo(c.Request.Context(), &publish.ListVideoRequest{})
	if err != nil {
		logger.WithFields(logrus.Fields{
			"UserId": req.UserId,
		}).Warnf("Error when trying to connect with PublishService")
		c.JSON(http.StatusOK, res)
		return
	}
	userId := req.UserId
	logger.WithFields(logrus.Fields{
		"UserId": userId,
	}).Infof("Publish List videos")

	c.JSON(http.StatusOK, res)
}
