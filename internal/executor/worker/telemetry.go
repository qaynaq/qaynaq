package worker

import (
	"context"
	"io"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/warpstreamlabs/bento/public/service"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/qaynaq/qaynaq/internal/persistence"
	pb "github.com/qaynaq/qaynaq/internal/protogen"
)

type TelemetryManager interface {
	ShipLogs(ctx context.Context)
	ShipMetrics(ctx context.Context)
}

type telemetryManager struct {
	coordinatorConnection CoordinatorConnection
	flowManager         FlowManager
}

func NewTelemetryManager(coordinatorConnection CoordinatorConnection, flowManager FlowManager) TelemetryManager {
	return &telemetryManager{
		coordinatorConnection: coordinatorConnection,
		flowManager:         flowManager,
	}
}

func (t *telemetryManager) ShipLogs(ctx context.Context) {
	retryDelay := time.Second

	for {
		eventStreamClient, err := t.coordinatorConnection.GetClient().IngestEvents(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create event stream client")
			retryDelay *= 2
			time.Sleep(min(retryDelay, EventStreamMaxDelay))
			continue
		}
		retryDelay = time.Second

		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("ShipLogs context done, closing event stream")
				if err := eventStreamClient.CloseSend(); err != nil {
					log.Error().Err(err).Msg("Error closing send event stream")
				}
				return
			default:
				sentEvents := 0
				flows := t.flowManager.GetAllFlows()

				for workerFlowID, flow := range flows {
					tracingSummary := flow.TracingSummary
					eventGetters := map[string]func(bool) map[string][]service.TracingEvent{
						string(persistence.FlowSectionInput):    tracingSummary.InputEvents,
						string(persistence.FlowSectionPipeline): tracingSummary.ProcessorEvents,
						string(persistence.FlowSectionOutput):   tracingSummary.OutputEvents,
					}

					for section, getEvents := range eventGetters {
						for componentLabel, events := range getEvents(true) {
							for _, event := range events {
								metaStruct, err := structpb.NewStruct(event.Meta)
								if err != nil {
									log.Error().
										Err(err).
										Int64("worker_flow_id", workerFlowID).
										Str("component_label", componentLabel).
										Str("event_type", string(event.Type)).
										Str("event_content", event.Content).
										Any("event_meta", event.Meta).
										Msg("Failed to convert meta field to pb struct")
									continue
								}

								if err := eventStreamClient.Send(&pb.Event{
									WorkerFlowId: workerFlowID,
									ComponentLabel: componentLabel,
									Section:        section,
									Type:           string(event.Type),
									Content:        event.Content,
									Meta:           metaStruct,
									TraceId:         event.FlowID,
								}); err != nil {
									log.Error().Err(err).Msg("Failed to send event, re-establishing connection")
									goto ReconnectEventStream
								}
								sentEvents++

								_, err = eventStreamClient.Recv()
								if err == io.EOF {
									log.Info().Msg("Server closed the connection")
									goto ReconnectEventStream
								}
								if err != nil {
									log.Error().Err(err).Msg("Failed to receive acknowledgment, re-establishing connection")
									goto ReconnectEventStream
								}
							}
						}
					}
				}

				if sentEvents == 0 {
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	ReconnectEventStream:
		log.Info().Msg("Attempting to reconnect flow...")
		time.Sleep(retryDelay)
	}
}

func (t *telemetryManager) ShipMetrics(ctx context.Context) {
	flows := t.flowManager.GetAllFlows()

	for workerFlowID, flow := range flows {
		tracingSummary := flow.TracingSummary
		err := t.coordinatorConnection.IngestMetrics(
			ctx,
			workerFlowID,
			tracingSummary.TotalInput(),
			tracingSummary.TotalProcessorErrors(),
			tracingSummary.TotalOutput(),
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send metrics")
		}
	}
}
