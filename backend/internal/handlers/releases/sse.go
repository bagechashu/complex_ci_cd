package releases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"built-and-deploy/internal/models"
	"built-and-deploy/internal/services"
	"built-and-deploy/pkg/logger"
	"built-and-deploy/pkg/responses"
)

// StreamReleaseEvents handles Server-Sent Events stream for release event tracking
//
// This endpoint streams release events in real-time using Server-Sent Events (SSE).
// When a client connects, it will receive all new events related to the release
// as they occur, without needing to poll.
//
// @Summary Stream Release Events (SSE)
// @Description Streams release events in real-time using Server-Sent Events
// @Tags Releases
// @Produce text/event-stream
// @Param id path int true "Release ID"
// @Success 200 "Event stream established"
// @Failure 400 {object} responses.ErrorResponse "Invalid release ID"
// @Failure 404 {object} responses.ErrorResponse "Release not found"
// @Failure 500 {object} responses.ErrorResponse "Internal server error"
// @Router /releases/{id}/stream [get]
//
// This endpoint:
// 1. First sends all existing events for the release
// 2. Then streams new events as they are published
// 3. Automatically closes when the release completes or on client disconnect
//
// Client-side example (JavaScript):
//
//	const eventSource = new EventSource(`/api/v1/releases/${releaseId}/stream`)
//	eventSource.onmessage = (event) => {
//	  const releaseEvent = JSON.parse(event.data)
//	  console.log('Event:', releaseEvent)
//	  if (['success', 'failed', 'rolled_back'].includes(releaseEvent.type)) {
//	    eventSource.close()
//	  }
//	}
//	eventSource.onerror = () => eventSource.close()
func StreamReleaseEvents(releaseService *services.ReleaseService, log *logger.Logger, eventBus services.EventBus) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get release ID from URL parameter
		releaseIDStr := r.PathValue("id")
		if releaseIDStr == "" {
			responses.BadRequestResponse(w, "release id is required")
			return
		}

		releaseID, err := strconv.Atoi(releaseIDStr)
		if err != nil {
			responses.BadRequestResponse(w, "invalid release id format")
			return
		}

		// Verify release exists
		release, err := releaseService.GetReleaseStatus(r.Context(), releaseID)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		if release == nil {
			responses.NotFoundResponse(w, "release not found")
			return
		}

		log.Info("SSE client connected", "releaseID", releaseID)

		// Set up SSE response headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("X-Accel-Buffering", "no") // Disable buffering for Nginx

		flusher, ok := w.(http.Flusher)
		if !ok {
			log.Error("Streaming not supported by client", "releaseID", releaseID)
			responses.InternalErrorResponse(w, "streaming not supported")
			return
		}

		// First, send all existing events for this release
		existingEvents, err := releaseService.GetReleaseEvents(r.Context(), releaseID)
		if err != nil {
			log.Error("Failed to get existing events", "releaseID", releaseID, "error", err)
			// Continue anyway, client might still get new events
		} else {
			for _, event := range existingEvents {
				sendSSEEvent(w, flusher, event, log)
			}
		}

		// Create channel for new events
		eventChan := make(chan *models.ReleaseEvent, 100)

		// Subscribe to new events
		unsubscribe := eventBus.Subscribe(fmt.Sprintf("release:%d", releaseID), eventChan)
		defer unsubscribe()

		log.Info("SSE client subscribed to events", "releaseID", releaseID)

		// Stream new events
		for {
			select {
			case <-r.Context().Done():
				// Client disconnected
				log.Info("SSE client disconnected", "releaseID", releaseID)
				return

			case event := <-eventChan:
				if event == nil {
					log.Debug("Event channel closed", "releaseID", releaseID)
					return
				}

				sendSSEEvent(w, flusher, event, log)

				// If release completed, send completion message and close
				if isReleaseCompleted(event) {
					log.Info("Release completed, closing SSE stream", "releaseID", releaseID, "status", event.Type)
					sendCommentSSE(w, flusher, "[STREAM_END] Release processing completed")
					return
				}
			}
		}
	}
}

// sendSSEEvent sends a single SSE event to the client
func sendSSEEvent(w http.ResponseWriter, flusher http.Flusher, event *models.ReleaseEvent, log *logger.Logger) {
	eventData, err := json.Marshal(event)
	if err != nil {
		log.Error("Failed to marshal event", "error", err)
		return
	}

	// Send SSE formatted data
	fmt.Fprintf(w, "id: %d\n", event.ID)
	fmt.Fprintf(w, "event: %s\n", event.Type)
	fmt.Fprintf(w, "data: %s\n\n", string(eventData))

	flusher.Flush()
}

// sendCommentSSE sends a comment-only SSE message (no data)
func sendCommentSSE(w http.ResponseWriter, flusher http.Flusher, comment string) {
	fmt.Fprintf(w, ": %s\n\n", comment)
	flusher.Flush()
}

// isReleaseCompleted checks if the event marks the release as completed
func isReleaseCompleted(event *models.ReleaseEvent) bool {
	completionTypes := map[string]bool{
		"success":     true,
		"failed":      true,
		"rolled_back": true,
		"completed":   true,
	}
	return completionTypes[event.Type]
}

// PollReleaseStatus provides a fallback polling endpoint for clients that don't support SSE
//
// This endpoint returns the current release status and recent events.
// Use this if the client doesn't support SSE or as a manual polling alternative.
//
// @Summary Poll Release Status
// @Description Returns current release status and recent events (polling alternative to SSE)
// @Tags Releases
// @Produce json
// @Param id path int true "Release ID"
// @Success 200 {object} ReleaseStatusResponse "Current status and events"
// @Failure 400 {object} responses.ErrorResponse "Invalid release ID"
// @Failure 404 {object} responses.ErrorResponse "Release not found"
// @Router /releases/{id}/status [get]
func PollReleaseStatus(releaseService *services.ReleaseService, log *logger.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		releaseIDStr := r.PathValue("id")
		if releaseIDStr == "" {
			responses.BadRequestResponse(w, "release id is required")
			return
		}

		releaseID, err := strconv.Atoi(releaseIDStr)
		if err != nil {
			responses.BadRequestResponse(w, "invalid release id format")
			return
		}

		// Get release status
		release, err := releaseService.GetReleaseStatus(r.Context(), releaseID)
		if err != nil {
			responses.InternalErrorResponse(w, err.Error())
			return
		}

		if release == nil {
			responses.NotFoundResponse(w, "release not found")
			return
		}

		// Get recent events
		events, err := releaseService.GetReleaseEvents(r.Context(), releaseID)
		if err != nil {
			log.Error("Failed to get release events", "releaseID", releaseID, "error", err)
			events = []*models.ReleaseEvent{}
		}

		response := ReleaseStatusResponse{
			Release: release,
			Events:  events,
			Time:    time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache")
		json.NewEncoder(w).Encode(response)
	}
}

// ReleaseStatusResponse combines release details with recent events
type ReleaseStatusResponse struct {
	Release *models.ReleaseRecord  `json:"release"`
	Events  []*models.ReleaseEvent `json:"events"`
	Time    time.Time              `json:"time"`
}
