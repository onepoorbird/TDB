package eventcoord

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/milvus-io/milvus/internal/metastore/kv/event"
	"github.com/milvus-io/milvus/internal/util/sessionutil"
	"github.com/milvus-io/milvus/pkg/v2/kv"
	"github.com/milvus-io/milvus/pkg/v2/log"
	"github.com/milvus-io/milvus/pkg/v2/models"
	"github.com/milvus-io/milvus/pkg/v2/util/typeutil"
)

// StateCode represents the state of the EventCoord.
type StateCode int32

const (
	StateCode_Initializing StateCode = 0
	StateCode_Healthy      StateCode = 1
	StateCode_Abnormal     StateCode = 2
	StateCode_Stopping     StateCode = 3
)

// EventCoord manages event logs, channels, and subscribers.
type EventCoord struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Dependencies
	catalog      *event.Catalog
	session      sessionutil.SessionInterface
	tsoAllocator typeutil.TimestampAllocator

	// State
	stateCode atomic.Int32
	initOnce  sync.Once
	startOnce sync.Once
	stopOnce  sync.Once

	// Background tasks
	ticker *time.Ticker

	// In-memory cache for active channels
	activeChannels sync.Map
}

// NewEventCoord creates a new EventCoord instance.
func NewEventCoord(ctx context.Context, metaKV kv.TxnKV, snapshotKV kv.SnapShotKV, session sessionutil.SessionInterface, tsoAllocator typeutil.TimestampAllocator) (*EventCoord, error) {
	ctx, cancel := context.WithCancel(ctx)

	catalog := event.NewCatalog(metaKV, snapshotKV)

	ec := &EventCoord{
		ctx:            ctx,
		cancel:         cancel,
		catalog:        catalog,
		session:        session,
		tsoAllocator:   tsoAllocator,
		ticker:         time.NewTicker(5 * time.Second),
	}

	ec.UpdateStateCode(StateCode_Initializing)
	return ec, nil
}

// UpdateStateCode updates the state code.
func (ec *EventCoord) UpdateStateCode(code StateCode) {
	ec.stateCode.Store(int32(code))
	log.Ctx(ec.ctx).Info("update eventcoord state", zap.String("state", code.String()))
}

// GetStateCode returns the current state code.
func (ec *EventCoord) GetStateCode() StateCode {
	return StateCode(ec.stateCode.Load())
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

// Init initializes the EventCoord.
func (ec *EventCoord) Init() error {
	var err error
	ec.initOnce.Do(func() {
		err = ec.init()
	})
	return err
}

func (ec *EventCoord) init() error {
	log.Ctx(ec.ctx).Info("EventCoord initializing")

	// TODO: Recover state from catalog if needed
	// TODO: Initialize event index

	ec.UpdateStateCode(StateCode_Healthy)
	log.Ctx(ec.ctx).Info("EventCoord initialized")
	return nil
}

// Start starts the EventCoord background tasks.
func (ec *EventCoord) Start() error {
	var err error
	ec.startOnce.Do(func() {
		err = ec.start()
	})
	return err
}

func (ec *EventCoord) start() error {
	log.Ctx(ec.ctx).Info("EventCoord starting")

	ec.wg.Add(1)
	go ec.backgroundTask()

	log.Ctx(ec.ctx).Info("EventCoord started")
	return nil
}

// Stop stops the EventCoord.
func (ec *EventCoord) Stop() error {
	var err error
	ec.stopOnce.Do(func() {
		err = ec.stop()
	})
	return err
}

func (ec *EventCoord) stop() error {
	log.Ctx(ec.ctx).Info("EventCoord stopping")
	ec.UpdateStateCode(StateCode_Stopping)

	ec.ticker.Stop()
	ec.cancel()
	ec.wg.Wait()

	log.Ctx(ec.ctx).Info("EventCoord stopped")
	return nil
}

func (ec *EventCoord) backgroundTask() {
	defer ec.wg.Done()

	for {
		select {
		case <-ec.ctx.Done():
			return
		case <-ec.ticker.C:
			ec.doBackgroundTasks()
		}
	}
}

func (ec *EventCoord) doBackgroundTasks() {
	// TODO: Implement background tasks
	// - Channel health check
	// - Subscriber timeout check
	// - Position cleanup
	// - Event retention policy enforcement
}

// getTimestamp returns the current timestamp.
func (ec *EventCoord) getTimestamp() (typeutil.Timestamp, error) {
	if ec.tsoAllocator != nil {
		return ec.tsoAllocator.AllocOne(ec.ctx)
	}
	return typeutil.Timestamp(time.Now().UnixNano()), nil
}

// ==================== Event Log Operations ====================

// AppendEvent appends an event to the event log.
func (ec *EventCoord) AppendEvent(ctx context.Context, channelName string, agentID, sessionID string, eventType models.EventType, payload []byte, parentEventID string, causalRefs []string, source string, importance float32, visibility string, metadata map[string]string) (*models.Event, uint64, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, 0, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to get timestamp")
	}

	eventID := generateEventID()
	event := &models.Event{
		EventID:       eventID,
		AgentID:       agentID,
		SessionID:     sessionID,
		EventType:     eventType,
		EventTime:     uint64(ts),
		IngestTime:    uint64(ts),
		VisibleTime:   uint64(ts),
		LogicalTs:     uint64(ts),
		ParentEventID: parentEventID,
		CausalRefs:    causalRefs,
		Payload:       payload,
		Source:        source,
		Importance:    importance,
		Visibility:    visibility,
		Version:       1,
		Metadata:      metadata,
	}

	logID, err := ec.catalog.AppendEvent(ctx, channelName, event, ts)
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to append event")
	}

	log.Ctx(ctx).Info("event appended", zap.String("eventID", eventID), zap.String("channel", channelName), zap.Uint64("logID", logID))
	return event, logID, nil
}

