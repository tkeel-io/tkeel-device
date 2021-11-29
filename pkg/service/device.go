package service

import (
	"context"
	pb "device/api/device/v1"
	json "encoding/json"
	"errors"
	"fmt"
	"github.com/tkeel-io/kit/log"
	go_struct "google.golang.org/protobuf/types/known/structpb"
)
const ResOK string = "ok"
const ResFailed string = "failed"

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
	log.Debug("req:", req)

	//1. verify Authentication in header and get user token map
	//tm, err := s.client.GetTokenMap(ctx)
	//if nil != err{
	//	return nil, err
	//}
	tm := map[string]string{
		//"id": "uuid",
		"entityType": "device",
		"owner": "abc",
		"source": "device",
	}

	//2. get url
	devId := GetUUID()
	url := s.client.GetCoreUrl("", tm) + "&id=" + devId
	log.Debug("get url: ", url)

	//3. build coreInfo and add system value
	coreInfo := new(pb.DeviceEntityCoreInfo)
	coreInfo.Dev = new(pb.DeviceEntity)
	coreInfo.Dev = req.Dev
	coreInfo.SysField = new(pb.DeviceEntitySysField)
	coreInfo.SysField.XId = devId
	coreInfo.SysField.XCreatedAt = GetTime()
	coreInfo.SysField.XUpdatedAt = GetTime()
	coreInfo.SysField.XEnable = true
	coreInfo.SysField.XStatus = false

	//4. create device token
	//token, err2 := s.client.CreatEntityToken("device", coreInfo.SysField.XId)
	//if nil != err2{
	//	return nil, err2
	//}
	token := "token"
	coreInfo.SysField.XToken = token

	//4. add internal values and marshal Dev
	dev, err3 := json.Marshal(coreInfo)
	if nil != err3 {
		return nil, err3
	}

	//5. core request
	res, err4 := s.client.Post(url, dev)
	if nil != err4 {
		log.Error("error post data to core", string(dev))
		return nil, err4
	}
	er := new(pb.EntityResponse)
	if err5 := json.Unmarshal(res, er); nil!=err5{
		return nil, err5
	}

	kv := er.Properties.GetStructValue().Fields
	devRes := new(pb.CreateDeviceResponse)
	fieldV, err6 := json.Marshal(kv)
	if nil != err6{
		return nil, err6
	}
	err7 := json.Unmarshal(fieldV, devRes)
	if nil != err7{
		return nil, err7
	}

	return devRes, nil
}

func (s *DeviceService) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.UpdateDeviceResponse, error) {
	log.Debug("UpdateDevice")
	log.Debug("req:", req)

	//tm, err := s.client.GetTokenMap(ctx)
	//if nil != err{
	//	return nil, err
	//}
	tm := map[string]string{
		//"id": "uuid",
		"entityType": "device",
		"owner": "abc",
		"source": "device",
	}
	midUrl := "/" + req.Id
	url := s.client.GetCoreUrl(midUrl, tm)
	log.Debug("get url :", url, req.Dev)

	coreInfo := new(pb.DeviceEntityCoreInfo)
	coreInfo.Dev = new(pb.DeviceEntity)
	coreInfo.Dev = req.Dev
	coreInfo.SysField = new(pb.DeviceEntitySysField)
	coreInfo.SysField.XUpdatedAt = GetTime()

	dev, err := json.Marshal(coreInfo)
	if err != nil {
		return nil, err
	}
	res, err2 := s.client.Put(url, dev)
	if nil != err2{
		return nil, err2
	}
	er := new(pb.EntityResponse)
	if err3 := json.Unmarshal(res, er); nil!=err3{
		return nil, err3
	}

	kv := er.Properties.GetStructValue().Fields
	devRes := new(pb.UpdateDeviceResponse)
	fieldV, err6 := json.Marshal(kv)
	if nil != err6{
		return nil, err6
	}
	err7 := json.Unmarshal(fieldV, devRes)
	if nil != err7{
		return nil, err7
	}

	return devRes, nil
}

func (s *DeviceService) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.DeleteDeviceResponse, error) {
	log.Debug("DeleteDevice")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err{
		return nil, err
	}
	ids := req.Ids.GetIds()
	for _, id :=range ids{
		midUrl := "/" + id
		url := s.client.GetCoreUrl(midUrl, tm)
		log.Debug("get url:", url)

		_, err2 := s.client.Delete(url)
		if nil != err2{
			return &pb.DeleteDeviceResponse{Result: ResFailed}, nil
		}
	}
	return &pb.DeleteDeviceResponse{Result: ResOK}, nil
}

func (s *DeviceService) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.GetDeviceResponse, error) {
	log.Debug("GetDevice")
	log.Debug("req:", req)
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err{
		return nil, err
	}
	midUrl := "/" + req.GetId()
	url := s.client.GetCoreUrl(midUrl, tm)
	log.Debug("get url :", url)

	res, err2 := s.client.Get(url)
	if nil != err2{
		return nil, err2
	}
	er := new(pb.EntityResponse)
	if err3 := json.Unmarshal(res, er); nil!=err3{
		return nil, err3
	}

	kv := er.Properties.GetStructValue().Fields
	devRes := new(pb.GetDeviceResponse)
	fieldV, err6 := json.Marshal(kv)
	if nil != err6{
		return nil, err6
	}
	err7 := json.Unmarshal(fieldV, devRes)
	if nil != err7{
		return nil, err7
	}

	return devRes, nil
}

