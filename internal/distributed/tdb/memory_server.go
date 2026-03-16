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

	"github.com/milvus-io/milvus/internal/memorycoord"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/proto/commonpb"
	"github.com/milvus-io/milvus/pkg/v2/proto/memorypb"
)

// MemoryServer implements the MemoryService gRPC interface.
type MemoryServer struct {
	memorypb.UnimplementedMemoryServiceServer
	coord *memorycoord.MemoryCoord
}

// NewMemoryServer creates a new MemoryServer instance.
func NewMemoryServer(coord *memorycoord.MemoryCoord) *MemoryServer {
	return &MemoryServer{
		coord: coord,
	}
}

// CreateMemory creates a new memory.
func (s *MemoryServer) CreateMemory(ctx context.Context, req *memorypb.CreateMemoryRequest) (*memorypb.CreateMemoryResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "CreateMemory"))

	memory, err := s.coord.CreateMemory(ctx, &models.Memory{
		AgentID:     req.GetAgentId(),
		SessionID:   req.GetSessionId(),
		MemoryType:  models.MemoryType(req.GetMemoryType()),
		Scope:       req.GetScope(),
		Content:     req.GetContent(),
		Summary:     req.GetSummary(),
		Confidence:  req.GetConfidence(),
		Importance:  req.GetImportance(),
		TTL:         req.GetTtl(),
		Metadata:    req.GetMetadata(),
	})
	if err != nil {
		log.Error("failed to create memory", zap.Error(err))
		return &memorypb.CreateMemoryResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &memorypb.CreateMemoryResponse{
		Status: &commonpb.Status{
			ErrorCode: commonpb.ErrorCode_Success,
		},
		MemoryId: memory.MemoryID,
		Version:  memory.Version,
	}, nil
}

// GetMemory gets a memory by ID.
func (s *MemoryServer) GetMemory(ctx context.Context, req *memorypb.GetMemoryRequest) (*memorypb.GetMemoryResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "GetMemory"))

	memory, policy, adaptation, err := s.coord.GetMemory(ctx, req.GetMemoryId())
	if err != nil {
		log.Error("failed to get memory", zap.Error(err))
		return &memorypb.GetMemoryResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &memorypb.GetMemoryResponse{
		Status:     &commonpb.Status{ErrorCode: commonpb.ErrorCode_Success},
		Memory:     convertMemoryToProto(memory),
		Policy:     convertMemoryPolicyToProto(policy),
		Adaptation: convertMemoryAdaptationToProto(adaptation),
	}, nil
}

// UpdateMemory updates an existing memory.
func (s *MemoryServer) UpdateMemory(ctx context.Context, req *memorypb.UpdateMemoryRequest) (*memorypb.UpdateMemoryResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "UpdateMemory"))

	updates := &models.Memory{
		Content:    req.GetContent(),
		Summary:    req.GetSummary(),
		Confidence: req.GetConfidence(),
		Importance: req.GetImportance(),
		TTL:        req.GetTtl(),
		Metadata:   req.GetMetadata(),
	}

	newVersion, err := s.coord.UpdateMemory(ctx, req.GetMemoryId(), updates)
	if err != nil {
		log.Error("failed to update memory", zap.Error(err))
		return &memorypb.UpdateMemoryResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &memorypb.UpdateMemoryResponse{
		Status:     &commonpb.Status{ErrorCode: commonpb.ErrorCode_Success},
		NewVersion: newVersion,
	}, nil
}

// DeleteMemory deletes a memory.
func (s *MemoryServer) DeleteMemory(ctx context.Context, req *memorypb.DeleteMemoryRequest) (*memorypb.DeleteMemoryResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "DeleteMemory"))

	err := s.coord.DeleteMemory(ctx, req.GetMemoryId(), req.GetHardDelete())
	if err != nil {
		log.Error("failed to delete memory", zap.Error(err))
		return &memorypb.DeleteMemoryResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &memorypb.DeleteMemoryResponse{
		Status: &commonpb.Status{ErrorCode: commonpb.ErrorCode_Success},
	}, nil
}

