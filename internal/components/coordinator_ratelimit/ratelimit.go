package coordinator_ratelimit

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/warpstreamlabs/bento/public/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func init() {
	err := service.RegisterRateLimit(
		"coordinator", Config(),
		func(conf *service.ParsedConfig, mgr *service.Resources) (service.RateLimit, error) {
			return NewFromConfig(conf, mgr)
		})
	if err != nil {
		panic(err)
	}
}

type RateLimit struct {
	client pb.CoordinatorClient
	conn   *grpc.ClientConn
	logger *service.Logger
	label  string
}

const rateLimitKeyContextKey = "rate_limit_key"

func NewFromConfig(conf *service.ParsedConfig, mgr *service.Resources) (*RateLimit, error) {
	label := mgr.Label()

	coordinatorAddr := os.Getenv("DISCOVERY_URI")
	if coordinatorAddr == "" {
		coordinatorAddr = "localhost:50000"
	}

	conn, err := grpc.NewClient(coordinatorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to coordinator: %w", err)
	}

	client := pb.NewCoordinatorClient(conn)

	return &RateLimit{
		client: client,
		conn:   conn,
		logger: mgr.Logger(),
		label:  label,
	}, nil
}

func (r *RateLimit) Access(ctx context.Context) (time.Duration, error) {
	return r.AccessWithCost(ctx, 1)
}

func (r *RateLimit) AccessWithCost(ctx context.Context, cost int64) (time.Duration, error) {
	keyStr := "global"
	if ctxKey, ok := ctx.Value(rateLimitKeyContextKey).(string); ok && ctxKey != "" {
		keyStr = ctxKey
	}

	req := &pb.RateLimitCheckRequest{
		Label: r.label,
		Key:   keyStr,
		Cost:  cost,
	}

	resp, err := r.client.CheckRateLimit(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if !resp.Allowed {
		waitDuration := time.Duration(resp.RetryAfterMs) * time.Millisecond
		return waitDuration, nil
	}

	return 0, nil
}

func (r *RateLimit) Close(ctx context.Context) error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
