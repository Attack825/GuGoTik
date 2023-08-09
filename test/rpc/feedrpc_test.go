package rpc

import (
	"GuGoTik/src/constant/config"
	"GuGoTik/src/rpc/feed"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"testing"
)

// 测试rpc
func TestListVideos(t *testing.T) {
	//初始化客户端
	var Client feed.FeedServiceClient
	req := feed.ListFeedRequest{
		LatestTime: proto.String("2023-08-04T12:34:56.789Z"), // Example timestamp in string format
		ActorId:    proto.Uint32(123),                        // Example user ID as uint32
	}
	//连接服务端
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1%s", config.FeedRpcServerPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	assert.Empty(t, err)
	Client = feed.NewFeedServiceClient(conn)
	//调用服务端方法
	res, err := Client.ListVideos(context.Background(), &req)
	assert.Empty(t, err)
	assert.Equal(t, int32(0), res.StatusCode)
}

// 创建一个 feed.FeedServiceClient 实例，用于与某种类型的服务进行通信。
// 创建一个 feed.ListFeedRequest 请求，其中 LatestTime 字段设置为 nil。
// 通过 gRPC 连接到本地地址并使用不安全的凭证（insecure credentials）。
// 创建一个 feed.NewFeedServiceClient 实例，用于处理与服务的通信。
// 调用 Client.ListVideos 方法，向服务器发送一个请求以获取视频列表。
// 使用断言（assert）来验证测试的条件是否满足：
// 使用 assert.Empty 来确保没有发生错误（err为空）。
// 使用 assert.Equal 来确保服务器返回的状态码（res.StatusCode）等于 0。
