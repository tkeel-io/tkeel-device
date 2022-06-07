package service

import (
	"context"

	json "encoding/json"
	v1 "github.com/tkeel-io/tkeel-device/api/openapi/v1"
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
	return &openapi_v1.AddonsIdentifyResponse{
		Res: util.GetV1ResultBadRequest("not declare addons"),
	}, nil
}

// Identify implements Identify.OpenapiServer.
func (s *OpenapiService) Identify(ctx context.Context, in *emptypb.Empty) (*openapi_v1.IdentifyResponse, error) {
	//device max
	devMax := make(map[string]interface{})
	devMax["key"] = "device_created_max"
	devMax["description"] = "创建设备最大数"
	devMax["value"] = 10000
	//tempalte max
	templateMax := make(map[string]interface{})
	templateMax["key"] = "device_template_max"
	templateMax["description"] = "创建设备模板最大数"
	templateMax["value"] = 10000

	profileArray := make([]interface{}, 2)
	profileArray = append(profileArray, devMax)
	profileArray = append(profileArray, templateMax)

	profilesBytes, _ := json.Marshal(profileArray)
	//if nil != err {
	//}

	return &openapi_v1.IdentifyResponse{
		Res:                     util.GetV1ResultOK(),
		PluginId:                "tkeel-device",
		Version:                 "v0.4.1",
		TkeelVersion:            "v0.4.0",
		DisableManualActivation: true,
		Profiles:                profilesBytes,
		//Profiles : profileArray,
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
