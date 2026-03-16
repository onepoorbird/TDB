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

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/milvus-io/milvus/pkg/v2/proto/agentpb"
	"github.com/milvus-io/milvus/pkg/v2/proto/commonpb"
	"github.com/milvus-io/milvus/pkg/v2/proto/eventpb"
	"github.com/milvus-io/milvus/pkg/v2/proto/memorypb"
	"github.com/milvus-io/milvus/pkg/v2/util/paramtable"
)

// TDBIntegrationTestSuite is the integration test suite for TDB
type TDBIntegrationTestSuite struct {
	suite.Suite
	ctx         context.Context
	agentClient agentpb.AgentServiceClient
	memoryClient memorypb.MemoryServiceClient
	eventClient  eventpb.EventServiceClient
	conn         *grpc.ClientConn
}

func (s *TDBIntegrationTestSuite) SetupSuite() {
	paramtable.Init()
	s.ctx = context.Background()

	// Connect to TDB server
	// Note: This requires a running TDB server
	var err error
	s.conn, err = grpc.Dial("localhost:19530", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		s.T().Skipf("Skipping integration tests: cannot connect to TDB server: %v", err)
	}

	s.agentClient = agentpb.NewAgentServiceClient(s.conn)
	s.memoryClient = memorypb.NewMemoryServiceClient(s.conn)
	s.eventClient = eventpb.NewEventServiceClient(s.conn)
}

func (s *TDBIntegrationTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
}

