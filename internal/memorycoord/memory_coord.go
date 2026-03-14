package memorycoord

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/milvus-io/milvus/internal/metastore/kv/memorycoord"
	"github.com/milvus-io/milvus/internal/util/sessionutil"
	"github.com/milvus-io/milvus/pkg/v2/kv"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

// StateCode represents the state of the MemoryCoord.
type StateCode int32

const (
	StateCode_Initializing StateCode = 0
	StateCode_Healthy      StateCode = 1
	StateCode_Abnormal     StateCode = 2
	StateCode_Stopping     StateCode = 3
)

// MemoryCoord manages memory, state, artifact, and relations.
type MemoryCoord struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Dependencies
	catalog  *memorycoord.Catalog
	session  sessionutil.SessionInterface
	tsoAllocator typeutil.TimestampAllocator

	// State
	stateCode atomic.Int32
	initOnce  sync.Once
	startOnce sync.Once
	stopOnce  sync.Once

	// Background tasks
	ticker *time.Ticker
}

// NewMemoryCoord creates a new MemoryCoord instance.
func NewMemoryCoord(ctx context.Context, metaKV kv.TxnKV, snapshotKV kv.SnapShotKV, session sessionutil.SessionInterface, tsoAllocator typeutil.TimestampAllocator) (*MemoryCoord, error) {
	ctx, cancel := context.WithCancel(ctx)

	catalog := memorycoord.NewCatalog(metaKV, snapshotKV)

	mc := &MemoryCoord{
		ctx:          ctx,
		cancel:       cancel,
		catalog:      catalog,
		session:      session,
		tsoAllocator: tsoAllocator,
		ticker:       time.NewTicker(5 * time.Second),
	}

	mc.UpdateStateCode(StateCode_Initializing)
	return mc, nil
}

// UpdateStateCode updates the state code.
func (mc *MemoryCoord) UpdateStateCode(code StateCode) {
	mc.stateCode.Store(int32(code))
	log.Ctx(mc.ctx).Info("update memorycoord state", zap.String("state", code.String()))
}

// GetStateCode returns the current state code.
func (mc *MemoryCoord) GetStateCode() StateCode {
	return StateCode(mc.stateCode.Load())
}

// String returns the string representation of StateCode.
func (s StateCode) String() string {
	switch s {
	case StateCode_Initializing:
		return "Initializing"
	case StateCode_Healthy:
		return "Healthy"
	case StateCode_Abnormal:
		return "Abnormal"
	case StateCode_Stopping:
		return "Stopping"
	default:
		return "Unknown"
	}
}

// Init initializes the MemoryCoord.
func (mc *MemoryCoord) Init() error {
	var err error
	mc.initOnce.Do(func() {
		err = mc.init()
	})
	return err
}

func (mc *MemoryCoord) init() error {
	log.Ctx(mc.ctx).Info("MemoryCoord initializing")

	// TODO: Recover state from catalog if needed
	// TODO: Initialize memory index

	mc.UpdateStateCode(StateCode_Healthy)
	log.Ctx(mc.ctx).Info("MemoryCoord initialized")
	return nil
}

// Start starts the MemoryCoord background tasks.
func (mc *MemoryCoord) Start() error {
	var err error
	mc.startOnce.Do(func() {
		err = mc.start()
	})
	return err
}

func (mc *MemoryCoord) start() error {
	log.Ctx(mc.ctx).Info("MemoryCoord starting")

	mc.wg.Add(1)
	go mc.backgroundTask()

	log.Ctx(mc.ctx).Info("MemoryCoord started")
	return nil
}

// Stop stops the MemoryCoord.
func (mc *MemoryCoord) Stop() error {
	var err error
	mc.stopOnce.Do(func() {
		err = mc.stop()
	})
	return err
}

func (mc *MemoryCoord) stop() error {
	log.Ctx(mc.ctx).Info("MemoryCoord stopping")
	mc.UpdateStateCode(StateCode_Stopping)

	mc.ticker.Stop()
	mc.cancel()
	mc.wg.Wait()

	log.Ctx(mc.ctx).Info("MemoryCoord stopped")
	return nil
}

func (mc *MemoryCoord) backgroundTask() {
	defer mc.wg.Done()

	for {
		select {
		case <-mc.ctx.Done():
			return
		case <-mc.ticker.C:
			mc.doBackgroundTasks()
		}
	}
}