// GetEvent retrieves an event by log ID.
func (ec *EventCoord) GetEvent(ctx context.Context, channelName string, logID uint64) (*models.Event, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	event, err := ec.catalog.GetEvent(ctx, channelName, logID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get event")
	}

	return event, nil
}

// GetEvents retrieves events in a range.
func (ec *EventCoord) GetEvents(ctx context.Context, channelName string, startLogID, endLogID uint64) ([]*models.Event, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	events, err := ec.catalog.GetEvents(ctx, channelName, startLogID, endLogID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get events")
	}

	return events, nil
}

// ListEventsByTimeRange lists events within a time range.
func (ec *EventCoord) ListEventsByTimeRange(ctx context.Context, channelName string, startTime, endTime uint64) ([]*models.Event, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	events, err := ec.catalog.ListEventsByTimeRange(ctx, channelName, startTime, endTime, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list events by time range")
	}

	return events, nil
}

// QueryEvents queries events based on a filter.
func (ec *EventCoord) QueryEvents(ctx context.Context, channelName string, filter *models.EventFilter, limit int64) ([]*models.Event, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	events, err := ec.catalog.QueryEvents(ctx, channelName, filter, limit, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query events")
	}

	return events, nil
}

// AppendEventBatch appends a batch of events.
func (ec *EventCoord) AppendEventBatch(ctx context.Context, channelName string, events []*models.Event) ([]uint64, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	logIDs, err := ec.catalog.AppendEventBatch(ctx, channelName, events, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to append event batch")
	}

	log.Ctx(ctx).Info("event batch appended", zap.String("channel", channelName), zap.Int("count", len(events)))
	return logIDs, nil
}

// ==================== Event Meta Operations ====================

