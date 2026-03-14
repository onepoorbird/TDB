package agentcoord

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/milvus-io/milvus/internal/metastore/kv/agentcoord"
	"github.com/milvus-io/milvus/internal/util/sessionutil"
	"github.com/milvus-io/milvus/pkg/v2/kv"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

// StateCode represents the state of the AgentCoord.
type StateCode int32

const (
	StateCode_Initializing StateCode = 0
	StateCode_Healthy      StateCode = 1
	StateCode_Abnormal     StateCode = 2
	StateCode_Stopping     StateCode = 3
)

// AgentCoord manages agents, sessions, and workspaces.
type AgentCoord struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Dependencies
	catalog  *agentcoord.Catalog
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

// NewAgentCoord creates a new AgentCoord instance.
func NewAgentCoord(ctx context.Context, metaKV kv.TxnKV, snapshotKV kv.SnapShotKV, session sessionutil.SessionInterface, tsoAllocator typeutil.TimestampAllocator) (*AgentCoord, error) {
	ctx, cancel := context.WithCancel(ctx)

	catalog := agentcoord.NewCatalog(metaKV, snapshotKV)

	ac := &AgentCoord{
		ctx:          ctx,
		cancel:       cancel,
		catalog:      catalog,
		session:      session,
		tsoAllocator: tsoAllocator,
		ticker:       time.NewTicker(5 * time.Second),
	}

	ac.UpdateStateCode(StateCode_Initializing)
	return ac, nil
}

// UpdateStateCode updates the state code.
func (ac *AgentCoord) UpdateStateCode(code StateCode) {
	ac.stateCode.Store(int32(code))
	log.Ctx(ac.ctx).Info("update agentcoord state", zap.String("state", code.String()))
}

