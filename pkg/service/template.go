package service

import (
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel-device/api/template/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
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

	//0. check device name repeated
	errRepeated := s.checkNameRepated(ctx, req.BasicInfo.Name)
	if nil != errRepeated {
		log.Debug("err:", errRepeated)
		return nil, errRepeated
	}

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
	//templateObject := &pb.EntityResponse{} // core define
	templateObject := make(map[string]interface{})
	err5 := json.Unmarshal(res, &templateObject)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}

	//return
	re, err7 := structpb.NewValue(templateObject)
	if nil != err7 {
		log.Error("convert tree failed ", err7)
		return nil, err7
	}
	out := &pb.CreateTemplateResponse{
		TemplateObject: re,
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
	updateT := &pb.UpdateTemplateEntityCoreInfo{}
	updateT.BasicInfo = req.BasicInfo
	data, err3 := json.Marshal(updateT)
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
	_, err5 := s.httpClient.CorePatchMethod(ctx, req.GetUid(), ma, "sysField.", "replace", "/patch")
	if nil != err5 {
		log.Error("error patch _updateAt ", err5)
		return nil, err5
	}

	//fmt response
	//templateObject := &pb.EntityResponse{} // core define
	templateObject := make(map[string]interface{})
	err6 := json.Unmarshal(res, &templateObject)
	if err6 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err6
	}

	//return
	re, err7 := structpb.NewValue(templateObject)
	if nil != err7 {
		log.Error("convert tree failed ", err7)
		return nil, err7
	}
	out := &pb.UpdateTemplateResponse{
		TemplateObject: re,
	}
	return out, nil
}

func (s *TemplateService) DeleteTemplate(ctx context.Context, req *pb.DeleteTemplateRequest) (*pb.DeleteTemplateResponse, error) {
	log.Debug("DelTemplate")
	log.Debug("req:", req)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	out := &pb.DeleteTemplateResponse{
		FaildDelTemplate: make([]*pb.FaildDelTemplate, 0),
	}

	for _, id := range req.Ids.GetIds() {
		//check clild
		err1 := s.checkChild(ctx, id)
		if err1 != nil {
			fd := &pb.FaildDelTemplate{
				Id:     id,
				Reason: err1.Error(),
			}
			out.FaildDelTemplate = append(out.FaildDelTemplate, fd)
			log.Error("have SubNode", id)
			continue
		}

		//get core url
		midUrl := "/" + id
		url := s.httpClient.GetCoreUrl(midUrl, tm, "template")
		log.Debug("get url :", url)

		//fmt request

		// do it
		_, err2 := s.httpClient.Delete(url)
		if nil != err2 {
			fd := &pb.FaildDelTemplate{
				Id:     id,
				Reason: err2.Error(),
			}
			out.FaildDelTemplate = append(out.FaildDelTemplate, fd)
			log.Error("error core return error", id)
			continue
		}
	}
	//fmt response
	return out, nil
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
	//templateObject := &pb.EntityResponse{} // core define
	templateObject := make(map[string]interface{})
	err3 := json.Unmarshal(res, &templateObject)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err3
	}

	//return
	re, err7 := structpb.NewValue(templateObject)
	if nil != err7 {
		log.Error("convert tree failed ", err7)
		return nil, err7
	}
	out := &pb.GetTemplateResponse{
		TemplateObject: re,
	}

	return out, nil
}

/*func (s *TemplateService) ListTemplate(ctx context.Context, req *pb.ListTemplateRequest) (*pb.ListTemplateResponse, error) {
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
}*/
func (s *TemplateService) AddTemplateAttribute(ctx context.Context, req *pb.AddTemplateAttributeRequest) (*emptypb.Empty, error) {
	log.Debug("AddTemplateAttribute")
	log.Debug("req:", req)

	return s.opTemplatePropConfig(ctx, req.GetUid(), req.Attr, "attributes.", "replace", "/configs/patch")
}

