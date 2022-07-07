package service

import (
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"github.com/tkeel-io/tkeel-device/pkg/service/openapi"
	"github.com/tkeel-io/tkeel/pkg/client/dapr"
	"strings"

	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel-device/api/device/v1"
	pbt "github.com/tkeel-io/tkeel-device/api/template/v1"

	//go_struct "google.golang.org/protobuf/types/known/structpb"
	"time"

	"github.com/tkeel-io/tkeel-device/pkg/service/metrics"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type DeviceService struct {
	pb.UnimplementedDeviceServer
	client     *CoreClient
	daprClient *dapr.HTTPClient
}

func NewDeviceService() *DeviceService {
	ds := &DeviceService{
		client: NewCoreClient(),
	}
	go ds.MetricsTimer()
	return ds
}

func (s *DeviceService) Init() {
	s.daprClient = dapr.NewHTTPClient("3500")
}

func (s *DeviceService) CreateDevice(ctx context.Context, req *pb.CreateDeviceRequest) (*pb.CreateDeviceResponse, error) {
	log.Debug("CreateDevice")
	log.Debug("req:", req.DevBasicInfo)

	//0. check device name repeated
	errRepeated := s.checkNameRepated(ctx, req.DevBasicInfo.Name)
	if nil != errRepeated {
		log.Debug("err:", errRepeated)
		return nil, errRepeated
	}

	//1. verify Authentication in header and get user token map
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		log.Debug("err:", err)
		return nil, err
	}

	// 2. get url
	devId := GetUUID()
	if req.DevBasicInfo.CustomId != "" {
		devId = req.DevBasicInfo.CustomId
	}
	url := s.client.GetCoreUrl("", tm, "device") + "&id=" + devId
	if req.DevBasicInfo.TemplateId != "" {
		url += "&from=" + req.DevBasicInfo.TemplateId
	}
	log.Debug("core Url: ", url)

	// 3. build coreInfo and add system value
	coreInfo := new(pb.DeviceEntityCoreInfo)
	coreInfo.BasicInfo = new(pb.DeviceEntityBasicInfo)
	coreInfo.BasicInfo = req.DevBasicInfo
	coreInfo.SysField = new(pb.DeviceEntitySysField)
	coreInfo.SysField.XId = devId
	coreInfo.SysField.XCreatedAt = GetTime()
	coreInfo.SysField.XUpdatedAt = GetTime()
	coreInfo.SysField.XEnable = true
	coreInfo.SysField.XStatus = "offline"
	coreInfo.SysField.XOwner = tm["owner"]
	coreInfo.SysField.XTenantId = tm["tenantId"]
	coreInfo.SysField.XSource = tm["source"]
	coreInfo.SysField.XSpacePath = devId
	coreInfo.SysField.XSubscribeAddr = ""

	//connectInfo
	coreInfo.ConnectInfo = new(pb.DeviceEntityConnectInfo)
	coreInfo.ConnectInfo.XClientId = ""
	coreInfo.ConnectInfo.XOnline = false
	coreInfo.ConnectInfo.XSockPort = ""
	coreInfo.ConnectInfo.XProtocol = ""
	coreInfo.ConnectInfo.XPeerHost = ""
	coreInfo.ConnectInfo.XUserName = ""

	//3.5 logical judgement
	if coreInfo.BasicInfo.DirectConnection == false && coreInfo.BasicInfo.TemplateId == "" {
		return nil, errors.New("non-direct connection must have template")
	}
	if coreInfo.BasicInfo.ParentId == coreInfo.SysField.XId {
		return nil, errors.New("error ParentId")
	}
	if coreInfo.BasicInfo.ParentId == "" { // 自动创建默认分组  先赋值 setmpper 里面会检查存在
		coreInfo.BasicInfo.ParentId = "iotd-" + tm["owner"] + "-defaultGroup"
		coreInfo.BasicInfo.ParentName = "默认分组"
		//return nil, errors.New("error ParentId")
	}

	//create  templateName mapper
	if coreInfo.BasicInfo.TemplateId != "" {
		s.client.setMapper(tm, "mapper_template_name", devId, "basicInfo.templateName", coreInfo.BasicInfo.TemplateId, "basicInfo.name")
	}

	//4. create device token
	token, err2 := s.client.CreatEntityToken("device", coreInfo.SysField.XId, tm["owner"], tm["userToken"])
	if nil != err2 {
		return nil, err2
	}
	//token := "token"
	coreInfo.SysField.XToken = token

	// 4. add internal values and marshal Dev
	dev, err3 := json.Marshal(coreInfo)
	if nil != err3 {
		return nil, err3
	}

	// 5. core request  and response
	log.Info("data: ", string(dev))
	res, err4 := s.client.Post(url, dev)
	if nil != err4 {
		log.Error("error post data to core", string(dev))
		return nil, err4
	}

	// 6. fmt response to user
	//deviceObject := &pb.EntityResponse{} // core define
	deviceObject := make(map[string]interface{}) // core define
	err5 := json.Unmarshal(res, &deviceObject)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}

	// 7 set mapper
	err6 := s.client.setSpacePathMapper(tm, devId, req.DevBasicInfo.ParentId, "device")
	if nil != err6 {
		log.Error("error addSpacePath mapper", err6)
		return nil, err6
	}

	// 8 return
	re, err7 := structpb.NewValue(deviceObject)
	if nil != err7 {
		log.Error("convert tree failed ", err7)
		return nil, err7
	}
	out := &pb.CreateDeviceResponse{
		DeviceObject: re,
	}

	return out, nil
}

