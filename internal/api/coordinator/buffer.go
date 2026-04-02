package coordinator

import (
	"context"

	"github.com/rs/zerolog/log"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/qaynaq/qaynaq/internal/persistence"
)

func (c *CoordinatorAPI) CreateBuffer(_ context.Context, in *pb.Buffer) (*pb.BufferResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	existing, err := c.bufferRepo.FindByLabel(in.GetLabel())
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing buffer")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if existing != nil {
		return nil, status.Error(codes.AlreadyExists, "Buffer with this label already exists")
	}

	buffer := &persistence.Buffer{}
	buffer.FromProto(in)
	buffer.ParentID = nil
	buffer.IsCurrent = true

	if err := c.bufferRepo.Create(buffer); err != nil {
		log.Error().Err(err).Msg("Failed to create buffer")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.BufferResponse{
		Data: buffer.ToProto(),
		Meta: &pb.CommonResponse{Message: "Buffer has been created successfully"},
	}, nil
}

func (c *CoordinatorAPI) GetBuffer(_ context.Context, in *pb.GetBufferRequest) (*pb.BufferResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	buffer, err := c.bufferRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find buffer")
		return nil, status.Error(codes.Internal, err.Error())
	} else if buffer == nil {
		return nil, status.Error(codes.NotFound, "Buffer not found")
	}

	return &pb.BufferResponse{
		Data: buffer.ToProto(),
		Meta: &pb.CommonResponse{Message: "OK"},
	}, nil
}

func (c *CoordinatorAPI) ListBuffers(_ context.Context, _ *emptypb.Empty) (*pb.ListBuffersResponse, error) {
	buffers, err := c.bufferRepo.ListAll()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list buffers")
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := &pb.ListBuffersResponse{
		Data: make([]*pb.Buffer, len(buffers)),
	}
	for i, buffer := range buffers {
		result.Data[i] = buffer.ToProto()
	}

	return result, nil
}

func (c *CoordinatorAPI) UpdateBuffer(_ context.Context, in *pb.Buffer) (*pb.BufferResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if in.GetId() == 0 {
		log.Debug().Msg("Invalid request: ID is required")
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	buffer, err := c.bufferRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find buffer")
		return nil, status.Error(codes.Internal, err.Error())
	} else if buffer == nil {
		return nil, status.Error(codes.NotFound, "Buffer not found")
	}

	existingByLabel, err := c.bufferRepo.FindByLabel(in.GetLabel())
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing buffer by label")
		return nil, status.Error(codes.Internal, err.Error())
	}

	if existingByLabel != nil && existingByLabel.ID != in.GetId() {
		sameLineage := false

		if existingByLabel.ParentID != nil && buffer.ParentID != nil {
			sameLineage = *existingByLabel.ParentID == *buffer.ParentID
		}

		if !sameLineage && existingByLabel.ParentID != nil {
			sameLineage = *existingByLabel.ParentID == buffer.ID
		}

		if !sameLineage && buffer.ParentID != nil {
			sameLineage = *buffer.ParentID == existingByLabel.ID
		}

		if sameLineage {
			return nil, status.Error(codes.FailedPrecondition, "Cannot update old version. You can only update the current version of this buffer")
		}
		return nil, status.Error(codes.AlreadyExists, "Another buffer with this label already exists")
	}

	newBuffer := &persistence.Buffer{}
	newBuffer.FromProto(in)
	if buffer.ParentID == nil {
		newBuffer.ParentID = &buffer.ID
	} else {
		newBuffer.ParentID = buffer.ParentID
	}
	newBuffer.IsCurrent = true

	if err = c.bufferRepo.Update(newBuffer); err != nil {
		log.Error().Err(err).Msg("Failed to update buffer")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.BufferResponse{
		Data: newBuffer.ToProto(),
		Meta: &pb.CommonResponse{Message: "Buffer has been updated successfully"},
	}, nil
}

func (c *CoordinatorAPI) DeleteBuffer(_ context.Context, in *pb.GetBufferRequest) (*pb.CommonResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	buffer, err := c.bufferRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find buffer")
		return nil, status.Error(codes.Internal, err.Error())
	} else if buffer == nil {
		return nil, status.Error(codes.NotFound, "Buffer not found")
	}

	if err := c.bufferRepo.Delete(in.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to delete buffer")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonResponse{
		Message: "Buffer has been deleted successfully",
	}, nil
}
