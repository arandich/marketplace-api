package main

import (
	"context"
	"github.com/arandich/marketplace-api/internal/config"
	"github.com/arandich/marketplace-api/pkg/metrics"
	sdkPrometheus "github.com/arandich/marketplace-sdk/prometheus"
	sdkRdb "github.com/arandich/marketplace-sdk/redis"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"net"
)

func initHTTP(ctx context.Context, cfg config.HttpConfig) (net.Listener, error) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("initializing HTTP server listener")

	lis, err := net.Listen(cfg.Network, cfg.Address)
	if err != nil {
		return nil, err
	}

	return lis, nil
}

func initGRPC(ctx context.Context, cfg config.GrpcConfig) (net.Listener, error) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("initializing GRPC server listener")

	lis, err := net.Listen(cfg.Network, cfg.Address)
	if err != nil {
		return nil, err
	}

	return lis, nil
}

func initMetrics(ctx context.Context, cfg config.PrometheusConfig) metrics.Metrics {
	logger := zerolog.Ctx(ctx)
	logger.Info().Str("namespace", cfg.Namespace).Str("subsystem", cfg.Subsystem).Msg("initializing prometheus metrics")

	promCfg := sdkPrometheus.Config{
		Namespace: cfg.Namespace,
		Subsystem: cfg.Subsystem,
	}

	baseMetrics := sdkPrometheus.New(promCfg)
	orderMetrics := metrics.NewOrderMetrics(promCfg)
	promMetrics := metrics.New(baseMetrics, orderMetrics)

	return promMetrics
}

func initRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	logger := zerolog.Ctx(ctx)
	logger.Info().Msg("initializing redis connection")

	redisCfg := sdkRdb.Config{
		Addr:     cfg.ConnAddr,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	rdb, err := sdkRdb.Connect(ctx, redisCfg)
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

func initAwsQueue(ctx context.Context, cfg config.QueueConfig) (*sqs.Client, error) {

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           cfg.URL,
			SigningRegion: cfg.Region,
		}, nil
	})

	cfgAws, err := awsConfig.LoadDefaultConfig(
		ctx,
		awsConfig.WithEndpointResolverWithOptions(customResolver),
		awsConfig.WithCredentialsProvider(&credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     cfg.AWSAccessKeyID,
				SecretAccessKey: cfg.AWSSecretAccessKey,
			}}),
	)
	if err != nil {
		return nil, err
	}

	client := sqs.NewFromConfig(cfgAws)

	return client, nil
}