type UpdateEntityReq struct {
	TemplateID  string                 `json:"template_id"`
	Description string                 `json:"description"`
	Configs     map[string]interface{} `json:"configs"`
	Properties  map[string]interface{} `json:"properties"`
}

func (s *DeviceService) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.UpdateDeviceResponse, error) {
	log.Debug("UpdateDevice")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/" + req.Id
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	if req.DevBasicInfo.TemplateId != "" {
		url += "&from=" + req.DevBasicInfo.TemplateId
	}
	log.Debug("get url :", url)

	updateData := &UpdateEntityReq{
		TemplateID:  req.DevBasicInfo.TemplateId,
		Description: req.DevBasicInfo.Description,
		Properties: map[string]interface{}{
			"basicInfo": req.DevBasicInfo,
		},
	}

	//check cycle.
	if req.DevBasicInfo.ParentId == req.Id {
		return nil, errors.New("error ParentId")
	}

	//do it
	//update BasicInfo
	dev, err1 := json.Marshal(updateData)
	if err1 != nil {
		return nil, err1
	}
	res, err2 := s.client.Put(url, dev)
	if nil != err2 {
		return nil, err2
	}
	//update updateAt
	ma := make(map[string]interface{})
	ma["_updatedAt"] = GetTime()
	_, err3 := s.client.CorePatchMethod(ctx, req.GetId(), ma, "sysField.", "replace", "/patch")
	if nil != err3 {
		log.Error("error patch _updateAt", err3)
		return nil, err3
	}

	//fmt response
	//er := new(pb.EntityResponse)
	deviceObject := make(map[string]interface{})
	if err4 := json.Unmarshal(res, &deviceObject); nil != err4 {
		return nil, err4
	}

	// 7 set mapper
	err6 := s.client.setSpacePathMapper(tm, req.Id, req.DevBasicInfo.ParentId, "device")
	if nil != err6 {
		log.Error("error addSpacePath mapper", err6)
		return nil, err6
	}

	//return
	re, err7 := structpb.NewValue(deviceObject)
	if nil != err7 {
		log.Error("convert tree failed ", err7)
		return nil, err7
	}
	out := &pb.UpdateDeviceResponse{
		DeviceObject: re,
	}

	return out, nil
}

func (s *DeviceService) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.DeleteDeviceResponse, error) {
	log.Debug("DeleteDevice")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	out := &pb.DeleteDeviceResponse{
		FaildDelDevice: make([]*pb.FaildDelDevice, 0),
	}
	ids := req.Ids.GetIds()
	for _, id := range ids {
		// delete subscribe entities
		urlSub := s.client.GetDeleleEntityFromSubUrl(id)
		_, _ = s.client.DeleteWithCtx(ctx, urlSub)

		// delete rule devices
		urlRule := s.client.GetDeleleEntityFromRuleUrl(id)
		_, _ = s.client.DeleteWithCtx(ctx,urlRule)

		midUrl := "/" + id
		url := s.client.GetCoreUrl(midUrl, tm, "device")
		log.Debug("get url:", url)

		_, err2 := s.client.Delete(url)
		if nil != err2 {
			fd := &pb.FaildDelDevice{
				Id:     id,
				Reason: err2.Error(),
			}
			out.FaildDelDevice = append(out.FaildDelDevice, fd)
			log.Error("error core return error", id)
			continue
		}
	}

	// addons
	openapiCli := NewDaprClientDefault(s.daprClient)
	if err = openapiCli.SchemaChangeAddons(ctx, tm["tenantId"],
		strings.Join(req.GetIds().GetIds(), ","), openapi.EventDeviceDelete, nil); err != nil {
		log.L().Error("call addons error")
	}
	return out, nil
}

func (s *DeviceService) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.GetDeviceResponse, error) {
	log.Debug("GetDevice")
	log.Debug("req:", req)
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/" + req.GetId()
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("get url :", url)

	res, err2 := s.client.Get(url)
	if nil != err2 {
		return nil, err2
	}

	//return
	//er := new(pb.EntityResponse)
	deviceObject := make(map[string]interface{})
	if err3 := json.Unmarshal(res, &deviceObject); nil != err3 {
		return nil, err3
	}

	re, err7 := structpb.NewValue(deviceObject)
	if nil != err7 {
		log.Error("convert tree failed ", err7)
		return nil, err7
	}

	out := &pb.GetDeviceResponse{
		DeviceObject: re,
	}

	return out, nil
}

func (s *DeviceService) SearchEntity(ctx context.Context, req *pb.ListDeviceRequest) (*pb.ListDeviceResponse, error) {
	log.Debug("SearchEntity")

	listDeviceObject, err3 := s.CoreSearchEntity(ctx, req.ListEntityQuery)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err3
	}

	re, err6 := structpb.NewValue(listDeviceObject)
	if nil != err6 {
		log.Error("convert tree failed ", err6)
		return nil, err6
	}
	out := &pb.ListDeviceResponse{
		ListDeviceObject: re,
	}

	return out, nil
}

func (s *DeviceService) EnableDevice(ctx context.Context, req *pb.EnableDeviceRequest) (*emptypb.Empty, error) {
	log.Debug("EnableDevice")
	log.Debug("req:", req)

	//todo, maybe add this in future
	//tm, err := s.client.GetTokenMap(ctx)
	//if nil != err{
	//	return nil, err
	//}
	//midUrl := "/" + req.GetId()
	//url := s.client.GetCoreUrl(midUrl, tm)
	//log.Debug("get url :", url)
	//
	//enable := map[string]interface{}{
	//	"_enable": req.Enable.Enable,
	//}
	//data, err := json.Marshal(enable)
	//if err != nil {
	//	return &pb.EnableDeviceResponse{Result: ResFailed}, err
	//}
	//_, err2 := s.client.Put(url, data)
	//if nil != err2{
	//	return &pb.EnableDeviceResponse{Result: ResFailed}, err2
	//}
	return &emptypb.Empty{}, nil
}

