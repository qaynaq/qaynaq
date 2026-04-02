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

func (c *CoordinatorAPI) CreateCache(_ context.Context, in *pb.Cache) (*pb.CacheResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	existing, err := c.cacheRepo.FindByLabel(in.GetLabel())
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing cache")
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existing != nil {
		return nil, status.Error(codes.AlreadyExists, "Cache with this label already exists")
	}

	cache := &persistence.Cache{}
	cache.FromProto(in)
	cache.ParentID = nil
	cache.IsCurrent = true

	if err := c.cacheRepo.Create(cache); err != nil {
		log.Error().Err(err).Msg("Failed to create cache")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CacheResponse{
		Data: cache.ToProto(),
		Meta: &pb.CommonResponse{Message: "Cache has been created successfully"},
	}, nil
}

func (c *CoordinatorAPI) GetCache(_ context.Context, in *pb.GetCacheRequest) (*pb.CacheResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	cache, err := c.cacheRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find cache")
		return nil, status.Error(codes.Internal, err.Error())
	} else if cache == nil {
		return nil, status.Error(codes.NotFound, "Cache not found")
	}

	return &pb.CacheResponse{
		Data: cache.ToProto(),
		Meta: &pb.CommonResponse{Message: "OK"},
	}, nil
}

func (c *CoordinatorAPI) ListCaches(_ context.Context, _ *emptypb.Empty) (*pb.ListCachesResponse, error) {
	caches, err := c.cacheRepo.ListAll()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list caches")
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := &pb.ListCachesResponse{
		Data: make([]*pb.Cache, len(caches)),
	}
	for i, cache := range caches {
		result.Data[i] = cache.ToProto()
	}

	return result, nil
}

func (c *CoordinatorAPI) UpdateCache(_ context.Context, in *pb.Cache) (*pb.CacheResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if in.GetId() == 0 {
		log.Debug().Msg("Invalid request: ID is required")
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	cache, err := c.cacheRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find cache")
		return nil, status.Error(codes.Internal, err.Error())
	} else if cache == nil {
		return nil, status.Error(codes.NotFound, "Cache not found")
	}

	existingByLabel, err := c.cacheRepo.FindByLabel(in.GetLabel())
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing cache by label")
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existingByLabel != nil && existingByLabel.ID != in.GetId() {
		sameLineage := false

		if existingByLabel.ParentID != nil && cache.ParentID != nil {
			sameLineage = *existingByLabel.ParentID == *cache.ParentID
		}

		if !sameLineage && existingByLabel.ParentID != nil {
			sameLineage = *existingByLabel.ParentID == cache.ID
		}

		if !sameLineage && cache.ParentID != nil {
			sameLineage = *cache.ParentID == existingByLabel.ID
		}

		if sameLineage {
			return nil, status.Error(codes.FailedPrecondition, "Cannot update old version. You can only update the current version of this cache")
		}
		return nil, status.Error(codes.AlreadyExists, "Another cache with this label already exists")
	}

	newCache := &persistence.Cache{}
	newCache.FromProto(in)
	if cache.ParentID == nil {
		newCache.ParentID = &cache.ID
	} else {
		newCache.ParentID = cache.ParentID
	}
	newCache.IsCurrent = true

	if err = c.cacheRepo.Update(newCache); err != nil {
		log.Error().Err(err).Msg("Failed to update cache")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CacheResponse{
		Data: newCache.ToProto(),
		Meta: &pb.CommonResponse{Message: "Cache has been updated successfully"},
	}, nil
}

func (c *CoordinatorAPI) DeleteCache(_ context.Context, in *pb.GetCacheRequest) (*pb.CommonResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	cache, err := c.cacheRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find cache")
		return nil, status.Error(codes.Internal, err.Error())
	} else if cache == nil {
		return nil, status.Error(codes.NotFound, "Cache not found")
	}

	if err := c.cacheRepo.Delete(in.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to delete cache")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonResponse{
		Message: "Cache has been deleted successfully",
	}, nil
}
