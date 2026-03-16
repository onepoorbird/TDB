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

package agentcoord

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/milvus-io/milvus/internal/metastore/kv/agentcoord"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

// MockCatalog is a mock implementation of the Catalog interface
type MockCatalog struct {
	mock.Mock
}

func (m *MockCatalog) CreateAgent(ctx context.Context, agent *models.Agent, ts typeutil.Timestamp) error {
	args := m.Called(ctx, agent, ts)
	return args.Error(0)
}

func (m *MockCatalog) GetAgent(ctx context.Context, agentID string, ts typeutil.Timestamp) (*models.Agent, error) {
	args := m.Called(ctx, agentID, ts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Agent), args.Error(1)
}

func (m *MockCatalog) ListAgents(ctx context.Context, tenantID, workspaceID string, ts typeutil.Timestamp) ([]*models.Agent, error) {
	args := m.Called(ctx, tenantID, workspaceID, ts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Agent), args.Error(1)
}

func (m *MockCatalog) UpdateAgent(ctx context.Context, agent *models.Agent, ts typeutil.Timestamp) error {
	args := m.Called(ctx, agent, ts)
	return args.Error(0)
}

func (m *MockCatalog) DeleteAgent(ctx context.Context, agentID string, ts typeutil.Timestamp) error {
	args := m.Called(ctx, agentID, ts)
	return args.Error(0)
}

func (m *MockCatalog) AgentExists(ctx context.Context, agentID string, ts typeutil.Timestamp) bool {
	args := m.Called(ctx, agentID, ts)
	return args.Bool(0)
}

func (m *MockCatalog) CreateSession(ctx context.Context, session *models.Session, ts typeutil.Timestamp) error {
	args := m.Called(ctx, session, ts)
	return args.Error(0)
}

