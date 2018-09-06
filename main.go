package main

import (
	"context"
	"errors"

	"go.uber.org/zap"
)

const (
	zapKey = "zap"
)

func fromCtx(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(zapKey).(*zap.Logger)
	if ok {
		return logger
	}
	return zap.NewNop()
}

func withName(ctx context.Context, name string) context.Context {
	return withZap(ctx, fromCtx(ctx).Named(name))
}

func withZap(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, zapKey, logger)
}

func with(ctx context.Context, fields ...zap.Field) context.Context {
	return withZap(ctx, fromCtx(ctx).With(fields...))
}

func auth(ctx context.Context, token string) (context.Context, error) {
	withNamed := withName(ctx, "auth")
	if token == "valid" {
		fromCtx(withNamed).Debug("token valid")
		return with(ctx, zap.Int64("user_id", 124)), nil
	} else {
		fromCtx(withNamed).Warn("token invalid")
		return with(ctx, zap.String("auth", "anonymous")), errors.New("not authorised")
	}
}

func process(ctx context.Context, requestID, token string) {
	ctx = with(withName(ctx, "process"),
		zap.String("request_id", requestID),
	)
	fromCtx(ctx).Info("start")
	authCtx, authErr := auth(ctx, token)
	if authErr != nil {
		fromCtx(ctx).Warn("auth failed", zap.Error(authErr))
		return
	}
	processBooking(authCtx, 9002)
}

func processBooking(ctx context.Context, bookingID int64) {
	ctx = with(withName(ctx, "booking"), zap.Int64("booking_id", bookingID))
	fromCtx(ctx).Info("processed")
}

/*

Output:

{"level":"info","msg":"logging enabled"}
{"level":"info","logger":"process","msg":"start","request_id":"00001885154"}
{"level":"info","ts":1536224417.8312197,"caller":"tracing-logs/main.go:103","msg":"production"}
{"level":"debug","logger":"process.auth","msg":"token valid","request_id":"00001885154"}
{"level":"info","ts":1536224417.8312333,"logger":"process","caller":"tracing-logs/main.go:49","msg":"start","request_id":"00001885154"}
{"level":"info","logger":"process.booking","msg":"processed","request_id":"00001885154","user_id":124,"booking_id":9002}
{"level":"info","logger":"process","msg":"start","request_id":"00009872658"}
{"level":"warn","logger":"process.auth","msg":"token invalid","request_id":"00009872658"}
{"level":"warn","logger":"process","msg":"auth failed","request_id":"00009872658","error":"not authorised"}
{"level":"info","msg":"<logging disabled>"}
{"level":"info","msg":"</logging disabled>"}
{"level":"info","ts":1536224417.8312402,"logger":"process.booking","caller":"tracing-logs/main.go:60","msg":"processed","request_id":"00001885154","user_id":124,"booking_id":9002}
{"level":"info","ts":1536224417.8312478,"logger":"process","caller":"tracing-logs/main.go:49","msg":"start","request_id":"00009872658"}
{"level":"warn","ts":1536224417.831251,"logger":"process.auth","caller":"tracing-logs/main.go:40","msg":"token invalid","request_id":"00009872658"}
{"level":"warn","ts":1536224417.8312564,"logger":"process","caller":"tracing-logs/main.go:52","msg":"auth failed","request_id":"00009872658","error":"not authorised"}
*/
func main() {
	l := zap.NewExample()
	ctx := withZap(context.Background(), l)
	l.Info("logging enabled")
	process(ctx, "00001885154", "valid")
	process(ctx, "00009872658", "invalid")

	l.Info("<logging disabled>")
	ctx = context.Background()
	process(ctx, "00001885154", "valid")
	process(ctx, "00009872658", "invalid")
	l.Info("</logging disabled>")
}