func (mc *MemoryCoord) doBackgroundTasks() {
	// TODO: Implement background tasks
	// - Memory decay and cleanup
	// - TTL expiration check
	// - Index maintenance
	// - Quarantine check
}

// getTimestamp returns the current timestamp.
func (mc *MemoryCoord) getTimestamp() (typeutil.Timestamp, error) {
	if mc.tsoAllocator != nil {
		return mc.tsoAllocator.AllocOne(mc.ctx)
	}
	return typeutil.Timestamp(time.Now().UnixNano()), nil
}

// ==================== Memory Management ====================

// CreateMemory creates a new memory.
func (mc *MemoryCoord) CreateMemory(ctx context.Context, agentID, sessionID string, memoryType models.MemoryType, scope, content, summary string, sourceEventIDs []string, confidence, importance float32, ttl int64, metadata map[string]string, embeddingVector []float32) (*models.Memory, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	memoryID := generateMemoryID()
	memory := &models.Memory{
		MemoryID:        memoryID,
		MemoryType:      memoryType,
		AgentID:         agentID,
		SessionID:       sessionID,
		Scope:           scope,
		Level:           models.LevelRaw,
		Content:         content,
		Summary:         summary,
		SourceEventIDs:  sourceEventIDs,
		Confidence:      confidence,
		Importance:      importance,
		FreshnessScore:  1.0, // Initial freshness
		TTL:             ttl,
		ValidFrom:       uint64(ts),
		Version:         1,
		IsActive:        true,
		State:           models.MemoryActive,
		CreatedTs:       uint64(ts),
		UpdatedTs:       uint64(ts),
		Metadata:        metadata,
		EmbeddingVector: embeddingVector,
	}

	if err := mc.catalog.CreateMemory(ctx, memory, ts); err != nil {
		return nil, errors.Wrap(err, "failed to create memory")
	}

	// Create default policy
	policy := &models.MemoryPolicy{
		MemoryID:         memoryID,
		SalienceWeight:   importance,
		TTL:              ttl,
		DecayFn:          "exponential",
		Confidence:       confidence,
		Verified:         false,
		Quarantined:      false,
		VisibilityPolicy: "private",
		CreatedTs:        uint64(ts),
		UpdatedTs:        uint64(ts),
	}
	if err := mc.catalog.CreateMemoryPolicy(ctx, policy, ts); err != nil {
		return nil, errors.Wrap(err, "failed to create memory policy")
	}

	log.Ctx(ctx).Info("memory created", zap.String("memoryID", memoryID), zap.String("agentID", agentID))
	return memory, nil
}

// GetMemory retrieves a memory by ID.
func (mc *MemoryCoord) GetMemory(ctx context.Context, memoryID string) (*models.Memory, *models.MemoryPolicy, *models.MemoryAdaptation, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, nil, nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get timestamp")
	}

	memory, err := mc.catalog.GetMemory(ctx, memoryID, ts)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to get memory")
	}

	policy, _ := mc.catalog.GetMemoryPolicy(ctx, memoryID, ts)
	adaptation, _ := mc.catalog.GetMemoryAdaptation(ctx, memoryID, ts)

	return memory, policy, adaptation, nil
}

// ListMemories lists all memories for an agent.
func (mc *MemoryCoord) ListMemories(ctx context.Context, agentID, sessionID string) ([]*models.Memory, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	memories, err := mc.catalog.ListMemories(ctx, agentID, sessionID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list memories")
	}

	return memories, nil
}

// ListMemoriesByType lists memories filtered by type.
func (mc *MemoryCoord) ListMemoriesByType(ctx context.Context, agentID string, memoryType models.MemoryType) ([]*models.Memory, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	memories, err := mc.catalog.ListMemoriesByType(ctx, agentID, memoryType, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list memories by type")
	}

	return memories, nil
}

