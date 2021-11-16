package service

import (
	"context"

	v1 "device/api/openapi/v1"
	"device/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// OpenapiService is a openapi service.
type OpenapiService struct {
	v1.UnimplementedOpenapiServer
}

// NewOpenapiService new a openapi service.
func NewOpenapiService() *OpenapiService {
	return &OpenapiService{
		UnimplementedOpenapiServer: v1.UnimplementedOpenapiServer{},
	}
}

// AddonsIdentify implements AddonsIdentify.OpenapiServer
func (s *OpenapiService) AddonsIdentify(ctx context.Context, in *v1.AddonsIdentifyRequest) (*v1.AddonsIdentifyResponse, error) {
	return &v1.AddonsIdentifyResponse{
		Res: util.GetV1ResultBadRequest("not declare addons"),
	}, nil
}

// Identify implements Identify.OpenapiServer
func (s *OpenapiService) Identify(ctx context.Context, in *emptypb.Empty) (*v1.IdentifyResponse, error) {
	return &v1.IdentifyResponse{
		Res:          util.GetV1ResultOK(),
		PluginID:     "tkeel-hello",
		Version:      "v0.2.0",
		TkeelVersion: "v0.2.0",
	}, nil
}

// Status implements Status.OpenapiServer
func (s *OpenapiService) Status(ctx context.Context, in *emptypb.Empty) (*v1.StatusResponse, error) {
	return &v1.StatusResponse{
		Res:    util.GetV1ResultOK(),
		Status: v1.PluginStatus_runing,
	}, nil
}

// TenantBind implements TenantBind.OpenapiServer
func (s *OpenapiService) TenantBind(ctx context.Context, in *v1.TenantBindRequst) (*v1.TenantBindResponse, error) {
	return &v1.TenantBindResponse{
		Res: util.GetV1ResultOK(),
	}, nil
}

// TenantUnbind implements TenantUnbind.OpenapiServer
func (s *OpenapiService) TenantUnbind(ctx context.Context, in *v1.TenantUnbindRequst) (*v1.TenantUnbindResponse, error) {
	return &v1.TenantUnbindResponse{
		Res: util.GetV1ResultOK(),
	}, nil
}
