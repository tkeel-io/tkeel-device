package service

import (
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel-device/api/template/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TemplateService struct {
	pb.UnimplementedTemplateServer
	httpClient *CoreClient
}

func NewTemplateService() *TemplateService {
	return &TemplateService{
		httpClient: NewCoreClient(),
	}
}

type Config struct {
	ID                string                 `json:"id" mapstructure:"id"`
	Type              string                 `json:"type" mapstructure:"type"`
	Weight            int                    `json:"weight" mapstructure:"weight"`
	Enabled           bool                   `json:"enabled" mapstructure:"enabled"`
	EnabledSearch     bool                   `json:"enabled_search" mapstructure:"enabled_search"`
	EnabledTimeSeries bool                   `json:"enabled_time_series" mapstructure:"enabled_time_series"`
	Description       string                 `json:"description" mapstructure:"description"`
	Define            map[string]interface{} `json:"define" mapstructure:"define"`
	LastTime          int64                  `json:"last_time" mapstructure:"last_time"`
}

func (s *TemplateService) CreateTemplate(ctx context.Context, req *pb.CreateTemplateRequest) (*pb.CreateTemplateResponse, error) {
	log.Debug("CreateTemplate")
	log.Debug("req:", req)

	//parse user token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	entityId := GetUUID()
	url := s.httpClient.GetCoreUrl("", tm, "template") + "&id=" + entityId
	log.Debug("get url: ", url)

	//fmt request
	sysField := &pb.TemplateEntitySysField{
		XId:        entityId,
		XCreatedAt: GetTime(),
		XUpdatedAt: GetTime(),
		XOwner:     tm["owner"],
		XSource:    tm["source"],
	}

	entityInfo := &pb.TemplateEntityCoreInfo{
		BasicInfo: req.BasicInfo,
		SysField:  sysField,
	}

	log.Debug("entityinfo : ", entityInfo)

	data, err3 := json.Marshal(entityInfo)
	if nil != err3 {
		return nil, err3
	}

	// do it
	res, err4 := s.httpClient.Post(url, data)
	if nil != err4 {
		log.Error("error post data to core", err4)
		return nil, err4
	}

	//fmt response
	templateObject := &pb.EntityResponse{} // core define
	err5 := json.Unmarshal(res, templateObject)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	out := &pb.CreateTemplateResponse{
		TemplateObject: templateObject,
	}

	return out, nil
}

func (s *TemplateService) UpdateTemplate(ctx context.Context, req *pb.UpdateTemplateRequest) (*pb.UpdateTemplateResponse, error) {
	log.Debug("UpdateTemplate")
	log.Debug("req:", req)

	//parse user token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + req.GetUid()
	url := s.httpClient.GetCoreUrl(midUrl, tm, "template")
	log.Debug("get url: ", url)

	//fmt request

	data, err3 := json.Marshal(req.BasicInfo)
	if nil != err3 {
		return nil, err3
	}

	// do it
	//update BasicInfo
	res, err4 := s.httpClient.Put(url, data)
	if nil != err4 {
		log.Error("error post data to core", err4)
		return nil, err4
	}
	//update updateAt
	ma := make(map[string]interface{})
	ma["_updatedAt"] = GetTime()
	_, err5 := s.CoreConfigPatchMethod(ctx, req.GetUid(), ma, "sysField.", "replace")
	if nil != err5 {
		log.Error("error patch _updateAt", err5)
		return nil, err5
	}

	//fmt response
	templateObject := &pb.EntityResponse{} // core define
	err6 := json.Unmarshal(res, templateObject)
	if err6 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err6
	}

	out := &pb.UpdateTemplateResponse{
		TemplateObject: templateObject,
	}
	return out, nil
}

