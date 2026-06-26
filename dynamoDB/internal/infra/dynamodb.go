package infra

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	appConfig "github.com/ramon/goals-tasks-api/internal/config"
)

type AWSClient struct {
	DynamoDB *dynamodb.Client
}

func NewAWSClient(ctx context.Context, cfg *appConfig.Config) (*AWSClient, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(cfg.AWSRegion),
	}

	if cfg.Env == "local" || cfg.AWSEndpoint != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider("mock_access_key", "mock_secret_key", "mock_session"),
		))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	var dbClient *dynamodb.Client
	if cfg.AWSEndpoint != "" {
		dbClient = dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
			o.BaseEndpoint = aws.String(cfg.AWSEndpoint)
		})
		slog.InfoContext(ctx, "DynamoDB client configured for local environment", "endpoint", cfg.AWSEndpoint)
	} else {
		dbClient = dynamodb.NewFromConfig(awsCfg)
		slog.InfoContext(ctx, "DynamoDB client configured for cloud environment")
	}

	return &AWSClient{
		DynamoDB: dbClient,
	}, nil
}
