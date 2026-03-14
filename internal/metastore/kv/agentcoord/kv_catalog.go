package agentcoord

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cockroachdb/errors"

	"github.com/milvus-io/milvus/pkg/v2/kv"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

// Catalog provides access to agent metadata stored in etcd.
type Catalog struct {
	Txn      kv.TxnKV
	Snapshot kv.SnapShotKV
}

// NewCatalog creates a new Catalog instance.
func NewCatalog(metaKV kv.TxnKV, ss kv.SnapShotKV) *Catalog {
	return &Catalog{Txn: metaKV, Snapshot: ss}
}

// ==================== Agent Operations ====================

// CreateAgent creates a new agent in the catalog.
func (c *Catalog) CreateAgent(ctx context.Context, agent *models.Agent, ts typeutil.Timestamp) error {
	key := BuildAgentKey(agent.AgentID)
	value, err := json.Marshal(agent)
	if err != nil {
		return errors.Wrap(err, "failed to marshal agent")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetAgent retrieves an agent by ID.
func (c *Catalog) GetAgent(ctx context.Context, agentID string, ts typeutil.Timestamp) (*models.Agent, error) {
	key := BuildAgentKey(agentID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("agent not found: %s", agentID)
		}
		return nil, errors.Wrap(err, "failed to load agent")
	}

	var agent models.Agent
	if err := json.Unmarshal([]byte(value), &agent); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal agent")
	}
	return &agent, nil
}

// ListAgents lists all agents in a workspace.
func (c *Catalog) ListAgents(ctx context.Context, tenantID, workspaceID string, ts typeutil.Timestamp) ([]*models.Agent, error) {
	prefix := BuildAgentPrefix(tenantID, workspaceID)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list agents")
	}

	agents := make([]*models.Agent, 0, len(values))
	for _, value := range values {
		var agent models.Agent
		if err := json.Unmarshal([]byte(value), &agent); err != nil {
			log.Warn("failed to unmarshal agent", log.Error(err))
			continue
		}
		agents = append(agents, &agent)
	}
	return agents, nil
}