// SaveEventMeta saves event metadata.
func (ec *EventCoord) SaveEventMeta(ctx context.Context, eventID string, meta map[string]string) error {
	if ec.GetStateCode() != StateCode_Healthy {
		return errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	if err := ec.catalog.SaveEventMeta(ctx, eventID, meta, ts); err != nil {
		return errors.Wrap(err, "failed to save event meta")
	}

	log.Ctx(ctx).Info("event meta saved", zap.String("eventID", eventID))
	return nil
}

// GetEventMeta retrieves event metadata.
func (ec *EventCoord) GetEventMeta(ctx context.Context, eventID string) (map[string]string, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	meta, err := ec.catalog.GetEventMeta(ctx, eventID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get event meta")
	}

	return meta, nil
}

// ==================== Channel Operations ====================

// CreateChannel creates a new event channel.
func (ec *EventCoord) CreateChannel(ctx context.Context, channelName string) error {
	if ec.GetStateCode() != StateCode_Healthy {
		return errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	if err := ec.catalog.CreateChannel(ctx, channelName, ts); err != nil {
		return errors.Wrap(err, "failed to create channel")
	}

	ec.activeChannels.Store(channelName, true)
	log.Ctx(ctx).Info("channel created", zap.String("channel", channelName))
	return nil
}

// GetChannel retrieves channel information.
func (ec *EventCoord) GetChannel(ctx context.Context, channelName string) (map[string]interface{}, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	channelInfo, err := ec.catalog.GetChannel(ctx, channelName, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get channel")
	}

	return channelInfo, nil
}

// ListChannels lists all event channels.
func (ec *EventCoord) ListChannels(ctx context.Context) ([]string, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	channels, err := ec.catalog.ListChannels(ctx, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list channels")
	}

	return channels, nil
}

// DeleteChannel deletes an event channel.
func (ec *EventCoord) DeleteChannel(ctx context.Context, channelName string) error {
	if ec.GetStateCode() != StateCode_Healthy {
		return errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	// Remove channel key
	key := event.BuildEventChannelKey(channelName)
	if err := ec.catalog.MultiSaveAndRemove(ctx, nil, []string{key}, ts); err != nil {
		return errors.Wrap(err, "failed to delete channel")
	}

	ec.activeChannels.Delete(channelName)
	log.Ctx(ctx).Info("channel deleted", zap.String("channel", channelName))
	return nil
}

// ==================== Subscriber Operations ====================

// RegisterSubscriber registers a new event subscriber.
func (ec *EventCoord) RegisterSubscriber(ctx context.Context, subscriberID string, filter *models.EventFilter) error {
	if ec.GetStateCode() != StateCode_Healthy {
		return errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	if err := ec.catalog.RegisterSubscriber(ctx, subscriberID, filter, ts); err != nil {
		return errors.Wrap(err, "failed to register subscriber")
	}

	log.Ctx(ctx).Info("subscriber registered", zap.String("subscriberID", subscriberID))
	return nil
}

// GetSubscriber retrieves subscriber information.
func (ec *EventCoord) GetSubscriber(ctx context.Context, subscriberID string) (map[string]interface{}, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	subscriberInfo, err := ec.catalog.GetSubscriber(ctx, subscriberID, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get subscriber")
	}

	return subscriberInfo, nil
}

// UnregisterSubscriber unregisters an event subscriber.
func (ec *EventCoord) UnregisterSubscriber(ctx context.Context, subscriberID string) error {
	if ec.GetStateCode() != StateCode_Healthy {
		return errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	if err := ec.catalog.UnregisterSubscriber(ctx, subscriberID, ts); err != nil {
		return errors.Wrap(err, "failed to unregister subscriber")
	}

	log.Ctx(ctx).Info("subscriber unregistered", zap.String("subscriberID", subscriberID))
	return nil
}

// ListSubscribers lists all registered subscribers.
func (ec *EventCoord) ListSubscribers(ctx context.Context) ([]string, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	prefix := event.EventSubscriberPrefix
	keys, _, err := ec.catalog.Snapshot.LoadWithPrefix(ctx, prefix, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list subscribers")
	}

	subscribers := make([]string, 0, len(keys))
	for _, key := range keys {
		subscriberID := key[len(prefix)+1:]
		subscribers = append(subscribers, subscriberID)
	}

	return subscribers, nil
}

// ==================== Position Operations ====================

// SavePosition saves the current position of a subscriber.
func (ec *EventCoord) SavePosition(ctx context.Context, subscriberID, channelName string, position *models.EventLogPosition) error {
	if ec.GetStateCode() != StateCode_Healthy {
		return errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return errors.Wrap(err, "failed to get timestamp")
	}

	if err := ec.catalog.SavePosition(ctx, subscriberID, channelName, position, ts); err != nil {
		return errors.Wrap(err, "failed to save position")
	}

	log.Ctx(ctx).Debug("position saved", zap.String("subscriberID", subscriberID), zap.String("channel", channelName), zap.Uint64("logID", position.LogID))
	return nil
}

// GetPosition retrieves the current position of a subscriber.
func (ec *EventCoord) GetPosition(ctx context.Context, subscriberID, channelName string) (*models.EventLogPosition, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	ts, err := ec.getTimestamp()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get timestamp")
	}

	position, err := ec.catalog.GetPosition(ctx, subscriberID, channelName, ts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get position")
	}

	return position, nil
}

// GetCurrentLogID gets the current log ID for a channel.
func (ec *EventCoord) GetCurrentLogID(ctx context.Context, channelName string) (uint64, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return 0, errors.New("EventCoord is not healthy")
	}

	logID, err := ec.catalog.GetCurrentLogID(ctx, channelName)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get current log ID")
	}

	return logID, nil
}

// ==================== Event Stream Operations ====================

// Subscribe creates a subscription to an event channel.
func (ec *EventCoord) Subscribe(ctx context.Context, subscriberID, channelName string, filter *models.EventFilter, startPosition *models.EventLogPosition) (chan *models.Event, error) {
	if ec.GetStateCode() != StateCode_Healthy {
		return nil, errors.New("EventCoord is not healthy")
	}

	// Register subscriber if not exists
	_, err := ec.GetSubscriber(ctx, subscriberID)
	if err != nil {
		if err := ec.RegisterSubscriber(ctx, subscriberID, filter); err != nil {
			return nil, errors.Wrap(err, "failed to register subscriber")
		}
	}

	// Save initial position if provided
	if startPosition != nil {
		if err := ec.SavePosition(ctx, subscriberID, channelName, startPosition); err != nil {
			return nil, errors.Wrap(err, "failed to save initial position")
		}
	}

	// Create event channel for streaming
	eventCh := make(chan *models.Event, 100)

	// Start goroutine to stream events
	ec.wg.Add(1)
	go ec.streamEvents(ctx, subscriberID, channelName, filter, eventCh)

	log.Ctx(ctx).Info("subscription created", zap.String("subscriberID", subscriberID), zap.String("channel", channelName))
	return eventCh, nil
}

// Unsubscribe closes a subscription.
func (ec *EventCoord) Unsubscribe(ctx context.Context, subscriberID, channelName string) error {
	// Position is already saved during streaming, just log here
	log.Ctx(ctx).Info("subscription closed", zap.String("subscriberID", subscriberID), zap.String("channel", channelName))
	return nil
}

// streamEvents streams events to the subscriber.
func (ec *EventCoord) streamEvents(ctx context.Context, subscriberID, channelName string, filter *models.EventFilter, eventCh chan<- *models.Event) {
	defer ec.wg.Done()
	defer close(eventCh)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ec.ctx.Done():
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Get current position
			position, err := ec.GetPosition(ec.ctx, subscriberID, channelName)
			if err != nil {
				log.Ctx(ec.ctx).Warn("failed to get position", zap.Error(err))
				continue
			}

			var startLogID uint64 = 1
			if position != nil {
				startLogID = position.LogID + 1
			}

			// Get current log ID
			currentLogID, err := ec.GetCurrentLogID(ec.ctx, channelName)
			if err != nil {
				log.Ctx(ec.ctx).Warn("failed to get current log ID", zap.Error(err))
				continue
			}

			if startLogID > currentLogID {
				// No new events
				continue
			}

			// Fetch new events
			events, err := ec.GetEvents(ec.ctx, channelName, startLogID, currentLogID)
			if err != nil {
				log.Ctx(ec.ctx).Warn("failed to get events", zap.Error(err))
				continue
			}

			// Send events to subscriber
			for _, event := range events {
				// Apply filter
				if !ec.matchesFilter(event, filter) {
					continue
				}

				select {
				case eventCh <- event:
					// Update position
					newPosition := &models.EventLogPosition{
						ChannelName: channelName,
						LogID:       currentLogID,
						Timestamp:   uint64(time.Now().UnixNano()),
					}
					if err := ec.SavePosition(ec.ctx, subscriberID, channelName, newPosition); err != nil {
						log.Ctx(ec.ctx).Warn("failed to save position", zap.Error(err))
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// matchesFilter checks if an event matches the filter criteria.
func (ec *EventCoord) matchesFilter(event *models.Event, filter *models.EventFilter) bool {
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

// ==================== Helper Functions ====================

func generateEventID() string {
	return typeutil.UniqueID(time.Now().UnixNano()).String()
}