func (s *TemplateService) UpdateTemplateAttribute(ctx context.Context, req *pb.UpdateTemplateAttributeRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateTemplateAttribute")
	log.Debug("req:", req)

	return s.opTemplatePropConfig(ctx, req.GetUid(), req.Attr, "attributes.", "replace", "/configs/patch")
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
	return s.httpClient.CorePatchMethod(ctx, req.GetUid(), attrMap, "attributes.", "remove", "/configs/patch")
}

func (s *TemplateService) GetTemplateAttribute(ctx context.Context, req *pb.GetTemplateAttributeRequest) (*pb.GetTemplateAttributeResponse, error) {
	log.Debug("GetTemplateAttribute")
	log.Debug("req:", req)

	templateAttrSingleObject, err5 := s.GetTemplatePropConfig(ctx, req.GetUid(), "attributes."+req.GetId())
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}

	re, err6 := structpb.NewValue(templateAttrSingleObject)
	if nil != err6 {
		log.Error("convert tree failed ", err6)
		return nil, err6
	}
	out := &pb.GetTemplateAttributeResponse{
		TemplateAttrSingleObject: re,
	}

	return out, nil
}

func (s *TemplateService) ListTemplateAttribute(ctx context.Context, req *pb.ListTemplateAttributeRequest) (*pb.ListTemplateAttributeResponse, error) {
	log.Debug("ListTemplateAttribute")
	log.Debug("req:", req)

	templateAttrObject, err5 := s.GetTemplatePropConfig(ctx, req.GetUid(), "attributes")
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	re, err6 := structpb.NewValue(templateAttrObject)
	if nil != err6 {
		log.Error("convert tree failed ", err6)
		return nil, err6
	}
	out := &pb.ListTemplateAttributeResponse{
		TemplateAttrObject: re,
	}
	return out, nil
}

func (s *TemplateService) AddTemplateTelemetry(ctx context.Context, req *pb.AddTemplateTelemetryRequest) (*emptypb.Empty, error) {
	log.Debug("AddTemplateTelemetry")
	log.Debug("req:", req)

	return s.opTemplatePropConfig(ctx, req.GetUid(), req.Tele, "telemetry.", "replace", "/configs/patch")
}

func (s *TemplateService) UpdateTemplateTelemetry(ctx context.Context, req *pb.UpdateTemplateTelemetryRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateTemplateTelemetry")
	log.Debug("req:", req)

	//do it
	return s.opTemplatePropConfig(ctx, req.GetUid(), req.Tele, "telemetry.", "replace", "/configs/patch")
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
	return s.httpClient.CorePatchMethod(ctx, req.GetUid(), attrMap, "telemetry.", "remove", "/configs/patch")
}
func (s *TemplateService) GetTemplateTelemetry(ctx context.Context, req *pb.GetTemplateTelemetryRequest) (*pb.GetTemplateTelemetryResponse, error) {
	log.Debug("GetTemplateTelemetry")
	log.Debug("req:", req)

	templateTeleSingleObject, err5 := s.GetTemplatePropConfig(ctx, req.GetUid(), "telemetry."+req.GetId())
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}

	re, err6 := structpb.NewValue(templateTeleSingleObject)
	if nil != err6 {
		log.Error("convert tree failed ", err6)
		return nil, err6
	}
	out := &pb.GetTemplateTelemetryResponse{
		TemplateTeleSingleObject: re,
	}
	return out, nil
}

func (s *TemplateService) ListTemplateTelemetry(ctx context.Context, req *pb.ListTemplateTelemetryRequest) (*pb.ListTemplateTelemetryResponse, error) {
	log.Debug("ListTemplateTelemetry")
	log.Debug("req:", req)

	templateTeleObject, err5 := s.GetTemplatePropConfig(ctx, req.GetUid(), "telemetry")
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	re, err6 := structpb.NewValue(templateTeleObject)
	if nil != err6 {
		log.Error("convert tree failed ", err6)
		return nil, err6
	}
	out := &pb.ListTemplateTelemetryResponse{
		TemplateTeleObject: re,
	}
	return out, nil
}

