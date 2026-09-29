package grpc

import (
	"sync"
	"time"
)

// LogEntry — запись внутреннего журнала для админки (GET /admin/logs
// забирает её по gRPC). Хранится в кольцевом буфере процесса; в stdout
// дублируется как раньше.
type LogEntry struct {
	AtMs    int64
	Level   string // INFO | WARN | ERROR
	Method  string // HTTP-метод или пусто
	Path    string // HTTP-путь или gRPC full method
	Code    string // HTTP-статус или gRPC-код
	MS      int64
	Message string
}

// LogRing — потокобезопасный кольцевой буфер последних записей журнала.
type LogRing struct {
	mu      sync.Mutex
	entries []LogEntry // новые в конце
	cap     int
}

func NewLogRing(capacity int) *LogRing {
	return &LogRing{entries: make([]LogEntry, 0, capacity), cap: capacity}
}

func (r *LogRing) Add(e LogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, e)
	if len(r.entries) > r.cap {
		r.entries = r.entries[len(r.entries)-r.cap:]
	}
}

// Latest возвращает последние n записей, новые первыми.
func (r *LogRing) Latest(n int) []LogEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n > len(r.entries) {
		n = len(r.entries)
	}
	out := make([]LogEntry, n)
	for i, e := range r.entries[len(r.entries)-n:] {
		out[n-1-i] = e
	}
	return out
}

func levelForGRPCCode(code string) string {
	switch code {
	case "OK", "Canceled":
		return "INFO"
	case "InvalidArgument", "NotFound", "AlreadyExists", "PermissionDenied", "Unauthenticated", "FailedPrecondition", "OutOfRange":
		return "WARN"
	default:
		return "ERROR"
	}
}

func nowMs() int64 { return time.Now().UnixMilli() }
