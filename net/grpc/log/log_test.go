package log_test

import (
	"bytes"
	"context"
	"log/slog"
	"net"
	"testing"

	"go.adoublef.dev/is"
	. "go.adoublef.dev/sdk/net/grpc/log"
	v1 "go.adoublef.dev/sdk/net/grpc/proto/acme/grpctest/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test_LogInterceptor(t *testing.T) {
	t.Skipf("test requires creating a good regex matcher")

	t.Run("OK", func(t *testing.T) {
		var (
			tc, _, ctx = newClient(t)
			is         = is.NewRelaxed(t)
		)

		err := tc.Ping(ctx)
		is.NoErr(err)
	})
}

func newClient(tb testing.TB) (*testClient, *bytes.Buffer, context.Context) {
	tb.Helper()
	var (
		is  = is.New(tb)
		ctx = context.Background()

		buf bytes.Buffer
	)

	ln, err := net.Listen("tcp", "localhost:0") // wrap in a must
	is.NoErr(err)                               // start a real listener

	// start connection
	insecure := grpc.WithTransportCredentials(insecure.NewCredentials())
	conn, err := grpc.DialContext(ctx, ln.Addr().String(), insecure)
	is.NoErr(err)

	tb.Cleanup(func() { conn.Close() })

	sl := slog.New(slog.NewTextHandler(&buf, nil))

	gs := grpc.NewServer(
		grpc.ChainUnaryInterceptor(LogInterceptor(sl)),
		grpc.ChainStreamInterceptor(LogStreamInterceptor(sl)),
	)

	v1.RegisterTestServiceServer(gs, &testServer{})
	go func() { is.NoErr(gs.Serve(ln)) }()

	tc := v1.NewTestServiceClient(conn)

	return &testClient{tc}, &buf, ctx
}

type testServer struct {
	v1.TestServiceServer
}

func (*testServer) Ping(ctx context.Context, r *v1.PingRequest) (*v1.PingResponse, error) {
	return &v1.PingResponse{}, nil
}

type testClient struct {
	c v1.TestServiceClient
}

func (c *testClient) Ping(ctx context.Context) error {
	_, err := c.c.Ping(ctx, &v1.PingRequest{})
	if err != nil {
		return err // status.Error
	}
	return nil
}
