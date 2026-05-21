package awsclients

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type LambdaClient struct {
	client *lambda.Client
}

func NewLambdaClient(cfg aws.Config) *LambdaClient {
	return &LambdaClient{
		client: lambda.NewFromConfig(cfg),
	}
}

func (c *LambdaClient) InvokeAsync(ctx context.Context, functionName string, payload []byte) error {
	_, err := c.client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: types.InvocationTypeEvent,
		Payload:        payload,
	})

	return err
}
