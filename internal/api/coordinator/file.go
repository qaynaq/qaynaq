package coordinator

import (
	"context"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func (c *CoordinatorAPI) CreateFile(_ context.Context, in *pb.File) (*pb.FileResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if err := c.validateFileKey(in.GetKey(), 0); err != nil {
		return nil, err
	}

	existing, err := c.fileRepo.FindByKey(in.GetKey())
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing file")
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existing != nil {
		return nil, status.Error(codes.AlreadyExists, "File with this key already exists")
	}

	file := &persistence.File{}
	file.FromProto(in)

	if err := c.fileRepo.Create(file); err != nil {
		log.Error().Err(err).Msg("Failed to create file")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.FileResponse{
		Data: file.ToProtoWithoutContent(),
		Meta: &pb.CommonResponse{Message: "File has been created successfully"},
	}, nil
}

func (c *CoordinatorAPI) GetFile(_ context.Context, in *pb.GetFileRequest) (*pb.FileResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	file, err := c.fileRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find file")
		return nil, status.Error(codes.Internal, err.Error())
	} else if file == nil {
		return nil, status.Error(codes.NotFound, "File not found")
	}

	return &pb.FileResponse{
		Data: file.ToProto(),
		Meta: &pb.CommonResponse{Message: "OK"},
	}, nil
}

func (c *CoordinatorAPI) ListFiles(_ context.Context, _ *emptypb.Empty) (*pb.ListFilesResponse, error) {
	files, err := c.fileRepo.ListAll()
	if err != nil {
		log.Error().Err(err).Msg("Failed to list files")
		return nil, status.Error(codes.Internal, err.Error())
	}

	result := &pb.ListFilesResponse{
		Data: make([]*pb.File, len(files)),
	}
	for i, file := range files {
		result.Data[i] = file.ToProtoWithoutContent()
	}

	return result, nil
}

func (c *CoordinatorAPI) UpdateFile(_ context.Context, in *pb.File) (*pb.FileResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	if in.GetId() == 0 {
		log.Debug().Msg("Invalid request: ID is required")
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}

	file, err := c.fileRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find file")
		return nil, status.Error(codes.Internal, err.Error())
	} else if file == nil {
		return nil, status.Error(codes.NotFound, "File not found")
	}

	if err := c.validateFileKey(in.GetKey(), in.GetId()); err != nil {
		return nil, err
	}

	existingByKey, err := c.fileRepo.FindByKey(in.GetKey())
	if err != nil {
		log.Error().Err(err).Msg("Failed to check existing file by key")
		return nil, status.Error(codes.Internal, err.Error())
	}
	if existingByKey != nil && existingByKey.ID != in.GetId() {
		return nil, status.Error(codes.AlreadyExists, "Another file with this key already exists")
	}

	file.Key = in.GetKey()
	file.Content = in.GetContent()
	file.Size = int64(len(in.GetContent()))

	if err = c.fileRepo.Update(file); err != nil {
		log.Error().Err(err).Msg("Failed to update file")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.FileResponse{
		Data: file.ToProtoWithoutContent(),
		Meta: &pb.CommonResponse{Message: "File has been updated successfully"},
	}, nil
}

func (c *CoordinatorAPI) DeleteFile(_ context.Context, in *pb.GetFileRequest) (*pb.CommonResponse, error) {
	if err := in.Validate(); err != nil {
		log.Debug().Err(err).Msg("Invalid request")
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	file, err := c.fileRepo.FindByID(in.GetId())
	if err != nil {
		log.Error().Err(err).Msg("Failed to find file")
		return nil, status.Error(codes.Internal, err.Error())
	} else if file == nil {
		return nil, status.Error(codes.NotFound, "File not found")
	}

	if err := c.fileRepo.Delete(in.GetId()); err != nil {
		log.Error().Err(err).Msg("Failed to delete file")
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.CommonResponse{
		Message: "File has been deleted successfully",
	}, nil
}

func (c *CoordinatorAPI) validateFileKey(key string, excludeID int64) error {
	if strings.HasSuffix(key, "/") {
		return status.Error(codes.InvalidArgument, "File key must include a filename (cannot end with '/')")
	}

	conflict, err := c.fileRepo.HasKeyConflict(key, excludeID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check file key conflict")
		return status.Error(codes.Internal, "Failed to validate file key")
	}
	if conflict {
		return status.Errorf(codes.InvalidArgument, "File key \"%s\" conflicts with an existing file or folder of the same name", key)
	}

	return nil
}
