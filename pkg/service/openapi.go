package service

import (
	"context"
	"fmt"
	"github.com/tkeel-io/tkeel-device/pkg/service/openapi"

	v1 "github.com/tkeel-io/tkeel-device/api/openapi/v1"
	pb "github.com/tkeel-io/tkeel-device/api/template/v1"
	"github.com/tkeel-io/tkeel-device/pkg/util"
	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
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

// AddonsIdentify implements AddonsIdentify.OpenapiServer.
func (s *OpenapiService) AddonsIdentify(ctx context.Context, in *openapi_v1.AddonsIdentifyRequest) (*openapi_v1.AddonsIdentifyResponse, error) {
	var endpoint string
	openapiCli := openapi.NewDaprClient("3500", "_systenant", "_user_tKeel_system")
	pluginId := in.GetPlugin().GetId()
	for _, addon := range in.ImplementedAddons {
		if addon.GetAddonsPoint() == openapi.DEVICE_SCHEMA_CHANGE {
			sendToPluginID := "keel"
			endpoint = addon.GetImplementedEndpoint()
			method := fmt.Sprintf("/apis/%s/%s", pluginId, endpoint)
			openapiCli.CallAddons(ctx, sendToPluginID, method, nil, &pb.UpdateTemplateResponse{})
		}
	}
	return &openapi_v1.AddonsIdentifyResponse{
		Res: util.GetV1ResultBadRequest("not declare addons"),
	}, nil
}

// Identify implements Identify.OpenapiServer.
func (s *OpenapiService) Identify(ctx context.Context, in *emptypb.Empty) (*openapi_v1.IdentifyResponse, error) {
	profiles := map[string]*openapi_v1.ProfileSchema{
		"device_created_max":  &openapi_v1.ProfileSchema{Type: "number", Title: "创建设备最大数", Default: 10000, MultipleOf: 1, Maximum: 10000, Minimum: 0},
		"device_template_max": &openapi_v1.ProfileSchema{Type: "number", Title: "创建设备模板最大数", Default: 10000, MultipleOf: 1, Maximum: 10000, Minimum: 0},
	}

	return &openapi_v1.IdentifyResponse{
		Res:                     util.GetV1ResultOK(),
		PluginId:                "tkeel-device",
		Version:                 "v0.4.2",
		TkeelVersion:            "v0.4.0",
		DisableManualActivation: true,
		Profiles:                profiles,
		//Profiles : profileArray,
		AddonsPoint: []*openapi_v1.AddonsPoint{
			&openapi_v1.AddonsPoint{
				Name: openapi.DEVICE_SCHEMA_CHANGE,
				Desc: "when device schema change.",
			},
		},
	}, nil
}

// Status implements Status.OpenapiServer.
func (s *OpenapiService) Status(ctx context.Context, in *emptypb.Empty) (*openapi_v1.StatusResponse, error) {
	return &openapi_v1.StatusResponse{
		Res:    util.GetV1ResultOK(),
		Status: openapi_v1.PluginStatus_RUNNING,
	}, nil
}

// TenantEnable implements TenantEnable.OpenapiServer.
func (s *OpenapiService) TenantEnable(ctx context.Context, in *openapi_v1.TenantEnableRequest) (*openapi_v1.TenantEnableResponse, error) {
	return &openapi_v1.TenantEnableResponse{
		Res: util.GetV1ResultOK(),
	}, nil
}

// TenantDisable implements TenantDisable.OpenapiServer.
func (s *OpenapiService) TenantDisable(ctx context.Context, in *openapi_v1.TenantDisableRequest) (*openapi_v1.TenantDisableResponse, error) {
	return &openapi_v1.TenantDisableResponse{
		Res: util.GetV1ResultOK(),
	}, nil
}
