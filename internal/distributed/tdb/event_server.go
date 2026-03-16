// Licensed to the LF AI & Data foundation under one
// or more contributor license agreements. See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership. The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tdb

import (
	"context"
	"io"

	"go.uber.org/zap"

	"github.com/milvus-io/milvus/internal/eventcoord"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/proto/commonpb"
	"github.com/milvus-io/milvus/pkg/v2/proto/eventpb"
)

// EventServer implements the EventService gRPC interface.
type EventServer struct {
	eventpb.UnimplementedEventServiceServer
	coord *eventcoord.EventCoord
}

// NewEventServer creates a new EventServer instance.
func NewEventServer(coord *eventcoord.EventCoord) *EventServer {
	return &EventServer{
		coord: coord,
	}
}

// AppendEvent appends a new event to the event log.
func (s *EventServer) AppendEvent(ctx context.Context, req *eventpb.AppendEventRequest) (*eventpb.AppendEventResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "AppendEvent"))

	event := &models.Event{
		TenantID:      req.GetTenantId(),
		WorkspaceID:   req.GetWorkspaceId(),
		AgentID:       req.GetAgentId(),
		SessionID:     req.GetSessionId(),
		EventType:     models.EventType(req.GetEventType()),
		Payload:       req.GetPayload(),
		Importance:    req.GetImportance(),
		Visibility:    req.GetVisibility(),
		ParentEventID: req.GetParentEventId(),
		CausalRefs:    req.GetCausalRefs(),
		Metadata:      req.GetMetadata(),
	}

	eventID, logicalTs, err := s.coord.AppendEvent(ctx, event)
	if err != nil {
		log.Error("failed to append event", zap.Error(err))
		return &eventpb.AppendEventResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &eventpb.AppendEventResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
		EventId:   eventID,
		LogicalTs: logicalTs,
	}, nil
}

// GetEvent gets an event by ID.
func (s *EventServer) GetEvent(ctx context.Context, req *eventpb.GetEventRequest) (*eventpb.GetEventResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "GetEvent"))

	event, err := s.coord.GetEvent(ctx, req.GetEventId())
	if err != nil {
		log.Error("failed to get event", zap.Error(err))
		return &eventpb.GetEventResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &eventpb.GetEventResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
		Event: convertEventToProto(event),
	}, nil
}

// QueryEvents queries events with filter.
func (s *EventServer) QueryEvents(ctx context.Context, req *eventpb.QueryEventsRequest) (*eventpb.QueryEventsResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "QueryEvents"))

	filter := &models.EventFilter{
		AgentIDs:   req.GetFilter().GetAgentIds(),
		SessionIDs: req.GetFilter().GetSessionIds(),
		StartTime:  req.GetFilter().GetStartTime(),
		EndTime:    req.GetFilter().GetEndTime(),
	}

	events, totalCount, err := s.coord.QueryEvents(ctx, filter, req.GetLimit(), req.GetOffset())
	if err != nil {
		log.Error("failed to query events", zap.Error(err))
		return &eventpb.QueryEventsResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	eventList := make([]*eventpb.Event, 0, len(events))
	for _, e := range events {
		eventList = append(eventList, convertEventToProto(e))
	}

	return &eventpb.QueryEventsResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
		Events:     eventList,
		TotalCount: totalCount,
	}, nil
}

// SubscribeEvents subscribes to events stream.
func (s *EventServer) SubscribeEvents(req *eventpb.SubscribeEventsRequest, stream eventpb.EventService_SubscribeEventsServer) error {
	log := log.Ctx(stream.Context()).With(
		zap.String("method", "SubscribeEvents"),
		zap.String("subscriber_id", req.GetSubscriberId()),
	)

	filter := &models.EventFilter{
		AgentIDs:   req.GetFilter().GetAgentIds(),
		SessionIDs: req.GetFilter().GetSessionIds(),
	}

	// Create a subscription
	subscription, err := s.coord.SubscribeEvents(stream.Context(), req.GetSubscriberId(), filter, req.GetStartTimestamp())
	if err != nil {
		log.Error("failed to subscribe events", zap.Error(err))
		return stream.Send(&eventpb.SubscribeEventsResponse{
			Response: &eventpb.SubscribeEventsResponse_Status{
				Status: &commonpb.Status{
					ErrorCode: commonpb.ErrorCode_UnexpectedError,
					Reason:    err.Error(),
				},
			},
		})
	}
	defer s.coord.UnsubscribeEvents(req.GetSubscriberId())

	log.Info("event subscription started")

	// Stream events to client
	for {
		select {
		case <-stream.Context().Done():
			log.Info("event subscription cancelled by client")
			return nil
		case event, ok := <-subscription.EventCh:
			if !ok {
				log.Info("event subscription channel closed")
				return nil
			}
			if err := stream.Send(&eventpb.SubscribeEventsResponse{
				Response: &eventpb.SubscribeEventsResponse_Event{
					Event: convertEventToProto(event),
				},
			}); err != nil {
				if err == io.EOF {
					log.Info("client disconnected")
					return nil
				}
				log.Error("failed to send event to subscriber", zap.Error(err))
				return err
			}
		}
	}
}

