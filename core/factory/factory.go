package factory

import (
	"audience-export-lambda/config"
	awsclients "audience-export-lambda/core/aws"
	"audience-export-lambda/core/logger"
	"audience-export-lambda/modules/repository"
	"context"
	"database/sql"
	"log/slog"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
)

type Factory struct {
	Config       *config.Config
	Logger       *slog.Logger
	DB           *sql.DB
	S3Client     *awsclients.S3Client
	LambdaClient *awsclients.LambdaClient
	AudienceRepo *repository.AudienceRepo
	JobRepo      *repository.JobRepo
}

func New(ctx context.Context) (*Factory, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New(cfg.LogLevel)

	db, err := sql.Open("pgx", cfg.DBDsn)
	if err != nil {
		return nil, err
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithBaseEndpoint(cfg.AWSEndpointURL),
	)
	if err != nil {
		return nil, err
	}

	s3Client := awsclients.NewS3Client(awsCfg)
	lambdaClient := awsclients.NewLambdaClient(awsCfg)

	audienceRepo := repository.NewAudienceRepo(db)
	jobRepo := repository.NewJobRepo(db)

	return &Factory{
		Config:       cfg,
		Logger:       log,
		DB:           db,
		S3Client:     s3Client,
		LambdaClient: lambdaClient,
		AudienceRepo: audienceRepo,
		JobRepo:      jobRepo,
	}, nil
}
