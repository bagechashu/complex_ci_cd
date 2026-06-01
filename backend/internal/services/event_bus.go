package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/repository"
	"built-and-deploy/pkg/logger"
)

// EventBus defines the interface for publishing and subscribing to events
type EventBus interface {
	// Publish sends an event to all subscribers of that topic
	Publish(topic string, event *models.ReleaseEvent)

	// Subscribe registers a channel to receive events for a topic
	// Returns an unsubscribe function to clean up the subscription
	Subscribe(topic string, eventChan chan *models.ReleaseEvent) func()
}

// SimpleEventBus is a basic in-memory implementation of EventBus
// It uses goroutines and channels for event distribution
type SimpleEventBus struct {
	mu          sync.RWMutex
	subscribers map[string][]chan *models.ReleaseEvent
	log         *logger.Logger
}

// NewSimpleEventBus creates a new SimpleEventBus
func NewSimpleEventBus(log *logger.Logger) EventBus {
	return &SimpleEventBus{
		subscribers: make(map[string][]chan *models.ReleaseEvent),
		log:         log,
	}
}

// Publish sends an event to all subscribers of that topic
func (eb *SimpleEventBus) Publish(topic string, event *models.ReleaseEvent) {
	eb.mu.RLock()
	subscribers, exists := eb.subscribers[topic]
	eb.mu.RUnlock()

	if !exists || len(subscribers) == 0 {
		return
	}

	eb.log.Debug("Publishing event to subscribers", "topic", topic, "subscriberCount", len(subscribers), "eventType", event.Type)

	// Send event to all subscribers in non-blocking way
	for i, ch := range subscribers {
		go func(index int, channel chan *models.ReleaseEvent) {
			select {
			case channel <- event:
				// Event sent successfully
			default:
				// Channel full, skip this subscriber to avoid blocking
				eb.log.Warn("Event channel full, skipping subscriber", "topic", topic, "index", index)
			}
		}(i, ch)
	}
}

// Subscribe registers a channel to receive events for a topic
func (eb *SimpleEventBus) Subscribe(topic string, eventChan chan *models.ReleaseEvent) func() {
	eb.mu.Lock()
	eb.subscribers[topic] = append(eb.subscribers[topic], eventChan)
	eb.log.Debug("New subscriber registered", "topic", topic, "totalSubscribers", len(eb.subscribers[topic]))
	eb.mu.Unlock()

	// Return unsubscribe function
	return func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()

		subscribers := eb.subscribers[topic]
		for i, ch := range subscribers {
			if ch == eventChan {
				// Remove from slice
				eb.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
				close(eventChan)
				eb.log.Debug("Subscriber unregistered", "topic", topic, "remainingSubscribers", len(eb.subscribers[topic]))
				return
			}
		}
	}
}

// ReleaseEventPublisher handles publishing release events
type ReleaseEventPublisher struct {
	eventRepo repository.ReleaseEventRepository
	eventBus  EventBus
	log       *logger.Logger
}

// NewReleaseEventPublisher creates a new ReleaseEventPublisher
func NewReleaseEventPublisher(eventRepo repository.ReleaseEventRepository, eventBus EventBus, log *logger.Logger) *ReleaseEventPublisher {
	return &ReleaseEventPublisher{
		eventRepo: eventRepo,
		eventBus:  eventBus,
		log:       log,
	}
}

// PublishEvent saves the event to database and publishes it to subscribers
func (rep *ReleaseEventPublisher) PublishEvent(ctx context.Context, releaseID int, eventType, message string, details interface{}) error {
	// Convert details to JSON string
	var detailsJSON string
	if details != nil {
		data, err := json.Marshal(details)
		if err != nil {
			rep.log.Warn("Failed to marshal event details", "error", err)
			detailsJSON = "{}"
		} else {
			detailsJSON = string(data)
		}
	} else {
		detailsJSON = "{}"
	}

	// Create event model
	event := &models.ReleaseEvent{
		ReleaseID: releaseID,
		Type:      eventType,
		Message:   message,
		Details:   detailsJSON,
	}

	// Save to database
	if err := rep.eventRepo.Create(ctx, event); err != nil {
		rep.log.Error("Failed to save release event", "releaseID", releaseID, "eventType", eventType, "error", err)
		return err
	}

	rep.log.Info("Release event published", "releaseID", releaseID, "eventType", eventType, "message", message)

	// Publish to event bus subscribers
	topic := fmt.Sprintf("release:%d", releaseID)
	rep.eventBus.Publish(topic, event)

	return nil
}

// PublishEventAsync publishes event asynchronously
func (rep *ReleaseEventPublisher) PublishEventAsync(ctx context.Context, releaseID int, eventType, message string, details interface{}) {
	go func() {
		if err := rep.PublishEvent(ctx, releaseID, eventType, message, details); err != nil {
			rep.log.Error("Failed to publish event asynchronously", "releaseID", releaseID, "error", err)
		}
	}()
}
