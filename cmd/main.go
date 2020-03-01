package main

import (
	"github.com/Rennbon/donself/application"
	"github.com/Rennbon/donself/domain"
	"github.com/Rennbon/donself/middleware"
	"github.com/Rennbon/donself/pb"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {

	//初始化svc逻辑层
	svc := domain.NewDonselfDomain()
	svc = middleware.WithMetric(svc)
	grpcServer := application.NewDoneselfServer(svc)
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
	)
	pb.RegisterDoneselfServer(s, grpcServer)
	log.Info("grpc start ", listener.Addr())

	go func() {
		err = http.ListenAndServe("0.0.0.0:10691", nil)
		if err != nil {
			os.Exit(2)
		}
	}()

	go func() {
		if err = s.Serve(listener); err != nil {
			log.Error(err)
			os.Exit(3)
		}
	}()
	select {}
}
