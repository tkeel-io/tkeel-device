package service

import (
	"context"
	pb "device/api/device/v1"
	json "encoding/json"
	"github.com/tkeel-io/kit/log"
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

func (s *DeviceService) CreateDevice(ctx context.Context, req *pb.CreateDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("CreateDevice")
	log.Debug("req:", req)


	if _, err := s.client.GetToken(ctx); nil != err{
		return nil, err
	}
	url := s.client.GetCoreUrl("")
	log.Debug("get url: ", url)

	dev, err3 := json.Marshal(req.Dev)
	if nil != err3 {
		return nil, err3
	}

	res, err4 := s.client.Post(ctx, dev)
	if nil != err4 {
		log.Error("error post data to core", dev)
		return nil, err4
	}
	return &pb.CommonResponse{Result: string(res)}, nil
}

func (s *DeviceService) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("UpdateDevice")
	log.Debug("req:", req)

	if _, err := s.client.GetToken(ctx); nil != err{
		return nil, err
	}
	midUrl := "/" + req.Dev.XId
	url := s.client.GetCoreUrl(midUrl)
	log.Debug("get url :", url)

	dev, err := json.Marshal(req.Dev)
	if err != nil {
		return nil, err
	}
	res, err2 := s.client.Put(ctx, dev)
	if nil != err2{
		log.Error("error put data to core", dev)
		return nil, err2
	}
	return &pb.CommonResponse{Result: string(res)}, nil
}

func (s *DeviceService) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("DeleteDevice")
	log.Debug("req:", req)

	if _, err := s.client.GetToken(ctx); nil != err{
		return nil, err
	}
	//fixme
	//midUrl := "/" + req.Ids.GetIds()
	//url := s.client.GetCoreUrl(midUrl)
	//log.Debug("get url:", url)

	ids, err := json.Marshal(req.Ids)
	if err != nil {
		return &pb.CommonResponse{Result: "failed"}, nil
	}
	_, err2 := s.client.Post(ctx, ids)
	if nil != err2{
		log.Error("error delete data", ids)
		return &pb.CommonResponse{Result: "failed"}, nil
	}
	return &pb.CommonResponse{Result: "ok"}, nil
}

func (s *DeviceService) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("GetDevice")
	log.Debug("req:", req)

	if _, err := s.client.GetToken(ctx); nil != err{
		return nil, err
	}
	midUrl := "/" + req.GetId()
	url := s.client.GetCoreUrl(midUrl)
	log.Debug("get url :", url)

	res, err := s.client.Get(ctx, req.Id)
	if nil != err{
		log.Error("error get data from core")
		return nil, err
	}
	return &pb.CommonResponse{Result: string(res)}, nil
}

func (s *DeviceService) ListDevice(ctx context.Context, req *pb.ListDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("ListDevice")
	log.Debug("req:", req)

	//fixme
	filter, err := json.Marshal(req.Filter)
	if err != nil {
		return nil, err
	}
	res, err2 := s.client.Post(ctx, filter)
	if nil != err2{
		log.Error("error list data from core")
		return nil, err2
	}
	return &pb.CommonResponse{Result: string(res)}, nil
}

func (s *DeviceService) EnableDevice(ctx context.Context, req *pb.EnableDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("EnableDevice")
	log.Debug("req:", req)
	ext := map[string]interface{}{
		"enable": req.Enable,
		//device id
		"id": req.Id,
	}
	data, err := json.Marshal(ext)
	if err != nil {
		return &pb.CommonResponse{Result: "failed"}, err
	}
	_, err2 := s.client.Put(ctx, data)
	if nil != err2{
		log.Error("error put data to core")
		return &pb.CommonResponse{Result: "failed"}, err2
	}
	return &pb.CommonResponse{Result: "ok"}, nil
}
func (s *DeviceService) AddDeviceExt(ctx context.Context, req *pb.AddDeviceExtRequest) (*pb.CommonResponse, error) {
	log.Debug("AddDeviceExt")
	log.Debug("req:", req)
	return &pb.CommonResponse{}, nil
}
func (s *DeviceService) DeleteDeviceExt(ctx context.Context, req *pb.DeleteDeviceExtRequest) (*pb.CommonResponse, error) {
	log.Debug("DeleteDeviceExt")
	log.Debug("req:", req)
	return &pb.CommonResponse{}, nil
}
func (s *DeviceService) UpdateDeviceExt(ctx context.Context, req *pb.UpdateDeviceExtRequest) (*pb.CommonResponse, error) {
	log.Debug("UpdateDeviceExt")
	log.Debug("req:", req)
	return &pb.CommonResponse{}, nil
}