func (s *DeviceService) AddDeviceExtBusiness(ctx context.Context, req *pb.AddDeviceExtBusinessRequest) (*emptypb.Empty, error) {
	log.Debug("AddDeviceExtBusiness")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/" + req.GetId() + "/patch"
	url := s.client.GetCoreUrl(midUrl, tm, "devcie")
	log.Debug("get url :", url)

	var exts []interface{}
	switch kv := req.ExtBusiness.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			e := map[string]interface{}{
				"path":     fmt.Sprintf("basicInfo.extBusiness.%s", k),
				"operator": "replace",
				"value":    v,
			}
			exts = append(exts, e)
		}
	default:
		return nil, errors.New("error Invalid payload")
	}

	log.Debug("ExtBusiness body: ", exts)
	data, err1 := json.Marshal(exts)
	if err1 != nil {
		return nil, err1
	}
	_, err2 := s.client.Put(url, data)
	if nil != err2 {
		return nil, err2
	}

	return &emptypb.Empty{}, nil
}
func (s *DeviceService) UpdateDeviceExtBusiness(ctx context.Context, req *pb.UpdateDeviceExtBusinessRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateDeviceExtBusiness")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/" + req.GetId() + "/patch"
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("get url :", url)

	var exts []interface{}
	switch kv := req.ExtBusiness.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			e := map[string]interface{}{
				"path":     fmt.Sprintf("basicInfo.extBusiness.%s", k),
				"operator": "replace",
				"value":    v,
			}
			exts = append(exts, e)
		}
	default:
		return nil, errors.New("error Invalid payload")
	}

	log.Debug("extBusiness body: ", exts)
	data, err1 := json.Marshal(exts)
	if err1 != nil {
		return nil, err1
	}
	_, err2 := s.client.Put(url, data)
	if nil != err2 {
		return nil, err2
	}

	return &emptypb.Empty{}, nil
}
func (s *DeviceService) DeleteDeviceExtBusiness(ctx context.Context, req *pb.DeleteDeviceExtBusinessRequest) (*emptypb.Empty, error) {
	log.Debug("DeleteDeviceExtBusiness")
	log.Debug("req:", req)
	// todo when core support
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/" + req.GetId() + "/patch"
	url := s.client.GetCoreUrl(midUrl, tm, "device")

	// var exts []interface{}
	keys := req.Keys.Keys
	exts := make([]interface{}, len(keys))
	for i, k := range keys {
		e := map[string]interface{}{
			"path":     fmt.Sprintf("basicInfo.extBusiness.%s", k),
			"operator": "remove",
			"value":    "",
		}
		exts[i] = e
	}

	log.Debug("ext body: ", exts)
	data, err1 := json.Marshal(exts)
	if err1 != nil {
		return nil, err1
	}
	_, err2 := s.client.Put(url, data)
	if nil != err2 {
		return nil, err2
	}
	return &emptypb.Empty{}, nil
}

func (s *DeviceService) AddDeviceExt(ctx context.Context, req *pb.AddDeviceExtRequest) (*emptypb.Empty, error) {
	log.Debug("AddDeviceExt")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/" + req.GetId() + "/patch"
	url := s.client.GetCoreUrl(midUrl, tm, "devcie")
	log.Debug("get url :", url)

	var exts []interface{}
	switch kv := req.Ext.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			e := map[string]interface{}{
				"path":     fmt.Sprintf("basicInfo.ext.%s", k),
				"operator": "replace",
				"value":    v,
			}
			exts = append(exts, e)
		}
	default:
		return nil, errors.New("error Invalid payload")
	}

	log.Debug("ext body: ", exts)
	data, err1 := json.Marshal(exts)
	if err1 != nil {
		return nil, err1
	}
	_, err2 := s.client.Put(url, data)
	if nil != err2 {
		return nil, err2
	}

	return &emptypb.Empty{}, nil
}

func (s *DeviceService) DeleteDeviceExt(ctx context.Context, req *pb.DeleteDeviceExtRequest) (*emptypb.Empty, error) {
	log.Debug("DeleteDeviceExt")
	log.Debug("req:", req)
	// todo when core support
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/" + req.GetId() + "/patch"
	url := s.client.GetCoreUrl(midUrl, tm, "device")

	// var exts []interface{}
	keys := req.Keys.Keys
	exts := make([]interface{}, len(keys))
	for i, k := range keys {
		e := map[string]interface{}{
			"path":     fmt.Sprintf("basicInfo.ext.%s", k),
			"operator": "remove",
			"value":    "",
		}
		exts[i] = e
	}

	log.Debug("ext body: ", exts)
	data, err1 := json.Marshal(exts)
	if err1 != nil {
		return nil, err1
	}
	_, err2 := s.client.Put(url, data)
	if nil != err2 {
		return nil, err2
	}
	return &emptypb.Empty{}, nil
}

