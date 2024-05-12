package service

import (
	"context"
	pb "github.com/arandich/marketplace-proto/api/proto/services"
)

type ApiRepository interface {
	SubmitOrder(ctx context.Context, req *pb.SubmitOrderRequest) (*pb.SubmitOrderResponse, error)
}

var _ ApiRepository = (*ApiService)(nil)

type ApiService struct {
	pb.UnimplementedApiServiceServer
	repository ApiRepository
}

func NewIdService(repository ApiRepository) ApiService {
	return ApiService{
		repository: repository,
	}
}

func (s ApiService) SubmitOrder(ctx context.Context, req *pb.SubmitOrderRequest) (*pb.SubmitOrderResponse, error) {
	return s.repository.SubmitOrder(ctx, req)
}
