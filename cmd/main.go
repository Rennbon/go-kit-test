package main

import (
	"github.com/Rennbon/donself/application"
	"github.com/Rennbon/donself/config"
	"github.com/Rennbon/donself/domain"
	"github.com/Rennbon/donself/middleware"
	"github.com/Rennbon/donself/pb"
	"github.com/go-kit/kit/log"
	kitlogrus "github.com/go-kit/kit/log/logrus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/urfave/cli"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	consulapi "github.com/hashicorp/consul/api"
	"net"
	"net/http"
	"os"
)

func main() {
	app := newApp()
	app.Run(os.Args)
}

const configPath = "c"

func newApp() (app *cli.App) {
	app = cli.NewApp()
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  configPath,
			Usage: "config path",
		},
	}
	return app

}

func run(cliCtx *cli.Context) {
	logger, loggerG := newLogger()
	cnfPath := cliCtx.String(configPath)
	//本地调试用

	cnfPath = "/justdo/bc/donself/config/config.toml"
	cnf, err := config.DecodeConfig(cnfPath)
	if err != nil {
		loggerG.Error(err)
		os.Exit(1)
	}

	consulCli, err := newConsulCli(cnf.Consul)
	if err != nil {
		loggerG.Error(err)
		os.Exit(2)
	}

	loggerG.Info("consulapi init success", consulCli)

	//zipkin
	// 创建环境变量
	reporter := zipkinhttp.NewReporter("http://www.rennbon.online:9411/api/v2/spans")
	defer reporter.Close()
	ep, err := zipkin.NewEndpoint("donself", "0.0.0.0:10690")
	if err != nil {
		logger.Log("tracer endpoint err:", err)
		os.Exit(5)
	}

	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(ep), zipkin.WithSharedSpans(true))
	if err != nil {
		logger.Log("tracer zptracer err:", err)
		os.Exit(5)
	}
	tracer := zipkinot.Wrap(nativeTracer)
	opentracing.SetGlobalTracer(tracer)
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

func newLogger() (log.Logger, *logrus.Logger) {
	logrusLogger := logrus.New()
	//设置控制台输出
	logrusLogger.Out = os.Stdout
	logrusLogger.Level = logrus.TraceLevel
	logrusLogger.Formatter = &logrus.TextFormatter{TimestampFormat: "02-01-2006 15:04:05", FullTimestamp: true, ForceColors: true}
	logger := kitlogrus.NewLogrusLogger(logrusLogger)
	return logger, logrusLogger
}
func newConsulCli(cnf *config.ConsulConfig) (*consulapi.Client, error) {
	c := &consulapi.Config{
		Address:    cnf.Address,
		Scheme:     cnf.Scheme,
		Datacenter: cnf.Datacenter,
		WaitTime:   cnf.WaitTime,
		HttpAuth: &consulapi.HttpBasicAuth{
			Username: cnf.User,
			Password: cnf.Password,
		},
	}
	if cnf.TLSconfig != nil && cnf.TLSconfig.Enable {
		c.TLSConfig = consulapi.TLSConfig{
			Address:            cnf.TLSconfig.Address,
			CAFile:             cnf.TLSconfig.CAFile,
			CAPath:             cnf.TLSconfig.CAPath,
			CertFile:           cnf.TLSconfig.CertFile,
			KeyFile:            cnf.TLSconfig.KeyFile,
			InsecureSkipVerify: true,
		}
	}
	return consulapi.NewClient(c)
}
