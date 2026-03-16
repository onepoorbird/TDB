package event

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cockroachdb/errors"
	"go.uber.org/zap"

	"github.com/milvus-io/milvus/pkg/v2/kv"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

// ErrKeyNotFound is returned when a key is not found in the catalog.
var ErrKeyNotFound = errors.New("key not found")

// Catalog provides access to event metadata stored in etcd.
type Catalog struct {
	Txn      kv.TxnKV
	Snapshot kv.SnapShotKV
}

// NewCatalog creates a new Catalog instance.
func NewCatalog(metaKV kv.TxnKV, ss kv.SnapShotKV) *Catalog {
	return &Catalog{Txn: metaKV, Snapshot: ss}
}

// ==================== Event Log Operations ====================

// AppendEvent appends an event to the event log.
func (c *Catalog) AppendEvent(ctx context.Context, channelName string, event *models.Event, ts typeutil.Timestamp) (uint64, error) {
	// Get next log ID
	logID, err := c.GetNextLogID(ctx, channelName)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get next log ID")
	}

	key := BuildEventLogKey(channelName, logID)
	value, err := json.Marshal(event)
	if err != nil {
		return 0, errors.Wrap(err, "failed to marshal event")
	}

	if err := c.Snapshot.Save(ctx, key, string(value), ts); err != nil {
		return 0, errors.Wrap(err, "failed to save event")
	}

	return logID, nil
}

// GetEvent retrieves an event by log ID.
func (c *Catalog) GetEvent(ctx context.Context, channelName string, logID uint64, ts typeutil.Timestamp) (*models.Event, error) {
	key := BuildEventLogKey(channelName, logID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil, errors.Errorf("event not found: channel=%s, logID=%d", channelName, logID)
		}
		return nil, errors.Wrap(err, "failed to load event")
	}

	var event models.Event
	if err := json.Unmarshal([]byte(value), &event); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal event")
	}
	return &event, nil
}

// GetEvents retrieves events in a range.
func (c *Catalog) GetEvents(ctx context.Context, channelName string, startLogID, endLogID uint64, ts typeutil.Timestamp) ([]*models.Event, error) {
	events := make([]*models.Event, 0)
	for logID := startLogID; logID <= endLogID; logID++ {
		event, err := c.GetEvent(ctx, channelName, logID, ts)
		if err != nil {
			if errors.Is(err, ErrKeyNotFound) {
				continue
			}
			return nil, err
		}
		events = append(events, event)
	}
	return events, nil
}

// ListEventsByTimeRange lists events within a time range.
func (c *Catalog) ListEventsByTimeRange(ctx context.Context, channelName string, startTime, endTime uint64, ts typeutil.Timestamp) ([]*models.Event, error) {
	prefix := BuildEventLogPrefix(channelName)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list events")
	}

	events := make([]*models.Event, 0, len(values))
	for _, value := range values {
		var event models.Event
		if err := json.Unmarshal([]byte(value), &event); err != nil {
			log.Warn("failed to unmarshal event", zap.Error(err))
			continue
		}
		if event.EventTime >= startTime && event.EventTime <= endTime {
			events = append(events, &event)
		}
	}
	return events, nil
}

// ==================== Event Meta Operations ====================

