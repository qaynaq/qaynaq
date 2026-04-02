package coordinator

import (
	"context"

	"github.com/rs/zerolog/log"

	coordinatorexecutor "github.com/qaynaq/qaynaq/internal/executor/coordinator"
	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func (c *CoordinatorAPI) ValidateFlow(_ context.Context, in *pb.ValidateFlowRequest) (*pb.ValidateFlowResponse, error) {
	flow := persistence.Flow{
		InputComponent:  in.GetInputComponent(),
		InputLabel:      in.GetInputLabel(),
		InputConfig:     []byte(in.GetInputConfig()),
		OutputComponent: in.GetOutputComponent(),
		OutputLabel:     in.GetOutputLabel(),
		OutputConfig:    []byte(in.GetOutputConfig()),
		Processors:      make([]persistence.FlowProcessor, len(in.GetProcessors())),
	}
	for i, p := range in.GetProcessors() {
		flow.Processors[i] = persistence.FlowProcessor{
			Label:     p.GetLabel(),
			Component: p.GetComponent(),
			Config:    []byte(p.GetConfig()),
		}
	}

	if err := coordinatorexecutor.ValidateFlow(flow); err != nil {
		log.Debug().Err(err).Msg("flow validation failed")
		return &pb.ValidateFlowResponse{Valid: false, Error: err.Error()}, nil
	}

	return &pb.ValidateFlowResponse{Valid: true}, nil
}