// QueryMemories queries memories with filter.
func (s *MemoryServer) QueryMemories(ctx context.Context, req *memorypb.QueryMemoriesRequest) (*memorypb.QueryMemoriesResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "QueryMemories"))

	filter := &models.MemoryFilter{
		AgentIDs:   req.GetFilter().GetAgentIds(),
		SessionIDs: req.GetFilter().GetSessionIds(),
		Scope:      req.GetFilter().GetScope(),
		StartTime:  req.GetFilter().GetStartTime(),
		EndTime:    req.GetFilter().GetEndTime(),
	}

	memories, totalCount, err := s.coord.QueryMemories(ctx, filter, req.GetLimit(), req.GetOffset())
	if err != nil {
		log.Error("failed to query memories", zap.Error(err))
		return &memorypb.QueryMemoriesResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	memoryList := make([]*memorypb.Memory, 0, len(memories))
	for _, m := range memories {
		memoryList = append(memoryList, convertMemoryToProto(m))
	}

	return &memorypb.QueryMemoriesResponse{
		Status:     &commonpb.Status{ErrorCode: commonpb.ErrorCode_Success},
		Memories:   memoryList,
		TotalCount: totalCount,
	}, nil
}

// SearchMemories searches memories by vector similarity.
func (s *MemoryServer) SearchMemories(ctx context.Context, req *memorypb.SearchMemoriesRequest) (*memorypb.SearchMemoriesResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "SearchMemories"))

	results, err := s.coord.SearchMemories(ctx, req.GetQueryVector(), int(req.GetTopK()), req.GetMinScore())
	if err != nil {
		log.Error("failed to search memories", zap.Error(err))
		return &memorypb.SearchMemoriesResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	searchResults := make([]*memorypb.SearchResult, 0, len(results))
	for _, r := range results {
		searchResults = append(searchResults, &memorypb.SearchResult{
			Memory:      convertMemoryToProto(r.Memory),
			Score:       r.Score,
			Explanation: r.Explanation,
		})
	}

	return &memorypb.SearchMemoriesResponse{
		Status:  &commonpb.Status{ErrorCode: commonpb.ErrorCode_Success},
		Results: searchResults,
	}, nil
}

// GetRelations gets relations for an object.
func (s *MemoryServer) GetRelations(ctx context.Context, req *memorypb.GetRelationsRequest) (*memorypb.GetRelationsResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "GetRelations"))

	relations, err := s.coord.GetRelations(ctx, req.GetObjectId(), req.GetObjectType(), req.GetRelationType(), req.GetHop())
	if err != nil {
		log.Error("failed to get relations", zap.Error(err))
		return &memorypb.GetRelationsResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	relationList := make([]*memorypb.Relation, 0, len(relations))
	for _, r := range relations {
		relationList = append(relationList, convertRelationToProto(r))
	}

	return &memorypb.GetRelationsResponse{
		Status:    &commonpb.Status{ErrorCode: commonpb.ErrorCode_Success},
		Relations: relationList,
	}, nil
}

// CreateRelation creates a new relation.
func (s *MemoryServer) CreateRelation(ctx context.Context, req *memorypb.CreateRelationRequest) (*memorypb.CreateRelationResponse, error) {
	log := log.Ctx(ctx).With(zap.String("method", "CreateRelation"))

	relation := &models.Relation{
		SrcObjectID: req.GetSrcObjectId(),
		SrcType:     req.GetSrcType(),
		DstObjectID: req.GetDstObjectId(),
		DstType:     req.GetDstType(),
		RelationType: req.GetRelationType(),
		Weight:      req.GetWeight(),
		Properties:  req.GetProperties(),
	}

	edgeID, err := s.coord.CreateRelation(ctx, relation)
	if err != nil {
		log.Error("failed to create relation", zap.Error(err))
		return &memorypb.CreateRelationResponse{
			Status: &commonpb.Status{
				ErrorCode: commonpb.ErrorCode_UnexpectedError,
				Reason:    err.Error(),
			},
		}, nil
	}

	return &memorypb.CreateRelationResponse{
		Status: &commonpb.Status{ErrorCode: commonpb.ErrorCode_Success},
		EdgeId: edgeID,
	}, nil
}

// Helper functions to convert models to protobuf messages