func (m *MockCatalog) GetSession(ctx context.Context, sessionID string, ts typeutil.Timestamp) (*models.Session, error) {
	args := m.Called(ctx, sessionID, ts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockCatalog) ListSessions(ctx context.Context, agentID string, ts typeutil.Timestamp) ([]*models.Session, error) {
	args := m.Called(ctx, agentID, ts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockCatalog) ListSessionsByState(ctx context.Context, agentID string, state models.SessionState, ts typeutil.Timestamp) ([]*models.Session, error) {
	args := m.Called(ctx, agentID, state, ts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockCatalog) UpdateSession(ctx context.Context, session *models.Session, ts typeutil.Timestamp) error {
	args := m.Called(ctx, session, ts)
	return args.Error(0)
}

func (m *MockCatalog) DeleteSession(ctx context.Context, sessionID string, ts typeutil.Timestamp) error {
	args := m.Called(ctx, sessionID, ts)
	return args.Error(0)
}

// AgentCoordTestSuite is the test suite for AgentCoord
type AgentCoordTestSuite struct {
	suite.Suite
	coord   *AgentCoord
	catalog *MockCatalog
	ctx     context.Context
}

func (s *AgentCoordTestSuite) SetupTest() {
	s.catalog = new(MockCatalog)
	s.coord = &AgentCoord{
		catalog: s.catalog,
	}
	s.ctx = context.Background()
}

func (s *AgentCoordTestSuite) TearDownTest() {
	s.catalog.AssertExpectations(s.T())
}

func (s *AgentCoordTestSuite) TestCreateAgent() {
	// Test creating a new agent
	s.catalog.On("CreateAgent", s.ctx, mock.AnythingOfType("*models.Agent"), mock.AnythingOfType("uint64")).
		Return(nil).Once()

	agent, err := s.coord.CreateAgent(
		s.ctx,
		"tenant-001",
		"workspace-001",
		"assistant",
		"helpful assistant",
		[]string{"chat", "search"},
		"standard",
		map[string]string{"key": "value"},
	)

	s.NoError(err)
	s.NotNil(agent)
	s.Equal("tenant-001", agent.TenantID)
	s.Equal("workspace-001", agent.WorkspaceID)
	s.Equal("assistant", agent.AgentType)
	s.Equal("helpful assistant", agent.RoleProfile)
	s.Equal([]string{"chat", "search"}, agent.CapabilitySet)
	s.Equal("standard", agent.DefaultMemoryPolicy)
	s.Equal(models.AgentStateCreating, agent.State)
	s.NotEmpty(agent.AgentID)
}

func (s *AgentCoordTestSuite) TestGetAgent() {
	// Test getting an existing agent
	expectedAgent := &models.Agent{
		AgentID:     "agent-001",
		TenantID:    "tenant-001",
		WorkspaceID: "workspace-001",
		AgentType:   "assistant",
		State:       models.AgentStateActive,
	}

	s.catalog.On("GetAgent", s.ctx, "agent-001", mock.AnythingOfType("uint64")).
		Return(expectedAgent, nil).Once()

	agent, err := s.coord.GetAgent(s.ctx, "agent-001")

	s.NoError(err)
	s.NotNil(agent)
	s.Equal("agent-001", agent.AgentID)
	s.Equal("tenant-001", agent.TenantID)
}

func (s *AgentCoordTestSuite) TestGetAgent_NotFound() {
	// Test getting a non-existent agent
	s.catalog.On("GetAgent", s.ctx, "agent-not-found", mock.AnythingOfType("uint64")).
		Return(nil, agentcoord.ErrKeyNotFound).Once()

	agent, err := s.coord.GetAgent(s.ctx, "agent-not-found")

	s.Error(err)
	s.Nil(agent)
}

func (s *AgentCoordTestSuite) TestListAgents() {
	// Test listing agents
	expectedAgents := []*models.Agent{
		{
			AgentID:     "agent-001",
			TenantID:    "tenant-001",
			WorkspaceID: "workspace-001",
			AgentType:   "assistant",
		},
		{
			AgentID:     "agent-002",
			TenantID:    "tenant-001",
			WorkspaceID: "workspace-001",
			AgentType:   "bot",
		},
	}

	s.catalog.On("ListAgents", s.ctx, "tenant-001", "workspace-001", mock.AnythingOfType("uint64")).
		Return(expectedAgents, nil).Once()

	agents, err := s.coord.ListAgents(s.ctx, "tenant-001", "workspace-001")

	s.NoError(err)
	s.Len(agents, 2)
	s.Equal("agent-001", agents[0].AgentID)
	s.Equal("agent-002", agents[1].AgentID)
}

func (s *AgentCoordTestSuite) TestUpdateAgent() {
	// Test updating an agent
	existingAgent := &models.Agent{
		AgentID:     "agent-001",
		TenantID:    "tenant-001",
		WorkspaceID: "workspace-001",
		AgentType:   "assistant",
		RoleProfile: "old profile",
		State:       models.AgentStateActive,
	}

	s.catalog.On("GetAgent", s.ctx, "agent-001", mock.AnythingOfType("uint64")).
		Return(existingAgent, nil).Once()
	s.catalog.On("UpdateAgent", s.ctx, mock.AnythingOfType("*models.Agent"), mock.AnythingOfType("uint64")).
		Return(nil).Once()

	updates := map[string]interface{}{
		"role_profile": "new profile",
	}

	agent, err := s.coord.UpdateAgent(s.ctx, "agent-001", updates)

	s.NoError(err)
	s.NotNil(agent)
	s.Equal("new profile", agent.RoleProfile)
}

func (s *AgentCoordTestSuite) TestDeleteAgent() {
	// Test deleting an agent
	s.catalog.On("DeleteAgent", s.ctx, "agent-001", mock.AnythingOfType("uint64")).
		Return(nil).Once()

	err := s.coord.DeleteAgent(s.ctx, "agent-001")

	s.NoError(err)
}

func (s *AgentCoordTestSuite) TestCreateSession() {
	// Test creating a new session
	s.catalog.On("CreateSession", s.ctx, mock.AnythingOfType("*models.Session"), mock.AnythingOfType("uint64")).
		Return(nil).Once()

	session, err := s.coord.CreateSession(
		s.ctx,
		"agent-001",
		"",
		"chat",
		"help the user",
		1000,
		3600000,
		map[string]string{"key": "value"},
	)

	s.NoError(err)
	s.NotNil(session)
	s.Equal("agent-001", session.AgentID)
	s.Equal("chat", session.TaskType)
	s.Equal("help the user", session.Goal)
	s.Equal(int64(1000), session.BudgetToken)
	s.Equal(int64(3600000), session.BudgetTimeMs)
	s.Equal(models.SessionStateCreating, session.State)
	s.NotEmpty(session.SessionID)
}

func (s *AgentCoordTestSuite) TestGetSession() {
	// Test getting an existing session
	expectedSession := &models.Session{
		SessionID: "session-001",
		AgentID:   "agent-001",
		TaskType:  "chat",
		Goal:      "help the user",
		State:     models.SessionStateActive,
	}

	s.catalog.On("GetSession", s.ctx, "session-001", mock.AnythingOfType("uint64")).
		Return(expectedSession, nil).Once()

	session, err := s.coord.GetSession(s.ctx, "session-001")

	s.NoError(err)
	s.NotNil(session)
	s.Equal("session-001", session.SessionID)
	s.Equal("agent-001", session.AgentID)
}

func (s *AgentCoordTestSuite) TestListSessions() {
	// Test listing sessions
	expectedSessions := []*models.Session{
		{
			SessionID: "session-001",
			AgentID:   "agent-001",
			TaskType:  "chat",
			State:     models.SessionStateActive,
		},
		{
			SessionID: "session-002",
			AgentID:   "agent-001",
			TaskType:  "search",
			State:     models.SessionStateCompleted,
		},
	}

	s.catalog.On("ListSessions", s.ctx, "agent-001", mock.AnythingOfType("uint64")).
		Return(expectedSessions, nil).Once()

	sessions, err := s.coord.ListSessions(s.ctx, "agent-001")

	s.NoError(err)
	s.Len(sessions, 2)
	s.Equal("session-001", sessions[0].SessionID)
	s.Equal("session-002", sessions[1].SessionID)
}

func (s *AgentCoordTestSuite) TestUpdateSession() {
	// Test updating a session
	existingSession := &models.Session{
		SessionID: "session-001",
		AgentID:   "agent-001",
		TaskType:  "chat",
		Goal:      "old goal",
		State:     models.SessionStateActive,
	}

	s.catalog.On("GetSession", s.ctx, "session-001", mock.AnythingOfType("uint64")).
		Return(existingSession, nil).Once()
	s.catalog.On("UpdateSession", s.ctx, mock.AnythingOfType("*models.Session"), mock.AnythingOfType("uint64")).
		Return(nil).Once()

	updates := map[string]interface{}{
		"goal": "new goal",
	}

	session, err := s.coord.UpdateSession(s.ctx, "session-001", updates)

	s.NoError(err)
	s.NotNil(session)
	s.Equal("new goal", session.Goal)
}

func TestAgentCoordSuite(t *testing.T) {
	suite.Run(t, new(AgentCoordTestSuite))
}

// Benchmark tests

func BenchmarkCreateAgent(b *testing.B) {
	coord := &AgentCoord{}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = coord.generateAgentID()
	}
}

func BenchmarkGenerateSessionID(b *testing.B) {
	coord := &AgentCoord{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = coord.generateSessionID()
	}
}
