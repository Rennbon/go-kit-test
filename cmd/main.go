package main

import (
	"github.com/Rennbon/donself/application"
	"github.com/Rennbon/donself/domain"
	"github.com/Rennbon/donself/middleware"
	"github.com/Rennbon/donself/pb"
	kitlogrus "github.com/go-kit/kit/log/logrus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	logrusLogger := logrus.New()
	//设置控制台输出
	logrusLogger.Out = os.Stdout
	logrusLogger.Level = logrus.TraceLevel
	logrusLogger.Formatter = &logrus.TextFormatter{TimestampFormat: "02-01-2006 15:04:05", FullTimestamp: true, ForceColors: true}
	logger := kitlogrus.NewLogrusLogger(logrusLogger)

	//zipkin
	// 创建环境变量
	zipkinURL := "http://www.rennbon.online:9411/api/v2/spans"
	zipkinBridge := true
	var zipkinTracer *zipkin.Tracer
	{
		var (
			err         error
			hostPort    = ":10690"
			serviceName = "donself"
			//zipkin上报管理器
			reporter = zipkinhttp.NewReporter(zipkinURL)
		)
		defer reporter.Close()
		zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
		//zipkin跟踪器
		zipkinTracer, err = zipkin.NewTracer(
			reporter, zipkin.WithLocalEndpoint(zEP),
		)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
		if !zipkinBridge {
			logger.Log("tracer", "Zipkin", "type", "Native", "URL", zipkinURL)
		}
	}
	// Determine which OpenTracing tracer to use. We'll pass the tracer to all the
	// components that use it, as a dependency.
	var tracer stdopentracing.Tracer
	{
		if zipkinBridge && zipkinTracer != nil {
			logger.Log("tracer", "Zipkin", "type", "OpenTracing", "URL", zipkinURL)
			tracer = zipkinot.Wrap(zipkinTracer)
			zipkinTracer = nil // do not instrument with both native tracer and opentracing bridge
		}
	}

	//初始化svc逻辑层
	svc := domain.NewDonselfDomain()
	svc = middleware.WithMetric(svc)
	svc = middleware.WithLogging(svc, logger)
	grpcServer := application.NewDoneselfServer(svc, logger, tracer)

	// The gRPC listener mounts the Go kit gRPC server we created.
	listener, err := net.Listen("tcp", "0.0.0.0:10690")
	if err != nil {
		os.Exit(1)
	}
	defer listener.Close()

	http.Handle("/metrics", promhttp.Handler())

	//todo 可以设置grpc基础配置
	s := grpc.NewServer(
		grpc.ConnectionTimeout(time.Minute),
		grpc.UnaryInterceptor(kitgrpc.Interceptor),
	)
	pb.RegisterDoneselfServer(s, grpcServer)
	logrus.Info("grpc start ", listener.Addr())

	go func() {
		err = http.ListenAndServe("0.0.0.0:10691", nil)
		if err != nil {
			os.Exit(2)
		}
	}()

	go func() {
		if err = s.Serve(listener); err != nil {
			logrus.Error(err)
			os.Exit(3)
		}
	}()
	select {}
}
