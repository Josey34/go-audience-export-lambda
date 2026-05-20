package dto

type AsyncEvent struct {
	Type      string `json:"type"`
	JobID     string `json:"job_id"`
	ProjectID string `json:"project_id"`
}