func (s *DeviceService) UpdateDeviceExt(ctx context.Context, req *pb.UpdateDeviceExtRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateDeviceExt")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/" + req.GetId() + "/patch"
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("get url :", url)

	var exts []interface{}
	switch kv := req.Ext.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			e := map[string]interface{}{
				"path":     fmt.Sprintf("basicInfo.ext.%s", k),
				"operator": "replace",
				"value":    v,
			}
			exts = append(exts, e)
		}
	default:
		return nil, errors.New("error Invalid payload")
	}

	log.Debug("ext body: ", exts)
	data, err1 := json.Marshal(exts)
	if err1 != nil {
		return nil, err1
	}
	_, err2 := s.client.Put(url, data)
	if nil != err2 {
		return nil, err2
	}

	return &emptypb.Empty{}, nil
}

func (s *DeviceService) expByConfigs(configObject map[string]interface{}, classify string, curName string, curId string, targetName string, targetId string) []*pb.Expression {
	//define
	exps := make([]*pb.Expression, 0)
	var attr map[string]interface{}

	//get
	if configs1, okc1 := configObject["configs"]; okc1 == true {
		if configs2, okc2 := configs1.(map[string]interface{}); okc2 == true {
			if attr1, ok1 := configs2[classify]; ok1 == true {
				if attr2, ok2 := attr1.(map[string]interface{}); ok2 == true {
					if define1, ok3 := attr2["define"]; ok3 == true {
						if define2, ok4 := define1.(map[string]interface{}); ok4 == true {
							if field1, ok5 := define2["fields"]; ok5 == true {
								if field2, ok6 := field1.(map[string]interface{}); ok6 == true {
									attr = field2
								}
							}
						}
					}
				}
			}
		}
	}

	//create
	for f := range attr {
		//to do :devices or properties are converted according to ext's configuration(key = mapper_alias)
		fAlias := f
		if conf, ok := attr[f].(map[string]interface{}); ok == true {
			if attrExt1, ok1 := conf["ext"]; ok1 == true {
				if attrExt2, ok2 := attrExt1.(map[string]interface{}); ok2 == true {
					if alias, ok3 := attrExt2["mapper_alias"]; ok3 == true {
						fAlias = alias.(string)
					}
				}
			}
		}
		//to do :devices or properties are converted according to ext's configuration(key = mappr_prefix)
		fPrefix := ""
		if conf, ok := attr[f].(map[string]interface{}); ok == true {
			if attrExt1, ok1 := conf["ext"]; ok1 == true {
				if attrExt2, ok2 := attrExt1.(map[string]interface{}); ok2 == true {
					if prefix, ok3 := attrExt2["mapper_prefix"]; ok3 == true {
						fPrefix = "." + prefix.(string)
					}
				}
			}
		}

		/***
		attrName := ""
		if conf, ok := attr[f].(map[string]interface{}); ok == true {
			if attrName1, ok1 := conf["name"]; ok1 == true {
				attrName = attrName1.(string)
			}
		}***/

		exp := &pb.Expression{
			Path:       classify + "." + f,
			Expression: targetId + "." + classify + fPrefix + "." + fAlias,
			Name:       f,
			//Description: curId + "=" + curName + "," + f + "=" + attrName + "," + targetId + "=" + targetName,
			Description: targetId + "=" + targetName,
		}
		exps = append(exps, exp)
	}
	return exps
}

func (s *DeviceService) CreateExpressionsByParseConfigs(tm map[string]string, curName string, curId string, targetName string, targetId string) ([]*pb.Expression, error) {

	//check targetId
	if targetId == "" {
		return nil, errors.New("targetId is empty")
	}

	//get cur configs
	configObject, err1 := s.client.GetCoreEntitySpecContent(tm, curId, "device", "configs", "")
	if nil != err1 {
		return nil, err1
	}

	//to do ?
	//get targetId confings
	//check TargetId configs
	//check targetId Properties

	//overlock
	//get attr
	expAttr := s.expByConfigs(configObject, "attributes", curName, curId, targetName, targetId)
	expTele := s.expByConfigs(configObject, "telemetry", curName, curId, targetName, targetId)
	expAll := append(expAttr, expTele...)
	log.Debug("expAll  = ", expAll)

	return expAll, nil
}
func (s *DeviceService) CreateDeviceDataRelationAuto(ctx context.Context, req *pb.CreateDeviceDataRelationAutoRequest) (*pb.CreateDeviceDataRelationAutoResponse, error) {
	log.Debug("CreateDataRelationAuto")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//create url
	midUrl := "/" + req.GetId() + "/expressions"
	url := s.client.GetCoreUrl(midUrl, tm, "device")

	//create expressions
	expressions, err2 := s.CreateExpressionsByParseConfigs(tm, req.Relation.GetCurName(), req.GetId(), req.Relation.GetTargetName(), req.Relation.GetTargetId())
	if nil != err2 {
		log.Error("error parse configs")
		return nil, err2
	}

	//do it
	ed := make(map[string]interface{})
	ed["expressions"] = expressions
	data, err3 := json.Marshal(ed)
	if nil != err3 {
		return nil, err3
	}
	log.Info("data: ", string(data))
	log.Debug("url :", url)
	_, err4 := s.client.Post(url, data)
	if err4 != nil {
		log.Error("error post data to core")
		return nil, err4
	}

	//get list
	res, err5 := s.client.Get(url)
	if nil != err5 {
		log.Error("error post data to core")
		return nil, err5
	}

	expressionObject := make(map[string]interface{}) // core define
	err6 := json.Unmarshal(res, &expressionObject)
	if err6 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err6
	}

	// return
	re, err7 := structpb.NewValue(expressionObject)
	if nil != err7 {
		log.Error("convert failed ", err7)
		return nil, err7
	}
	out := &pb.CreateDeviceDataRelationAutoResponse{
		ExpressionObject: re,
	}
	return out, nil
}

