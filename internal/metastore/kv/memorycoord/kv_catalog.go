package memorycoord

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/errors"

	"github.com/milvus-io/milvus/pkg/v2/kv"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

// Catalog provides access to memory metadata stored in etcd.
type Catalog struct {
	Txn      kv.TxnKV
	Snapshot kv.SnapShotKV
}

// NewCatalog creates a new Catalog instance.
func NewCatalog(metaKV kv.TxnKV, ss kv.SnapShotKV) *Catalog {
	return &Catalog{Txn: metaKV, Snapshot: ss}
}

// ==================== Memory Operations ====================

// CreateMemory creates a new memory in the catalog.
func (c *Catalog) CreateMemory(ctx context.Context, memory *models.Memory, ts typeutil.Timestamp) error {
	key := BuildMemoryKey(memory.MemoryID)
	value, err := json.Marshal(memory)
	if err != nil {
		return errors.Wrap(err, "failed to marshal memory")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetMemory retrieves a memory by ID.
func (c *Catalog) GetMemory(ctx context.Context, memoryID string, ts typeutil.Timestamp) (*models.Memory, error) {
	key := BuildMemoryKey(memoryID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("memory not found: %s", memoryID)
		}
		return nil, errors.Wrap(err, "failed to load memory")
	}

	var memory models.Memory
	if err := json.Unmarshal([]byte(value), &memory); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal memory")
	}
	return &memory, nil
}

// ListMemories lists all memories for an agent or session.
func (c *Catalog) ListMemories(ctx context.Context, agentID, sessionID string, ts typeutil.Timestamp) ([]*models.Memory, error) {
	prefix := BuildMemoryPrefix(agentID, sessionID)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list memories")
	}

	memories := make([]*models.Memory, 0, len(values))
	for _, value := range values {
		var memory models.Memory
		if err := json.Unmarshal([]byte(value), &memory); err != nil {
			log.Warn("failed to unmarshal memory", log.Error(err))
			continue
		}
		memories = append(memories, &memory)
	}
	return memories, nil
}

// ListMemoriesByType lists memories filtered by type.
func (c *Catalog) ListMemoriesByType(ctx context.Context, agentID string, memoryType models.MemoryType, ts typeutil.Timestamp) ([]*models.Memory, error) {
	memories, err := c.ListMemories(ctx, agentID, "", ts)
	if err != nil {
		return nil, err
	}

	filtered := make([]*models.Memory, 0)
	for _, memory := range memories {
		if memory.MemoryType == memoryType {
			filtered = append(filtered, memory)
		}
	}
	return filtered, nil
}

// ListMemoriesByState lists memories filtered by state.
func (c *Catalog) ListMemoriesByState(ctx context.Context, agentID string, state models.MemoryState, ts typeutil.Timestamp) ([]*models.Memory, error) {
	memories, err := c.ListMemories(ctx, agentID, "", ts)
	if err != nil {
		return nil, err
	}

	filtered := make([]*models.Memory, 0)
	for _, memory := range memories {
		if memory.State == state {
			filtered = append(filtered, memory)
		}
	}
	return filtered, nil
}

