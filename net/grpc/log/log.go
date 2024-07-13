package log

import (
	"context"
	"log"
	"log/slog"
	"os"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var debug = debugT(false)

type debugT bool

func (debugT) Printf(format string, v ...any) {
	if debug {
		log.Default().Printf(format, v...)
	}
}

type contextKey struct{ string }

func (k *contextKey) String() string { return "grpc: context value " + k.string }

var (
	LoggerContextKey = &contextKey{"grpc-log"}
)

type logger struct {
	l *slog.Logger
}

func newLogger[I grpc.UnaryServerInfo | grpc.StreamServerInfo](ctx context.Context, l *slog.Logger, info *I) *logger {
	// https://appliedgo.com/blog/a-tip-and-a-trick-when-working-with-generics
	var method string
	switch info := any(info).(type) {
	case *grpc.UnaryServerInfo:
		method = info.FullMethod
	case *grpc.StreamServerInfo:
		method = info.FullMethod
	}
	service, method := path.Split(method)
	l = l.With(
		slog.String("service", path.Clean(service)),
		slog.String("method", "/"+method),
	)
	ip := realIP(ctx)
	if ip != "" {
		l = l.With(slog.String("realIP", ip))
	}
	return &logger{l: l}
}

func (l *logger) Write(ctx context.Context, status *status.Status, elapsed time.Duration) {
	code := status.Code()

	l.l = l.l.With(
		// likely do not need to print the code here
		slog.Int("status", int(code)),
		// slog.Int("bytes", bytes),
		slog.Duration("elapsed", elapsed),
	)
	if code != codes.OK {
		// only read up to 512 characters
		l.l = l.l.With(slog.String("body", status.Message()))
	}
	// https://grpc.io/docs/guides/error/
	//
	// https://www.reddit.com/r/golang/comments/13k3ne7/get_first_n_characters_of_a_string/
	l.l.Log(ctx, logLevel(code), code.String())
}

func LogInterceptor(l *slog.Logger) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// https://jbrandhorst.com/post/grpc-errors/
	// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/main/interceptors/server.go
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (rs any, err error) {
		ll := newLogger(ctx, l, info)
		t1 := time.Now()
		ctx = context.WithValue(ctx, LoggerContextKey, ll)
		rs, err = handler(ctx, req)
		ll.Write(ctx, status.Convert(err), time.Since(t1))
		return rs, err
	}
}

// StreamServerInterceptor is a gRPC server-side interceptor that provides reporting for Streaming RPCs.
func LogStreamInterceptor(l *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ll := newLogger(ss.Context(), l, info)
		t1 := time.Now()
		// pass the logger into the context
		ctx := context.WithValue(ss.Context(), LoggerContextKey, ll)
		err := handler(srv, &serverStream{ss, ctx})
		ll.Write(ss.Context(), status.Convert(err), time.Since(t1))
		return err
	}
}

// Logger returns the in-context Logger for a request.
func Logger(ctx context.Context) *slog.Logger {
	entry, ok := ctx.Value(LoggerContextKey).(*logger)
	if !ok || entry == nil {
		opts := &slog.HandlerOptions{
			AddSource: true,
			// LevelError+1 will be higher than all levels
			// hence logs would be skipped
			Level: slog.LevelError + 1,
		}
		return slog.New(slog.NewTextHandler(os.Stderr, opts))
	} else {
		return entry.l
	}
}

// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/main/interceptors/logging/options.go#L104
func logLevel(code codes.Code) slog.Level {
	switch code {
	case codes.OK, codes.NotFound, codes.Canceled, codes.AlreadyExists, codes.InvalidArgument, codes.Unauthenticated:
		return slog.LevelInfo
	case codes.DeadlineExceeded, codes.PermissionDenied, codes.ResourceExhausted, codes.FailedPrecondition, codes.Aborted, codes.OutOfRange, codes.Unavailable:
		return slog.LevelWarn
	case codes.Unknown, codes.Unimplemented, codes.Internal, codes.DataLoss:
		return slog.LevelError
	default:
		return slog.LevelError
	}
}

func ErrAttr(err error) slog.Attr {
	return slog.Any("err", err)
}

var _ grpc.ServerStream = (*serverStream)(nil)

type serverStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context implements grpc.ServerStream.
func (s *serverStream) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return s.ServerStream.Context()
}