// UpdateMemory updates an existing memory.
func (mc *MemoryCoord) UpdateMemory(ctx context.Context, memoryID string, updates map[string]interface{}) (*models.Memory, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	memory, err := mc.catalog.GetMemory(ctx, memoryID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get memory")
	}

	// Apply updates
	if content, ok := updates["content"].(string); ok {
		memory.Content = content
	}
	if summary, ok := updates["summary"].(string); ok {
		memory.Summary = summary
	}
	if confidence, ok := updates["confidence"].(float32); ok {
		memory.Confidence = confidence
	}
	if importance, ok := updates["importance"].(float32); ok {
		memory.Importance = importance
	}
	if ttl, ok := updates["ttl"].(int64); ok {
		memory.TTL = ttl
	}
	if metadata, ok := updates["metadata"].(map[string]string); ok {
		memory.Metadata = metadata
	}
	if embeddingVector, ok := updates["embedding_vector"].([]float32); ok {
		memory.EmbeddingVector = embeddingVector
	}

	memory.Version++
	memory.UpdatedTs = uint64(ts)

	if err := mc.catalog.UpdateMemory(ctx, memory, ts); err != nil {
		return nil, errors.Wrap(err, "failed to update memory")
	}

	log.Ctx(ctx).Info("memory updated", zap.String("memoryID", memoryID), zap.Int64("version", memory.Version))
	return memory, nil
}

