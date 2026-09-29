package grpc

import (
	"context"

	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
)

const maxLogsLimit = 500

// GetLogs отдаёт последние записи кольцевого буфера журнала процесса
// (в gateway раздаётся через GET /admin/logs, закрытый ролью owner).
func (h *CatalogGRPCHandler) GetLogs(ctx context.Context, req *catalogv1.GetLogsRequest) (*catalogv1.GetLogsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 || limit > maxLogsLimit {
		limit = maxLogsLimit
	}
	entries := requestLogs.Latest(limit)
	out := make([]*catalogv1.LogEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, &catalogv1.LogEntry{
			AtMs: e.AtMs, Level: e.Level, Method: e.Method,
			Path: e.Path, Code: e.Code, Ms: e.MS, Message: e.Message,
		})
	}
	return &catalogv1.GetLogsResponse{Entries: out}, nil
}
