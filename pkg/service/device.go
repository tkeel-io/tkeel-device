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

	//1. verify Authentication in header and get user token map
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		log.Debug("err:", err)
		return nil, err
	}

	// 2. get url
	devId := GetUUID()
	url := s.client.GetCoreUrl("", tm, "device") + "&id=" + devId
	log.Debug("get url: ", url)

	// 3. build coreInfo and add system value
	coreInfo := new(pb.DeviceEntityCoreInfo)
	coreInfo.BasicInfo = new(pb.DeviceEntityBasicInfo)
	coreInfo.BasicInfo = req.DevBasicInfo
	coreInfo.SysField = new(pb.DeviceEntitySysField)
	coreInfo.SysField.XId = devId
	coreInfo.SysField.XCreatedAt = GetTime()
	coreInfo.SysField.XUpdatedAt = GetTime()
	coreInfo.SysField.XEnable = true
	coreInfo.SysField.XStatus = false
	coreInfo.SysField.XOwner = tm["owner"]
	coreInfo.SysField.XSource = tm["source"]

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
	deviceObject := &pb.EntityResponse{} // core define
	err5 := json.Unmarshal(res, deviceObject)
	if err5 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err5
	}
	out := &pb.CreateDeviceResponse{
		DeviceObject: deviceObject,
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

	updateBasicInfo := &pb.DeviceEntityBasicInfo{} 
    updateBasicInfo = req.DevBasicInfo 
    
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
    _, err3 := s.CorePatchMethod(ctx, req.GetId(), ma, "sysField.", "replace")
	if nil != err3 {
		log.Error("error patch _updateAt", err3)
		return nil, err1 
	}

    //fmt response 
	er := new(pb.EntityResponse)
	if err4 := json.Unmarshal(res, er); nil != err4 {
		return nil, err4 
	}

	out := &pb.UpdateDeviceResponse{
		DeviceObject: er,
	}

	return out, nil
}

func (s *DeviceService) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*emptypb.Empty, error) {
	log.Debug("DeleteDevice")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	ids := req.Ids.GetIds()
	for _, id := range ids {
		midUrl := "/" + id
		url := s.client.GetCoreUrl(midUrl, tm, "device")
		log.Debug("get url:", url)

		_, err2 := s.client.Delete(url)
		if nil != err2 {
			return nil, err2
		}
	}
	return &emptypb.Empty{}, nil
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
	er := new(pb.EntityResponse)
	if err3 := json.Unmarshal(res, er); nil != err3 {
		return nil, err3
	}

	out := &pb.GetDeviceResponse{
		DeviceObject: er,
	}

	return out, nil
}

func (s *DeviceService) ListDevice(ctx context.Context, req *pb.ListDeviceRequest) (*pb.ListDeviceResponse, error) {
	log.Debug("ListDevice")
	//req.Filter.Page.Reverse = req.Filter.Page.GetReverse()
	//req.Filter.Page.Limit = req.Filter.Page.GetLimit()
	//req.Filter.Page.Offset = req.Filter.Page.GetOffset()
	//log.Debug("req:", req, req.Filter.Page)
	log.Debug("req:", req, req.ListEntityQuery)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}
	midUrl := "/search"
	url := s.client.GetCoreUrl(midUrl, tm, "device")
	log.Debug("get url :", url)

	filter, err1 := json.Marshal(req.ListEntityQuery)
	if err1 != nil {
		return nil, err1
	}
	res, err2 := s.client.Post(url, filter)
	if nil != err2 {
		return nil, err2
	}

	listDeviceObject := &pb.ListEntityResponse{}
	err3 := json.Unmarshal(res, listDeviceObject)
	if err3 != nil {
		log.Error("error Unmarshal data from core")
		return nil, err3
	}
	out := &pb.ListDeviceResponse{
		ListDeviceObject: listDeviceObject,
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

func (s *DeviceService) CorePatchMethod(ctx context.Context, entityId string, kv map[string]interface{}, path string, operator string) (*emptypb.Empty, error) {
	log.Debug("CorePatchMethod")
	log.Debug("path:", path)
	log.Debug("operator:", operator)

	//get token
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err {
		return nil, err
	}

	//get core url
	midUrl := "/" + entityId + "/patch"
	url := s.client.GetCoreUrl(midUrl, tm, "group")
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
	_, err4 := s.client.Put(url, data)
	if nil != err4 {
		log.Error("error post data to core", data)
		return nil, err4
	}

	//fmt response
	return &emptypb.Empty{}, nil
}
