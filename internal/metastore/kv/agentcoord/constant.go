package agentcoord

import "fmt"

const (
	// ComponentPrefix prefix for agentcoord component
	ComponentPrefix = "agent-coord"

	// AgentMetaPrefix prefix for agent meta
	AgentMetaPrefix = ComponentPrefix + "/agent"

	// SessionMetaPrefix prefix for session meta
	SessionMetaPrefix = ComponentPrefix + "/session"

	// WorkspaceMetaPrefix prefix for workspace meta
	WorkspaceMetaPrefix = ComponentPrefix + "/workspace"

	// SnapshotPrefix prefix for snapshots
	SnapshotPrefix = ComponentPrefix + "/snapshots"
)

// BuildAgentKey builds agent key
func BuildAgentKey(agentID string) string {
	return fmt.Sprintf("%s/%s", AgentMetaPrefix, agentID)
}

// BuildAgentPrefix builds agent prefix for listing
func BuildAgentPrefix(tenantID, workspaceID string) string {
	if tenantID != "" && workspaceID != "" {
		return fmt.Sprintf("%s/%s/%s", AgentMetaPrefix, tenantID, workspaceID)
	}
	if tenantID != "" {
		return fmt.Sprintf("%s/%s", AgentMetaPrefix, tenantID)
	}
	return AgentMetaPrefix
}

// BuildSessionKey builds session key
func BuildSessionKey(sessionID string) string {
	return fmt.Sprintf("%s/%s", SessionMetaPrefix, sessionID)
}

// BuildSessionPrefix builds session prefix for listing
func BuildSessionPrefix(agentID string) string {
	if agentID != "" {
		return fmt.Sprintf("%s/%s", SessionMetaPrefix, agentID)
	}
	return SessionMetaPrefix
}

// BuildWorkspaceKey builds workspace key
func BuildWorkspaceKey(workspaceID string) string {
	return fmt.Sprintf("%s/%s", WorkspaceMetaPrefix, workspaceID)
}

// BuildWorkspacePrefix builds workspace prefix for listing
func BuildWorkspacePrefix(tenantID string) string {
	if tenantID != "" {
		return fmt.Sprintf("%s/%s", WorkspaceMetaPrefix, tenantID)
	}
	return WorkspaceMetaPrefix
}

// BuildSnapshotKey builds snapshot key
func BuildSnapshotKey(objectType, objectID string, version int64) string {
	return fmt.Sprintf("%s/%s/%s/%d", SnapshotPrefix, objectType, objectID, version)
}
