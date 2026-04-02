package coordinator

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func (c *CoordinatorAPI) CreateRateLimit(_ context.Context, in *pb.RateLimit) (*pb.RateLimitResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	existing, err := c.rateLimitRepo.FindByLabel(in.GetLabel())
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing rate limit")
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existing != nil {
		return nil, status.Error(codes.AlreadyExists, "Rate limit with this label already exists")
	}

	rateLimit := &persistence.RateLimit{}
	rateLimit.FromProto(in)
	rateLimit.ParentID = nil
	rateLimit.IsCurrent = true

	if err := c.rateLimitRepo.Create(rateLimit); err != nil {
		log.Error().Err(err).Msg("Failed to create rate limit")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RateLimitResponse{
		Data: rateLimit.ToProto(),
		Meta: &pb.CommonResponse{Message: "Rate limit has been created successfully"},
	}, nil
}

func (c *CoordinatorAPI) GetRateLimit(_ context.Context, in *pb.GetRateLimitRequest) (*pb.RateLimitResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rateLimit, err := c.rateLimitRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find rate limit")
		return nil, status.Error(codes.Internal, err.Error())
	} else if rateLimit == nil {
		return nil, status.Error(codes.NotFound, "Rate limit not found")
	}

	return &pb.RateLimitResponse{
		Data: rateLimit.ToProto(),
		Meta: &pb.CommonResponse{Message: "OK"},
	}, nil
}

func (c *CoordinatorAPI) ListRateLimits(_ context.Context, _ *emptypb.Empty) (*pb.ListRateLimitsResponse, error) {
	rateLimits, err := c.rateLimitRepo.ListAll()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list rate limits")
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := &pb.ListRateLimitsResponse{
		Data: make([]*pb.RateLimit, len(rateLimits)),
	}
	for i, rateLimit := range rateLimits {
		result.Data[i] = rateLimit.ToProto()
	}

	return result, nil
}

func (c *CoordinatorAPI) UpdateRateLimit(_ context.Context, in *pb.RateLimit) (*pb.RateLimitResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if in.GetId() == 0 {
		log.Debug().Msg("Invalid request: ID is required")
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	rateLimit, err := c.rateLimitRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find rate limit")
		return nil, status.Error(codes.Internal, err.Error())
	} else if rateLimit == nil {
		return nil, status.Error(codes.NotFound, "Rate limit not found")
	}

	existingByLabel, err := c.rateLimitRepo.FindByLabel(in.GetLabel())
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing rate limit by label")
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existingByLabel != nil && existingByLabel.ID != in.GetId() {
		sameLineage := false

		if existingByLabel.ParentID != nil && rateLimit.ParentID != nil {
			sameLineage = *existingByLabel.ParentID == *rateLimit.ParentID
		}

		if !sameLineage && existingByLabel.ParentID != nil {
			sameLineage = *existingByLabel.ParentID == rateLimit.ID
		}

		if !sameLineage && rateLimit.ParentID != nil {
			sameLineage = *rateLimit.ParentID == existingByLabel.ID
		}

		if sameLineage {
			return nil, status.Error(codes.FailedPrecondition, "Cannot update old version. You can only update the current version of this rate limit")
		}
		return nil, status.Error(codes.AlreadyExists, "Another rate limit with this label already exists")
	}

	newRateLimit := &persistence.RateLimit{}
	newRateLimit.FromProto(in)
	if rateLimit.ParentID == nil {
		newRateLimit.ParentID = &rateLimit.ID
	} else {
		newRateLimit.ParentID = rateLimit.ParentID
	}
	newRateLimit.IsCurrent = true

	if err = c.rateLimitRepo.Update(newRateLimit); err != nil {
		log.Error().Err(err).Msg("Failed to update rate limit")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RateLimitResponse{
		Data: newRateLimit.ToProto(),
		Meta: &pb.CommonResponse{Message: "Rate limit has been updated successfully"},
	}, nil
}

func (c *CoordinatorAPI) DeleteRateLimit(_ context.Context, in *pb.GetRateLimitRequest) (*pb.CommonResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rateLimit, err := c.rateLimitRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find rate limit")
		return nil, status.Error(codes.Internal, err.Error())
	} else if rateLimit == nil {
		return nil, status.Error(codes.NotFound, "Rate limit not found")
	}

	if err := c.rateLimitRepo.Delete(in.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to delete rate limit")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonResponse{
		Message: "Rate limit has been deleted successfully",
	}, nil
}

func (c *CoordinatorAPI) CheckRateLimit(_ context.Context, in *pb.RateLimitCheckRequest) (*pb.RateLimitCheckResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := c.rateLimiterEngine.Check(in.GetLabel(), in.GetKey(), in.GetCost())
	if err != nil {
		log.Error().Err(err).Str("label", in.GetLabel()).Str("key", in.GetKey()).Msg("Failed to check rate limit")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RateLimitCheckResponse{
		Allowed:      result.Allowed,
		RetryAfterMs: result.RetryAfterMs,
		Remaining:    result.Remaining,
		Limit:        result.Limit,
		ResetAt:      result.ResetAt,
	}, nil
}
