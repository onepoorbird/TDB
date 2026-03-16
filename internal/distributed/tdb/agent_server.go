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

	"go.uber.org/zap"

	"github.com/milvus-io/milvus/internal/agentcoord"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/proto/agentpb"
	"github.com/milvus-io/milvus/pkg/v2/proto/commonpb"
)

// AgentServer implements the AgentService gRPC interface.
type AgentServer struct {
	agentpb.UnimplementedAgentServiceServer
	coord *agentcoord.AgentCoord
}

// NewAgentServer creates a new AgentServer instance.
func NewAgentServer(coord *agentcoord.AgentCoord) *AgentServer {
	return &AgentServer{
		coord: coord,
	}
}

// CreateAgent creates a new agent.
func (s *AgentServer) CreateAgent(ctx context.Context, req *agentpb.CreateAgentRequest) (*agentpb.CreateAgentResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "CreateAgent"))
	
	agent, err := s.coord.CreateAgent(
		ctx,
		req.GetTenantId(),
		req.GetWorkspaceId(),
		req.GetAgentType(),
		req.GetRoleProfile(),
		req.GetCapabilitySet(),
		req.GetDefaultMemoryPolicy(),
		req.GetMetadata(),
	)
	if err != nil {
		log.Error("failed to create agent", zap.Error(err))
		return &agentpb.CreateAgentResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &agentpb.CreateAgentResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
		AgentId: agent.AgentID,
	}, nil
}

// GetAgent gets an agent by ID.
func (s *AgentServer) GetAgent(ctx context.Context, req *agentpb.GetAgentRequest) (*agentpb.GetAgentResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "GetAgent"))
	
	agent, err := s.coord.GetAgent(ctx, req.GetAgentId())
	if err != nil {
		log.Error("failed to get agent", zap.Error(err))
		return &agentpb.GetAgentResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &agentpb.GetAgentResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
		Agent: convertAgentToProto(agent),
	}, nil
}

// ListAgents lists all agents for a tenant/workspace.
func (s *AgentServer) ListAgents(ctx context.Context, req *agentpb.ListAgentsRequest) (*agentpb.ListAgentsResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "ListAgents"))
	
	agents, err := s.coord.ListAgents(ctx, req.GetTenantId(), req.GetWorkspaceId())
	if err != nil {
		log.Error("failed to list agents", zap.Error(err))
		return &agentpb.ListAgentsResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	agentInfos := make([]*agentpb.AgentInfo, 0, len(agents))
	for _, agent := range agents {
		agentInfos = append(agentInfos, convertAgentToProto(agent))
	}

	return &agentpb.ListAgentsResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
		Agents: agentInfos,
	}, nil
}

// UpdateAgent updates an existing agent.
func (s *AgentServer) UpdateAgent(ctx context.Context, req *agentpb.UpdateAgentRequest) (*agentpb.UpdateAgentResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "UpdateAgent"))
	
	updates := make(map[string]interface{})
	if req.GetRoleProfile() != "" {
		updates["role_profile"] = req.GetRoleProfile()
	}
	if len(req.GetCapabilitySet()) > 0 {
		updates["capability_set"] = req.GetCapabilitySet()
	}
	if req.GetDefaultMemoryPolicy() != "" {
		updates["default_memory_policy"] = req.GetDefaultMemoryPolicy()
	}
	if len(req.GetMetadata()) > 0 {
		updates["metadata"] = req.GetMetadata()
	}

	_, err := s.coord.UpdateAgent(ctx, req.GetAgentId(), updates)
	if err != nil {
		log.Error("failed to update agent", zap.Error(err))
		return &agentpb.UpdateAgentResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &agentpb.UpdateAgentResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
	}, nil
}

// DeleteAgent deletes an agent.
func (s *AgentServer) DeleteAgent(ctx context.Context, req *agentpb.DeleteAgentRequest) (*agentpb.DeleteAgentResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "DeleteAgent"))
	
	err := s.coord.DeleteAgent(ctx, req.GetAgentId())
	if err != nil {
		log.Error("failed to delete agent", zap.Error(err))
		return &agentpb.DeleteAgentResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &agentpb.DeleteAgentResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
	}, nil
}

// CreateSession creates a new session.
func (s *AgentServer) CreateSession(ctx context.Context, req *agentpb.CreateSessionRequest) (*agentpb.CreateSessionResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "CreateSession"))
	
	session, err := s.coord.CreateSession(
		ctx,
		req.GetAgentId(),
		req.GetParentSessionId(),
		req.GetTaskType(),
		req.GetGoal(),
		req.GetBudgetToken(),
		req.GetBudgetTimeMs(),
		req.GetMetadata(),
	)
	if err != nil {
		log.Error("failed to create session", zap.Error(err))
		return &agentpb.CreateSessionResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &agentpb.CreateSessionResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
		SessionId: session.SessionID,
	}, nil
}

