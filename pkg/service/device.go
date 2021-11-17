package service

import (
	"context"

	pb "device/api/device/v1"
)

type DeviceService struct {
	pb.UnimplementedDeviceServer
}

func NewDeviceService() *DeviceService {
	return &DeviceService{}
}

func (s *DeviceService) CreateDevice(ctx context.Context, req *pb.CreateDeviceRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}
func (s *DeviceService) UpdateDevice(ctx context.Context, req *pb.UpdateDeviceRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}
func (s *DeviceService) DeleteDevice(ctx context.Context, req *pb.DeleteDeviceRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}
func (s *DeviceService) GetDevice(ctx context.Context, req *pb.GetDeviceRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}
func (s *DeviceService) ListDevice(ctx context.Context, req *pb.ListDeviceRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}
