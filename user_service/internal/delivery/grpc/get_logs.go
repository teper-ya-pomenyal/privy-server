package grpc

import (
	"context"

	userv1 "github.com/teper-ya-pomenyal/privy_stream/proto/user/v1"
)

const maxLogsLimit = 500

// GetLogs отдаёт последние записи кольцевого буфера журнала процесса
// (в gateway раздаётся через GET /admin/logs, закрытый ролью owner).
func (h *UserGRPCHandler) GetLogs(ctx context.Context, req *userv1.GetLogsRequest) (*userv1.GetLogsResponse, error) {
	limit := int(req.Limit)
	if limit <= 0 || limit > maxLogsLimit {
		limit = maxLogsLimit
	}
	entries := requestLogs.Latest(limit)
	out := make([]*userv1.LogEntry, 0, len(entries))
	for _, e := range entries {
		out = append(out, &userv1.LogEntry{
			AtMs: e.AtMs, Level: e.Level, Method: e.Method,
			Path: e.Path, Code: e.Code, Ms: e.MS, Message: e.Message,
		})
	}
	return &userv1.GetLogsResponse{Entries: out}, nil
}
