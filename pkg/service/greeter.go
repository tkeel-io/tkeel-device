package service

import (
	"context"

	v1 "device/api/helloworld/v1"
)

// GreeterService is a greeter service.
type GreeterService struct {
	v1.UnimplementedGreeterServer
}

// NewGreeterService new a greeter service.
func NewGreeterService() *GreeterService {
	return &GreeterService{
		UnimplementedGreeterServer: v1.UnimplementedGreeterServer{},
	}
}

// SayHello implements helloworld.GreeterServer
func (s *GreeterService) SayHello(ctx context.Context, in *v1.HelloRequest) (*v1.HelloResponse, error) {
	return &v1.HelloResponse{Message: "Hello " + in.GetName()}, nil
}
