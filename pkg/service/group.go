package service

import (
	"context"

	pb "github.com/tkeel-io/tkeel-device/api/group/v1"
)

type GroupService struct {
	pb.UnimplementedGroupServer
}

func NewGroupService() *GroupService {
	return &GroupService{}
}

func (s *GroupService) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}

func (s *GroupService) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}

func (s *GroupService) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}

func (s *GroupService) GetGroup(ctx context.Context, req *pb.GetGroupRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}

func (s *GroupService) ListGroup(ctx context.Context, req *pb.ListGroupRequest) (*pb.CommonResponse, error) {
	return &pb.CommonResponse{}, nil
}
