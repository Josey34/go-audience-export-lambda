package usecase

import (
	awsclients "audience-export-lambda/core/aws"
	"audience-export-lambda/core/vars"
	"audience-export-lambda/errors"
	"audience-export-lambda/modules/dto"
	"audience-export-lambda/modules/repository"
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type EnqueueExport struct {
	jobRepo      *repository.JobRepo
	lambdaClient *awsclients.LambdaClient
	funcName     string
}

func NewEnqueueExport(jobRepo *repository.JobRepo, lambdaClient *awsclients.LambdaClient, funcName string) *EnqueueExport {
	return &EnqueueExport{
		jobRepo:      jobRepo,
		lambdaClient: lambdaClient,
		funcName:     funcName,
	}
}

func (u *EnqueueExport) Execute(ctx context.Context, req dto.Request) (dto.Response, error) {
	if req.ProjectID == "" {
		return dto.Response{}, errors.ErrInvalidInput
	}

	jobId := uuid.New().String()
	if err := u.jobRepo.Create(ctx, jobId, req.ProjectID); err != nil {
		return dto.Response{}, err
	}

	event := dto.AsyncEvent{Type: "async_export", JobID: jobId, ProjectID: req.ProjectID}

	payload, err := json.Marshal(event)
	if err != nil {
		return dto.Response{}, err
	}

	if err := u.lambdaClient.InvokeAsync(ctx, u.funcName, payload); err != nil {
		return dto.Response{}, err
	}

	return dto.Response{JobID: jobId, Status: vars.StatusPending}, nil
}
