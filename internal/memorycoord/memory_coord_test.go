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

package memorycoord

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

func (m *MockCatalog) CreateMemory(ctx context.Context, memory *models.Memory, ts typeutil.Timestamp) error {
	args := m.Called(ctx, memory, ts)
	return args.Error(0)
}

func (m *MockCatalog) GetMemory(ctx context.Context, memoryID string, ts typeutil.Timestamp) (*models.Memory, *models.MemoryPolicy, *models.MemoryAdaptation, error) {
	args := m.Called(ctx, memoryID, ts)
	if args.Get(0) == nil {
		return nil, nil, nil, args.Error(3)
	}
	return args.Get(0).(*models.Memory), args.Get(1).(*models.MemoryPolicy), args.Get(2).(*models.MemoryAdaptation), args.Error(3)
}

func (m *MockCatalog) UpdateMemory(ctx context.Context, memory *models.Memory, ts typeutil.Timestamp) error {
	args := m.Called(ctx, memory, ts)
	return args.Error(0)
}

func (m *MockCatalog) DeleteMemory(ctx context.Context, memoryID string, hardDelete bool, ts typeutil.Timestamp) error {
	args := m.Called(ctx, memoryID, hardDelete, ts)
	return args.Error(0)
}

func (m *MockCatalog) QueryMemories(ctx context.Context, filter *models.MemoryFilter, limit, offset int, ts typeutil.Timestamp) ([]*models.Memory, int64, error) {
	args := m.Called(ctx, filter, limit, offset, ts)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*models.Memory), args.Get(1).(int64), args.Error(2)
}

func (m *MockCatalog) SearchMemories(ctx context.Context, queryVector []float32, topK int, minScore float32, ts typeutil.Timestamp) ([]*models.MemorySearchResult, error) {
	args := m.Called(ctx, queryVector, topK, minScore, ts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MemorySearchResult), args.Error(1)
}

func (m *MockCatalog) GetRelations(ctx context.Context, objectID, objectType, relationType string, hop int, ts typeutil.Timestamp) ([]*models.Relation, error) {
	args := m.Called(ctx, objectID, objectType, relationType, hop, ts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Relation), args.Error(1)
}

func (m *MockCatalog) CreateRelation(ctx context.Context, relation *models.Relation, ts typeutil.Timestamp) error {
	args := m.Called(ctx, relation, ts)
	return args.Error(0)
}

// MemoryCoordTestSuite is the test suite for MemoryCoord
type MemoryCoordTestSuite struct {
	suite.Suite
	coord   *MemoryCoord
	catalog *MockCatalog
	ctx     context.Context
}

func (s *MemoryCoordTestSuite) SetupTest() {
	s.catalog = new(MockCatalog)
	s.coord = &MemoryCoord{
		catalog: s.catalog,
	}
	s.ctx = context.Background()
}

func (s *MemoryCoordTestSuite) TearDownTest() {
	s.catalog.AssertExpectations(s.T())
}

func (s *MemoryCoordTestSuite) TestCreateMemory() {
	s.catalog.On("CreateMemory", s.ctx, mock.AnythingOfType("*models.Memory"), mock.AnythingOfType("uint64")).
		Return(nil).Once()

	memory := &models.Memory{
		AgentID:    "agent-001",
		SessionID:  "session-001",
		MemoryType: models.MemoryTypeEpisodic,
		Scope:      "session",
		Content:    "test content",
		Summary:    "test summary",
		Confidence: 0.9,
		Importance: 0.8,
		TTL:        3600,
		Metadata:   map[string]string{"key": "value"},
	}

	result, err := s.coord.CreateMemory(s.ctx, memory)

	s.NoError(err)
	s.NotNil(result)
	s.Equal("agent-001", result.AgentID)
	s.Equal(models.MemoryTypeEpisodic, result.MemoryType)
	s.Equal(models.MemoryLevelRaw, result.Level)
	s.NotEmpty(result.MemoryID)
}

func (s *MemoryCoordTestSuite) TestGetMemory() {
	expectedMemory := &models.Memory{
		MemoryID:   "memory-001",
		AgentID:    "agent-001",
		MemoryType: models.MemoryTypeEpisodic,
		Content:    "test content",
	}
	expectedPolicy := &models.MemoryPolicy{
		MemoryID: "memory-001",
		TTL:      3600,
	}
	expectedAdaptation := &models.MemoryAdaptation{
		MemoryID: "memory-001",
	}

	s.catalog.On("GetMemory", s.ctx, "memory-001", mock.AnythingOfType("uint64")).
		Return(expectedMemory, expectedPolicy, expectedAdaptation, nil).Once()

	memory, policy, adaptation, err := s.coord.GetMemory(s.ctx, "memory-001")

	s.NoError(err)
	s.NotNil(memory)
	s.NotNil(policy)
	s.NotNil(adaptation)
	s.Equal("memory-001", memory.MemoryID)
}