func (s *TemplateService) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*emptypb.Empty, error) {
	log.Debug("DelTemplate")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	for _, id := range req.Ids.GetIds() {
		//get core url
		midUrl := "/" + id
		url := s.httpClient.GetCoreUrl(midUrl, tm, "template")
		log.Debug("get url :", url)

		//fmt request

		// do it
		_, err1 := s.httpClient.Delete(url)
		if nil != err1 {
			log.Error("error post data to core", id)
			return nil, err1
		}
	}
	//fmt response
	return &emptypb.Empty{}, nil
}

func (s *TemplateService) GetTemplate(ctx context.Context, req *pb.GetTemplateRequest) (*pb.GetTemplateResponse, error) {
	log.Debug("GetTemplate")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + req.GetUid()
	url := s.httpClient.GetCoreUrl(midUrl, tm, "template")
	log.Debug("get url :", url)

	//do it
	res, err2 := s.httpClient.Get(url)
	if nil != err2 {
		log.Error("error get data from core : ", err2)
		return nil, err2
	}

	//fmt response
	templateObject := &pb.EntityResponse{} // core define
	err3 := json.Unmarshal(res, templateObject)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err3
	}
	out := &pb.GetTemplateResponse{
		TemplateObject: templateObject,
	}

	return out, nil
}

func (s *TemplateService) ListTemplate(ctx context.Context, req *pb.ListTemplateRequest) (*pb.ListTemplateResponse, error) {
	log.Debug("ListTemplate")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/search"
	url := s.httpClient.GetCoreUrl(midUrl, tm, "template")
	log.Debug("url :", url)

	data, err := json.Marshal(req.ListEntityQuery)
	if nil != err {
		return nil, err
	}

	//do it
	res, err2 := s.httpClient.Post(url, data)
	if nil != err2 {
		log.Error("error get data from core : ", err2)
		return nil, err2
	}

	//fmt response
	listEntityTotalInfo := &pb.ListEntityResponse{}
	err3 := json.Unmarshal(res, listEntityTotalInfo)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err3
	}
	out := &pb.ListTemplateResponse{
		ListTemplateObject: listEntityTotalInfo,
	}

	return out, nil
}
func (s *TemplateService) AddTemplateAttribute(ctx context.Context, req *pb.AddTemplateAttributeRequest) (*emptypb.Empty, error) {
	log.Debug("AddTemplateAttribute")
	log.Debug("req:", req)

	//fmt request
	attrMap := make(map[string]interface{})
	attrMap[req.Attr.Id] = req.Attr

	//do it
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), attrMap, "attributes.", "add")

}
func (s *TemplateService) UpdateTemplateAttribute(ctx context.Context, req *pb.UpdateTemplateAttributeRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateTemplateAttribute")
	log.Debug("req:", req)

	//fmt request
	attrMap := make(map[string]interface{})
	attrMap[req.Attr.Id] = req.Attr

	//do it
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), attrMap, "attributes.", "replace")
}

func (s *TemplateService) DeleteTemplateAttribute(ctx context.Context, req *pb.DeleteTemplateAttributeRequest) (*emptypb.Empty, error) {
	log.Debug("DelTemplateAttribute")
	log.Debug("req:", req)

	//fmt request
	attrMap := make(map[string]interface{})
	for _, id := range req.Ids.Ids {
		attrMap[id] = "del"
	}
	//do it
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), attrMap, "attributes.", "remove")
}

func (s *TemplateService) GetTemplateAttribute(ctx context.Context, req *pb.GetTemplateAttributeRequest) (*pb.GetTemplateAttributeResponse, error) {
	log.Debug("GetTemplateAttribute")
	log.Debug("req:", req)

	templateAttrSingleObject, err5 := s.GetTemplatePropConfig(ctx, req.GetUid(), req.GetId(), "attributes")
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	out := &pb.GetTemplateAttributeResponse{
		TemplateAttrSingleObject: templateAttrSingleObject,
	}

	return out, nil
}

