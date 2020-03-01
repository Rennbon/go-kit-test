package application

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/transport/grpc"
)

type GrpcBaseProperty struct {
	endpoint.Endpoint
	grpc.DecodeRequestFunc
	grpc.EncodeResponseFunc
}
