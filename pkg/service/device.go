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
	dev, err := json.Marshal(req.Dev)
	if err != nil {
		return nil, err
	}
	res, err := s.client.Post(dev)
	if nil != err {
		log.Error("error post data to core", dev)
		return nil, err
	}
	return &pb.CommonResponse{Result: string(res)}, nil
}
func (s *DeviceService) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("UpdateDevice")
	log.Debug("req:", req)
	dev, err := json.Marshal(req.Dev)
	if err != nil {
		return nil, err
	}
	res, err2 := s.client.Put(dev)
	if nil != err2{
		log.Error("error put data to core", dev)
		return nil, err2
	}
	return &pb.CommonResponse{Result: string(res)}, nil
}
func (s *DeviceService) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("DeleteDevice")
	log.Debug("req:", req)
	ids, err := json.Marshal(req.Ids)
	if err != nil {
		return &pb.CommonResponse{Result: "failed"}, nil
	}
	_, err2 := s.client.Post(ids)
	if nil != err2{
		log.Error("error delete data", ids)
		return &pb.CommonResponse{Result: "failed"}, nil
	}
	return &pb.CommonResponse{Result: "ok"}, nil
}
func (s *DeviceService) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("GetDevice")
	log.Debug("req:", req)
	res, err := s.client.Get(req.Id)
	if nil != err{
		log.Error("error get data from core")
		return nil, err
	}
	return &pb.CommonResponse{Result: string(res)}, nil
}
func (s *DeviceService) ListDevice(ctx context.Context, req *pb.ListDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("ListDevice")
	log.Debug("req:", req)
	filter, err := json.Marshal(req.Filter)
	if err != nil {
		return nil, err
	}
	res, err2 := s.client.Post(filter)
	if nil != err2{
		log.Error("error list data from core")
		return nil, err2
	}
	return &pb.CommonResponse{Result: string(res)}, nil
}
