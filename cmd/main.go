package main

import (
	"fmt"
	"github.com/Rennbon/donself/application"
	"github.com/Rennbon/donself/config"
	"github.com/Rennbon/donself/health"
	"github.com/Rennbon/donself/pb"
	"github.com/Rennbon/donself/service"
	"github.com/go-kit/kit/log"
	kitlogrus "github.com/go-kit/kit/log/logrus"
	kitconsul "github.com/go-kit/kit/sd/consul"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/opentracing/opentracing-go"
	zipkinot "github.com/openzipkin-contrib/zipkin-go-opentracing"
	"github.com/openzipkin/zipkin-go"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"net/http"
	"os"
	"strconv"
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
	//logger 初始化
	logger, loggerG := newLogger()
	cnfPath := cliCtx.String(configPath)
	//本地调试用

	cnfPath = "/justdo/bc/donself/config/config.toml"
	cnf, err := config.DecodeConfig(cnfPath)
	if err != nil {
		loggerG.Error(err)
		os.Exit(1)
	}
	loggerG.Info(cnf.Server)

	//grpc地址
	port, _ := strconv.Atoi(cnf.Server.Port)
	grpcAddr := fmt.Sprintf("%v:%v", cnf.Server.Host, port)

	//metric指标地址
	metricAddr := fmt.Sprintf("%v:%v", cnf.Server.Host, cnf.Server.MetricsPort)

	//zipkin 创建环境变量
	reporter := zipkinhttp.NewReporter(cnf.Server.ZipkinReporter)
	defer reporter.Close()
	ep, err := zipkin.NewEndpoint(cnf.Server.Name, grpcAddr)
	if err != nil {
		loggerG.WithField("zipkin", "newEndPoint").Error(err)
		os.Exit(5)
	}
	nativeTracer, err := zipkin.NewTracer(reporter, zipkin.WithLocalEndpoint(ep), zipkin.WithSharedSpans(true))
	if err != nil {
		loggerG.WithField("zipkin", "newTracer").Error(err)
		os.Exit(5)
	}
	tracer := zipkinot.Wrap(nativeTracer)
	opentracing.SetGlobalTracer(tracer)

	//初始化svc逻辑层 并开始套壳，一层包一层
	svc := service.NewDonselfService()
	//svc = middleware.WithMetric(svc)
	//svc = middleware.WithLogging(svc, logger)
	grpcServer := application.NewDoneselfServer(svc, logger, tracer)

	// 创建grpc tcp 监听
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		loggerG.WithField("net", "listen").Error(err)
		os.Exit(1)
	}
	defer listener.Close()

	//grpc 启动
	//creds, _ := credentials.NewServerTLSFromFile(cnf.Server.CertFile, cnf.Server.KeyFile)
	s := grpc.NewServer(
		grpc.UnaryInterceptor(kitgrpc.Interceptor),
		//grpc.Creds(creds),
	)
	//注册测试服务
	pb.RegisterDoneselfServer(s, grpcServer)
	//服务发现用健康检查逻辑注册
	grpc_health_v1.RegisterHealthServer(s, &health.HealthImpl{})
	logrus.Info("grpc start ", listener.Addr())
	go func() {
		if err = s.Serve(listener); err != nil {
			loggerG.WithField("grpc", "serve").Error(err)
			os.Exit(3)
		}
	}()

	//metrics http 启动
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err = http.ListenAndServe(metricAddr, nil)
		if err != nil {
			loggerG.WithField("http", "listenAndServe").Error(err)
			os.Exit(2)
		}
	}()

	//consul服务发现逻辑注册
	reg, err := newConsulRegister(cnf.Consul, &checkConfig{
		serviceName: cnf.Server.Name,
		port:        port,
		ip:          cnf.Server.Local, //因为监听consul和本服务都在一台服务器的docker中
		interval:    cnf.Server.HealthInterval.String(),
		deregister:  cnf.Server.Deregister.String(),
	}, logger)
	if err != nil {
		loggerG.WithField("consul", "newConsulRegister").Error(err)
		os.Exit(9)
	}
	reg.Register()
	defer reg.Deregister()

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

type checkConfig struct {
	serviceName string
	port        int
	ip          string
	interval    string
	deregister  string
}

func newConsulRegister(cnf *config.ConsulConfig, checkCnf *checkConfig, logger log.Logger) (*kitconsul.Registrar, error) {
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
	//本地默认配置
	//c = consulapi.DefaultConfig()
	if cnf.TLSconfig != nil && cnf.TLSconfig.Enable {
		c.TLSConfig = consulapi.TLSConfig{
			Address:            cnf.TLSconfig.Address,
			CAFile:             cnf.TLSconfig.CAFile,
			CAPath:             cnf.TLSconfig.CAPath,
			CertFile:           cnf.TLSconfig.CertFile,
			KeyFile:            cnf.TLSconfig.KeyFile,
			InsecureSkipVerify: false,
		}
	}

	consulCli, err := consulapi.NewClient(c)
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("%v-%v-%v", checkCnf.serviceName, checkCnf.ip, checkCnf.port)
	reg := &consulapi.AgentServiceRegistration{
		ID:      id,
		Name:    checkCnf.serviceName, //fmt.Sprintf("grpc.health.v1.%v", checkCnf.serviceName),
		Port:    checkCnf.port,
		Tags:    []string{"this is tag"},
		Address: checkCnf.ip,
		Check: &consulapi.AgentServiceCheck{
			Interval:                       checkCnf.interval,
			GRPC:                           fmt.Sprintf("%s:%d/%s", checkCnf.ip, checkCnf.port, checkCnf.serviceName),
			DeregisterCriticalServiceAfter: checkCnf.deregister,
		},
	}
	kitcli := kitconsul.NewClient(consulCli)
	register := kitconsul.NewRegistrar(kitcli, reg, logger)
	return register, nil
}

//docker环境下不要中这个，container内部localIP和linux localIP是隔离的
func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}
