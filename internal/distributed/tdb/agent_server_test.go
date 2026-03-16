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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/milvus-io/milvus/internal/agentcoord"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/proto/agentpb"
	"github.com/milvus-io/milvus/pkg/v2/proto/commonpb"
)

// MockAgentCoord is a mock implementation of the AgentCoord
type MockAgentCoord struct {
	mock.Mock
}

func (m *MockAgentCoord) CreateAgent(ctx context.Context, tenantID, workspaceID, agentType, roleProfile string, capabilitySet []string, defaultMemoryPolicy string, metadata map[string]string) (*models.Agent, error) {
	args := m.Called(ctx, tenantID, workspaceID, agentType, roleProfile, capabilitySet, defaultMemoryPolicy, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func (m *MockAgentCoord) GetAgent(ctx context.Context, agentID string) (*models.Agent, error) {
	args := m.Called(ctx, agentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func (m *MockAgentCoord) ListAgents(ctx context.Context, tenantID, workspaceID string) ([]*models.Agent, error) {
	args := m.Called(ctx, tenantID, workspaceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Agent), args.Error(1)
}

func (m *MockAgentCoord) UpdateAgent(ctx context.Context, agentID string, updates map[string]interface{}) (*models.Agent, error) {
	args := m.Called(ctx, agentID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func (m *MockAgentCoord) DeleteAgent(ctx context.Context, agentID string) error {
	args := m.Called(ctx, agentID)
	return args.Error(0)
}

func (m *MockAgentCoord) CreateSession(ctx context.Context, agentID, parentSessionID, taskType, goal string, budgetToken, budgetTimeMs int64, metadata map[string]string) (*models.Session, error) {
	args := m.Called(ctx, agentID, parentSessionID, taskType, goal, budgetToken, budgetTimeMs, metadata)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockAgentCoord) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockAgentCoord) ListSessions(ctx context.Context, agentID string) ([]*models.Session, error) {
	args := m.Called(ctx, agentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockAgentCoord) UpdateSession(ctx context.Context, sessionID string, updates map[string]interface{}) (*models.Session, error) {
	args := m.Called(ctx, sessionID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

// AgentServerTestSuite is the test suite for AgentServer
type AgentServerTestSuite struct {
	suite.Suite
	server *AgentServer
	coord  *MockAgentCoord
	ctx    context.Context
}

func (s *AgentServerTestSuite) SetupTest() {
	s.coord = new(MockAgentCoord)
	s.server = &AgentServer{
		coord: (*agentcoord.AgentCoord)(s.coord), // Type assertion for testing
	}
	s.ctx = context.Background()
}

func (s *AgentServerTestSuite) TearDownTest() {
	s.coord.AssertExpectations(s.T())
}

func (s *AgentServerTestSuite) TestCreateAgent() {
	expectedAgent := &models.Agent{
		AgentID:     "agent-001",
		TenantID:    "tenant-001",
		WorkspaceID: "workspace-001",
		AgentType:   "assistant",
	}

	s.coord.On("CreateAgent", s.ctx, "tenant-001", "workspace-001", "assistant", "helpful assistant", []string{"chat"}, "standard", mock.Anything).
		Return(expectedAgent, nil).Once()

	req := &agentpb.CreateAgentRequest{
		TenantId:            "tenant-001",
		WorkspaceId:         "workspace-001",
		AgentType:           "assistant",
		RoleProfile:         "helpful assistant",
		CapabilitySet:       []string{"chat"},
		DefaultMemoryPolicy: "standard",
	}

	resp, err := s.server.CreateAgent(s.ctx, req)

	s.NoError(err)
	s.NotNil(resp)
	s.Equal(commonpb.ErrorCode_Success, resp.Status.ErrorCode)
	s.Equal("agent-001", resp.AgentId)
}

func (s *AgentServerTestSuite) TestGetAgent() {
	expectedAgent := &models.Agent{
		AgentID:     "agent-001",
		TenantID:    "tenant-001",
		WorkspaceID: "workspace-001",
		AgentType:   "assistant",
		State:       models.AgentStateActive,
	}

	s.coord.On("GetAgent", s.ctx, "agent-001").
		Return(expectedAgent, nil).Once()

	req := &agentpb.GetAgentRequest{
		AgentId: "agent-001",
	}

	resp, err := s.server.GetAgent(s.ctx, req)

	s.NoError(err)
	s.NotNil(resp)
	s.Equal(commonpb.ErrorCode_Success, resp.Status.ErrorCode)
	s.NotNil(resp.Agent)
	s.Equal("agent-001", resp.Agent.AgentId)
}

func TestAgentServerSuite(t *testing.T) {
	suite.Run(t, new(AgentServerTestSuite))
}
