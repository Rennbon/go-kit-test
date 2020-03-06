package main

import (
	"context"
	"fmt"
	"github.com/Rennbon/donself/pb"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	kitlogrus "github.com/go-kit/kit/log/logrus"
	"github.com/go-kit/kit/tracing/opentracing"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	nativeopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"os"
	"time"
)

type Client interface {
	AllMyTargets(ctx context.Context, in *pb.AllMyTargetsRequest) (*pb.AllMyTargetsResponse, error)
}
type client struct {
	AllMyTargetsEndpoint endpoint.Endpoint
}

//客户端封装没有附加parentId，如果要的话需要改源码
func NewGrpcClient(tracer nativeopentracing.Tracer, conn *grpc.ClientConn, logger log.Logger) Client {
	var options []kitgrpc.ClientOption

	c := kitgrpc.NewClient(
		conn,
		"pb.Doneself",  //必须和pb相对应，框架会组合ServiceName+method反射，有点坑
		"AllMyTargets", //因为没有直接使用pb.NewXXXClinet()，做了公用化处理
		func(ctx context.Context, request interface{}) (interface{}, error) {
			return request.(*pb.AllMyTargetsRequest), nil
		},
		func(ctx context.Context, response interface{}) (interface{}, error) {
			return response.(*pb.AllMyTargetsResponse), nil
		},
		&pb.AllMyTargetsResponse{},
		append(options, kitgrpc.ClientBefore(opentracing.ContextToGRPC(tracer, logger)))...,
	).Endpoint()
	c = opentracing.TraceClient(tracer, "AllMyTargets")(c)
	return &client{AllMyTargetsEndpoint: c}
}

func (c *client) AllMyTargets(ctx context.Context, in *pb.AllMyTargetsRequest) (*pb.AllMyTargetsResponse, error) {
	resp, err := c.AllMyTargetsEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.AllMyTargetsResponse), nil
}
func main() {
	logrusLogger := logrus.New()
	//设置控制台输出
	logrusLogger.Out = os.Stdout
	logrusLogger.Level = logrus.TraceLevel
	logrusLogger.Formatter = &logrus.TextFormatter{TimestampFormat: "02-01-2006 15:04:05", FullTimestamp: true, ForceColors: true}
	logger := kitlogrus.NewLogrusLogger(logrusLogger)

	reporter := zipkinhttp.NewReporter("http://www.rennbon.online:9411/api/v2/spans")
	defer reporter.Close()
	zep, err := zipkin.NewEndpoint("donself", "")
	if err != nil {
		os.Exit(5)
	}
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(zep), zipkin.WithSharedSpans(true))
	if err != nil {
		os.Exit(5)
	}
	tracer := zipkinot.Wrap(nativeTracer)

	nativeopentracing.SetGlobalTracer(tracer)

	ctx1, _ := context.WithTimeout(context.Background(), time.Second*5)
	address := "www.rennbon.online:10690"
	//address = "127.0.0.1:10690"

	//credentials.n
	//creds, _ := credentials.NewClientTLSFromFile("/justdo/bc/donself/config/testdata/server.pem", "")
	conn, err := grpc.DialContext(ctx1, address, grpc.WithInsecure()) //, grpc.WithTransportCredentials(creds))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	c := NewGrpcClient(tracer, conn, logger)
	req := &pb.AllMyTargetsRequest{
		PageIndex: 1,
		PageSize:  10,
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Minute*10)
	for i := uint32(1); i < 2; i++ {
		time.Sleep(time.Millisecond * 250)
		go func(num uint32) {
			req.PageIndex = num

			res, err := c.AllMyTargets(ctx, req)
			if err != nil {
				logger.Log("num:", i, err)
			} else {
				logger.Log("num:", i, res)
			}
		}(i)
	}
	select {}
}
