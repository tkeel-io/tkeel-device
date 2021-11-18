package service

import (
	"context"
	pb "device/api/device/v1"
	json "encoding/json"
	"github.com/tkeel-io/kit/log"
)

type DeviceService struct {
	pb.UnimplementedDeviceServer
}

func NewDeviceService() *DeviceService {
	return &DeviceService{}
}

func (s *DeviceService) CreateDevice(ctx context.Context, req *pb.CreateDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("CreateDevice")
	log.Debug("req:", req)
	dev, err := json.Marshal(req.Dev)
	if err != nil {
		return nil, err
	}
	//todo publish data to core
	return &pb.CommonResponse{Result: string(dev)}, nil
}
func (s *DeviceService) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("UpdateDevice")
	log.Debug("req:", req)
	return &pb.CommonResponse{}, nil
}
func (s *DeviceService) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("DeleteDevice")
	log.Debug("req:", req)
	return &pb.CommonResponse{}, nil
}
func (s *DeviceService) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("GetDevice")
	log.Debug("req:", req)
	return &pb.CommonResponse{}, nil
}
func (s *DeviceService) ListDevice(ctx context.Context, req *pb.ListDeviceRequest) (*pb.CommonResponse, error) {
	log.Debug("ListDevice")
	log.Debug("req:", req)
	return &pb.CommonResponse{}, nil
}
