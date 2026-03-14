package memorycoord

import "fmt"

const (
	// ComponentPrefix prefix for memorycoord component
	ComponentPrefix = "memory-coord"

	// MemoryMetaPrefix prefix for memory meta
	MemoryMetaPrefix = ComponentPrefix + "/memory"

	// MemoryPolicyPrefix prefix for memory policy
	MemoryPolicyPrefix = ComponentPrefix + "/memory-policy"

	// MemoryAdaptationPrefix prefix for memory adaptation
	MemoryAdaptationPrefix = ComponentPrefix + "/memory-adaptation"

	// StateMetaPrefix prefix for state meta
	StateMetaPrefix = ComponentPrefix + "/state"

	// ArtifactMetaPrefix prefix for artifact meta
	ArtifactMetaPrefix = ComponentPrefix + "/artifact"

	// RelationMetaPrefix prefix for relation meta
	RelationMetaPrefix = ComponentPrefix + "/relation"

	// ShareContractPrefix prefix for share contract
	ShareContractPrefix = ComponentPrefix + "/share-contract"

	// MemoryIndexPrefix prefix for memory index
	MemoryIndexPrefix = ComponentPrefix + "/index"

	// SnapshotPrefix prefix for snapshots
	SnapshotPrefix = ComponentPrefix + "/snapshots"
)

// BuildMemoryKey builds memory key
func BuildMemoryKey(memoryID string) string {
	return fmt.Sprintf("%s/%s", MemoryMetaPrefix, memoryID)
}

// BuildMemoryPrefix builds memory prefix for listing
func BuildMemoryPrefix(agentID, sessionID string) string {
	if agentID != "" && sessionID != "" {
		return fmt.Sprintf("%s/%s/%s", MemoryMetaPrefix, agentID, sessionID)
	}
	if agentID != "" {
		return fmt.Sprintf("%s/%s", MemoryMetaPrefix, agentID)
	}
	return MemoryMetaPrefix
}

// BuildMemoryPolicyKey builds memory policy key
func BuildMemoryPolicyKey(memoryID string) string {
	return fmt.Sprintf("%s/%s", MemoryPolicyPrefix, memoryID)
}

// BuildMemoryAdaptationKey builds memory adaptation key
func BuildMemoryAdaptationKey(memoryID string) string {
	return fmt.Sprintf("%s/%s", MemoryAdaptationPrefix, memoryID)
}

// BuildStateKey builds state key
func BuildStateKey(stateID string) string {
	return fmt.Sprintf("%s/%s", StateMetaPrefix, stateID)
}

// BuildStatePrefix builds state prefix for listing
func BuildStatePrefix(agentID, sessionID string) string {
	if agentID != "" && sessionID != "" {
		return fmt.Sprintf("%s/%s/%s", StateMetaPrefix, agentID, sessionID)
	}
	if agentID != "" {
		return fmt.Sprintf("%s/%s", StateMetaPrefix, agentID)
	}
	return StateMetaPrefix
}

// BuildArtifactKey builds artifact key
func BuildArtifactKey(artifactID string) string {
	return fmt.Sprintf("%s/%s", ArtifactMetaPrefix, artifactID)
}

// BuildArtifactPrefix builds artifact prefix for listing
func BuildArtifactPrefix(sessionID string) string {
	if sessionID != "" {
		return fmt.Sprintf("%s/%s", ArtifactMetaPrefix, sessionID)
	}
	return ArtifactMetaPrefix
}

// BuildRelationKey builds relation key
func BuildRelationKey(edgeID string) string {
	return fmt.Sprintf("%s/%s", RelationMetaPrefix, edgeID)
}

// BuildRelationPrefix builds relation prefix for listing
func BuildRelationPrefix(objectID string) string {
	if objectID != "" {
		return fmt.Sprintf("%s/%s", RelationMetaPrefix, objectID)
	}
	return RelationMetaPrefix
}

// BuildShareContractKey builds share contract key
func BuildShareContractKey(contractID string) string {
	return fmt.Sprintf("%s/%s", ShareContractPrefix, contractID)
}

// BuildShareContractPrefix builds share contract prefix for listing
func BuildShareContractPrefix(scope string) string {
	if scope != "" {
		return fmt.Sprintf("%s/%s", ShareContractPrefix, scope)
	}
	return ShareContractPrefix
}

// BuildMemoryIndexKey builds memory index key
func BuildMemoryIndexKey(collectionID, indexID int64) string {
	return fmt.Sprintf("%s/%d/%d", MemoryIndexPrefix, collectionID, indexID)
}

// BuildSnapshotKey builds snapshot key
func BuildSnapshotKey(objectType, objectID string, version int64) string {
	return fmt.Sprintf("%s/%s/%s/%d", SnapshotPrefix, objectType, objectID, version)
}
