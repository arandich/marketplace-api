package queue

import (
	"context"
	"github.com/arandich/marketplace-api/internal/model"
	"github.com/arandich/marketplace-api/pkg/metrics"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"log"
	"time"
)

type OrderQueue struct {
	client   *sqs.Client
	queueURL string
	metrics  metrics.Metrics
	log      *zerolog.Logger
	rdb      *redis.Client
}

func NewAwsQueue(ctx context.Context, client *sqs.Client, queueURL string, metrics metrics.Metrics, rdb *redis.Client) *OrderQueue {
	return &OrderQueue{
		client:   client,
		queueURL: queueURL,
		metrics:  metrics,
		log:      zerolog.Ctx(ctx),
		rdb:      rdb,
	}
}

func (o *OrderQueue) StartReceiving(ctx context.Context) error {

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			o.log.Info().Msg("receiving orders from queue")
			orders, err := o.GetOrders(ctx)
			if err != nil {
				o.log.Error().Err(err).Msg("failed to get orders from queue")
				continue
			}

			for _, order := range orders {
				valTime := o.rdb.Get(ctx, order.ActionID)
				if valTime.Err() != nil {
					o.log.Error().Err(valTime.Err()).Msg("failed to get order time from redis")
					continue
				}

				// get time from redis, example time in redis: 1715461032
				parsed, err := valTime.Int64()
				if err != nil {
					o.log.Error().Err(err).Msg("failed to parse order time from redis")
					continue
				}

				if parsed == 0 {
					continue
				}

				// convert time to time.Time
				parsedTimeTime := time.Unix(parsed, 0)

				// calculate time difference
				diff := order.Time.Sub(parsedTimeTime)

				// push time metric
				o.metrics.OrderTimeMetric.RecordOrderTime(ctx, diff)

				o.rdb.Del(ctx, order.ActionID)

				o.log.Info().Str("action_id", order.ActionID).Str("time", diff.String()).Msg("order processed")
			}
		}
	}
}

func (o *OrderQueue) PublishOrder(ctx context.Context, order model.OrderMsg) error {

	jsonParsed, err := order.ToJson()
	if err != nil {
		return err
	}

	_, err = o.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    &o.queueURL,
		MessageBody: aws.String(jsonParsed),
	})

	return err
}

func (o *OrderQueue) GetOrders(ctx context.Context) ([]model.OrderFromQueue, error) {
	orders, err := o.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:        &o.queueURL,
		WaitTimeSeconds: 20,
	})

	if err != nil {
		return nil, err
	}

	ordersList := make([]model.OrderFromQueue, len(orders.Messages))

	for i, order := range orders.Messages {
		orderModel := model.OrderMsg{}

		err := json.Unmarshal([]byte(*order.Body), &orderModel)
		if err != nil {
			continue
		}

		orderModelFromQueue := model.OrderFromQueue{
			ActionID: orderModel.ActionID,
			Time:     time.Now(),
		}

		ordersList[i] = orderModelFromQueue

		if _, err := o.client.DeleteMessage(
			ctx,
			&sqs.DeleteMessageInput{
				QueueUrl:      &o.queueURL,
				ReceiptHandle: order.ReceiptHandle,
			},
		); err != nil {
			log.Fatalln(err)
		}
	}

	return ordersList, nil
}
