package entity

import "time"

type ExportJob struct {
	JobID        string
	ProjectID    string
	Status       string
	S3Key        *string
	RowCount     *int64
	ErrorMessage *string
	StartedAt    time.Time
	UpdatedAt    time.Time
	FinishedAt   *time.Time
}