func (s *DeviceService) ListDevice(ctx context.Context, req *pb.ListDeviceRequest) (*pb.ListDeviceResponse, error) {
	log.Debug("ListDevice")
	req.Filter.Page.Reverse = req.Filter.Page.GetReverse()
	req.Filter.Page.Limit = req.Filter.Page.GetLimit()
	req.Filter.Page.Offset = req.Filter.Page.GetOffset()
	log.Debug("req:", req, req.Filter.Page)

	//fixme
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err{
		return nil, err
	}
	midUrl := "/search"
	url := s.client.GetCoreUrl(midUrl, tm)
	log.Debug("get url :", url)

	filter, err1 := json.Marshal(req.Filter)
	if err1 != nil {
		return nil, err1
	}
	res, err2 := s.client.Post(url, filter)
	if nil != err2{
		return nil, err2
	}
	var er interface{}
	if err3 := json.Unmarshal(res, &er); nil!=err3{
		log.Error("error Unmarshal", err3)
		return nil, err3
	}
	value, err4 := go_struct.NewValue(er)
	if nil != err4{
		return nil, err4
	}
	return &pb.ListDeviceResponse{Result: value}, nil
}

func (s *DeviceService) EnableDevice(ctx context.Context, req *pb.EnableDeviceRequest) (*pb.EnableDeviceResponse, error) {
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
	return &pb.EnableDeviceResponse{Result: ResOK}, nil
}

func (s *DeviceService) AddDeviceExt(ctx context.Context, req *pb.AddDeviceExtRequest) (*pb.AddDeviceExtResponse, error) {
	log.Debug("AddDeviceExt")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err{
		return nil, err
	}
	midUrl := "/" + req.GetId()
	url := s.client.GetCoreUrl(midUrl, tm)
	log.Debug("get url :", url)

	var exts []interface{}
	switch kv := req.Ext.AsInterface().(type) {
	case map[string]interface{}:
		for k, v := range kv {
			e := map[string]interface{}{
				"path": fmt.Sprintf("dev.ext.%s", k),
				"operator": "replace",
				"value": v,
			}
			exts = append(exts, e)
		}
	default:
		return nil, errors.New("error Invalid payload")
	}

	log.Debug("ext body: ", exts)
	data, err1 := json.Marshal(exts)
	if err1 != nil {
		return &pb.AddDeviceExtResponse{Result: ResFailed}, err1
	}
	_, err2 := s.client.Patch(url, data)
	if nil != err2{
		return &pb.AddDeviceExtResponse{Result: ResFailed}, err2
	}

	return &pb.AddDeviceExtResponse{Result: ResOK}, nil
}

func (s *DeviceService) DeleteDeviceExt(ctx context.Context, req *pb.DeleteDeviceExtRequest) (*pb.DeleteDeviceExtResponse, error) {
	log.Debug("DeleteDeviceExt")
	log.Debug("req:", req)
	//todo when core support
	tm, err := s.client.GetTokenMap(ctx)
	if nil != err{
		return nil, err
	}
	midUrl := "/" + req.GetId()
	url := s.client.GetCoreUrl(midUrl, tm)

	//var exts []interface{}
	keys := req.Keys.Keys
	exts := make([]interface{}, len(keys))
	for i, k := range keys {
		e := map[string]interface{}{
			"path": fmt.Sprintf("dev.ext.%s", k),
			"operator": "remove",
			"value": "",
		}
		exts[i] = e
	}

	log.Debug("ext body: ", exts)
	data, err1 := json.Marshal(exts)
	if err1 != nil {
		return &pb.DeleteDeviceExtResponse{Result: ResFailed}, err1
	}
	_, err2 := s.client.Patch(url, data)
	if nil != err2{
		return &pb.DeleteDeviceExtResponse{Result: ResFailed}, err2
	}
	return &pb.DeleteDeviceExtResponse{Result: ResOK}, nil
}

func (s *DeviceService) UpdateDeviceExt(ctx context.Context, req *pb.UpdateDeviceExtRequest) (*pb.UpdateDeviceExtResponse, error) {
	log.Debug("UpdateDeviceExt")
	log.Debug("req:", req)

	tm, err := s.client.GetTokenMap(ctx)
	if nil != err{
		return nil, err
	}
	midUrl := "/" + req.GetId()
	url := s.client.GetCoreUrl(midUrl, tm)
	log.Debug("get url :", url)

	var exts [1]interface{}
	exts[0]= map[string]interface{}{
		"path": fmt.Sprintf("dev.ext.%s", req.Ext.GetKey()),
		"operator": "replace",
		"value": req.Ext.GetValue(),
	}
	log.Debug("ext body: ", exts)
	data, err1 := json.Marshal(exts)
	if err1 != nil {
		return &pb.UpdateDeviceExtResponse{Result: ResFailed}, err1
	}
	_, err2 := s.client.Patch(url, data)
	if nil != err2{
		return &pb.UpdateDeviceExtResponse{Result: ResFailed}, err2
	}
	return &pb.UpdateDeviceExtResponse{Result: ResOK}, nil
}
