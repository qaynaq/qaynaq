package coordinator

import (
	"context"

	"github.com/rs/zerolog/log"

	coordinatorexecutor "github.com/qaynaq/qaynaq/internal/executor/coordinator"
	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

func (c *CoordinatorAPI) TryFlow(ctx context.Context, in *pb.TryFlowRequest) (*pb.TryFlowResponse, error) {
	processors := make([]persistence.FlowProcessor, len(in.GetProcessors()))
	for i, p := range in.GetProcessors() {
		processors[i] = persistence.FlowProcessor{
			Label:     p.GetLabel(),
			Component: p.GetComponent(),
			Config:    []byte(p.GetConfig()),
		}
	}

	messages := make([]coordinatorexecutor.TryMessage, len(in.GetMessages()))
	for i, m := range in.GetMessages() {
		messages[i] = coordinatorexecutor.TryMessage{Content: m.GetContent()}
	}

	envVarLookupFn := func(key string) (string, bool) {
		secret, err := c.secretRepo.GetByKey(key)
		if err != nil {
			return "", false
		}
		decrypted, err := c.aesgcm.Decrypt(secret.EncryptedValue)
		if err != nil {
			return "", false
		}
		return decrypted, true
	}

	result := coordinatorexecutor.TryStream(ctx, processors, messages, coordinatorexecutor.TryFlowOptions{
		EnvVarLookupFn: envVarLookupFn,
		FileRepo:       c.fileRepo,
	})
	if result.Error != "" {
		log.Debug().Str("error", result.Error).Msg("flow try failed")
	}

	outputs := make([]*pb.TryFlowResponse_TryOutput, len(result.Outputs))
	for i, o := range result.Outputs {
		outputs[i] = &pb.TryFlowResponse_TryOutput{Content: o.Content}
	}

	return &pb.TryFlowResponse{
		Outputs: outputs,
		Error:   result.Error,
	}, nil
}