func (s *TemplateService) AddTemplateTelemetryExt(ctx context.Context, req *pb.AddTemplateTelemetryExtRequest) (*emptypb.Empty, error) {
	log.Debug("AddTemplateTelemetryExt")
	log.Debug("req:", req)

	//get proConfig define
	propConfig, err := s.GetTemplatePropConfig(ctx, req.GetUid(), "telemetry."+req.GetId())
	if err != nil {
		return nil, err
	}
	ext, err := s.GetSinglePropConfExt(propConfig, "telemetry."+req.GetId())
	if err != nil {
		return nil, err
	}

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
	propConfig["telemetry."+req.GetId()].(map[string]interface{})["define"].(map[string]interface{})["ext"] = ext
	eMap := make(map[string]interface{})
	eMap[req.GetId()] = propConfig["telemetry."+req.GetId()]
	return s.httpClient.CorePatchMethod(ctx, req.GetUid(), eMap, "telemetry.", "replace", "/configs/patch")
}

func (s *TemplateService) UpdateTemplateTelemetryExt(ctx context.Context, req *pb.UpdateTemplateTelemetryExtRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateTemplateTelemetryExt")
	log.Debug("req:", req)

	//get proConfig define
	propConfig, err := s.GetTemplatePropConfig(ctx, req.GetUid(), "telemetry."+req.GetId())
	if err != nil {
		return nil, err
	}
	ext, err := s.GetSinglePropConfExt(propConfig, "telemetry."+req.GetId())
	if err != nil {
		return nil, err
	}

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
	propConfig["telemetry."+req.GetId()].(map[string]interface{})["define"].(map[string]interface{})["ext"] = ext
	eMap := make(map[string]interface{})
	eMap[req.GetId()] = propConfig["telemetry."+req.GetId()]
	return s.httpClient.CorePatchMethod(ctx, req.GetUid(), eMap, "telemetry.", "replace", "/configs/patch")
}
func (s *TemplateService) DeleteTemplateTelemetryExt(ctx context.Context, req *pb.DeleteTemplateTelemetryExtRequest) (*emptypb.Empty, error) {
	log.Debug("DeleteTemplateTelemetryExt")
	log.Debug("req:", req)

	//get proConfig define
	propConfig, err := s.GetTemplatePropConfig(ctx, req.GetUid(), "telemetry."+req.GetId())
	if err != nil {
		return nil, err
	}

	ext, err := s.GetSinglePropConfExt(propConfig, "telemetry."+req.GetId())
	if err != nil {
		return nil, err
	}

	//midfy
	for _, k := range req.Keys.Keys {
		delete(ext, k)
	}
	log.Debug("new ext :", ext)

	//patch
	propConfig["telemetry."+req.GetId()].(map[string]interface{})["define"].(map[string]interface{})["ext"] = ext
	eMap := make(map[string]interface{})
	eMap[req.GetId()] = propConfig["telemetry."+req.GetId()]
	return s.httpClient.CorePatchMethod(ctx, req.GetUid(), eMap, "telemetry.", "replace", "/configs/patch")
}

func (s *TemplateService) AddTemplateCommand(ctx context.Context, req *pb.AddTemplateCommandRequest) (*emptypb.Empty, error) {
	log.Debug("AddTemplateCommand")
	log.Debug("req:", req)

	//do it
	return s.opTemplatePropConfig(ctx, req.GetUid(), req.Cmd, "commands.", "replace", "/configs/patch")
}
func (s *TemplateService) UpdateTemplateCommand(ctx context.Context, req *pb.UpdateTemplateCommandRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateTemplateCommand")
	log.Debug("req:", req)

	//do it
	return s.opTemplatePropConfig(ctx, req.GetUid(), req.Cmd, "commands.", "replace", "/configs/patch")
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
	return s.httpClient.CorePatchMethod(ctx, req.GetUid(), cmdMap, "commands.", "remove", "/configs/patch")
}