func (s *MemoryCoordTestSuite) TestUpdateMemory() {
	existingMemory := &models.Memory{
		MemoryID:   "memory-001",
		AgentID:    "agent-001",
		Content:    "old content",
		Summary:    "old summary",
		Confidence: 0.8,
		Importance: 0.7,
		Version:    1,
	}
	expectedPolicy := &models.MemoryPolicy{MemoryID: "memory-001"}
	expectedAdaptation := &models.MemoryAdaptation{MemoryID: "memory-001"}

	s.catalog.On("GetMemory", s.ctx, "memory-001", mock.AnythingOfType("uint64")).
		Return(existingMemory, expectedPolicy, expectedAdaptation, nil).Once()
	s.catalog.On("UpdateMemory", s.ctx, mock.AnythingOfType("*models.Memory"), mock.AnythingOfType("uint64")).
		Return(nil).Once()

	updates := &models.Memory{
		Content:    "new content",
		Confidence: 0.95,
	}

	newVersion, err := s.coord.UpdateMemory(s.ctx, "memory-001", updates)

	s.NoError(err)
	s.Equal(int64(2), newVersion)
}

func (s *MemoryCoordTestSuite) TestDeleteMemory() {
	s.catalog.On("DeleteMemory", s.ctx, "memory-001", false, mock.AnythingOfType("uint64")).
		Return(nil).Once()

	err := s.coord.DeleteMemory(s.ctx, "memory-001", false)

	s.NoError(err)
}

func (s *MemoryCoordTestSuite) TestQueryMemories() {
	expectedMemories := []*models.Memory{
		{
			MemoryID:   "memory-001",
			AgentID:    "agent-001",
			MemoryType: models.MemoryTypeEpisodic,
		},
		{
			MemoryID:   "memory-002",
			AgentID:    "agent-001",
			MemoryType: models.MemoryTypeSemantic,
		},
	}

	filter := &models.MemoryFilter{
		AgentIDs: []string{"agent-001"},
	}

	s.catalog.On("QueryMemories", s.ctx, filter, 10, 0, mock.AnythingOfType("uint64")).
		Return(expectedMemories, int64(2), nil).Once()

	memories, totalCount, err := s.coord.QueryMemories(s.ctx, filter, 10, 0)

	s.NoError(err)
	s.Len(memories, 2)
	s.Equal(int64(2), totalCount)
}

func (s *MemoryCoordTestSuite) TestSearchMemories() {
	queryVector := []float32{0.1, 0.2, 0.3}
	expectedResults := []*models.MemorySearchResult{
		{
			Memory: &models.Memory{
				MemoryID:   "memory-001",
				AgentID:    "agent-001",
				MemoryType: models.MemoryTypeEpisodic,
			},
			Score:       0.95,
			Explanation: "high similarity",
		},
	}

	s.catalog.On("SearchMemories", s.ctx, queryVector, 10, float32(0.5), mock.AnythingOfType("uint64")).
		Return(expectedResults, nil).Once()

	results, err := s.coord.SearchMemories(s.ctx, queryVector, 10, 0.5)

	s.NoError(err)
	s.Len(results, 1)
	s.Equal(float32(0.95), results[0].Score)
}

func (s *MemoryCoordTestSuite) TestGetRelations() {
	expectedRelations := []*models.Relation{
		{
			EdgeID:       "edge-001",
			SrcObjectID:  "memory-001",
			DstObjectID:  "memory-002",
			RelationType: "similar_to",
			Weight:       0.8,
		},
	}

	s.catalog.On("GetRelations", s.ctx, "memory-001", "memory", "similar_to", 1, mock.AnythingOfType("uint64")).
		Return(expectedRelations, nil).Once()

	relations, err := s.coord.GetRelations(s.ctx, "memory-001", "memory", "similar_to", 1)

	s.NoError(err)
	s.Len(relations, 1)
	s.Equal("edge-001", relations[0].EdgeID)
}

func (s *MemoryCoordTestSuite) TestCreateRelation() {
	s.catalog.On("CreateRelation", s.ctx, mock.AnythingOfType("*models.Relation"), mock.AnythingOfType("uint64")).
		Return(nil).Once()

	relation := &models.Relation{
		SrcObjectID:  "memory-001",
		SrcType:      "memory",
		DstObjectID:  "memory-002",
		DstType:      "memory",
		RelationType: "similar_to",
		Weight:       0.8,
		Properties:   map[string]string{"reason": "semantic similarity"},
	}

	edgeID, err := s.coord.CreateRelation(s.ctx, relation)

	s.NoError(err)
	s.NotEmpty(edgeID)
}

func TestMemoryCoordSuite(t *testing.T) {
	suite.Run(t, new(MemoryCoordTestSuite))
}

// Benchmark tests

func BenchmarkCreateMemory(b *testing.B) {
	coord := &MemoryCoord{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = coord.generateMemoryID()
	}
}

func BenchmarkGenerateRelationID(b *testing.B) {
	coord := &MemoryCoord{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = coord.generateRelationID()
	}
}
