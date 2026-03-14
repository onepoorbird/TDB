package event

import "fmt"

const (
	// ComponentPrefix prefix for event component
	ComponentPrefix = "event"

	// EventLogPrefix prefix for event log
	EventLogPrefix = ComponentPrefix + "/log"

	// EventMetaPrefix prefix for event meta
	EventMetaPrefix = ComponentPrefix + "/meta"

	// EventChannelPrefix prefix for event channel
	EventChannelPrefix = ComponentPrefix + "/channel"

	// EventSubscriberPrefix prefix for event subscriber
	EventSubscriberPrefix = ComponentPrefix + "/subscriber"

	// EventPositionPrefix prefix for event position
	EventPositionPrefix = ComponentPrefix + "/position"
)

// BuildEventLogKey builds event log key
func BuildEventLogKey(channelName string, logID uint64) string {
	return fmt.Sprintf("%s/%s/%d", EventLogPrefix, channelName, logID)
}

// BuildEventLogPrefix builds event log prefix
func BuildEventLogPrefix(channelName string) string {
	if channelName != "" {
		return fmt.Sprintf("%s/%s", EventLogPrefix, channelName)
	}
	return EventLogPrefix
}

// BuildEventMetaKey builds event meta key
func BuildEventMetaKey(eventID string) string {
	return fmt.Sprintf("%s/%s", EventMetaPrefix, eventID)
}

// BuildEventChannelKey builds event channel key
func BuildEventChannelKey(channelName string) string {
	return fmt.Sprintf("%s/%s", EventChannelPrefix, channelName)
}

// BuildEventSubscriberKey builds event subscriber key
func BuildEventSubscriberKey(subscriberID string) string {
	return fmt.Sprintf("%s/%s", EventSubscriberPrefix, subscriberID)
}

// BuildEventPositionKey builds event position key
func BuildEventPositionKey(subscriberID, channelName string) string {
	return fmt.Sprintf("%s/%s/%s", EventPositionPrefix, subscriberID, channelName)
}
