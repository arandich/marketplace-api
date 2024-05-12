package main

import (
	"context"
	"fmt"
	"github.com/arandich/marketplace-api/internal/config"
	"github.com/arandich/marketplace-api/internal/model"
	"github.com/arandich/marketplace-api/internal/queue"
	"github.com/arandich/marketplace-api/internal/repository"
	"github.com/arandich/marketplace-api/internal/service"
	grpcTransport "github.com/arandich/marketplace-api/internal/transport/grpc"
	"github.com/arandich/marketplace-api/internal/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"syscall"
)

func runApp(ctx context.Context, cfg config.Config) {
	logger := zerolog.Ctx(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt, os.Kill)

	// Prometheus.
	promMetrics := initMetrics(ctx, cfg.Prometheus)

	// HTTP
	httpLis, err := initHTTP(ctx, cfg.HTTP)
	if err != nil {
		logger.Fatal().Err(err).Msg("error connecting to HTTP server")
	}
	defer func() {
		if err = httpLis.Close(); err != nil {
			logger.Error().Err(err).Msg("error closing HTTP listener")
		}
	}()

	srv := http.NewServer(httpLis, cfg.HTTP)
	httpErrCh := srv.StartAndServe(promMetrics)

	// GRPC
	grpcLis, err := initGRPC(ctx, cfg.GRPC)
	if err != nil {
		logger.Fatal().Err(err).Msg("error initializing GRPC listener")
	}
	defer func() {
		if err = grpcLis.Close(); err != nil {
			logger.Error().Err(err).Msg("error closing GRPC listener")
		}
	}()

	redisConn, err := initRedis(ctx, cfg.Redis)
	if err != nil {
		logger.Fatal().Err(err).Msg("error initializing Redis connection")
	}
	defer redisConn.Close()

	// AWS Queue
	awsClient, err := initAwsQueue(ctx, cfg.Queue)
	if err != nil {
		logger.Fatal().Err(err).Msg("error initializing AWS Queue")
	}

	queueName := "order_submit"

	orderQueueURL, err := awsClient.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: &queueName,
		Attributes: map[string]string{
			"DelaySeconds":                  "0",
			"ReceiveMessageWaitTimeSeconds": "20",
		},
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("error creating AWS Queue")
	}

	fmt.Println("Queue created, URL: " + *orderQueueURL.QueueUrl)

	orderQueue := queue.NewAwsQueue(ctx, awsClient, *orderQueueURL.QueueUrl, promMetrics, redisConn)

	for i := 0; i < cfg.Queue.Workers; i++ {
		go func() {
			err := orderQueue.StartReceiving(ctx)
			if err != nil {
				logger.Error().Err(err).Msg("error receiving from queue")
			}
		}()
	}

	clients := model.Clients{}

	services := model.Services{
		ApiService: service.NewIdService(repository.NewApiRepository(ctx, redisConn, promMetrics, cfg, clients, orderQueue)),
	}
	// GRPC.
	grpcTrSrv := grpcTransport.New(ctx, cfg.GRPC)
	grpcSrv, grpcErrCh := grpcTrSrv.Start(ctx, grpcLis, services, promMetrics)
	defer grpcSrv.GracefulStop()

	logger.Info().Str("service", cfg.App.Name).Msg("service started")

	for {
		select {
		case err = <-grpcErrCh:
			logger.Error().Err(err).Msg("retrieved error from GRPC server")
			c <- os.Kill
		case err = <-httpErrCh:
			logger.Error().Err(err).Msg("retrieved error from HTTP server")
			c <- os.Kill
		case sig := <-c:
			logger.Warn().Str("signal", sig.String()).Msg("received shutdown signal")
			return
		}
	}
}