// SaveEventMeta saves event metadata.
func (c *Catalog) SaveEventMeta(ctx context.Context, eventID string, meta map[string]string, ts typeutil.Timestamp) error {
	key := BuildEventMetaKey(eventID)
	value, err := json.Marshal(meta)
	if err != nil {
		return errors.Wrap(err, "failed to marshal event meta")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetEventMeta retrieves event metadata.
func (c *Catalog) GetEventMeta(ctx context.Context, eventID string, ts typeutil.Timestamp) (map[string]string, error) {
	key := BuildEventMetaKey(eventID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to load event meta")
	}

	var meta map[string]string
	if err := json.Unmarshal([]byte(value), &meta); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal event meta")
	}
	return meta, nil
}

// ==================== Channel Operations ====================

// CreateChannel creates a new event channel.
func (c *Catalog) CreateChannel(ctx context.Context, channelName string, ts typeutil.Timestamp) error {
	key := BuildEventChannelKey(channelName)
	channelInfo := map[string]interface{}{
		"channel_name": channelName,
		"created_ts":   ts,
		"status":       "active",
	}
	value, err := json.Marshal(channelInfo)
	if err != nil {
		return errors.Wrap(err, "failed to marshal channel info")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetChannel retrieves channel information.
func (c *Catalog) GetChannel(ctx context.Context, channelName string, ts typeutil.Timestamp) (map[string]interface{}, error) {
	key := BuildEventChannelKey(channelName)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil, errors.Errorf("channel not found: %s", channelName)
		}
		return nil, errors.Wrap(err, "failed to load channel")
	}

	var channelInfo map[string]interface{}
	if err := json.Unmarshal([]byte(value), &channelInfo); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal channel info")
	}
	return channelInfo, nil
}

// ListChannels lists all event channels.
func (c *Catalog) ListChannels(ctx context.Context, ts typeutil.Timestamp) ([]string, error) {
	prefix := EventChannelPrefix
	keys, _, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list channels")
	}

	channels := make([]string, 0, len(keys))
	for _, key := range keys {
		channelName := key[len(prefix)+1:]
		channels = append(channels, channelName)
	}
	return channels, nil
}

// ==================== Subscriber Operations ====================

// RegisterSubscriber registers a new event subscriber.
func (c *Catalog) RegisterSubscriber(ctx context.Context, subscriberID string, filter *models.EventFilter, ts typeutil.Timestamp) error {
	key := BuildEventSubscriberKey(subscriberID)
	subscriberInfo := map[string]interface{}{
		"subscriber_id": subscriberID,
		"filter":        filter,
		"registered_ts": ts,
		"status":        "active",
	}
	value, err := json.Marshal(subscriberInfo)
	if err != nil {
		return errors.Wrap(err, "failed to marshal subscriber info")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetSubscriber retrieves subscriber information.
func (c *Catalog) GetSubscriber(ctx context.Context, subscriberID string, ts typeutil.Timestamp) (map[string]interface{}, error) {
	key := BuildEventSubscriberKey(subscriberID)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil, errors.Errorf("subscriber not found: %s", subscriberID)
		}
		return nil, errors.Wrap(err, "failed to load subscriber")
	}

	var subscriberInfo map[string]interface{}
	if err := json.Unmarshal([]byte(value), &subscriberInfo); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal subscriber info")
	}
	return subscriberInfo, nil
}

// UnregisterSubscriber unregisters an event subscriber.
func (c *Catalog) UnregisterSubscriber(ctx context.Context, subscriberID string, ts typeutil.Timestamp) error {
	key := BuildEventSubscriberKey(subscriberID)
	return c.Snapshot.MultiSaveAndRemove(ctx, nil, []string{key}, ts)
}

// ==================== Position Operations ====================

// SavePosition saves the current position of a subscriber.
func (c *Catalog) SavePosition(ctx context.Context, subscriberID, channelName string, position *models.EventLogPosition, ts typeutil.Timestamp) error {
	key := BuildEventPositionKey(subscriberID, channelName)
	value, err := json.Marshal(position)
	if err != nil {
		return errors.Wrap(err, "failed to marshal position")
	}
	return c.Snapshot.Save(ctx, key, string(value), ts)
}

// GetPosition retrieves the current position of a subscriber.
func (c *Catalog) GetPosition(ctx context.Context, subscriberID, channelName string, ts typeutil.Timestamp) (*models.EventLogPosition, error) {
	key := BuildEventPositionKey(subscriberID, channelName)
	value, err := c.Snapshot.Load(ctx, key, ts)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "failed to load position")
	}

	var position models.EventLogPosition
	if err := json.Unmarshal([]byte(value), &position); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal position")
	}
	return &position, nil
}

// ==================== Log ID Management ====================