// UpdateMemory updates an existing memory.
func (c *Catalog) UpdateMemory(ctx context.Context, memory *models.Memory, ts typeutil.Timestamp) error {
	key := BuildMemoryKey(memory.MemoryID)
	value, err := json.Marshal(memory)
	if err != nil {
		return errors.Wrap(err, "failed to marshal memory")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteMemory deletes a memory (soft delete by default).
func (c *Catalog) DeleteMemory(ctx context.Context, memoryID string, hardDelete bool, ts typeutil.Timestamp) error {
	key := BuildMemoryKey(memoryID)
	if hardDelete {
		return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
	}
	// Soft delete: update state to deleted
	memory, err := c.GetMemory(ctx, memoryID, ts)
	if err != nil {
		return err
	}
	memory.State = models.MemoryDeleted
	memory.IsActive = false
	return c.UpdateMemory(ctx, memory, ts)
}

// ==================== MemoryPolicy Operations ====================

// CreateMemoryPolicy creates a new memory policy.
func (c *Catalog) CreateMemoryPolicy(ctx context.Context, policy *models.MemoryPolicy, ts typeutil.Timestamp) error {
	key := BuildMemoryPolicyKey(policy.MemoryID)
	value, err := json.Marshal(policy)
	if err != nil {
		return errors.Wrap(err, "failed to marshal memory policy")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetMemoryPolicy retrieves a memory policy by memory ID.
func (c *Catalog) GetMemoryPolicy(ctx context.Context, memoryID string, ts typeutil.Timestamp) (*models.MemoryPolicy, error) {
	key := BuildMemoryPolicyKey(memoryID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("memory policy not found: %s", memoryID)
		}
		return nil, errors.Wrap(err, "failed to load memory policy")
	}

	var policy models.MemoryPolicy
	if err := json.Unmarshal([]byte(value), &policy); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal memory policy")
	}
	return &policy, nil
}

// UpdateMemoryPolicy updates an existing memory policy.
func (c *Catalog) UpdateMemoryPolicy(ctx context.Context, policy *models.MemoryPolicy, ts typeutil.Timestamp) error {
	key := BuildMemoryPolicyKey(policy.MemoryID)
	value, err := json.Marshal(policy)
	if err != nil {
		return errors.Wrap(err, "failed to marshal memory policy")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteMemoryPolicy deletes a memory policy.
func (c *Catalog) DeleteMemoryPolicy(ctx context.Context, memoryID string, ts typeutil.Timestamp) error {
	key := BuildMemoryPolicyKey(memoryID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== MemoryAdaptation Operations ====================

// CreateMemoryAdaptation creates a new memory adaptation.
func (c *Catalog) CreateMemoryAdaptation(ctx context.Context, adaptation *models.MemoryAdaptation, ts typeutil.Timestamp) error {
	key := BuildMemoryAdaptationKey(adaptation.MemoryID)
	value, err := json.Marshal(adaptation)
	if err != nil {
		return errors.Wrap(err, "failed to marshal memory adaptation")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetMemoryAdaptation retrieves a memory adaptation by memory ID.
func (c *Catalog) GetMemoryAdaptation(ctx context.Context, memoryID string, ts typeutil.Timestamp) (*models.MemoryAdaptation, error) {
	key := BuildMemoryAdaptationKey(memoryID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("memory adaptation not found: %s", memoryID)
		}
		return nil, errors.Wrap(err, "failed to load memory adaptation")
	}

	var adaptation models.MemoryAdaptation
	if err := json.Unmarshal([]byte(value), &adaptation); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal memory adaptation")
	}
	return &adaptation, nil
}

// UpdateMemoryAdaptation updates an existing memory adaptation.
func (c *Catalog) UpdateMemoryAdaptation(ctx context.Context, adaptation *models.MemoryAdaptation, ts typeutil.Timestamp) error {
	key := BuildMemoryAdaptationKey(adaptation.MemoryID)
	value, err := json.Marshal(adaptation)
	if err != nil {
		return errors.Wrap(err, "failed to marshal memory adaptation")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteMemoryAdaptation deletes a memory adaptation.
func (c *Catalog) DeleteMemoryAdaptation(ctx context.Context, memoryID string, ts typeutil.Timestamp) error {
	key := BuildMemoryAdaptationKey(memoryID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== State Operations ====================

// CreateState creates a new state in the catalog.
func (c *Catalog) CreateState(ctx context.Context, state *models.State, ts typeutil.Timestamp) error {
	key := BuildStateKey(state.StateID)
	value, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "failed to marshal state")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetState retrieves a state by ID.
func (c *Catalog) GetState(ctx context.Context, stateID string, ts typeutil.Timestamp) (*models.State, error) {
	key := BuildStateKey(stateID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("state not found: %s", stateID)
		}
		return nil, errors.Wrap(err, "failed to load state")
	}

	var state models.State
	if err := json.Unmarshal([]byte(value), &state); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal state")
	}
	return &state, nil
}

// ListStates lists all states for an agent or session.
func (c *Catalog) ListStates(ctx context.Context, agentID, sessionID string, ts typeutil.Timestamp) ([]*models.State, error) {
	prefix := BuildStatePrefix(agentID, sessionID)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list states")
	}

	states := make([]*models.State, 0, len(values))
	for _, value := range values {
		var state models.State
		if err := json.Unmarshal([]byte(value), &state); err != nil {
			log.Warn("failed to unmarshal state", log.Error(err))
			continue
		}
		states = append(states, &state)
	}
	return states, nil
}

// UpdateState updates an existing state.
func (c *Catalog) UpdateState(ctx context.Context, state *models.State, ts typeutil.Timestamp) error {
	key := BuildStateKey(state.StateID)
	value, err := json.Marshal(state)
	if err != nil {
		return errors.Wrap(err, "failed to marshal state")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteState deletes a state.
func (c *Catalog) DeleteState(ctx context.Context, stateID string, ts typeutil.Timestamp) error {
	key := BuildStateKey(stateID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== Artifact Operations ====================

// CreateArtifact creates a new artifact in the catalog.
func (c *Catalog) CreateArtifact(ctx context.Context, artifact *models.Artifact, ts typeutil.Timestamp) error {
	key := BuildArtifactKey(artifact.ArtifactID)
	value, err := json.Marshal(artifact)
	if err != nil {
		return errors.Wrap(err, "failed to marshal artifact")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetArtifact retrieves an artifact by ID.
func (c *Catalog) GetArtifact(ctx context.Context, artifactID string, ts typeutil.Timestamp) (*models.Artifact, error) {
	key := BuildArtifactKey(artifactID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("artifact not found: %s", artifactID)
		}
		return nil, errors.Wrap(err, "failed to load artifact")
	}

	var artifact models.Artifact
	if err := json.Unmarshal([]byte(value), &artifact); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal artifact")
	}
	return &artifact, nil
}

// ListArtifacts lists all artifacts for a session.
func (c *Catalog) ListArtifacts(ctx context.Context, sessionID string, ts typeutil.Timestamp) ([]*models.Artifact, error) {
	prefix := BuildArtifactPrefix(sessionID)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list artifacts")
	}

	artifacts := make([]*models.Artifact, 0, len(values))
	for _, value := range values {
		var artifact models.Artifact
		if err := json.Unmarshal([]byte(value), &artifact); err != nil {
			log.Warn("failed to unmarshal artifact", log.Error(err))
			continue
		}
		artifacts = append(artifacts, &artifact)
	}
	return artifacts, nil
}

// UpdateArtifact updates an existing artifact.
func (c *Catalog) UpdateArtifact(ctx context.Context, artifact *models.Artifact, ts typeutil.Timestamp) error {
	key := BuildArtifactKey(artifact.ArtifactID)
	value, err := json.Marshal(artifact)
	if err != nil {
		return errors.Wrap(err, "failed to marshal artifact")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteArtifact deletes an artifact.
func (c *Catalog) DeleteArtifact(ctx context.Context, artifactID string, ts typeutil.Timestamp) error {
	key := BuildArtifactKey(artifactID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== Relation Operations ====================

// CreateRelation creates a new relation in the catalog.
func (c *Catalog) CreateRelation(ctx context.Context, relation *models.Relation, ts typeutil.Timestamp) error {
	key := BuildRelationKey(relation.EdgeID)
	value, err := json.Marshal(relation)
	if err != nil {
		return errors.Wrap(err, "failed to marshal relation")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetRelation retrieves a relation by ID.
func (c *Catalog) GetRelation(ctx context.Context, edgeID string, ts typeutil.Timestamp) (*models.Relation, error) {
	key := BuildRelationKey(edgeID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("relation not found: %s", edgeID)
		}
		return nil, errors.Wrap(err, "failed to load relation")
	}

	var relation models.Relation
	if err := json.Unmarshal([]byte(value), &relation); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal relation")
	}
	return &relation, nil
}

// ListRelations lists all relations for an object.
func (c *Catalog) ListRelations(ctx context.Context, objectID string, ts typeutil.Timestamp) ([]*models.Relation, error) {
	prefix := BuildRelationPrefix(objectID)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list relations")
	}

	relations := make([]*models.Relation, 0, len(values))
	for _, value := range values {
		var relation models.Relation
		if err := json.Unmarshal([]byte(value), &relation); err != nil {
			log.Warn("failed to unmarshal relation", log.Error(err))
			continue
		}
		relations = append(relations, &relation)
	}
	return relations, nil
}

// DeleteRelation deletes a relation.
func (c *Catalog) DeleteRelation(ctx context.Context, edgeID string, ts typeutil.Timestamp) error {
	key := BuildRelationKey(edgeID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== ShareContract Operations ====================

// CreateShareContract creates a new share contract in the catalog.
func (c *Catalog) CreateShareContract(ctx context.Context, contract *models.ShareContract, ts typeutil.Timestamp) error {
	key := BuildShareContractKey(contract.ContractID)
	value, err := json.Marshal(contract)
	if err != nil {
		return errors.Wrap(err, "failed to marshal share contract")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetShareContract retrieves a share contract by ID.
func (c *Catalog) GetShareContract(ctx context.Context, contractID string, ts typeutil.Timestamp) (*models.ShareContract, error) {
	key := BuildShareContractKey(contractID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("share contract not found: %s", contractID)
		}
		return nil, errors.Wrap(err, "failed to load share contract")
	}

	var contract models.ShareContract
	if err := json.Unmarshal([]byte(value), &contract); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal share contract")
	}
	return &contract, nil
}

// ListShareContracts lists all share contracts for a scope.
func (c *Catalog) ListShareContracts(ctx context.Context, scope string, ts typeutil.Timestamp) ([]*models.ShareContract, error) {
	prefix := BuildShareContractPrefix(scope)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list share contracts")
	}

	contracts := make([]*models.ShareContract, 0, len(values))
	for _, value := range values {
		var contract models.ShareContract
		if err := json.Unmarshal([]byte(value), &contract); err != nil {
			log.Warn("failed to unmarshal share contract", log.Error(err))
			continue
		}
		contracts = append(contracts, &contract)
	}
	return contracts, nil
}

// UpdateShareContract updates an existing share contract.
func (c *Catalog) UpdateShareContract(ctx context.Context, contract *models.ShareContract, ts typeutil.Timestamp) error {
	key := BuildShareContractKey(contract.ContractID)
	value, err := json.Marshal(contract)
	if err != nil {
		return errors.Wrap(err, "failed to marshal share contract")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteShareContract deletes a share contract.
func (c *Catalog) DeleteShareContract(ctx context.Context, contractID string, ts typeutil.Timestamp) error {
	key := BuildShareContractKey(contractID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== Transaction Operations ====================

// MultiSave saves multiple key-value pairs in a transaction.
func (c *Catalog) MultiSave(ctx context.Context, kvs map[string]string, ts typeutil.Timestamp) error {
	return c.Snapshot.MultiSave(ctx, kvs, ts)
}

// MultiSaveAndRemove saves and removes multiple keys in a transaction.
func (c *Catalog) MultiSaveAndRemove(ctx context.Context, saves map[string]string, removals []string, ts typeutil.Timestamp) error {
	return c.Snapshot.MultiSaveAndRemove(ctx, saves, removals, ts)
}