func (s *TemplateService) ListTemplateAttribute(ctx context.Context, req *pb.ListTemplateAttributeRequest) (*pb.ListTemplateAttributeResponse, error) {
	log.Debug("ListTemplateAttribute")
	log.Debug("req:", req)

	templateAttrObject, err5 := s.ListTemplatePropConfig(ctx, req.GetUid(), "attributes")
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	out := &pb.ListTemplateAttributeResponse{
		TemplateAttrObject: templateAttrObject,
	}
	return out, nil
}

func (s *TemplateService) AddTemplateTelemetry(ctx context.Context, req *pb.AddTemplateTelemetryRequest) (*emptypb.Empty, error) {
	log.Debug("AddTemplateTelemetry")
	log.Debug("req:", req)

	//fmt request
	teleMap := make(map[string]interface{})
	teleMap[req.Tele.Id] = req.Tele

	//do it
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), teleMap, "telemetry.", "add")
}

func (s *TemplateService) UpdateTemplateTelemetry(ctx context.Context, req *pb.UpdateTemplateTelemetryRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateTemplateTelemetry")
	log.Debug("req:", req)

	//fmt request
	teleMap := make(map[string]interface{})
	teleMap[req.Tele.Id] = req.Tele

	//do it
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), teleMap, "telemetry.", "replace")
}
func (s *TemplateService) DeleteTemplateTelemetry(ctx context.Context, req *pb.DeleteTemplateTelemetryRequest) (*emptypb.Empty, error) {
	log.Debug("DeleteTemplateTelemetry")
	log.Debug("req:", req)

	//fmt request
	attrMap := make(map[string]interface{})
	for _, id := range req.Ids.Ids {
		attrMap[id] = "del"
	}
	//do it
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), attrMap, "telemetry.", "remove")
}
func (s *TemplateService) GetTemplateTelemetry(ctx context.Context, req *pb.GetTemplateTelemetryRequest) (*pb.GetTemplateTelemetryResponse, error) {
	log.Debug("GetTemplateTelemetry")
	log.Debug("req:", req)

	templateTeleSingleObject, err5 := s.GetTemplatePropConfig(ctx, req.GetUid(), req.GetId(), "telemetry")
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	out := &pb.GetTemplateTelemetryResponse{
		TemplateTeleSingleObject: templateTeleSingleObject,
	}
	return out, nil
}

func (s *TemplateService) ListTemplateTelemetry(ctx context.Context, req *pb.ListTemplateTelemetryRequest) (*pb.ListTemplateTelemetryResponse, error) {
	log.Debug("ListTemplateTelemetry")
	log.Debug("req:", req)

	templateTeleObject, err5 := s.ListTemplatePropConfig(ctx, req.GetUid(), "telemetry")
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	out := &pb.ListTemplateTelemetryResponse{
		TemplateTeleObject: templateTeleObject,
	}
	return out, nil
}

func (s *TemplateService) AddTemplateTelemetryExt(ctx context.Context, req *pb.AddTemplateTelemetryExtRequest) (*emptypb.Empty, error) {
	log.Debug("AddTemplateTelemetryExt")
	log.Debug("req:", req)

	//get proConfig define
	define, err := s.GetTemplateTelemetryDefine(ctx, req.GetUid(), req.GetId())
	if err != nil {
		return nil, err
	}
	log.Debug("define :", define)

	ext, ok := define["define"].(map[string]interface{})["ext"].(map[string]interface{})
	if !ok {
		return nil, errors.New("error get propConfig ext")
	}
	log.Debug("old ext :", ext)

	//midfy
	switch extKV := req.Kvs.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range extKV {
			ext[k] = v
		}
	default:
		return nil, errors.New("error ext params")
	}
	log.Debug("new ext :", ext)

	//patch
	define["define"].(map[string]interface{})["ext"] = ext
	eMap := make(map[string]interface{})
	eMap[req.GetId()] = define
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), eMap, "telemetry.", "replace")
}