func (s *TemplateService) GetTemplateCommand(ctx context.Context, req *pb.GetTemplateCommandRequest) (*pb.GetTemplateCommandResponse, error) {
	log.Debug("GetTemplateCommand")
	log.Debug("req:", req)

	templateCmdSingleObject, err5 := s.GetTemplatePropConfig(ctx, req.GetUid(), "commands."+req.GetId())
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}

	re, err6 := structpb.NewValue(templateCmdSingleObject)
	if nil != err6 {
		log.Error("convert tree failed ", err6)
		return nil, err6
	}
	out := &pb.GetTemplateCommandResponse{
		TemplateCmdSingleObject: re,
	}
	return out, nil
}

func (s *TemplateService) ListTemplateCommand(ctx context.Context, req *pb.ListTemplateCommandRequest) (*pb.ListTemplateCommandResponse, error) {
	log.Debug("ListTemplateComand")
	log.Debug("req:", req)

	templateCmdObject, err5 := s.GetTemplatePropConfig(ctx, req.GetUid(), "commands")
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	re, err6 := structpb.NewValue(templateCmdObject)
	if nil != err6 {
		log.Error("convert tree failed ", err6)
		return nil, err6
	}
	out := &pb.ListTemplateCommandResponse{
		TemplateCmdObject: re,
	}
	return out, nil
}

//abstraction
/*func (s *TemplateService) CorePatchMethod(ctx context.Context, entityId string, kv map[string]interface{}, path string, operator string, pathClassify string) (*emptypb.Empty, error) {
	log.Debug("CorePatchMethod")
	log.Debug("path:", path)
	log.Debug("operator:", operator)

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + entityId + pathClassify
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
	_, err4 := s.httpClient.Put(url, data)
	if nil != err4 {
		log.Error("error post data to core", data)
		return nil, err4
	}

	//fmt response
	return &emptypb.Empty{}, nil
}*/
/*func (s *TemplateService) ListTemplatePropConfig(ctx context.Context, entityId string, classify string) (*pb.EntityResponse, error) {

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
}*/

func (s *TemplateService) GetTemplatePropConfig(ctx context.Context, entityId string, pid string) (map[string]interface{}, error) {

	//get token
	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + entityId + "/configs"
	url := s.httpClient.GetCoreUrl(midUrl, tm, "template") + fmt.Sprintf("&property_ids=%s", pid)
	log.Debug("url :", url)

	//fmt request

	// do it
	res, err1 := s.httpClient.Get(url)
	if nil != err1 {
		log.Error("error post data to core", err1)
		return nil, err1
	}

	//fmt response
	templateSinglePropConfigObject := make(map[string]interface{})
	err5 := json.Unmarshal(res, &templateSinglePropConfigObject)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	return templateSinglePropConfigObject, nil
	/*kv := templateSinglePropConfigObject.Configs.GetStructValue().Fields
	propConfig, err3 := json.Marshal(kv)
	if nil != err3 {
		return nil, err3
	}
	pr := make(map[string]interface{})
	err4 := json.Unmarshal(propConfig, &pr)
	if nil != err4 {
		return nil, err4
	}
	return pr, nil*/

	/*config, ok := pr[pid]
	    if !ok {
	        return nil, errors.New("error get pid")
	    }
		return config, nil*/
}
func (s *TemplateService) GetSinglePropConfExt(propConfig map[string]interface{}, pid string) (map[string]interface{}, error) {

	ext := make(map[string]interface{})
	v, ok := propConfig[pid]
	if !ok {
		return nil, errors.New("error get propCofig")
	}
	v1, ok1 := v.(map[string]interface{})
	if !ok1 {
		return nil, errors.New("error trans propCofig")
	}
	define, ok2 := v1["define"]
	if !ok2 {
		return nil, errors.New("error get define")
	}
	define1, ok3 := define.(map[string]interface{})
	if !ok3 {
		return nil, errors.New("error trans define")
	}
	extOld, ok4 := define1["ext"]
	if !ok4 {
		return nil, errors.New("error get ext")
	}
	extOld1, ok5 := extOld.(map[string]interface{})
	if !ok5 {
		return nil, errors.New("error trans ext")
	}

	ext = extOld1
	log.Debug("old ext :", ext)
	return ext, nil
}

