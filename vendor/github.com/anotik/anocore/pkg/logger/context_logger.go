package logger

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/anotik/anocore/pkg/consts"
	"github.com/anotik/anocore/pkg/response"
	"github.com/anotik/anocore/pkg/util"
)

// ContextLogger is a wrapper around slog.Logger that adds context values to log entries
type ContextLogger struct {
	logger   *slog.Logger
	ctx      context.Context
	LineInfo bool
}

// NewContextLogger creates a new ContextLogger with the given slog.Logger and context
func NewContextLogger(ctx context.Context) (*ContextLogger, error) {
	handlerOpts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == "msg" {
				a.Key = "message"
			}
			return a
		},
	}
	jSONLogger := slog.New(slog.NewJSONHandler(os.Stdout, handlerOpts))
	if jSONLogger == nil {
		// TODO: INIT_ERROR -> panic
		return nil, fmt.Errorf("SetServiceLoggerAsDefault: failed to create JSON logger")
	}
	slog.SetDefault(jSONLogger)
	return &ContextLogger{
		logger:   jSONLogger,
		ctx:      ctx,
		LineInfo: false,
	}, nil
}

// WithContext creates a new ContextLogger with the given context
func (l *ContextLogger) WithContext(ctx context.Context) (*ContextLogger, error) {
	newLogger, err := NewContextLogger(ctx)
	if err != nil {
		return nil, err
	}
	return newLogger, nil
}

// getContextAttrs returns slog.Attr slice from context values
func (l *ContextLogger) getContextAttrs() []slog.Attr {
	var attrs []slog.Attr
	if l.ctx == nil {
		return attrs
	}

	// Get all values from context
	keysValue := l.ctx.Value(ContextKeysKey{})
	if keysValue == nil {
		return attrs
	}

	keys, ok := keysValue.([]any)
	if !ok || keys == nil {
		return attrs
	}

	for _, key := range keys {
		if val := l.ctx.Value(key); val != nil {
			if keyStr, ok := key.(string); ok {
				attrs = append(attrs, slog.Any(keyStr, val))
			}
		}
	}
	return attrs
}

// convertToAttrs converts []any to []slog.Attr
// It handles various types including structs, maps, slices, and primitive types.
// The args should be in key-value pairs where keys are strings.
// Example usage:
//
//	logger.Info("message", "key1", value1, "key2", struct{...})
func convertToAttrs(args []any) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(args))
	for i := 0; i < len(args); i += 2 {
		if i+1 >= len(args) {
			break
		}
		key, ok := args[i].(string)
		if !ok {
			continue
		}
		value := args[i+1]
		attrs = append(attrs, slog.Any(key, value))
	}
	return attrs
}

// convertAttrsToAny converts []slog.Attr to []any
func convertAttrsToAny(attrs []slog.Attr) []any {
	result := make([]any, len(attrs))
	for i, attr := range attrs {
		result[i] = attr
	}
	return result
}

// getRuntimeInfo returns file, line number and function name information
func (l *ContextLogger) getRuntimeInfo() []any {
	var args []any
	pc, file, line, ok := runtime.Caller(2) // Using 2 to skip the logging method itself
	if ok {
		args = append(args, "file", file, "line", line)
		if fn := runtime.FuncForPC(pc); fn != nil {
			args = append(args, "func", fn.Name())
		}
	}
	return args
}

// Info logs a message at Info level with context values
func (l *ContextLogger) Info(msg string, args ...any) {
	if l.LineInfo {
		runtimeInfo := l.getRuntimeInfo()
		args = append(runtimeInfo, args...)
	}

	contextAttrs := l.getContextAttrs()
	argAttrs := convertToAttrs(args)
	allAttrs := append(contextAttrs, argAttrs...)
	l.logger.Info(msg, convertAttrsToAny(allAttrs)...)
}

// Error logs a message at Error level with context values
func (l *ContextLogger) Error(msg string, args ...any) {
	if l.LineInfo {
		runtimeInfo := l.getRuntimeInfo()
		args = append(runtimeInfo, args...)
	}

	contextAttrs := l.getContextAttrs()
	argAttrs := convertToAttrs(args)
	allAttrs := append(contextAttrs, argAttrs...)
	l.logger.Error(msg, convertAttrsToAny(allAttrs)...)
}

// Debug logs a message at Debug level with context values
func (l *ContextLogger) Debug(msg string, args ...any) {
	if l.LineInfo {
		runtimeInfo := l.getRuntimeInfo()
		args = append(runtimeInfo, args...)
	}

	contextAttrs := l.getContextAttrs()
	argAttrs := convertToAttrs(args)
	allAttrs := append(contextAttrs, argAttrs...)
	l.logger.Debug(msg, convertAttrsToAny(allAttrs)...)
}

