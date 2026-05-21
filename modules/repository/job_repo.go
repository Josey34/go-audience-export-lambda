package repository

import (
	"context"
	"database/sql"
)

type JobRepo struct {
	db *sql.DB
}

func NewJobRepo(db *sql.DB) *JobRepo {
	return &JobRepo{
		db: db,
	}
}

func (r *JobRepo) Create(ctx context.Context, jobID, projectID string) error {
	query := `INSERT INTO export_jobs (job_id, project_id, status) VALUES ($1, $2, 'PENDING')`

	_, err := r.db.ExecContext(ctx, query, jobID, projectID)
	if err != nil {
		return err
	}

	return nil
}

func (r *JobRepo) MarkRunning(ctx context.Context, jobID string) error {
	query := `UPDATE export_jobs SET status='RUNNING', updated_at=NOW() WHERE job_id=$1`

	_, err := r.db.ExecContext(ctx, query, jobID)
	if err != nil {
		return err
	}

	return nil
}

func (r *JobRepo) MarkSuccess(ctx context.Context, jobID, s3Key string, rowCount int) error {
	query := `UPDATE export_jobs SET status='SUCCESS', s3_key=$2, row_count=$3, updated_at=NOW(), finished_at=NOW() WHERE job_id=$1`

	_, err := r.db.ExecContext(ctx, query, jobID, s3Key, rowCount)
	if err != nil {
		return err
	}

	return nil
}

func (r *JobRepo) MarkFailed(ctx context.Context, jobID, errMsg string) error {
	query := `UPDATE export_jobs SET status='FAILED', error_message=$2, updated_at=NOW(), finished_at=NOW() WHERE job_id=$1`

	_, err := r.db.ExecContext(ctx, query, jobID, errMsg)
	if err != nil {
		return err
	}

	return nil
}
