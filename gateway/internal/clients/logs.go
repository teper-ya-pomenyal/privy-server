package clients

import (
	"context"

	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	userv1 "github.com/teper-ya-pomenyal/privy_stream/proto/user/v1"
)

// LogEntry — запись внутреннего журнала сервиса (кольцевой буфер в памяти).
// Единый формат для gateway, user_service и catalog_service.
type LogEntry struct {
	At      int64  `json:"at"` // unix ms
	Level   string `json:"level"`
	Service string `json:"service"`
	Method  string `json:"method,omitempty"`
	Path    string `json:"path,omitempty"`
	Code    string `json:"code,omitempty"`
	MS      int64  `json:"ms,omitempty"`
	Message string `json:"message,omitempty"`
}

// GetUserServiceLogs возвращает последние записи журнала user_service.
func (u *UserClient) GetLogs(ctx context.Context, limit int32) ([]LogEntry, error) {
	res, err := u.grpcClient.GetLogs(ctx, &userv1.GetLogsRequest{Limit: limit})
	if err != nil {
		return nil, err
	}
	entries := make([]LogEntry, 0, len(res.Entries))
	for _, e := range res.Entries {
		entries = append(entries, LogEntry{
			At: e.AtMs, Level: e.Level, Service: "user_service",
			Method: e.Method, Path: e.Path, Code: e.Code, MS: e.Ms, Message: e.Message,
		})
	}
	return entries, nil
}

// GetCatalogLogs возвращает последние записи журнала catalog_service.
func (c *CatalogClient) GetLogs(ctx context.Context, limit int32) ([]LogEntry, error) {
	res, err := c.grpcClient.GetLogs(ctx, &catalogv1.GetLogsRequest{Limit: limit})
	if err != nil {
		return nil, err
	}
	entries := make([]LogEntry, 0, len(res.Entries))
	for _, e := range res.Entries {
		entries = append(entries, LogEntry{
			At: e.AtMs, Level: e.Level, Service: "catalog_service",
			Method: e.Method, Path: e.Path, Code: e.Code, MS: e.Ms, Message: e.Message,
		})
	}
	return entries, nil
}