func (s *DeviceService) CreateDeviceDataRelation(ctx context.Context, req *pb.CreateDeviceDataRelationRequest) (*emptypb.Empty, error) {
	log.Debug("CreateDataRelation")
	log.Debug("req:", req)
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//create url
	midUrl := "/" + req.GetId() + "/expressions"
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("url :", url)

	//do it
	data, err3 := json.Marshal(req.Expressions)
	if nil != err3 {
		return nil, err3
	}
	log.Info("data: ", string(data))
	_, err4 := s.client.Post(url, data)
	if nil != err4 {
		log.Error("error post data to core")
		return nil, err4
	}
	return &emptypb.Empty{}, nil
}
func (s *DeviceService) UpdateDeviceDataRelation(ctx context.Context, req *pb.UpdateDeviceDataRelationRequest) (*emptypb.Empty, error) {
	log.Debug("UpdateDataRelation")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//create url
	midUrl := "/" + req.GetId() + "/expressions"
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("url :", url)

	//do it
	data, err3 := json.Marshal(req.Expressions)
	if nil != err3 {
		return nil, err3
	}
	log.Info("data: ", string(data))
	_, err4 := s.client.Post(url, data)
	if nil != err4 {
		log.Error("error post data to core")
		return nil, err4
	}
	return &emptypb.Empty{}, nil
}
func (s *DeviceService) DeleteDeviceDataRelation(ctx context.Context, req *pb.DeleteDeviceDataRelationRequest) (*emptypb.Empty, error) {
	log.Debug("DeleteDataRelation")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//create url
	midUrl := "/" + req.GetId() + "/expressions"
	url := s.client.GetCoreUrl(midUrl, tm, "device")

	//create query
	paths := ""
	for _, path := range req.Paths.Paths {
		paths += path + ","
	}
	url += "&paths=" + paths
	log.Debug("url :", url)

	// do it
	_, err4 := s.client.Delete(url)
	if nil != err4 {
		log.Error("error post data to core")
		return nil, err4
	}
	return &emptypb.Empty{}, nil
}

func (s *DeviceService) GetDeviceDataRelation(ctx context.Context, req *pb.GetDeviceDataRelationRequest) (*pb.GetDeviceDataRelationResponse, error) {
	log.Debug("GetDeviceDataRelation")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//create url
	midUrl := "/" + req.GetId() + "/expressions/" + req.Path
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("url :", url)

	res, err4 := s.client.Get(url)
	if nil != err4 {
		log.Error("error post data to core")
		return nil, err4
	}

	expression := make(map[string]interface{}) // core define
	err5 := json.Unmarshal(res, &expression)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}

	// return
	re, err7 := structpb.NewValue(expression)
	if nil != err7 {
		log.Error("convert failed ", err7)
		return nil, err7
	}
	out := &pb.GetDeviceDataRelationResponse{
		Expressions: re,
	}

	return out, nil
}

func (s *DeviceService) ListDeviceDataRelation(ctx context.Context, req *pb.ListDeviceDataRelationRequest) (*pb.ListDeviceDataRelationResponse, error) {
	log.Debug("ListDeviceDataRelation")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//create url
	midUrl := "/" + req.GetId() + "/expressions"
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("url :", url)

	res, err4 := s.client.Get(url)
	if nil != err4 {
		log.Error("error post data to core")
		return nil, err4
	}

	expressionObject := make(map[string]interface{}) // core define
	err5 := json.Unmarshal(res, &expressionObject)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}

	// return
	re, err7 := structpb.NewValue(expressionObject)
	if nil != err7 {
		log.Error("convert failed ", err7)
		return nil, err7
	}
	out := &pb.ListDeviceDataRelationResponse{
		ExpressionObject: re,
	}

	return out, nil
}

func (s *DeviceService) SetDeviceRaw(ctx context.Context, req *pb.SetDeviceRawRequest) (*emptypb.Empty, error) {
	log.Debug("SetDeviceRaw")
	log.Debug("req:", req)

	ma := make(map[string]interface{})
	ma["rawDown"] = req.Value
	_, err3 := s.client.CorePatchMethod(ctx, req.GetId(), ma, "", "replace", "/patch")
	if nil != err3 {
		log.Error("error patch rawDown", err3)
		return nil, err3
	}

	// set rawData  for ws
	/*
		down := make(map[string]interface{})
		down["id"] = req.GetId()
		down["ts"] = GetTime()
		down["Values"] = req.Value
		down["path"] = ""
		down["type"] = "rawDown"
		down["mark"] = "downstream"

		md := make(map[string]interface{})
		md["rawData"] = down
		_, err4 := s.client.CorePatchMethod(ctx, req.GetId(), md, "", "replace", "/patch")
		if nil != err4 {
			log.Error("error patch rawData", err4)
			return nil, err4
		}*/

	return &emptypb.Empty{}, nil
}