// Helper functions to convert models to protobuf messages

func convertEventToProto(e *models.Event) *eventpb.Event {
	if e == nil {
		return nil
	}

	eventType := eventpb.EventType_EVENT_UNKNOWN
	switch e.EventType {
	case models.EventTypeUserMessage:
		eventType = eventpb.EventType_USER_MESSAGE
	case models.EventTypeAssistantMessage:
		eventType = eventpb.EventType_ASSISTANT_MESSAGE
	case models.EventTypeToolCallIssued:
		eventType = eventpb.EventType_TOOL_CALL_ISSUED
	case models.EventTypeToolResultReturned:
		eventType = eventpb.EventType_TOOL_RESULT_RETURNED
	case models.EventTypeRetrievalExecuted:
		eventType = eventpb.EventType_RETRIEVAL_EXECUTED
	case models.EventTypeMemoryWriteRequested:
		eventType = eventpb.EventType_MEMORY_WRITE_REQUESTED
	case models.EventTypeMemoryConsolidated:
		eventType = eventpb.EventType_MEMORY_CONSOLIDATED
	case models.EventTypeMemoryUpdated:
		eventType = eventpb.EventType_MEMORY_UPDATED
	case models.EventMemoryDeleted:
		eventType = eventpb.EventType_MEMORY_DELETED
	case models.EventTypePlanUpdated:
		eventType = eventpb.EventType_PLAN_UPDATED
	case models.EventTypePlanExecuted:
		eventType = eventpb.EventType_PLAN_EXECUTED
	case models.EventTypeCritiqueGenerated:
		eventType = eventpb.EventType_CRITIQUE_GENERATED
	case models.EventTypeReflectionCreated:
		eventType = eventpb.EventType_REFLECTION_CREATED
	case models.EventTypeTaskStarted:
		eventType = eventpb.EventType_TASK_STARTED
	case models.EventTypeTaskFinished:
		eventType = eventpb.EventType_TASK_FINISHED
	case models.EventTypeTaskFailed:
		eventType = eventpb.EventType_TASK_FAILED
	case models.EventTypeHandoffOccurred:
		eventType = eventpb.EventType_HANDOFF_OCCURRED
	case models.EventTypeSharedMemoryAccessed:
		eventType = eventpb.EventType_SHARED_MEMORY_ACCESSED
	case models.EventTypeSessionCreated:
		eventType = eventpb.EventType_SESSION_CREATED
	case models.EventTypeSessionEnded:
		eventType = eventpb.EventType_SESSION_ENDED
	case models.EventTypeAgentRegistered:
		eventType = eventpb.EventType_AGENT_REGISTERED
	case models.EventTypeAgentDeregistered:
		eventType = eventpb.EventType_AGENT_DEREGISTERED
	}

	return &eventpb.Event{
		EventId:       e.EventID,
		TenantId:      e.TenantID,
		WorkspaceId:   e.WorkspaceID,
		AgentId:       e.AgentID,
		SessionId:     e.SessionID,
		EventType:     eventType,
		EventTime:     e.EventTime,
		IngestTime:    e.IngestTime,
		VisibleTime:   e.VisibleTime,
		LogicalTs:     e.LogicalTs,
		ParentEventId: e.ParentEventID,
		CausalRefs:    e.CausalRefs,
		Payload:       e.Payload,
		Source:        e.Source,
		Importance:    e.Importance,
		Visibility:    e.Visibility,
		Version:       e.Version,
		Metadata:      e.Metadata,
	}
}