// GetStateCode returns the current state code.
func (ac *AgentCoord) GetStateCode() StateCode {
	return StateCode(ac.stateCode.Load())
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

// Init initializes the AgentCoord.
func (ac *AgentCoord) Init() error {
	var err error
	ac.initOnce.Do(func() {
		err = ac.init()
	})
	return err
}

func (ac *AgentCoord) init() error {
	log.Ctx(ac.ctx).Info("AgentCoord initializing")

	// TODO: Recover state from catalog if needed

	ac.UpdateStateCode(StateCode_Healthy)
	log.Ctx(ac.ctx).Info("AgentCoord initialized")
	return nil
}

// Start starts the AgentCoord background tasks.
func (ac *AgentCoord) Start() error {
	var err error
	ac.startOnce.Do(func() {
		err = ac.start()
	})
	return err
}

func (ac *AgentCoord) start() error {
	log.Ctx(ac.ctx).Info("AgentCoord starting")

	ac.wg.Add(1)
	go ac.backgroundTask()

	log.Ctx(ac.ctx).Info("AgentCoord started")
	return nil
}

// Stop stops the AgentCoord.
func (ac *AgentCoord) Stop() error {
	var err error
	ac.stopOnce.Do(func() {
		err = ac.stop()
	})
	return err
}

func (ac *AgentCoord) stop() error {
	log.Ctx(ac.ctx).Info("AgentCoord stopping")
	ac.UpdateStateCode(StateCode_Stopping)

	ac.ticker.Stop()
	ac.cancel()
	ac.wg.Wait()

	log.Ctx(ac.ctx).Info("AgentCoord stopped")
	return nil
}

func (ac *AgentCoord) backgroundTask() {
	defer ac.wg.Done()

	for {
		select {
		case <-ac.ctx.Done():
			return
		case <-ac.ticker.C:
			ac.doBackgroundTasks()
		}
	}
}

func (ac *AgentCoord) doBackgroundTasks() {
	// TODO: Implement background tasks
	// - Cleanup expired sessions
	// - Update agent statistics
	// - Check agent health
}

// getTimestamp returns the current timestamp.
func (ac *AgentCoord) getTimestamp() (typeutil.Timestamp, error) {
	if ac.tsoAllocator != nil {
		return ac.tsoAllocator.AllocOne(ac.ctx)
	}
	return typeutil.Timestamp(time.Now().UnixNano()), nil
}

// ==================== Agent Management ====================

// CreateAgent creates a new agent.
func (ac *AgentCoord) CreateAgent(ctx context.Context, tenantID, workspaceID, agentType, roleProfile string, capabilitySet []string, defaultMemoryPolicy string, metadata map[string]string) (*models.Agent, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	agentID := generateAgentID()
	agent := &models.Agent{
		AgentID:             agentID,
		TenantID:            tenantID,
		WorkspaceID:         workspaceID,
		AgentType:           agentType,
		RoleProfile:         roleProfile,
		CapabilitySet:       capabilitySet,
		DefaultMemoryPolicy: defaultMemoryPolicy,
		CreatedTs:           uint64(ts),
		UpdatedTs:           uint64(ts),
		State:               models.AgentCreating,
		Metadata:            metadata,
	}

	if err := ac.catalog.CreateAgent(ctx, agent, ts); err != nil {
		return nil, errors.Wrap(err, "failed to create agent")
	}

	// Update state to active
	agent.State = models.AgentActive
	agent.UpdatedTs = uint64(ts)
	if err := ac.catalog.UpdateAgent(ctx, agent, ts); err != nil {
		return nil, errors.Wrap(err, "failed to update agent state")
	}

	log.Ctx(ctx).Info("agent created", zap.String("agentID", agentID))
	return agent, nil
}

// GetAgent retrieves an agent by ID.
func (ac *AgentCoord) GetAgent(ctx context.Context, agentID string) (*models.Agent, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	agent, err := ac.catalog.GetAgent(ctx, agentID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get agent")
	}

	return agent, nil
}

// ListAgents lists all agents in a workspace.
func (ac *AgentCoord) ListAgents(ctx context.Context, tenantID, workspaceID string) ([]*models.Agent, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	agents, err := ac.catalog.ListAgents(ctx, tenantID, workspaceID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list agents")
	}

	return agents, nil
}

// UpdateAgent updates an existing agent.
func (ac *AgentCoord) UpdateAgent(ctx context.Context, agentID string, updates map[string]interface{}) (*models.Agent, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	agent, err := ac.catalog.GetAgent(ctx, agentID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get agent")
	}

	// Apply updates
	if roleProfile, ok := updates["role_profile"].(string); ok {
		agent.RoleProfile = roleProfile
	}
	if capabilitySet, ok := updates["capability_set"].([]string); ok {
		agent.CapabilitySet = capabilitySet
	}
	if defaultMemoryPolicy, ok := updates["default_memory_policy"].(string); ok {
		agent.DefaultMemoryPolicy = defaultMemoryPolicy
	}
	if metadata, ok := updates["metadata"].(map[string]string); ok {
		agent.Metadata = metadata
	}

	agent.UpdatedTs = uint64(ts)
	if err := ac.catalog.UpdateAgent(ctx, agent, ts); err != nil {
		return nil, errors.Wrap(err, "failed to update agent")
	}

	log.Ctx(ctx).Info("agent updated", zap.String("agentID", agentID))
	return agent, nil
}

// DeleteAgent deletes an agent.
func (ac *AgentCoord) DeleteAgent(ctx context.Context, agentID string) error {
	if ac.GetStateCode() != StateCode_Healthy {
		return errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	// Check if agent exists
	if !ac.catalog.AgentExists(ctx, agentID, ts) {
		return errors.Errorf("agent not found: %s", agentID)
	}

	// TODO: Check if agent has active sessions

	if err := ac.catalog.DeleteAgent(ctx, agentID, ts); err != nil {
		return errors.Wrap(err, "failed to delete agent")
	}

	log.Ctx(ctx).Info("agent deleted", zap.String("agentID", agentID))
	return nil
}

// PauseAgent pauses an agent.
func (ac *AgentCoord) PauseAgent(ctx context.Context, agentID string) error {
	if ac.GetStateCode() != StateCode_Healthy {
		return errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	agent, err := ac.catalog.GetAgent(ctx, agentID, ts)
	if err != nil {
		return errors.Wrap(err, "failed to get agent")
	}

	agent.State = models.AgentPaused
	agent.UpdatedTs = uint64(ts)

	if err := ac.catalog.UpdateAgent(ctx, agent, ts); err != nil {
		return errors.Wrap(err, "failed to pause agent")
	}

	log.Ctx(ctx).Info("agent paused", zap.String("agentID", agentID))
	return nil
}

// ResumeAgent resumes a paused agent.
func (ac *AgentCoord) ResumeAgent(ctx context.Context, agentID string) error {
	if ac.GetStateCode() != StateCode_Healthy {
		return errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	agent, err := ac.catalog.GetAgent(ctx, agentID, ts)
	if err != nil {
		return errors.Wrap(err, "failed to get agent")
	}

	agent.State = models.AgentActive
	agent.UpdatedTs = uint64(ts)

	if err := ac.catalog.UpdateAgent(ctx, agent, ts); err != nil {
		return errors.Wrap(err, "failed to resume agent")
	}

	log.Ctx(ctx).Info("agent resumed", zap.String("agentID", agentID))
	return nil
}

// ==================== Session Management ====================

// CreateSession creates a new session.
func (ac *AgentCoord) CreateSession(ctx context.Context, agentID, parentSessionID, taskType, goal string, budgetToken, budgetTimeMs int64, metadata map[string]string) (*models.Session, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	// Check if agent exists
	if !ac.catalog.AgentExists(ctx, agentID, ts) {
		return nil, errors.Errorf("agent not found: %s", agentID)
	}

	sessionID := generateSessionID()
	session := &models.Session{
		SessionID:       sessionID,
		AgentID:         agentID,
		ParentSessionID: parentSessionID,
		TaskType:        taskType,
		Goal:            goal,
		StartTs:         uint64(ts),
		State:           models.SessionCreating,
		BudgetToken:     budgetToken,
		BudgetTimeMs:    budgetTimeMs,
		Metadata:        metadata,
	}

	if err := ac.catalog.CreateSession(ctx, session, ts); err != nil {
		return nil, errors.Wrap(err, "failed to create session")
	}

	// Update state to active
	session.State = models.SessionActive
	if err := ac.catalog.UpdateSession(ctx, session, ts); err != nil {
		return nil, errors.Wrap(err, "failed to update session state")
	}

	log.Ctx(ctx).Info("session created", zap.String("sessionID", sessionID), zap.String("agentID", agentID))
	return session, nil
}

// GetSession retrieves a session by ID.
func (ac *AgentCoord) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	session, err := ac.catalog.GetSession(ctx, sessionID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}

	return session, nil
}

// ListSessions lists all sessions for an agent.
func (ac *AgentCoord) ListSessions(ctx context.Context, agentID string) ([]*models.Session, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	sessions, err := ac.catalog.ListSessions(ctx, agentID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list sessions")
	}

	return sessions, nil
}

// ListActiveSessions lists active sessions for an agent.
func (ac *AgentCoord) ListActiveSessions(ctx context.Context, agentID string) ([]*models.Session, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	sessions, err := ac.catalog.ListSessionsByState(ctx, agentID, models.SessionActive, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list active sessions")
	}

	return sessions, nil
}

// UpdateSession updates an existing session.
func (ac *AgentCoord) UpdateSession(ctx context.Context, sessionID string, updates map[string]interface{}) (*models.Session, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	session, err := ac.catalog.GetSession(ctx, sessionID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get session")
	}

	// Apply updates
	if goal, ok := updates["goal"].(string); ok {
		session.Goal = goal
	}
	if state, ok := updates["state"].(models.SessionState); ok {
		session.State = state
	}
	if metadata, ok := updates["metadata"].(map[string]string); ok {
		session.Metadata = metadata
	}

	if err := ac.catalog.UpdateSession(ctx, session, ts); err != nil {
		return nil, errors.Wrap(err, "failed to update session")
	}

	log.Ctx(ctx).Info("session updated", zap.String("sessionID", sessionID))
	return session, nil
}

// CompleteSession completes a session.
func (ac *AgentCoord) CompleteSession(ctx context.Context, sessionID string) error {
	if ac.GetStateCode() != StateCode_Healthy {
		return errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	session, err := ac.catalog.GetSession(ctx, sessionID, ts)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	session.State = models.SessionCompleted
	session.EndTs = uint64(ts)

	if err := ac.catalog.UpdateSession(ctx, session, ts); err != nil {
		return errors.Wrap(err, "failed to complete session")
	}

	log.Ctx(ctx).Info("session completed", zap.String("sessionID", sessionID))
	return nil
}

// FailSession marks a session as failed.
func (ac *AgentCoord) FailSession(ctx context.Context, sessionID string) error {
	if ac.GetStateCode() != StateCode_Healthy {
		return errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	session, err := ac.catalog.GetSession(ctx, sessionID, ts)
	if err != nil {
		return errors.Wrap(err, "failed to get session")
	}

	session.State = models.SessionFailed
	session.EndTs = uint64(ts)

	if err := ac.catalog.UpdateSession(ctx, session, ts); err != nil {
		return errors.Wrap(err, "failed to fail session")
	}

	log.Ctx(ctx).Info("session failed", zap.String("sessionID", sessionID))
	return nil
}

// ==================== Workspace Management ====================

// CreateWorkspace creates a new workspace.
func (ac *AgentCoord) CreateWorkspace(ctx context.Context, tenantID, name, description string, metadata map[string]string) (*models.Workspace, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	workspaceID := generateWorkspaceID()
	workspace := &models.Workspace{
		WorkspaceID: workspaceID,
		TenantID:    tenantID,
		Name:        name,
		Description: description,
		AgentIDs:    []string{},
		Metadata:    metadata,
		CreatedTs:   uint64(ts),
		UpdatedTs:   uint64(ts),
	}

	if err := ac.catalog.CreateWorkspace(ctx, workspace, ts); err != nil {
		return nil, errors.Wrap(err, "failed to create workspace")
	}

	log.Ctx(ctx).Info("workspace created", zap.String("workspaceID", workspaceID))
	return workspace, nil
}

// GetWorkspace retrieves a workspace by ID.
func (ac *AgentCoord) GetWorkspace(ctx context.Context, workspaceID string) (*models.Workspace, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	workspace, err := ac.catalog.GetWorkspace(ctx, workspaceID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get workspace")
	}

	return workspace, nil
}

// ListWorkspaces lists all workspaces for a tenant.
func (ac *AgentCoord) ListWorkspaces(ctx context.Context, tenantID string) ([]*models.Workspace, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	workspaces, err := ac.catalog.ListWorkspaces(ctx, tenantID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list workspaces")
	}

	return workspaces, nil
}

// UpdateWorkspace updates an existing workspace.
func (ac *AgentCoord) UpdateWorkspace(ctx context.Context, workspaceID string, updates map[string]interface{}) (*models.Workspace, error) {
	if ac.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	workspace, err := ac.catalog.GetWorkspace(ctx, workspaceID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get workspace")
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		workspace.Name = name
	}
	if description, ok := updates["description"].(string); ok {
		workspace.Description = description
	}
	if agentIDs, ok := updates["agent_ids"].([]string); ok {
		workspace.AgentIDs = agentIDs
	}
	if metadata, ok := updates["metadata"].(map[string]string); ok {
		workspace.Metadata = metadata
	}

	workspace.UpdatedTs = uint64(ts)

	if err := ac.catalog.UpdateWorkspace(ctx, workspace, ts); err != nil {
		return nil, errors.Wrap(err, "failed to update workspace")
	}

	log.Ctx(ctx).Info("workspace updated", zap.String("workspaceID", workspaceID))
	return workspace, nil
}

// DeleteWorkspace deletes a workspace.
func (ac *AgentCoord) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	if ac.GetStateCode() != StateCode_Healthy {
		return errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	// TODO: Check if workspace has agents

	if err := ac.catalog.DeleteWorkspace(ctx, workspaceID, ts); err != nil {
		return errors.Wrap(err, "failed to delete workspace")
	}

	log.Ctx(ctx).Info("workspace deleted", zap.String("workspaceID", workspaceID))
	return nil
}

// AddAgentToWorkspace adds an agent to a workspace.
func (ac *AgentCoord) AddAgentToWorkspace(ctx context.Context, workspaceID, agentID string) error {
	if ac.GetStateCode() != StateCode_Healthy {
		return errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	workspace, err := ac.catalog.GetWorkspace(ctx, workspaceID, ts)
	if err != nil {
		return errors.Wrap(err, "failed to get workspace")
	}

	// Check if agent already exists in workspace
	for _, id := range workspace.AgentIDs {
		if id == agentID {
			return nil // Already exists
		}
	}

	workspace.AgentIDs = append(workspace.AgentIDs, agentID)
	workspace.UpdatedTs = uint64(ts)

	if err := ac.catalog.UpdateWorkspace(ctx, workspace, ts); err != nil {
		return errors.Wrap(err, "failed to add agent to workspace")
	}

	log.Ctx(ctx).Info("agent added to workspace", zap.String("workspaceID", workspaceID), zap.String("agentID", agentID))
	return nil
}

// RemoveAgentFromWorkspace removes an agent from a workspace.
func (ac *AgentCoord) RemoveAgentFromWorkspace(ctx context.Context, workspaceID, agentID string) error {
	if ac.GetStateCode() != StateCode_Healthy {
		return errors.New("AgentCoord is not healthy")
	}

	ts, err := ac.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	workspace, err := ac.catalog.GetWorkspace(ctx, workspaceID, ts)
	if err != nil {
		return errors.Wrap(err, "failed to get workspace")
	}

	// Remove agent from list
	newAgentIDs := make([]string, 0, len(workspace.AgentIDs))
	for _, id := range workspace.AgentIDs {
		if id != agentID {
			newAgentIDs = append(newAgentIDs, id)
		}
	}

	workspace.AgentIDs = newAgentIDs
	workspace.UpdatedTs = uint64(ts)

	if err := ac.catalog.UpdateWorkspace(ctx, workspace, ts); err != nil {
		return errors.Wrap(err, "failed to remove agent from workspace")
	}

	log.Ctx(ctx).Info("agent removed from workspace", zap.String("workspaceID", workspaceID), zap.String("agentID", agentID))
	return nil
}

// ==================== Helper Functions ====================

func generateAgentID() string {
	return typeutil.UniqueID(time.Now().UnixNano()).String()
}

func generateSessionID() string {
	return typeutil.UniqueID(time.Now().UnixNano()).String()
}

func generateWorkspaceID() string {
	return typeutil.UniqueID(time.Now().UnixNano()).String()
}