func convertMemoryToProto(m *models.Memory) *memorypb.Memory {
	if m == nil {
		return nil
	}

	memoryType := memorypb.MemoryType_MEMORY_UNKNOWN
	switch m.MemoryType {
	case models.MemoryTypeEpisodic:
		memoryType = memorypb.MemoryType_EPISODIC
	case models.MemoryTypeSemantic:
		memoryType = memorypb.MemoryType_SEMANTIC
	case models.MemoryTypeProcedural:
		memoryType = memorypb.MemoryType_PROCEDURAL
	case models.MemoryTypeSocial:
		memoryType = memorypb.MemoryType_SOCIAL
	case models.MemoryTypeReflective:
		memoryType = memorypb.MemoryType_REFLECTIVE
	}

	level := memorypb.MemoryLevel_LEVEL_UNKNOWN
	switch m.Level {
	case models.MemoryLevelRaw:
		level = memorypb.MemoryLevel_LEVEL_RAW
	case models.MemoryLevelSummary:
		level = memorypb.MemoryLevel_LEVEL_SUMMARY
	case models.MemoryLevelPattern:
		level = memorypb.MemoryLevel_LEVEL_PATTERN
	}

	state := memorypb.MemoryState_MEMORY_STATE_UNKNOWN
	switch m.State {
	case models.MemoryStateActive:
		state = memorypb.MemoryState_MEMORY_ACTIVE
	case models.MemoryStateFading:
		state = memorypb.MemoryState_MEMORY_FADING
	case models.MemoryStateArchived:
		state = memorypb.MemoryState_MEMORY_ARCHIVED
	case models.MemoryStateQuarantined:
		state = memorypb.MemoryState_MEMORY_QUARANTINED
	case models.MemoryStateDeleted:
		state = memorypb.MemoryState_MEMORY_DELETED
	}

	return &memorypb.Memory{
		MemoryId:    m.MemoryID,
		MemoryType:  memoryType,
		AgentId:     m.AgentID,
		SessionId:   m.SessionID,
		Scope:       m.Scope,
		Level:       level,
		Content:     m.Content,
		Summary:     m.Summary,
		Confidence:  m.Confidence,
		Importance:  m.Importance,
		Ttl:         m.TTL,
		Version:     m.Version,
		IsActive:    m.IsActive,
		State:       state,
		CreatedTs:   m.CreatedAt,
		UpdatedTs:   m.UpdatedAt,
		Metadata:    m.Metadata,
	}
}

func convertMemoryPolicyToProto(p *models.MemoryPolicy) *memorypb.MemoryPolicy {
	if p == nil {
		return nil
	}
	return &memorypb.MemoryPolicy{
		MemoryId:          p.MemoryID,
		SalienceWeight:    p.SalienceWeight,
		Ttl:               p.TTL,
		DecayFn:           p.DecayFn,
		Confidence:        p.Confidence,
		Verified:          p.Verified,
		VerifiedBy:        p.VerifiedBy,
		VerifiedAt:        p.VerifiedAt,
		Quarantined:       p.Quarantined,
		QuarantineReason:  p.QuarantineReason,
		VisibilityPolicy:  p.VisibilityPolicy,
		ReadAcl:           p.ReadACL,
		WriteAcl:          p.WriteACL,
		DeriveAcl:         p.DeriveACL,
		PolicyReason:      p.PolicyReason,
		PolicySource:      p.PolicySource,
		PolicyEventId:     p.PolicyEventID,
	}
}

func convertMemoryAdaptationToProto(a *models.MemoryAdaptation) *memorypb.MemoryAdaptation {
	if a == nil {
		return nil
	}
	return &memorypb.MemoryAdaptation{
		MemoryId:            a.MemoryID,
		RetrievalProfile:    a.RetrievalProfile,
		RankingParams:       a.RankingParams,
		FilteringThresholds: a.FilteringThresholds,
		ProjectionWeights:   a.ProjectionWeights,
		EmbeddingFamily:     a.EmbeddingFamily,
		ModelId:             a.ModelID,
		AdaptationReason:    a.AdaptationReason,
		AdaptationSource:    a.AdaptationSource,
	}
}

func convertRelationToProto(r *models.Relation) *memorypb.Relation {
	if r == nil {
		return nil
	}
	return &memorypb.Relation{
		EdgeId:         r.EdgeID,
		SrcObjectId:    r.SrcObjectID,
		SrcType:        r.SrcType,
		DstObjectId:    r.DstObjectID,
		DstType:        r.DstType,
		RelationType:   r.RelationType,
		Weight:         r.Weight,
		Properties:     r.Properties,
		CreatedTs:      r.CreatedAt,
		CreatedByEventId: r.CreatedByEventID,
	}
}