func (s *DeviceService) SetDeviceAttribte(ctx context.Context, req *pb.SetDeviceAttributeRequest) (*emptypb.Empty, error) {
	log.Debug("SetDeviceAttribte")
	log.Debug("req:", req)

	ma := make(map[string]interface{})
	ma[req.Content.Id] = req.Content.Value
	_, err3 := s.client.CorePatchMethod(ctx, req.GetId(), ma, "attributes.", "replace", "/patch")
	if nil != err3 {
		log.Error("error patch attribute", err3)
		return nil, err3
	}
	// set rawData  for ws
	/*down := make(map[string]interface{})
	down["id"] = req.GetId()
	down["ts"] = GetTime()
	down["Values"] = req.Content
	down["path"] = ""
	down["type"] = "shareAttribute"
	down["mark"] = "downstream"

	md := make(map[string]interface{})
	md["rawData"] = down
	_, err4 := s.client.CorePatchMethod(ctx, req.GetId(), md, "", "replace", "/patch")
	if nil != err4 {
		log.Error("error patch rawData", err4)
		return nil, err4
	}*/

	return &emptypb.Empty{}, nil
}
func (s *DeviceService) SetDeviceCommand(ctx context.Context, req *pb.SetDeviceCommandRequest) (*emptypb.Empty, error) {
	log.Debug("SetDeviceCommand")
	log.Debug("req:", req)

	ma := make(map[string]interface{})
	ma[req.Content.Id] = req.Content.Value
	_, err3 := s.client.CorePatchMethod(ctx, req.GetId(), ma, "commands.", "replace", "/patch")
	if nil != err3 {
		log.Error("error patch commands", err3)
		return nil, err3
	}
	// set rawData  for ws
	/*down := make(map[string]interface{})
	down["id"] = req.GetId()
	down["ts"] = GetTime()
	down["Values"] = req.Content
	down["path"] = ""
	down["type"] = "command"
	down["mark"] = "downstream"

	md := make(map[string]interface{})
	md["rawData"] = down
	_, err4 := s.client.CorePatchMethod(ctx, req.GetId(), md, "", "replace", "/patch")
	if nil != err4 {
		log.Error("error patch rawData", err4)
		return nil, err4
	}*/

	return &emptypb.Empty{}, nil
}

func (s *DeviceService) SaveDeviceConfAsSelfTemplte(ctx context.Context, req *pb.SaveDeviceConfAsSelfTemplteRequest) (*emptypb.Empty, error) {
	log.Debug("SaveDeviceConfAsSelfTemplte")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get configs
	configObject, err1 := s.client.GetCoreEntitySpecContent(tm, req.Id, "device", "configs", "")
	if nil != err1 {
		return nil, err1
	}
	configs, ok := configObject["configs"]
	if !ok {
		log.Error("error config non exist")
		return nil, errors.New("error config non exist")
	}
	log.Debug(configs)

	//get TemplateId
	templateIdObject, err2 := s.client.GetCoreEntitySpecContent(tm, req.Id, "device", "properties", "basicInfo.templateId")
	if nil != err2 {
		return nil, err2
	}
	templateId := ""
	if prop, ok := templateIdObject["properties"]; ok {
		if prop1, ok1 := prop.(map[string]interface{}); ok1 {
			if id, ok2 := prop1["basicInfo.templateId"]; ok2 {
				if idstr, ok3 := id.(string); ok3 {
					templateId = idstr
				}
			}
		}
	}

	// if tmeplateId  NON exist
	if templateId == "" {
		return nil, errors.New("templateID non exist")
	} else {
		log.Debug(templateId)

		//patch
		midUrl := "/" + templateId
		url := s.client.GetCoreUrl(midUrl, tm, "device")
		log.Debug("put url :", url)

		data := make(map[string]interface{})
		data["configs"] = configs
		log.Debug(data)
		dt, err3 := json.Marshal(data)
		if err3 != nil {
			return nil, err3
		}
		_, err4 := s.client.Put(url, dt)
		if nil != err4 {
			return nil, err4
		}

		return &emptypb.Empty{}, nil
	}
}