func (s *TemplateService) UpdateTemplateTelemetryExt(ctx context.Context, req *pb.UpdateTemplateTelemetryExtRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateTemplateTelemetryExt")
	log.Debug("req:", req)

	//get proConfig define
	define, err := s.GetTemplateTelemetryDefine(ctx, req.GetUid(), req.GetId())
	if err != nil {
		return nil, err
	}
	log.Debug("define :", define)

	ext, ok := define["define"].(map[string]interface{})["ext"].(map[string]interface{})
	if !ok {
		return nil, errors.New("error get propConfig ext")
	}
	log.Debug("old ext :", ext)

	//midfy
	switch extKV := req.Kvs.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range extKV {
			ext[k] = v
		}
	default:
		return nil, errors.New("error ext params")
	}
	log.Debug("new ext :", ext)

	//patch
	define["define"].(map[string]interface{})["ext"] = ext
	eMap := make(map[string]interface{})
	eMap[req.GetId()] = define
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), eMap, "telemetry.", "replace")
}
func (s *TemplateService) DeleteTemplateTelemetryExt(ctx context.Context, req *pb.DeleteTemplateTelemetryExtRequest) (*emptypb.Empty, error) {
	log.Debug("DeleteTemplateTelemetryExt")
	log.Debug("req:", req)

	//get proConfig define
	define, err := s.GetTemplateTelemetryDefine(ctx, req.GetUid(), req.GetId())
	if err != nil {
		return nil, err
	}
	log.Debug("define :", define)

	ext, ok := define["define"].(map[string]interface{})["ext"].(map[string]interface{})
	if !ok {
		return nil, errors.New("error get propConfig ext")
	}
	log.Debug("old ext :", ext)

	//midfy
	for _, k := range req.Keys.Keys {
		delete(ext, k)
	}
	log.Debug("new ext :", ext)

	//patch
	define["define"].(map[string]interface{})["ext"] = ext
	eMap := make(map[string]interface{})
	eMap[req.GetId()] = define
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), eMap, "telemetry.", "replace")
}

func (s *TemplateService) AddTemplateCommand(ctx context.Context, req *pb.AddTemplateCommandRequest) (*emptypb.Empty, error) {
	log.Debug("AddTemplateCommand")
	log.Debug("req:", req)

	//fmt request
	cmdMap := make(map[string]interface{})
	cmdMap[req.Cmd.Id] = req.Cmd

	//do it
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), cmdMap, "commands.", "add")
}
func (s *TemplateService) UpdateTemplateCommand(ctx context.Context, req *pb.UpdateTemplateCommandRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateTemplateCommand")
	log.Debug("req:", req)

	//fmt request
	cmdMap := make(map[string]interface{})
	cmdMap[req.Cmd.Id] = req.Cmd

	//do it
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), cmdMap, "commands.", "replace")
}
func (s *TemplateService) DeleteTemplateCommand(ctx context.Context, req *pb.DeleteTemplateCommandRequest) (*emptypb.Empty, error) {
	log.Debug("DeleteTemplateCommand")
	log.Debug("req:", req)

	//fmt request
	cmdMap := make(map[string]interface{})
	for _, id := range req.Ids.Ids {
		cmdMap[id] = "del"
	}
	//do it
	return s.CoreConfigPatchMethod(ctx, req.GetUid(), cmdMap, "commands.", "remove")
}

func (s *TemplateService) GetTemplateCommand(ctx context.Context, req *pb.GetTemplateCommandRequest) (*pb.GetTemplateCommandResponse, error) {
	log.Debug("GetTemplateCommand")
	log.Debug("req:", req)

	templateCmdSingleObject, err5 := s.GetTemplatePropConfig(ctx, req.GetUid(), req.GetId(), "commands")
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	out := &pb.GetTemplateCommandResponse{
		TemplateCmdSingleObject: templateCmdSingleObject,
	}
	return out, nil
}

