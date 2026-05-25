package main

import (
	"audience-export-lambda/core/factory"
	"audience-export-lambda/handler"
	"audience-export-lambda/modules/usecase"
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	ctx := context.Background()
	f, err := factory.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	enqueue := usecase.NewEnqueueExport(f.JobRepo, f.LambdaClient, f.Config.LambdaFunctionName)
	run := usecase.NewRunExport(f.JobRepo, f.AudienceRepo, f.S3Client, f.Config.S3Bucket)

	h := handler.New(enqueue, run)
	lambda.Start(h.Handle)
}
