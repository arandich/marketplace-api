package repository

import (
	"context"
	"errors"
	"github.com/arandich/marketplace-api/internal/config"
	"github.com/arandich/marketplace-api/internal/model"
	"github.com/arandich/marketplace-api/internal/queue"
	"github.com/arandich/marketplace-api/pkg/metrics"
	pb "github.com/arandich/marketplace-proto/api/proto/services"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"time"
)

type ApiRepository struct {
	promMetrics metrics.Metrics
	logger      *zerolog.Logger
	cfg         config.Config
	clients     model.Clients
	orderQueue  queue.IOrderQueue
	redisClient *redis.Client
}

func NewApiRepository(ctx context.Context, redisConn *redis.Client, promMetrics metrics.Metrics, cfg config.Config, clients model.Clients, queueClient queue.IOrderQueue) *ApiRepository {
	return &ApiRepository{
		promMetrics: promMetrics,
		logger:      zerolog.Ctx(ctx),
		cfg:         cfg,
		clients:     clients,
		orderQueue:  queueClient,
		redisClient: redisConn,
	}
}

func (a *ApiRepository) SubmitOrder(ctx context.Context, req *pb.SubmitOrderRequest) (*pb.SubmitOrderResponse, error) {

	if err := validateRequest(req); err != nil {
		a.logger.Error().Err(err).Msg("request validation failed")
		return nil, err
	}

	timeStart := time.Now()

	dto := model.OrderMsg{
		ActionID: req.ActionID,
		ItemIDs:  req.ItemIDs,
	}

	err := a.orderQueue.PublishOrder(ctx, dto)
	if err != nil {
		a.logger.Error().Err(err).Msg("failed to push order to queue")
		return nil, err
	}

	a.redisClient.Set(ctx, req.ActionID, timeStart.Unix(), time.Minute*3)

	return &pb.SubmitOrderResponse{Status: model.InProgress}, nil
}

func validateRequest(req *pb.SubmitOrderRequest) error {
	if req == nil {
		return errors.New("request is nil")
	}

	if req.ClientID == "" {
		return errors.New("client id is empty")
	}

	if req.ItemIDs == nil || len(req.ItemIDs) == 0 {
		return errors.New("item ids are empty")
	}

	if req.ActionID == "" {
		return errors.New("action id is empty")
	}

	return nil
}