// GetNextLogID gets the next log ID for a channel.
func (c *Catalog) GetNextLogID(ctx context.Context, channelName string) (uint64, error) {
	key := fmt.Sprintf("%s/%s/next_log_id", EventLogPrefix, channelName)
	value, err := c.Txn.Load(ctx, key)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			// First event, start from 1
			if err := c.Txn.Save(ctx, key, "1"); err != nil {
				return 0, errors.Wrap(err, "failed to initialize log ID")
			}
			return 1, nil
		}
		return 0, errors.Wrap(err, "failed to load log ID")
	}

	logID, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse log ID")
	}

	nextLogID := logID + 1
	if err := c.Txn.Save(ctx, key, strconv.FormatUint(nextLogID, 10)); err != nil {
		return 0, errors.Wrap(err, "failed to update log ID")
	}

	return nextLogID, nil
}

// GetCurrentLogID gets the current log ID for a channel.
func (c *Catalog) GetCurrentLogID(ctx context.Context, channelName string) (uint64, error) {
	key := fmt.Sprintf("%s/%s/next_log_id", EventLogPrefix, channelName)
	value, err := c.Txn.Load(ctx, key)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) {
			return 0, nil
		}
		return 0, errors.Wrap(err, "failed to load log ID")
	}

	logID, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, errors.Wrap(err, "failed to parse log ID")
	}

	return logID, nil
}

// ==================== Batch Operations ====================

// AppendEventBatch appends a batch of events.
func (c *Catalog) AppendEventBatch(ctx context.Context, channelName string, events []*models.Event, ts typeutil.Timestamp) ([]uint64, error) {
	logIDs := make([]uint64, 0, len(events))
	kvs := make(map[string]string)

	for _, event := range events {
		logID, err := c.GetNextLogID(ctx, channelName)
		if err != nil {
			return nil, errors.Wrap(err, "failed to get next log ID")
		}

		key := BuildEventLogKey(channelName, logID)
		value, err := json.Marshal(event)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal event")
		}

		kvs[key] = string(value)
		logIDs = append(logIDs, logID)
	}

	if err := c.Snapshot.MultiSave(ctx, kvs, ts); err != nil {
		return nil, errors.Wrap(err, "failed to save event batch")
	}

	return logIDs, nil
}

// ==================== Query Operations ====================

// QueryEvents queries events based on a filter.
func (c *Catalog) QueryEvents(ctx context.Context, channelName string, filter *models.EventFilter, limit int64, ts typeutil.Timestamp) ([]*models.Event, error) {
	prefix := BuildEventLogPrefix(channelName)
	_, values, err := c.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query events")
	}

	events := make([]*models.Event, 0)
	for _, value := range values {
		var event models.Event
		if err := json.Unmarshal([]byte(value), &event); err != nil {
			log.Warn("failed to unmarshal event", zap.Error(err))
			continue
		}

		// Apply filters
		if !c.matchesFilter(&event, filter) {
			continue
		}

		events = append(events, &event)

		if limit > 0 && int64(len(events)) >= limit {
			break
		}
	}

	return events, nil
}

// matchesFilter checks if an event matches the filter criteria.
func (c *Catalog) matchesFilter(event *models.Event, filter *models.EventFilter) bool {
	if filter == nil {
		return true
	}

	// Filter by agent IDs
	if len(filter.AgentIDs) > 0 {
		found := false
		for _, agentID := range filter.AgentIDs {
			if event.AgentID == agentID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by session IDs
	if len(filter.SessionIDs) > 0 {
		found := false
		for _, sessionID := range filter.SessionIDs {
			if event.SessionID == sessionID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by event types
	if len(filter.EventTypes) > 0 {
		found := false
		for _, eventType := range filter.EventTypes {
			if event.EventType == eventType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Filter by time range
	if filter.StartTime > 0 && event.EventTime < filter.StartTime {
		return false
	}
	if filter.EndTime > 0 && event.EventTime > filter.EndTime {
		return false
	}

	// Filter by importance
	if event.Importance < filter.MinImportance {
		return false
	}
	if filter.MaxImportance > 0 && event.Importance > filter.MaxImportance {
		return false
	}

	return true
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