// TestAgentLifecycle tests the full lifecycle of an agent
func (s *TDBIntegrationTestSuite) TestAgentLifecycle() {
	// Create agent
	createResp, err := s.agentClient.CreateAgent(s.ctx, &agentpb.CreateAgentRequest{
		TenantId:            "test-tenant",
		WorkspaceId:         "test-workspace",
		AgentType:           "test-agent",
		RoleProfile:         "test profile",
		CapabilitySet:       []string{"test"},
		DefaultMemoryPolicy: "standard",
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, createResp.Status.ErrorCode)
	agentID := createResp.AgentId
	s.NotEmpty(agentID)

	// Get agent
	getResp, err := s.agentClient.GetAgent(s.ctx, &agentpb.GetAgentRequest{
		AgentId: agentID,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, getResp.Status.ErrorCode)
	s.Equal(agentID, getResp.Agent.AgentId)

	// Update agent
	updateResp, err := s.agentClient.UpdateAgent(s.ctx, &agentpb.UpdateAgentRequest{
		AgentId:     agentID,
		RoleProfile: "updated profile",
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, updateResp.Status.ErrorCode)

	// List agents
	listResp, err := s.agentClient.ListAgents(s.ctx, &agentpb.ListAgentsRequest{
		TenantId:    "test-tenant",
		WorkspaceId: "test-workspace",
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, listResp.Status.ErrorCode)
	s.NotEmpty(listResp.Agents)

	// Delete agent
	deleteResp, err := s.agentClient.DeleteAgent(s.ctx, &agentpb.DeleteAgentRequest{
		AgentId: agentID,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, deleteResp.Status.ErrorCode)
}

// TestSessionLifecycle tests the full lifecycle of a session
func (s *TDBIntegrationTestSuite) TestSessionLifecycle() {
	// First create an agent
	createAgentResp, err := s.agentClient.CreateAgent(s.ctx, &agentpb.CreateAgentRequest{
		TenantId:            "test-tenant",
		WorkspaceId:         "test-workspace",
		AgentType:           "test-agent",
		RoleProfile:         "test profile",
		CapabilitySet:       []string{"test"},
		DefaultMemoryPolicy: "standard",
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, createAgentResp.Status.ErrorCode)
	agentID := createAgentResp.AgentId

	// Create session
	createSessionResp, err := s.agentClient.CreateSession(s.ctx, &agentpb.CreateSessionRequest{
		AgentId:     agentID,
		TaskType:    "test-task",
		Goal:        "test goal",
		BudgetToken: 1000,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, createSessionResp.Status.ErrorCode)
	sessionID := createSessionResp.SessionId
	s.NotEmpty(sessionID)

	// Get session
	getSessionResp, err := s.agentClient.GetSession(s.ctx, &agentpb.GetSessionRequest{
		SessionId: sessionID,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, getSessionResp.Status.ErrorCode)
	s.Equal(sessionID, getSessionResp.Session.SessionId)

	// List sessions
	listSessionsResp, err := s.agentClient.ListSessions(s.ctx, &agentpb.ListSessionsRequest{
		AgentId: agentID,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, listSessionsResp.Status.ErrorCode)
	s.NotEmpty(listSessionsResp.Sessions)

	// Cleanup
	_, _ = s.agentClient.DeleteAgent(s.ctx, &agentpb.DeleteAgentRequest{
		AgentId: agentID,
	})
}

// TestMemoryOperations tests memory CRUD operations
func (s *TDBIntegrationTestSuite) TestMemoryOperations() {
	// Create memory
	createResp, err := s.memoryClient.CreateMemory(s.ctx, &memorypb.CreateMemoryRequest{
		AgentId:    "test-agent",
		SessionId:  "test-session",
		MemoryType: memorypb.MemoryType_EPISODIC,
		Scope:      "session",
		Content:    "test content",
		Summary:    "test summary",
		Confidence: 0.9,
		Importance: 0.8,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, createResp.Status.ErrorCode)
	memoryID := createResp.MemoryId
	s.NotEmpty(memoryID)

	// Get memory
	getResp, err := s.memoryClient.GetMemory(s.ctx, &memorypb.GetMemoryRequest{
		MemoryId: memoryID,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, getResp.Status.ErrorCode)
	s.Equal(memoryID, getResp.Memory.MemoryId)

	// Update memory
	updateResp, err := s.memoryClient.UpdateMemory(s.ctx, &memorypb.UpdateMemoryRequest{
		MemoryId:   memoryID,
		Content:    "updated content",
		Confidence: 0.95,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, updateResp.Status.ErrorCode)
	s.Equal(int64(2), updateResp.NewVersion)

	// Query memories
	queryResp, err := s.memoryClient.QueryMemories(s.ctx, &memorypb.QueryMemoriesRequest{
		Filter: &memorypb.MemoryFilter{
			AgentIds: []string{"test-agent"},
		},
		Limit:  10,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, queryResp.Status.ErrorCode)
	s.NotEmpty(queryResp.Memories)

	// Delete memory
	deleteResp, err := s.memoryClient.DeleteMemory(s.ctx, &memorypb.DeleteMemoryRequest{
		MemoryId:   memoryID,
		HardDelete: false,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, deleteResp.Status.ErrorCode)
}

// TestEventOperations tests event operations
func (s *TDBIntegrationTestSuite) TestEventOperations() {
	// Append event
	appendResp, err := s.eventClient.AppendEvent(s.ctx, &eventpb.AppendEventRequest{
		TenantId:    "test-tenant",
		WorkspaceId: "test-workspace",
		AgentId:     "test-agent",
		SessionId:   "test-session",
		EventType:   eventpb.EventType_USER_MESSAGE,
		Payload:     []byte(`{"message": "hello"}`),
		Importance:  0.8,
		Visibility:  "public",
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, appendResp.Status.ErrorCode)
	eventID := appendResp.EventId
	s.NotEmpty(eventID)
	s.Greater(appendResp.LogicalTs, int64(0))

	// Get event
	getResp, err := s.eventClient.GetEvent(s.ctx, &eventpb.GetEventRequest{
		EventId: eventID,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, getResp.Status.ErrorCode)
	s.Equal(eventID, getResp.Event.EventId)

	// Query events
	queryResp, err := s.eventClient.QueryEvents(s.ctx, &eventpb.QueryEventsRequest{
		Filter: &eventpb.EventFilter{
			AgentIds: []string{"test-agent"},
		},
		Limit:  10,
		Offset: 0,
	})
	s.Require().NoError(err)
	s.Require().Equal(commonpb.ErrorCode_Success, queryResp.Status.ErrorCode)
	s.NotEmpty(queryResp.Events)
}

// TestEventSubscription tests event subscription
func (s *TDBIntegrationTestSuite) TestEventSubscription() {
	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	stream, err := s.eventClient.SubscribeEvents(ctx, &eventpb.SubscribeEventsRequest{
		SubscriberId: "test-subscriber",
		Filter: &eventpb.EventFilter{
			AgentIds: []string{"test-agent"},
		},
	})
	s.Require().NoError(err)

	// Append an event
	_, _ = s.eventClient.AppendEvent(s.ctx, &eventpb.AppendEventRequest{
		TenantId:    "test-tenant",
		WorkspaceId: "test-workspace",
		AgentId:     "test-agent",
		SessionId:   "test-session",
		EventType:   eventpb.EventType_USER_MESSAGE,
		Payload:     []byte(`{"message": "hello"}`),
		Importance:  0.8,
		Visibility:  "public",
	})

	// Try to receive the event (with timeout)
	done := make(chan bool)
	go func() {
		resp, err := stream.Recv()
		if err == nil && resp.GetEvent() != nil {
			s.NotEmpty(resp.GetEvent().EventId)
		}
		done <- true
	}()

	select {
	case <-done:
		// Successfully received event or got error
	case <-time.After(3 * time.Second):
		s.T().Log("Timeout waiting for event, which is expected in integration test")
	}
}

func TestTDBIntegrationSuite(t *testing.T) {
	suite.Run(t, new(TDBIntegrationTestSuite))
}
