package main

import (
	"context"
	"github.com/Rennbon/donself/log"
	"github.com/Rennbon/donself/pb"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	"github.com/go-kit/kit/sd/lb"
	"io"

	kitlog "github.com/go-kit/kit/log"
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/tracing/opentracing"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/hashicorp/consul/api"
	nativeopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
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

func factoryFor(tracer nativeopentracing.Tracer, logger kitlog.Logger) sd.Factory {
	return func(instance string) (endpoint.Endpoint, io.Closer, error) {
		conn, err := grpc.Dial(instance, grpc.WithInsecure())
		if err != nil {
			return nil, nil, err
		}
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
		return c, conn, nil
	}
}

//客户端封装没有附加parentId，如果要的话需要改源码
func NewGrpcClient(tracer nativeopentracing.Tracer, kitClient kitconsul.Client, logger kitlog.Logger) Client {
	var (
		serviceName  = "donself"
		tags         = []string{}
		passingOnly  = true
		retryTimeout = time.Millisecond * 500
		retryNum     = 3
		endpoints    = &client{}
	)

	instancer := kitconsul.NewInstancer(kitClient, logger, serviceName, tags, passingOnly)
	factory := factoryFor(tracer, logger)
	endpointer := sd.NewEndpointer(instancer, factory, logger)
	balancer := lb.NewRoundRobin(endpointer)
	retry := lb.Retry(retryNum, retryTimeout, balancer)
	endpoints.AllMyTargetsEndpoint = retry
	return endpoints
	/*	var options []kitgrpc.ClientOption

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
		return &client{AllMyTargetsEndpoint: c}*/
}

func (c *client) AllMyTargets(ctx context.Context, in *pb.AllMyTargetsRequest) (*pb.AllMyTargetsResponse, error) {
	resp, err := c.AllMyTargetsEndpoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.AllMyTargetsResponse), nil
}
func main() {

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

	/*  这里不重要了，服务发现只需要有服务名和consul地址
	address := "www.rennbon.online:10690"
	address = "127.0.0.1:10690"
	*/
	//ctx1, _ := context.WithTimeout(context.Background(), time.Second*5)
	/*//credentials.n
	//creds, _ := credentials.NewClientTLSFromFile("/justdo/bc/donself/config/testdata/server.pem", "")
	conn, err := grpc.DialContext(ctx1, address, grpc.WithInsecure()) //, grpc.WithTransportCredentials(creds))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	c := NewGrpcClient(tracer, conn, log.Logger)*/

	//consul 服务发现
	consulConfig := api.DefaultConfig()
	consulCli, err := api.NewClient(consulConfig)
	if err != nil {
		log.Logrus.Error(err)
		os.Exit(1)
	}

	cli := kitconsul.NewClient(consulCli)
	c := NewGrpcClient(tracer, cli, log.Logger)

	req := &pb.AllMyTargetsRequest{
		PageIndex: 1,
		PageSize:  10,
	}
	ctx, _ := context.WithTimeout(context.Background(), time.Minute*10)
	for i := uint32(1); i < 1000; i++ {
		time.Sleep(time.Millisecond * 250)
		go func(num uint32) {
			req.PageIndex = num

			res, err := c.AllMyTargets(ctx, req)
			if err != nil {
				log.Logger.Log("num:", i, err)
			} else {
				log.Logger.Log("num:", i, res)
			}
		}(i)
	}
	select {}
}