// DeleteMemory deletes a memory.
func (mc *MemoryCoord) DeleteMemory(ctx context.Context, memoryID string, hardDelete bool) error {
	if mc.GetStateCode() != StateCode_Healthy {
		return errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	if err := mc.catalog.DeleteMemory(ctx, memoryID, hardDelete, ts); err != nil {
		return errors.Wrap(err, "failed to delete memory")
	}

	if hardDelete {
		// Also delete policy and adaptation
		mc.catalog.DeleteMemoryPolicy(ctx, memoryID, ts)
		mc.catalog.DeleteMemoryAdaptation(ctx, memoryID, ts)
	}

	log.Ctx(ctx).Info("memory deleted", zap.String("memoryID", memoryID), zap.Bool("hardDelete", hardDelete))
	return nil
}

// ArchiveMemory archives a memory.
func (mc *MemoryCoord) ArchiveMemory(ctx context.Context, memoryID string) error {
	if mc.GetStateCode() != StateCode_Healthy {
		return errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	memory, err := mc.catalog.GetMemory(ctx, memoryID, ts)
	if err != nil {
		return errors.Wrap(err, "failed to get memory")
	}

	memory.State = models.MemoryArchived
	memory.IsActive = false
	memory.UpdatedTs = uint64(ts)

	if err := mc.catalog.UpdateMemory(ctx, memory, ts); err != nil {
		return errors.Wrap(err, "failed to archive memory")
	}

	log.Ctx(ctx).Info("memory archived", zap.String("memoryID", memoryID))
	return nil
}

// QuarantineMemory quarantines a memory.
func (mc *MemoryCoord) QuarantineMemory(ctx context.Context, memoryID, reason string) error {
	if mc.GetStateCode() != StateCode_Healthy {
		return errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	memory, err := mc.catalog.GetMemory(ctx, memoryID, ts)
	if err != nil {
		return errors.Wrap(err, "failed to get memory")
	}

	memory.State = models.MemoryQuarantined
	memory.UpdatedTs = uint64(ts)

	if err := mc.catalog.UpdateMemory(ctx, memory, ts); err != nil {
		return errors.Wrap(err, "failed to quarantine memory")
	}

	// Update policy
	policy, _ := mc.catalog.GetMemoryPolicy(ctx, memoryID, ts)
	if policy != nil {
		policy.Quarantined = true
		policy.QuarantineReason = reason
		policy.UpdatedTs = uint64(ts)
		mc.catalog.UpdateMemoryPolicy(ctx, policy, ts)
	}

	log.Ctx(ctx).Info("memory quarantined", zap.String("memoryID", memoryID), zap.String("reason", reason))
	return nil
}

// UpdateMemoryPolicy updates memory policy.
func (mc *MemoryCoord) UpdateMemoryPolicy(ctx context.Context, memoryID string, updates map[string]interface{}) (*models.MemoryPolicy, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	policy, err := mc.catalog.GetMemoryPolicy(ctx, memoryID, ts)
	if err != nil {
		// Create new policy if not exists
		policy = &models.MemoryPolicy{
			MemoryID: memoryID,
		}
	}

	// Apply updates
	if salienceWeight, ok := updates["salience_weight"].(float32); ok {
		policy.SalienceWeight = salienceWeight
	}
	if ttl, ok := updates["ttl"].(int64); ok {
		policy.TTL = ttl
	}
	if decayFn, ok := updates["decay_fn"].(string); ok {
		policy.DecayFn = decayFn
	}
	if verified, ok := updates["verified"].(bool); ok {
		policy.Verified = verified
	}
	if visibilityPolicy, ok := updates["visibility_policy"].(string); ok {
		policy.VisibilityPolicy = visibilityPolicy
	}
	if readACL, ok := updates["read_acl"].([]string); ok {
		policy.ReadACL = readACL
	}
	if writeACL, ok := updates["write_acl"].([]string); ok {
		policy.WriteACL = writeACL
	}

	policy.UpdatedTs = uint64(ts)

	if err := mc.catalog.UpdateMemoryPolicy(ctx, policy, ts); err != nil {
		return nil, errors.Wrap(err, "failed to update memory policy")
	}

	log.Ctx(ctx).Info("memory policy updated", zap.String("memoryID", memoryID))
	return policy, nil
}

// UpdateMemoryAdaptation updates memory adaptation.
func (mc *MemoryCoord) UpdateMemoryAdaptation(ctx context.Context, memoryID string, updates map[string]interface{}) (*models.MemoryAdaptation, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	adaptation, err := mc.catalog.GetMemoryAdaptation(ctx, memoryID, ts)
	if err != nil {
		// Create new adaptation if not exists
		adaptation = &models.MemoryAdaptation{
			MemoryID: memoryID,
		}
	}

	// Apply updates
	if retrievalProfile, ok := updates["retrieval_profile"].(map[string]float32); ok {
		adaptation.RetrievalProfile = retrievalProfile
	}
	if rankingParams, ok := updates["ranking_params"].(map[string]float32); ok {
		adaptation.RankingParams = rankingParams
	}
	if embeddingFamily, ok := updates["embedding_family"].(string); ok {
		adaptation.EmbeddingFamily = embeddingFamily
	}
	if modelID, ok := updates["model_id"].(string); ok {
		adaptation.ModelID = modelID
	}

	adaptation.UpdatedTs = uint64(ts)

	if err := mc.catalog.UpdateMemoryAdaptation(ctx, adaptation, ts); err != nil {
		return nil, errors.Wrap(err, "failed to update memory adaptation")
	}

	log.Ctx(ctx).Info("memory adaptation updated", zap.String("memoryID", memoryID))
	return adaptation, nil
}

// ==================== State Management ====================

// CreateState creates a new state.
func (mc *MemoryCoord) CreateState(ctx context.Context, agentID, sessionID, stateType, stateKey string, stateValue []byte, derivedFromEventID string, metadata map[string]string) (*models.State, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	stateID := generateStateID()
	state := &models.State{
		StateID:            stateID,
		AgentID:            agentID,
		SessionID:          sessionID,
		StateType:          stateType,
		StateKey:           stateKey,
		StateValue:         stateValue,
		Version:            1,
		DerivedFromEventID: derivedFromEventID,
		CheckpointTs:       uint64(ts),
		CreatedTs:          uint64(ts),
		UpdatedTs:          uint64(ts),
		Metadata:           metadata,
	}

	if err := mc.catalog.CreateState(ctx, state, ts); err != nil {
		return nil, errors.Wrap(err, "failed to create state")
	}

	log.Ctx(ctx).Info("state created", zap.String("stateID", stateID), zap.String("agentID", agentID))
	return state, nil
}

// GetState retrieves a state by ID.
func (mc *MemoryCoord) GetState(ctx context.Context, stateID string) (*models.State, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	state, err := mc.catalog.GetState(ctx, stateID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get state")
	}

	return state, nil
}

// ListStates lists all states for an agent or session.
func (mc *MemoryCoord) ListStates(ctx context.Context, agentID, sessionID string) ([]*models.State, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	states, err := mc.catalog.ListStates(ctx, agentID, sessionID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list states")
	}

	return states, nil
}

// UpdateState updates an existing state.
func (mc *MemoryCoord) UpdateState(ctx context.Context, stateID string, stateValue []byte) (*models.State, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	state, err := mc.catalog.GetState(ctx, stateID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get state")
	}

	state.StateValue = stateValue
	state.Version++
	state.UpdatedTs = uint64(ts)

	if err := mc.catalog.UpdateState(ctx, state, ts); err != nil {
		return nil, errors.Wrap(err, "failed to update state")
	}

	log.Ctx(ctx).Info("state updated", zap.String("stateID", stateID), zap.Int64("version", state.Version))
	return state, nil
}

// DeleteState deletes a state.
func (mc *MemoryCoord) DeleteState(ctx context.Context, stateID string) error {
	if mc.GetStateCode() != StateCode_Healthy {
		return errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	if err := mc.catalog.DeleteState(ctx, stateID, ts); err != nil {
		return errors.Wrap(err, "failed to delete state")
	}

	log.Ctx(ctx).Info("state deleted", zap.String("stateID", stateID))
	return nil
}

// ==================== Artifact Management ====================

// CreateArtifact creates a new artifact.
func (mc *MemoryCoord) CreateArtifact(ctx context.Context, sessionID, ownerAgentID, artifactType, uri, contentRef, mimeType string, metadata map[string]string, hash string) (*models.Artifact, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	artifactID := generateArtifactID()
	artifact := &models.Artifact{
		ArtifactID:   artifactID,
		SessionID:    sessionID,
		OwnerAgentID: ownerAgentID,
		ArtifactType: artifactType,
		URI:          uri,
		ContentRef:   contentRef,
		MimeType:     mimeType,
		Metadata:     metadata,
		Hash:         hash,
		Version:      1,
		CreatedTs:    uint64(ts),
	}

	if err := mc.catalog.CreateArtifact(ctx, artifact, ts); err != nil {
		return nil, errors.Wrap(err, "failed to create artifact")
	}

	log.Ctx(ctx).Info("artifact created", zap.String("artifactID", artifactID), zap.String("sessionID", sessionID))
	return artifact, nil
}

// GetArtifact retrieves an artifact by ID.
func (mc *MemoryCoord) GetArtifact(ctx context.Context, artifactID string) (*models.Artifact, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	artifact, err := mc.catalog.GetArtifact(ctx, artifactID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get artifact")
	}

	return artifact, nil
}

// ListArtifacts lists all artifacts for a session.
func (mc *MemoryCoord) ListArtifacts(ctx context.Context, sessionID string) ([]*models.Artifact, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	artifacts, err := mc.catalog.ListArtifacts(ctx, sessionID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list artifacts")
	}

	return artifacts, nil
}

// DeleteArtifact deletes an artifact.
func (mc *MemoryCoord) DeleteArtifact(ctx context.Context, artifactID string) error {
	if mc.GetStateCode() != StateCode_Healthy {
		return errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	if err := mc.catalog.DeleteArtifact(ctx, artifactID, ts); err != nil {
		return errors.Wrap(err, "failed to delete artifact")
	}

	log.Ctx(ctx).Info("artifact deleted", zap.String("artifactID", artifactID))
	return nil
}

// ==================== Relation Management ====================

// CreateRelation creates a new relation between objects.
func (mc *MemoryCoord) CreateRelation(ctx context.Context, srcObjectID, srcType, dstObjectID, dstType, relationType string, weight float32, properties map[string]string, createdByEventID string) (*models.Relation, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	edgeID := generateEdgeID()
	relation := &models.Relation{
		EdgeID:           edgeID,
		SrcObjectID:      srcObjectID,
		SrcType:          srcType,
		DstObjectID:      dstObjectID,
		DstType:          dstType,
		RelationType:     relationType,
		Weight:           weight,
		Properties:       properties,
		CreatedTs:        uint64(ts),
		CreatedByEventID: createdByEventID,
	}

	if err := mc.catalog.CreateRelation(ctx, relation, ts); err != nil {
		return nil, errors.Wrap(err, "failed to create relation")
	}

	log.Ctx(ctx).Info("relation created", zap.String("edgeID", edgeID), zap.String("relationType", relationType))
	return relation, nil
}

// GetRelation retrieves a relation by ID.
func (mc *MemoryCoord) GetRelation(ctx context.Context, edgeID string) (*models.Relation, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	relation, err := mc.catalog.GetRelation(ctx, edgeID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get relation")
	}

	return relation, nil
}

// ListRelations lists all relations for an object.
func (mc *MemoryCoord) ListRelations(ctx context.Context, objectID string) ([]*models.Relation, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	relations, err := mc.catalog.ListRelations(ctx, objectID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list relations")
	}

	return relations, nil
}

// DeleteRelation deletes a relation.
func (mc *MemoryCoord) DeleteRelation(ctx context.Context, edgeID string) error {
	if mc.GetStateCode() != StateCode_Healthy {
		return errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	if err := mc.catalog.DeleteRelation(ctx, edgeID, ts); err != nil {
		return errors.Wrap(err, "failed to delete relation")
	}

	log.Ctx(ctx).Info("relation deleted", zap.String("edgeID", edgeID))
	return nil
}

// ==================== ShareContract Management ====================

// CreateShareContract creates a new share contract.
func (mc *MemoryCoord) CreateShareContract(ctx context.Context, scope, ownerAgentID string, readACL, writeACL, deriveACL []string, ttlPolicy int64, consistencyLevel models.ConsistencyLevel, mergePolicy models.MergePolicy, metadata map[string]string) (*models.ShareContract, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	contractID := generateContractID()
	contract := &models.ShareContract{
		ContractID:        contractID,
		Scope:             scope,
		OwnerAgentID:      ownerAgentID,
		ReadACL:           readACL,
		WriteACL:          writeACL,
		DeriveACL:         deriveACL,
		TTLPolicy:         ttlPolicy,
		ConsistencyLevel:  consistencyLevel,
		MergePolicy:       mergePolicy,
		QuarantineEnabled: false,
		AuditEnabled:      true,
		Metadata:          metadata,
		CreatedTs:         uint64(ts),
		UpdatedTs:         uint64(ts),
	}

	if err := mc.catalog.CreateShareContract(ctx, contract, ts); err != nil {
		return nil, errors.Wrap(err, "failed to create share contract")
	}

	log.Ctx(ctx).Info("share contract created", zap.String("contractID", contractID), zap.String("scope", scope))
	return contract, nil
}

// GetShareContract retrieves a share contract by ID.
func (mc *MemoryCoord) GetShareContract(ctx context.Context, contractID string) (*models.ShareContract, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	contract, err := mc.catalog.GetShareContract(ctx, contractID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get share contract")
	}

	return contract, nil
}

// ListShareContracts lists all share contracts for a scope.
func (mc *MemoryCoord) ListShareContracts(ctx context.Context, scope string) ([]*models.ShareContract, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	contracts, err := mc.catalog.ListShareContracts(ctx, scope, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list share contracts")
	}

	return contracts, nil
}

// UpdateShareContract updates an existing share contract.
func (mc *MemoryCoord) UpdateShareContract(ctx context.Context, contractID string, updates map[string]interface{}) (*models.ShareContract, error) {
	if mc.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	contract, err := mc.catalog.GetShareContract(ctx, contractID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get share contract")
	}

	// Apply updates
	if readACL, ok := updates["read_acl"].([]string); ok {
		contract.ReadACL = readACL
	}
	if writeACL, ok := updates["write_acl"].([]string); ok {
		contract.WriteACL = writeACL
	}
	if deriveACL, ok := updates["derive_acl"].([]string); ok {
		contract.DeriveACL = deriveACL
	}
	if ttlPolicy, ok := updates["ttl_policy"].(int64); ok {
		contract.TTLPolicy = ttlPolicy
	}
	if consistencyLevel, ok := updates["consistency_level"].(models.ConsistencyLevel); ok {
		contract.ConsistencyLevel = consistencyLevel
	}
	if mergePolicy, ok := updates["merge_policy"].(models.MergePolicy); ok {
		contract.MergePolicy = mergePolicy
	}

	contract.UpdatedTs = uint64(ts)

	if err := mc.catalog.UpdateShareContract(ctx, contract, ts); err != nil {
		return nil, errors.Wrap(err, "failed to update share contract")
	}

	log.Ctx(ctx).Info("share contract updated", zap.String("contractID", contractID))
	return contract, nil
}

// DeleteShareContract deletes a share contract.
func (mc *MemoryCoord) DeleteShareContract(ctx context.Context, contractID string) error {
	if mc.GetStateCode() != StateCode_Healthy {
		return errors.New("MemoryCoord is not healthy")
	}

	ts, err := mc.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	if err := mc.catalog.DeleteShareContract(ctx, contractID, ts); err != nil {
		return errors.Wrap(err, "failed to delete share contract")
	}

	log.Ctx(ctx).Info("share contract deleted", zap.String("contractID", contractID))
	return nil
}

// ==================== Helper Functions ====================

func generateMemoryID() string {
	return typeutil.UniqueID(time.Now().UnixNano()).String()
}

func generateStateID() string {
	return typeutil.UniqueID(time.Now().UnixNano()).String()
}

func generateArtifactID() string {
	return typeutil.UniqueID(time.Now().UnixNano()).String()
}

func generateEdgeID() string {
	return typeutil.UniqueID(time.Now().UnixNano()).String()
}

func generateContractID() string {
	return typeutil.UniqueID(time.Now().UnixNano()).String()
}
