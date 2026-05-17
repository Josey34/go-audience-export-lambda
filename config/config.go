package config

import (
	"fmt"
	"os"
)

type Config struct {
	DBDsn              string
	AWSRegion          string
	AWSEndpointURL     string
	S3Bucket           string
	LambdaFunctionName string
	LogLevel           string
}

func Load() (*Config, error) {
	dbDsn := os.Getenv("DB_DSN")
	if dbDsn == "" {
		return nil, fmt.Errorf("Required env var DB_DSN is not set")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		return nil, fmt.Errorf("Required env var AWS_REGION is not set")
	}

	awsEndpointUrl := os.Getenv("AWS_ENDPOINT_URL")
	if awsEndpointUrl == "" {
		return nil, fmt.Errorf("Required env var AWS_ENDPOINT_URL is not set")
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		return nil, fmt.Errorf("Required env var S3_BUCKET is not set")
	}

	lambdaFunctionName := os.Getenv("LAMBDA_FUNCTION_NAME")
	if lambdaFunctionName == "" {
		return nil, fmt.Errorf("Required env var LAMBDA_FUNCTION_NAME is not set")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	return &Config{
		DBDsn:              dbDsn,
		AWSRegion:          awsRegion,
		AWSEndpointURL:     awsEndpointUrl,
		S3Bucket:           s3Bucket,
		LambdaFunctionName: lambdaFunctionName,
		LogLevel:           logLevel,
	}, nil
}