// UpdateAgent updates an existing agent.
func (c *Catalog) UpdateAgent(ctx context.Context, agent *models.Agent, ts typeutil.Timestamp) error {
	key := BuildAgentKey(agent.AgentID)
	value, err := json.Marshal(agent)
	if err != nil {
		return errors.Wrap(err, "failed to marshal agent")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteAgent deletes an agent.
func (c *Catalog) DeleteAgent(ctx context.Context, agentID string, ts typeutil.Timestamp) error {
	key := BuildAgentKey(agentID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// AgentExists checks if an agent exists.
func (c *Catalog) AgentExists(ctx context.Context, agentID string, ts typeutil.Timestamp) bool {
	key := BuildAgentKey(agentID)
	_, err := c.Snapshot.Load(ctx, key, ts)
	return err == nil
}

// ==================== Session Operations ====================

// CreateSession creates a new session in the catalog.
func (c *Catalog) CreateSession(ctx context.Context, session *models.Session, ts typeutil.Timestamp) error {
	key := BuildSessionKey(session.SessionID)
	value, err := json.Marshal(session)
	if err != nil {
		return errors.Wrap(err, "failed to marshal session")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetSession retrieves a session by ID.
func (c *Catalog) GetSession(ctx context.Context, sessionID string, ts typeutil.Timestamp) (*models.Session, error) {
	key := BuildSessionKey(sessionID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("session not found: %s", sessionID)
		}
		return nil, errors.Wrap(err, "failed to load session")
	}

	var session models.Session
	if err := json.Unmarshal([]byte(value), &session); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal session")
	}
	return &session, nil
}

// ListSessions lists all sessions for an agent.
func (c *Catalog) ListSessions(ctx context.Context, agentID string, ts typeutil.Timestamp) ([]*models.Session, error) {
	prefix := BuildSessionPrefix(agentID)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list sessions")
	}

	sessions := make([]*models.Session, 0, len(values))
	for _, value := range values {
		var session models.Session
		if err := json.Unmarshal([]byte(value), &session); err != nil {
			log.Warn("failed to unmarshal session", log.Error(err))
			continue
		}
		sessions = append(sessions, &session)
	}
	return sessions, nil
}

// ListSessionsByState lists sessions filtered by state.
func (c *Catalog) ListSessionsByState(ctx context.Context, agentID string, state models.SessionState, ts typeutil.Timestamp) ([]*models.Session, error) {
	sessions, err := c.ListSessions(ctx, agentID, ts)
	if err != nil {
		return nil, err
	}

	filtered := make([]*models.Session, 0)
	for _, session := range sessions {
		if session.State == state {
			filtered = append(filtered, session)
		}
	}
	return filtered, nil
}

// UpdateSession updates an existing session.
func (c *Catalog) UpdateSession(ctx context.Context, session *models.Session, ts typeutil.Timestamp) error {
	key := BuildSessionKey(session.SessionID)
	value, err := json.Marshal(session)
	if err != nil {
		return errors.Wrap(err, "failed to marshal session")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteSession deletes a session.
func (c *Catalog) DeleteSession(ctx context.Context, sessionID string, ts typeutil.Timestamp) error {
	key := BuildSessionKey(sessionID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== Workspace Operations ====================

// CreateWorkspace creates a new workspace in the catalog.
func (c *Catalog) CreateWorkspace(ctx context.Context, workspace *models.Workspace, ts typeutil.Timestamp) error {
	key := BuildWorkspaceKey(workspace.WorkspaceID)
	value, err := json.Marshal(workspace)
	if err != nil {
		return errors.Wrap(err, "failed to marshal workspace")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetWorkspace retrieves a workspace by ID.
func (c *Catalog) GetWorkspace(ctx context.Context, workspaceID string, ts typeutil.Timestamp) (*models.Workspace, error) {
	key := BuildWorkspaceKey(workspaceID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			return nil, errors.Errorf("workspace not found: %s", workspaceID)
		}
		return nil, errors.Wrap(err, "failed to load workspace")
	}

	var workspace models.Workspace
	if err := json.Unmarshal([]byte(value), &workspace); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal workspace")
	}
	return &workspace, nil
}

// ListWorkspaces lists all workspaces for a tenant.
func (c *Catalog) ListWorkspaces(ctx context.Context, tenantID string, ts typeutil.Timestamp) ([]*models.Workspace, error) {
	prefix := BuildWorkspacePrefix(tenantID)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list workspaces")
	}

	workspaces := make([]*models.Workspace, 0, len(values))
	for _, value := range values {
		var workspace models.Workspace
		if err := json.Unmarshal([]byte(value), &workspace); err != nil {
			log.Warn("failed to unmarshal workspace", log.Error(err))
			continue
		}
		workspaces = append(workspaces, &workspace)
	}
	return workspaces, nil
}

// UpdateWorkspace updates an existing workspace.
func (c *Catalog) UpdateWorkspace(ctx context.Context, workspace *models.Workspace, ts typeutil.Timestamp) error {
	key := BuildWorkspaceKey(workspace.WorkspaceID)
	value, err := json.Marshal(workspace)
	if err != nil {
		return errors.Wrap(err, "failed to marshal workspace")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// DeleteWorkspace deletes a workspace.
func (c *Catalog) DeleteWorkspace(ctx context.Context, workspaceID string, ts typeutil.Timestamp) error {
	key := BuildWorkspaceKey(workspaceID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== Snapshot Operations ====================

// CreateSnapshot creates a snapshot of an object.
func (c *Catalog) CreateSnapshot(ctx context.Context, objectType, objectID string, data []byte, version int64, ts typeutil.Timestamp) error {
	key := BuildSnapshotKey(objectType, objectID, version)
	return c.Snapshot.Save(ctx, key, string(data), ts)
}

// GetSnapshot retrieves a snapshot.
func (c *Catalog) GetSnapshot(ctx context.Context, objectType, objectID string, version int64, ts typeutil.Timestamp) ([]byte, error) {
	key := BuildSnapshotKey(objectType, objectID, version)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load snapshot")
	}
	return []byte(value), nil
}

// ListSnapshots lists all snapshots for an object.
func (c *Catalog) ListSnapshots(ctx context.Context, objectType, objectID string, ts typeutil.Timestamp) ([]int64, error) {
	prefix := fmt.Sprintf("%s/%s/%s", SnapshotPrefix, objectType, objectID)
	keys, _, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list snapshots")
	}

	versions := make([]int64, 0, len(keys))
	for _, key := range keys {
		var version int64
		if _, err := fmt.Sscanf(key, prefix+"/%d", &version); err == nil {
			versions = append(versions, version)
		}
	}
	return versions, nil
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
