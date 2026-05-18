package appcontext

import "context"

type contextKey string

const jobIDKey contextKey = "job_id"

func WithJobID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, jobIDKey, id)
}

func JobIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(jobIDKey).(string)
	return id
}