// GetSession gets a session by ID.
func (s *AgentServer) GetSession(ctx context.Context, req *agentpb.GetSessionRequest) (*agentpb.GetSessionResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "GetSession"))
	
	session, err := s.coord.GetSession(ctx, req.GetSessionId())
	if err != nil {
		log.Error("failed to get session", zap.Error(err))
		return &agentpb.GetSessionResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &agentpb.GetSessionResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
		Session: convertSessionToProto(session),
	}, nil
}

// ListSessions lists all sessions for an agent.
func (s *AgentServer) ListSessions(ctx context.Context, req *agentpb.ListSessionsRequest) (*agentpb.ListSessionsResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "ListSessions"))
	
	sessions, err := s.coord.ListSessions(ctx, req.GetAgentId())
	if err != nil {
		log.Error("failed to list sessions", zap.Error(err))
		return &agentpb.ListSessionsResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	sessionInfos := make([]*agentpb.SessionInfo, 0, len(sessions))
	for _, session := range sessions {
		sessionInfos = append(sessionInfos, convertSessionToProto(session))
	}

	return &agentpb.ListSessionsResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
		Sessions: sessionInfos,
	}, nil
}

// UpdateSession updates an existing session.
func (s *AgentServer) UpdateSession(ctx context.Context, req *agentpb.UpdateSessionRequest) (*agentpb.UpdateSessionResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "UpdateSession"))
	
	updates := make(map[string]interface{})
	if req.GetGoal() != "" {
		updates["goal"] = req.GetGoal()
	}
	if len(req.GetMetadata()) > 0 {
		updates["metadata"] = req.GetMetadata()
	}

	_, err := s.coord.UpdateSession(ctx, req.GetSessionId(), updates)
	if err != nil {
		log.Error("failed to update session", zap.Error(err))
		return &agentpb.UpdateSessionResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &agentpb.UpdateSessionResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
	}, nil
}

// Helper functions to convert models to protobuf messages

func convertAgentToProto(agent *models.Agent) *agentpb.AgentInfo {
	if agent == nil {
		return nil
	}

	state := agentpb.AgentState_AGENT_UNKNOWN
	switch agent.State {
	case models.AgentStateCreating:
		state = agentpb.AgentState_AGENT_CREATING
	case models.AgentStateActive:
		state = agentpb.AgentState_AGENT_ACTIVE
	case models.AgentStatePaused:
		state = agentpb.AgentState_AGENT_PAUSED
	case models.AgentStateTerminated:
		state = agentpb.AgentState_AGENT_TERMINATED
	}

	return &agentpb.AgentInfo{
		AgentId:             agent.AgentID,
		TenantId:            agent.TenantID,
		WorkspaceId:         agent.WorkspaceID,
		AgentType:           agent.AgentType,
		RoleProfile:         agent.RoleProfile,
		PolicyRef:           agent.PolicyRef,
		CapabilitySet:       agent.CapabilitySet,
		DefaultMemoryPolicy: agent.DefaultMemoryPolicy,
		CreatedTs:           agent.CreatedAt,
		UpdatedTs:           agent.UpdatedAt,
		State:               state,
		Metadata:            agent.Metadata,
	}
}

func convertSessionToProto(session *models.Session) *agentpb.SessionInfo {
	if session == nil {
		return nil
	}

	state := agentpb.SessionState_SESSION_UNKNOWN
	switch session.State {
	case models.SessionStateCreating:
		state = agentpb.SessionState_SESSION_CREATING
	case models.SessionStateActive:
		state = agentpb.SessionState_SESSION_ACTIVE
	case models.SessionStatePaused:
		state = agentpb.SessionState_SESSION_PAUSED
	case models.SessionStateCompleted:
		state = agentpb.SessionState_SESSION_COMPLETED
	case models.SessionStateFailed:
		state = agentpb.SessionState_SESSION_FAILED
	}

	return &agentpb.SessionInfo{
		SessionId:       session.SessionID,
		AgentId:         session.AgentID,
		ParentSessionId: session.ParentSessionID,
		TaskType:        session.TaskType,
		Goal:            session.Goal,
		ContextRef:      session.ContextRef,
		StartTs:         session.StartAt,
		EndTs:           session.EndAt,
		State:           state,
		BudgetToken:     session.BudgetToken,
		BudgetTimeMs:    session.BudgetTimeMs,
		Metadata:        session.Metadata,
	}
}