func (s *DeviceService) SaveDeviceConfAsOtherTemplte(ctx context.Context, req *pb.SaveDeviceConfAsOtherTemplateRequest) (*pb.CreateTemplateResponse, error) {
	log.Debug("SaveDeviceConfAsOtherTemplte")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	//get configs
	configObject, err1 := s.client.GetCoreEntitySpecContent(tm, req.Id, "device", "configs", "")
	if nil != err1 {
		return nil, err1
	}
	configs, ok := configObject["configs"]
	if !ok {
		log.Error("error config non exist")
		return nil, errors.New("error config non exist")
	}

	templateObject, err2, _ := s.SaveConfAsOtherTemplte(ctx, tm, configs, req.OtherTemplateInfo)
	if nil != err2 {
		log.Error("SaveConfAsOtherTemplte failed ", err2)
		return nil, err2
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

func (s *DeviceService) SaveDeviceConfAsTemplteAndRef(ctx context.Context, req *pb.SaveDeviceConfAsOtherTemplateRequest) (*pb.CreateTemplateResponse, error) {
	log.Debug("SaveDeviceConfAsTemplteAndRef")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	//get configs
	configObject, err1 := s.client.GetCoreEntitySpecContent(tm, req.Id, "device", "configs", "")
	if nil != err1 {
		return nil, err1
	}
	configs, ok := configObject["configs"]
	if !ok {
		log.Error("error config non exist")
		return nil, errors.New("error config non exist")
	}

	templateObject, err2, templateId := s.SaveConfAsOtherTemplte(ctx, tm, configs, req.OtherTemplateInfo)
	if nil != err2 {
		log.Error("SaveConfAsOtherTemplte failed ", err2)
		return nil, err2
	}
	//ref
	ma := make(map[string]interface{})
	ma["sysField._updatedAt"] = GetTime()
	ma["basicInfo.templateId"] = templateId
	ma["basicInfo.templateName"] = req.OtherTemplateInfo.Name
	_, err3 := s.client.CorePatchMethod(ctx, req.GetId(), ma, "", "replace", "/patch")
	if nil != err3 {
		log.Error("error patch dev entity", err3)
		return nil, err3
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

func (s *DeviceService) SaveConfAsOtherTemplte(ctx context.Context, tm map[string]string, configs interface{}, otherTemplateInfo *pb.TemplateBasicInfo) (map[string]interface{}, error, string) {
	errRepeated := s.checkNameRepated(ctx, otherTemplateInfo.Name)
	if nil != errRepeated {
		log.Debug("err:", errRepeated)
		return nil, errRepeated, ""
	}
	//get core url
	entityId := GetUUID()
	url := s.client.GetCoreUrl("", tm, "template") + "&id=" + entityId

	//fmt request
	sysField := &pbt.TemplateEntitySysField{
		XId:        entityId,
		XCreatedAt: GetTime(),
		XUpdatedAt: GetTime(),
		XOwner:     tm["owner"],
		XSource:    tm["source"],
	}

	templateInfo := &pbt.TemplateBasicInfo{
		Name:        otherTemplateInfo.Name,
		Description: otherTemplateInfo.Description,
	}

	entityInfo := &pbt.TemplateEntityCoreInfo{
		BasicInfo: templateInfo,
		SysField:  sysField,
	}
	log.Debug("entityinfo : ", entityInfo)
	data, err2 := json.Marshal(entityInfo)
	if nil != err2 {
		return nil, err2, ""
	}
	// do it
	res, err3 := s.client.Post(url, data)
	if nil != err3 {
		log.Error("error post data to core", err3)
		return nil, err3, ""
	}
	//fmt response
	templateObject := make(map[string]interface{})
	err4 := json.Unmarshal(res, &templateObject)
	if err4 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err4, ""
	}

	templateId := ""
	if id, ok := templateObject["id"]; ok {
		if idstr, ok1 := id.(string); ok1 {
			templateId = idstr
		}
	}
	if templateId == "" {
		log.Error("error templateId non exist")
		return nil, errors.New("error templateId non exist"), ""
	}
	log.Debug(templateId)

	//patch configs
	midUrl := "/" + templateId
	url1 := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("put url :", url1)

	data1 := make(map[string]interface{})
	data1["configs"] = configs
	log.Debug(data1)
	dt, err5 := json.Marshal(data1)
	if err5 != nil {
		return nil, err5, ""
	}
	resNew, err6 := s.client.Put(url1, dt)
	if nil != err6 {
		return nil, err6, ""
	}

	templateObjectNew := make(map[string]interface{})
	err7 := json.Unmarshal(resNew, &templateObjectNew)
	if err7 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err7, ""
	}
	return templateObjectNew, nil, templateId
}

func (s *DeviceService) CoreSearchEntity(ctx context.Context, listEntityQuery *pb.ListEntityQuery) (map[string]interface{}, error) {
	log.Debug("CoreSearchEntity")

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/search"
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("core url :", url)

	//Data isolation
	user := &pb.Condition{
		Field:    "owner",
		Operator: "$eq",
		Value:    structpb.NewStringValue(tm["owner"]),
		//Value:    tm["owner"],
	}
	listEntityQuery.Condition = append(listEntityQuery.Condition, user)

	log.Debug("Query:", listEntityQuery)

	//do it
	filter, err1 := json.Marshal(listEntityQuery)
	if err1 != nil {
		return nil, err1
	}
	res, err2 := s.client.Post(url, filter)
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

func (s *DeviceService) CoreSearchEntity2(listEntityQuery *pb.ListEntityQuery) (map[string]interface{}, error) {
	//log.Debug("CoreSearchEntity2")

	midUrl := "/search"
	url := coreUrl + midUrl
	//log.Debug("core url :", url)
	//log.Debug("Query:", listEntityQuery)

	//do it
	filter, err1 := json.Marshal(listEntityQuery)
	if err1 != nil {
		return nil, err1
	}
	res, err2 := s.client.Post(url, filter)
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

func (s *DeviceService) checkNameRepated(ctx context.Context, name string) error {
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
		Value:    structpb.NewStringValue(name),
	}
	condition2 := &pb.Condition{
		Field:    "type",
		Operator: "$eq",
		Value:    structpb.NewStringValue("device"),
		//Value:    "device",
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

func (s *DeviceService) MetricsSetDevNum(tenant string) error {
	//log.Debug(tenant)
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
		Field:    "sysField._tenantId",
		Operator: "$eq",
		Value:    structpb.NewStringValue(tenant),
	}
	condition2 := &pb.Condition{
		Field:    "type",
		Operator: "$eq",
		Value:    structpb.NewStringValue("device"),
		//Value:    "device",
	}
	query.Condition = append(query.Condition, condition1)
	query.Condition = append(query.Condition, condition2)
	//search
	listObject, err := s.CoreSearchEntity2(query)
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
	//check
	//log.Debug(tl)
	metrics.CollectorDeviceNumRequest.WithLabelValues(tenant).Set(tl)
	return nil
}

func (s *DeviceService) MetricsSetTemplateNum(tenant string) error {
	//log.Debug(tenant)
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
		Field:    "sysField._tenantId",
		Operator: "$eq",
		Value:    structpb.NewStringValue(tenant),
	}
	condition2 := &pb.Condition{
		Field:    "type",
		Operator: "$eq",
		Value:    structpb.NewStringValue("template"),
		//Value:    "device",
	}
	query.Condition = append(query.Condition, condition1)
	query.Condition = append(query.Condition, condition2)
	//search
	listObject, err := s.CoreSearchEntity2(query)
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
	//set
	//log.Debug(tl)
	metrics.CollectorDeviceTemplateRequest.WithLabelValues(tenant).Set(tl)
	return nil
}
func (s *DeviceService) MetricsSetDevOnlineNum(tenant string) error {
	//log.Debug(tenant)
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
		Field:    "sysField._tenantId",
		Operator: "$eq",
		Value:    structpb.NewStringValue(tenant),
	}
	condition2 := &pb.Condition{
		Field:    "connectInfo._online",
		Operator: "$eq",
		Value:    structpb.NewBoolValue(true),
		//Value:    "device",
	}
	query.Condition = append(query.Condition, condition1)
	query.Condition = append(query.Condition, condition2)
	//search
	listObject, err := s.CoreSearchEntity2(query)
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
	//set
	//log.Debug(tl)
	metrics.CollectorDeviceOnlineRequest.WithLabelValues(tenant).Set(tl)
	return nil
}
func (s *DeviceService) MetricsTimer() {
	for true {
		//log.Debug("MetricsTimer")
		//get tenants list
		tenants, err := s.client.GetTenantsList()
		if err != nil {
			log.Error(err)
		}
		//log.Debug(tenants)
		for _, v := range tenants {
			//get device num
			s.MetricsSetDevNum(v)
			//get template num
			s.MetricsSetTemplateNum(v)
			//get online num
			s.MetricsSetDevOnlineNum(v)
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
		//sleep
		time.Sleep(time.Duration(10) * time.Second)
	}
	return
}

func (s *DeviceService) GetDeviceBasicInfo(ctx context.Context, req *pb.GetDeviceBasicInfoRequest) (*pb.GetDeviceBasicInfoResponse, error) {
	log.Debug("GetDeviceBasicInfo")
	log.Debug("req:", req)

	re, err := s.GetDeviceDetailInfo(ctx, req.Id, "properties", "basicInfo")
	if nil != err {
		return nil, err
	}
	out := &pb.GetDeviceBasicInfoResponse{
		BasicInfoObject: re,
	}

	return out, nil
}

func (s *DeviceService) GetDeviceSysInfo(ctx context.Context, req *pb.GetDeviceSysInfoRequest) (*pb.GetDeviceSysInfoResponse, error) {
	log.Debug("GetDeviceSysInfo")
	log.Debug("req:", req)

	re, err := s.GetDeviceDetailInfo(ctx, req.Id, "properties", "sysField")
	if nil != err {
		return nil, err
	}
	out := &pb.GetDeviceSysInfoResponse{
		SysInfoObject: re,
	}

	return out, nil
}

func (s *DeviceService) GetDeviceConnectInfo(ctx context.Context, req *pb.GetDeviceConnectInfoRequest) (*pb.GetDeviceConnectInfoResponse, error) {
	log.Debug("GetDeviceConnectInfo")
	log.Debug("req:", req)

	re, err := s.GetDeviceDetailInfo(ctx, req.Id, "properties", "connectInfo")
	if nil != err {
		return nil, err
	}
	out := &pb.GetDeviceConnectInfoResponse{
		ConnectInfoObject: re,
	}

	return out, nil
}

func (s *DeviceService) GetDeviceRawData(ctx context.Context, req *pb.GetDeviceRawDataRequest) (*pb.GetDeviceRawDataResponse, error) {
	log.Debug("GetDeviceRawData")
	log.Debug("req:", req)

	re, err := s.GetDeviceDetailInfo(ctx, req.Id, "properties", "rawData")
	if nil != err {
		return nil, err
	}
	out := &pb.GetDeviceRawDataResponse{
		RawDataObject: re,
	}

	return out, nil
}
func (s *DeviceService) GetDeviceAttributeData(ctx context.Context, req *pb.GetDeviceAttributeDataRequest) (*pb.GetDeviceAttributeDataResponse, error) {
	log.Debug("GetDeviceAttributeData")
	log.Debug("req:", req)

	re, err := s.GetDeviceDetailInfo(ctx, req.Id, "properties", "attributes")
	if nil != err {
		return nil, err
	}
	out := &pb.GetDeviceAttributeDataResponse{
		AttributeDataObject: re,
	}

	return out, nil
}
func (s *DeviceService) GetDeviceTelemetryData(ctx context.Context, req *pb.GetDeviceTelemetryDataRequest) (*pb.GetDeviceTelemetryDataResponse, error) {
	log.Debug("GetDeviceTelemetryData")
	log.Debug("req:", req)

	re, err := s.GetDeviceDetailInfo(ctx, req.Id, "properties", "telemetry")
	if nil != err {
		return nil, err
	}
	out := &pb.GetDeviceTelemetryDataResponse{
		TelemetryDataObject: re,
	}

	return out, nil
}

func (s *DeviceService) GetDeviceDetailInfo(ctx context.Context, id string, classify string, pids string) (*structpb.Value, error) {
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core data
	obj, err1 := s.client.GetCoreEntitySpecContent(tm, id, "device", classify, pids)
	if nil != err1 {
		return nil, err1
	}
	//cut out the excess
	var infoObj map[string]interface{}
	if prop, ok := obj["properties"]; ok {
		if prop1, ok1 := prop.(map[string]interface{}); ok1 {
			if info, ok2 := prop1[pids]; ok2 {
				if info1, ok3 := info.(map[string]interface{}); ok3 {
					infoObj = info1
				}
			}
		}
	}

	//return
	re, err2 := structpb.NewValue(infoObj)
	if nil != err2 {
		log.Error("convert tree failed ", err2)
		return nil, err2
	}

	return re, nil
}