// Warn logs a message at Warn level with context values
func (l *ContextLogger) Warn(msg string, args ...any) {
	if l.LineInfo {
		runtimeInfo := l.getRuntimeInfo()
		args = append(runtimeInfo, args...)
	}

	contextAttrs := l.getContextAttrs()
	argAttrs := convertToAttrs(args)
	allAttrs := append(contextAttrs, argAttrs...)
	l.logger.Warn(msg, convertAttrsToAny(allAttrs)...)
}

// WithContextValues creates a new context with values that will be included in logs
func WithContextValues(ctx context.Context, values map[string]any) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	// Get existing keys or create new slice
	keysValue := ctx.Value(ContextKeysKey{})
	if keysValue == nil {
		keysValue = make([]any, 0)
	}
	keys, ok := keysValue.([]any)
	if !ok || keys == nil {
		keys = make([]any, 0)
	}

	// Add new keys to the slice
	for k := range values {
		keys = append(keys, k)
	}

	// Create new context with keys and values
	ctx = context.WithValue(ctx, ContextKeysKey{}, keys)
	for k, v := range values {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx
}

func ContextWithTracHeaders(r *http.Request) context.Context {

	m := make(map[string]any, len(consts.TRACE_HEADERS_MAP))
	for reqHeader, targetKey := range consts.TRACE_HEADERS_MAP {
		if v := r.Header.Get(reqHeader); v != "" {
			m[targetKey] = v
		}
	}
	ctx := WithContextValues(context.Background(), m)

	return ctx
}

func WithTraceHeaders(r *http.Request) (*ContextLogger, error) {
	ctx := ContextWithTracHeaders(r)
	newLogger, err := NewContextLogger(ctx)
	if err != nil {
		return nil, err
	}
	return newLogger, nil
}

func filterHeaders(reqHeader http.Header, unwantedKeys []string) map[string]string {
	headers := make(map[string]string)
	for k := range reqHeader {
		if slices.Contains(unwantedKeys, strings.ToLower(k)) {
			continue
		}
		headers[k] = reqHeader.Get(k)
	}
	return headers
}

// ContextKeysKey is used to store the list of context keys in the context
type ContextKeysKey struct{}

// LogRequest logs details of an HTTP request including method, URL, headers and trace headers.
// The request is cloned to avoid modifying the original request.
func (l *ContextLogger) LogRequest(r *http.Request) {
	req := r.Clone(context.Background())

	// Create headers map including trace headers
	headers := filterHeaders(req.Header, []string{"x-api-key"})

	// Log the request with all headers in the headers map and trace headers as separate fields
	l.Info(
		"request_received",
		"method", req.Method,
		"url", util.GetFullURL(req),
		"user_agent", req.Header.Get("User-Agent"),
		"headers", headers,
	)
}

// LogResponse logs details of an HTTP response including status code, duration, headers and trace headers.
func (l *ContextLogger) LogResponse(w http.ResponseWriter, resp response.APIResponse, req *http.Request, startTime time.Time) {
	duration := time.Since(startTime).Seconds() * 1000

	statusCode := http.StatusOK
	if resp.Err != nil {
		statusCode = resp.Err.StatusCode
	}

	headers := make(map[string]string)
	for k, v := range w.Header() {
		headers[k] = v[0]
	}

	if statusCode == 500 {
		l.Error(
			"response_sent",
			"status_code", statusCode,
			"duration_ms", duration,
			"method", req.Method,
			"url", util.GetFullURL(req),
			"user_agent", req.Header.Get("User-Agent"),
			"headers", headers,
		)
	} else {
		l.Info(
			"response_sent",
			"status_code", statusCode,
			"duration_ms", duration,
			"method", req.Method,
			"url", util.GetFullURL(req),
			"user_agent", req.Header.Get("User-Agent"),
			"headers", headers,
		)
	}

}

func (l *ContextLogger) LogFailedResponse(w http.ResponseWriter, resp response.APIResponse, req *http.Request, startTime time.Time, bytesWritten int, err error) {
	duration := time.Since(startTime).Seconds() * 1000

	statusCode := 0

	headers := make(map[string]string)
	for k, v := range w.Header() {
		headers[k] = v[0]
	}

	l.Error(
		"response_sent",
		"error", err,
		"status_code", statusCode,
		"duration_ms", duration,
		"method", req.Method,
		"url", util.GetFullURL(req),
		"user_agent", req.Header.Get("User-Agent"),
		"headers", headers,
	)
}
