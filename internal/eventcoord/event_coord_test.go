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

package eventcoord

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

// MockCatalog is a mock implementation of the Catalog interface
type MockCatalog struct {
	mock.Mock
}

func (m *MockCatalog) AppendEvent(ctx context.Context, event *models.Event, ts typeutil.Timestamp) (string, int64, error) {
	args := m.Called(ctx, event, ts)
	return args.String(0), args.Get(1).(int64), args.Error(2)
}

func (m *MockCatalog) GetEvent(ctx context.Context, eventID string, ts typeutil.Timestamp) (*models.Event, error) {
	args := m.Called(ctx, eventID, ts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Event), args.Error(1)
}

func (m *MockCatalog) QueryEvents(ctx context.Context, filter *models.EventFilter, limit, offset int, ts typeutil.Timestamp) ([]*models.Event, int64, error) {
	args := m.Called(ctx, filter, limit, offset, ts)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Event), args.Get(1).(int64), args.Error(2)
}

func (m *MockCatalog) GetNextLogID(ctx context.Context, channelName string, ts typeutil.Timestamp) (int64, error) {
	args := m.Called(ctx, channelName, ts)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCatalog) CreateSubscriber(ctx context.Context, subscriber *models.EventSubscriber, ts typeutil.Timestamp) error {
	args := m.Called(ctx, subscriber, ts)
	return args.Error(0)
}

func (m *MockCatalog) GetSubscriber(ctx context.Context, subscriberID string, ts typeutil.Timestamp) (*models.EventSubscriber, error) {
	args := m.Called(ctx, subscriberID, ts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EventSubscriber), args.Error(1)
}

func (m *MockCatalog) UpdateSubscriberPosition(ctx context.Context, subscriberID string, position *models.EventPosition, ts typeutil.Timestamp) error {
	args := m.Called(ctx, subscriberID, position, ts)
	return args.Error(0)
}

func (m *MockCatalog) DeleteSubscriber(ctx context.Context, subscriberID string, ts typeutil.Timestamp) error {
	args := m.Called(ctx, subscriberID, ts)
	return args.Error(0)
}

// EventCoordTestSuite is the test suite for EventCoord
type EventCoordTestSuite struct {
	suite.Suite
	coord   *EventCoord
	catalog *MockCatalog
	ctx     context.Context
}

func (s *EventCoordTestSuite) SetupTest() {
	s.catalog = new(MockCatalog)
	s.coord = &EventCoord{
		catalog:       s.catalog,
		subscriptions: make(map[string]*EventSubscription),
	}
	s.ctx = context.Background()
}

func (s *EventCoordTestSuite) TearDownTest() {
	s.catalog.AssertExpectations(s.T())
}

func (s *EventCoordTestSuite) TestAppendEvent() {
	s.catalog.On("AppendEvent", s.ctx, mock.AnythingOfType("*models.Event"), mock.AnythingOfType("uint64")).
		Return("event-001", int64(1000), nil).Once()

	event := &models.Event{
		TenantID:    "tenant-001",
		WorkspaceID: "workspace-001",
		AgentID:     "agent-001",
		SessionID:   "session-001",
		EventType:   models.EventTypeUserMessage,
		Payload:     []byte(`{"message": "hello"}`),
		Importance:  0.8,
		Visibility:  "public",
	}

	eventID, logicalTs, err := s.coord.AppendEvent(s.ctx, event)

	s.NoError(err)
	s.Equal("event-001", eventID)
	s.Equal(int64(1000), logicalTs)
}

func (s *EventCoordTestSuite) TestGetEvent() {
	expectedEvent := &models.Event{
		EventID:     "event-001",
		TenantID:    "tenant-001",
		WorkspaceID: "workspace-001",
		AgentID:     "agent-001",
		EventType:   models.EventTypeUserMessage,
		Payload:     []byte(`{"message": "hello"}`),
		LogicalTs:   1000,
	}

	s.catalog.On("GetEvent", s.ctx, "event-001", mock.AnythingOfType("uint64")).
		Return(expectedEvent, nil).Once()

	event, err := s.coord.GetEvent(s.ctx, "event-001")

	s.NoError(err)
	s.NotNil(event)
	s.Equal("event-001", event.EventID)
	s.Equal("agent-001", event.AgentID)
	s.Equal(models.EventTypeUserMessage, event.EventType)
}

func (s *EventCoordTestSuite) TestQueryEvents() {
	expectedEvents := []*models.Event{
		{
			EventID:   "event-001",
			AgentID:   "agent-001",
			EventType: models.EventTypeUserMessage,
			LogicalTs: 1000,
		},
		{
			EventID:   "event-002",
			AgentID:   "agent-001",
			EventType: models.EventTypeAssistantMessage,
			LogicalTs: 1001,
		},
	}

	filter := &models.EventFilter{
		AgentIDs: []string{"agent-001"},
	}

	s.catalog.On("QueryEvents", s.ctx, filter, 10, 0, mock.AnythingOfType("uint64")).
		Return(expectedEvents, int64(2), nil).Once()

	events, totalCount, err := s.coord.QueryEvents(s.ctx, filter, 10, 0)

	s.NoError(err)
	s.Len(events, 2)
	s.Equal(int64(2), totalCount)
}

func TestEventCoordSuite(t *testing.T) {
	suite.Run(t, new(EventCoordTestSuite))
}

// Benchmark tests

func BenchmarkAppendEvent(b *testing.B) {
	coord := &EventCoord{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = coord.generateEventID()
	}
}