func (s *TemplateService) opTemplatePropConfig(ctx context.Context, templateId string, prop *pb.PropConfig, item string, op string, pathClassify string) (*emptypb.Empty, error) {
	//fmt request
	propMap := make(map[string]interface{})
	/*for _, p := range prop.PropAarry {
		propMap[p.Id] = p
	}*/
	propMap[prop.Id] = prop
	//do it
	return s.httpClient.CorePatchMethod(ctx, templateId, propMap, item, op, pathClassify)
}

func (s *TemplateService) checkChild(ctx context.Context, id string) error {
	log.Debug("checkChild")
	//create query
	query := &pb.ListEntityQuery{
		PageNum:      1,
		PageSize:     0,
		OrderBy:      "name",
		IsDescending: false,
		Query:        "",
		Condition:    make([]*pb.Condition, 0),
	}
	condition1 := &pb.Condition{
		Field:    "basicInfo.templateId",
		Operator: "$eq",
		Value:    id,
	}
	query.Condition = append(query.Condition, condition1)
	//search
	listObject, err := s.CoreSearchEntity(ctx, query)
	if err != nil {
		log.Error("error Core return ", err)
		return err
	}

	//check
	total, ok := listObject["total"]
	if !ok {
		log.Error("error  total field does not exist")
		return errors.New("total field does not exist")
	}

	tl, ok1 := total.(float64)
	if !ok1 {
		return errors.New("total is not int type")
	}

	if tl == 0 {
		return nil
	} else {
		return errors.New("have SubNode")
	}
}
func (s *TemplateService) CoreSearchEntity(ctx context.Context, listEntityQuery *pb.ListEntityQuery) (map[string]interface{}, error) {
	log.Debug("CoreSearchEntity")

	tm, err := s.httpClient.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/search"
	url := s.httpClient.GetCoreUrl(midUrl, tm, "template")
	log.Debug("core url :", url)

	//Data isolation
	user := &pb.Condition{
		Field:    "owner",
		Operator: "$eq",
		Value:    tm["owner"],
	}
	listEntityQuery.Condition = append(listEntityQuery.Condition, user)
	log.Debug("Query:", listEntityQuery)

	//do it
	filter, err1 := json.Marshal(listEntityQuery)
	if err1 != nil {
		return nil, err1
	}
	res, err2 := s.httpClient.Post(url, filter)
	if nil != err2 {
		return nil, err2
	}

	listObject := make(map[string]interface{})
	err3 := json.Unmarshal(res, &listObject)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err3
	}
	return listObject, nil
}
func (s *TemplateService) checkNameRepated(ctx context.Context, name string) error {
	log.Debug("checkNameRepated")
	if name == "" {
		return errors.New("name cannot be empty")
	}
	//create query
	query := &pb.ListEntityQuery{
		PageNum:      1,
		PageSize:     0,
		OrderBy:      "name",
		IsDescending: false,
		Query:        "",
		Condition:    make([]*pb.Condition, 0),
	}
	condition1 := &pb.Condition{
		Field:    "basicInfo.name",
		Operator: "$eq",
		Value:    name,
	}
	condition2 := &pb.Condition{
		Field:    "type",
		Operator: "$eq",
		Value:    "template",
	}
	query.Condition = append(query.Condition, condition1)
	query.Condition = append(query.Condition, condition2)
	//search
	listObject, err := s.CoreSearchEntity(ctx, query)
	if err != nil {
		log.Error("error Core return ", err)
		return err
	}

	//check
	total, ok := listObject["total"]
	if !ok {
		log.Error("error  total field does not exist")
		return errors.New("total field does not exist")
	}

	tl, ok1 := total.(float64)
	if !ok1 {
		return errors.New("total is not int type")
	}

	if tl == 0 {
		return nil
	} else {
		return errors.New("have repeated")
	}
}