func (s *TemplateService) ListTemplateCommand(ctx context.Context, req *pb.ListTemplateCommandRequest) (*pb.ListTemplateCommandResponse, error) {
	log.Debug("ListTemplateComand")
	log.Debug("req:", req)

	templateCmdObject, err5 := s.ListTemplatePropConfig(ctx, req.GetUid(), "commands")
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	out := &pb.ListTemplateCommandResponse{
		TemplateCmdObject: templateCmdObject,
	}
	return out, nil
}

//abstraction
func (s *TemplateService) CoreConfigPatchMethod(ctx context.Context, entityId string, kv map[string]interface{}, path string, operator string) (*emptypb.Empty, error) {
	log.Debug("CoreConfigPatchMethod")
	log.Debug("path:", path)
	log.Debug("operator:", operator)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + entityId + "/configs/patch"
	url := s.httpClient.GetCoreUrl(midUrl, tm, "template")
	log.Debug("patch url :", url)

	//fmt request
	var patchArray []CorePatch

	for k, v := range kv {
		ph := CorePatch{
			Path:     path + k,
			Operator: operator,
			Value:    v,
		}
		patchArray = append(patchArray, ph)
	}
	log.Debug("patch Array :", patchArray)

	data, err3 := json.Marshal(patchArray)
	if nil != err3 {
		return nil, err3
	}

	// do it
	_, err4 := s.httpClient.Post(url, data)
	if nil != err4 {
		log.Error("error post data to core", data)
		return nil, err4
	}

	//fmt response
	return &emptypb.Empty{}, nil
}
func (s *TemplateService) ListTemplatePropConfig(ctx context.Context, entityId string, classify string) (*pb.EntityResponse, error) {

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + entityId + "/configs"
	url := s.httpClient.GetCoreUrl(midUrl, tm, "template") + fmt.Sprintf("&property_ids=%s", classify)
	log.Debug("url :", url)

	//fmt request

	// do it
	res, err1 := s.httpClient.Get(url)
	if nil != err1 {
		log.Error("error post data to core", err1)
		return nil, err1
	}

	//fmt response
	templatePropConfigObject := &pb.EntityResponse{} // core define
	err5 := json.Unmarshal(res, templatePropConfigObject)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	return templatePropConfigObject, nil
}
func (s *TemplateService) GetTemplatePropConfig(ctx context.Context, entityId string, propId string, classify string) (*pb.EntityResponse, error) {

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + entityId + "/configs"
	url := s.httpClient.GetCoreUrl(midUrl, tm, "template") + fmt.Sprintf("&property_ids=%s", classify+"."+propId)
	log.Debug("url :", url)

	//fmt request

	// do it
	res, err1 := s.httpClient.Get(url)
	if nil != err1 {
		log.Error("error post data to core", err1)
		return nil, err1
	}

	//fmt response
	templateSinglePropConfigObject := &pb.EntityResponse{} // core define
	err5 := json.Unmarshal(res, templateSinglePropConfigObject)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}

	return templateSinglePropConfigObject, nil
}
func (s *TemplateService) GetTemplateTelemetryDefine(ctx context.Context, entityId string, propConfigId string) (map[string]interface{}, error) {

	//get proConfig
	templateTeleSingleObject, err := s.GetTemplatePropConfig(ctx, entityId, propConfigId, "telemetry")
	if err != nil {
		log.Error("error Unmarshal data from core")
		return nil, err
	}
	kv := templateTeleSingleObject.Configs.GetStructValue().Fields
	propConfig, err3 := json.Marshal(kv)
	if nil != err3 {
		return nil, err3
	}
	pr := make(map[string]interface{})
	err4 := json.Unmarshal(propConfig, &pr)
	if nil != err4 {
		return nil, err4
	}

	define, ok := pr["telemetry."+propConfigId].(map[string]interface{})
	if !ok {
		return nil, errors.New("error get propConfig")
	}

	return define, nil
}
