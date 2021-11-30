package service

import (
	"context"

	pb "github.com/tkeel-io/tkeel-device/api/measure/v1"
)

type MeasureService struct {
	pb.UnimplementedMeasureServer
}

func NewMeasureService() *MeasureService {
	return &MeasureService{}
}

func (s *MeasureService) CreateMeasure(ctx context.Context, req *pb.CreateMeasureRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}

func (s *MeasureService) UpdateMeasure(ctx context.Context, req *pb.UpdateMeasureRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}

func (s *MeasureService) DeleteMeasure(ctx context.Context, req *pb.DeleteMeasureRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}

func (s *MeasureService) GetMeasure(ctx context.Context, req *pb.GetMeasureRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}

func (s *MeasureService) ListMeasure(ctx context.Context, req *pb.ListMeasureRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}
