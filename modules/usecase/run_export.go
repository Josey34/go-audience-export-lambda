package usecase

import (
	awsclients "audience-export-lambda/core/aws"
	"audience-export-lambda/modules/dto"
	"audience-export-lambda/modules/repository"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
)

type RunExport struct {
	jobRepo      *repository.JobRepo
	audienceRepo *repository.AudienceRepo
	s3Client     *awsclients.S3Client
	bucket       string
}

func NewRunExport(jobRepo *repository.JobRepo, audienceRepo *repository.AudienceRepo, s3Client *awsclients.S3Client, bucket string) *RunExport {
	return &RunExport{
		jobRepo:      jobRepo,
		audienceRepo: audienceRepo,
		s3Client:     s3Client,
		bucket:       bucket,
	}
}

func (u *RunExport) Execute(ctx context.Context, event dto.AsyncEvent) (err error) {
	if err := u.jobRepo.MarkRunning(ctx, event.JobID); err != nil {
		return err
	}

	defer func() {
		if err != nil {
			u.jobRepo.MarkFailed(ctx, event.JobID, err.Error())
		}
	}()

	rows, err := u.audienceRepo.FindByProjectID(ctx, event.ProjectID)
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	w := csv.NewWriter(&buf)
	w.Write([]string{"project_id", "retailer_code", "retailer_name", "product_flagging", "product_1_name", "product_1_price"})
	for _, row := range rows {
		w.Write([]string{
			row.ProjectID, row.RetailerCode, row.RetailerName,
			row.ProductFlagging, row.ProductName,
			strconv.FormatFloat(row.ProductPrice, 'f', 2, 64),
		})

	}
	w.Flush()

	key := fmt.Sprintf("exports/project-%s/audience.csv", event.ProjectID)

	if err = u.s3Client.Upload(ctx, u.bucket, key, &buf); err != nil {
		return err
	}

	if err = u.jobRepo.MarkSuccess(ctx, event.JobID, key, len(rows)); err != nil {
		return err
	}

	return nil
}
