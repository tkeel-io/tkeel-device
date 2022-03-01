package service

import (
	"context"
	json "encoding/json"
	"errors"
	"fmt"
	"github.com/tkeel-io/kit/log"
	pb "github.com/tkeel-io/tkeel-device/api/device/v1"
	//go_struct "google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

type DeviceService struct {
	pb.UnimplementedDeviceServer
	client *CoreClient
}

func NewDeviceService() *DeviceService {
	return &DeviceService{
		client: NewCoreClient(),
	}
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
	err6 := s.client.setSpacePathMapper(tm, devId, req.DevBasicInfo.ParentId)
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

func (s *DeviceService) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.UpdateDeviceResponse, error) {
	log.Debug("UpdateDevice")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/" + req.Id
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("get url :", url)

	updateBasicInfo := &pb.UpdateDeviceEntityCoreInfo{}
	updateBasicInfo.BasicInfo = req.DevBasicInfo

	//do it
	//update BasicInfo
	dev, err1 := json.Marshal(updateBasicInfo)
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
	err6 := s.client.setSpacePathMapper(tm, req.Id, req.DevBasicInfo.ParentId)
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

func (s *DeviceService) CreateDeviceDataRelation(ctx context.Context, req *pb.CreateDeviceDataRelationRequest) (*emptypb.Empty, error) {
	log.Debug("CreateDataRelation")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/" + req.GetId() + "/patch"
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("get url :", url)

	//write relation

	//create mapper

	return &emptypb.Empty{}, nil
}
func (s *DeviceService) UpdateDeviceDataRelation(ctx context.Context, req *pb.UpdateDeviceDataRelationRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *DeviceService) DeleteDeviceDataRelation(ctx context.Context, req *pb.DeleteDeviceDataRelationRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *DeviceService) ListDeviceDataRelation(ctx context.Context, req *pb.ListDeviceDataRelationRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *DeviceService) SetDeviceRaw(ctx context.Context, req *pb.SetDeviceRawRequest) (*emptypb.Empty, error) {
	log.Debug("SetDeviceRaw")
	log.Debug("req:", req)

	ma := make(map[string]interface{})
	ma["value"] = req.Value
	_, err3 := s.client.CorePatchMethod(ctx, req.GetId(), ma, "rawDown.", "replace", "/patch")
	if nil != err3 {
		log.Error("error patch _updateAt", err3)
		return nil, err3
	}
	return &emptypb.Empty{}, nil
}

func (s *DeviceService) SetDeviceAttribte(ctx context.Context, req *pb.SetDeviceAttributeRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (s *DeviceService) SetDeviceCommand(ctx context.Context, req *pb.SetDeviceCommandRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
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
