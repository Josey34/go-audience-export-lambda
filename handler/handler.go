package handler

import (
	"audience-export-lambda/modules/dto"
	"audience-export-lambda/modules/usecase"
	"context"
	"encoding/json"
)

type Handler struct {
	enqueue *usecase.EnqueueExport
	run     *usecase.RunExport
}

func New(enqueue *usecase.EnqueueExport, run *usecase.RunExport) *Handler {
	return &Handler{enqueue: enqueue, run: run}
}

func (h *Handler) Handle(ctx context.Context, raw json.RawMessage) (interface{}, error) {
	var event dto.AsyncEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return nil, err
	}

	if event.Type == "async_export" {
		return nil, h.run.Execute(ctx, event)
	}

	var req dto.Request
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, err
	}

	return h.enqueue.Execute(ctx, req)
}
