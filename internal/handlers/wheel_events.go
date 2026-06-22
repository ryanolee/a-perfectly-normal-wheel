package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/components"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/services"
	"go.uber.org/zap"
)

type (
	WheelEventsHandler struct {
		wheelService       WheelService
		wheelEventsService WheelEventsService
		sessionService     SessionService
		logger             *zap.Logger
	}
)

func NewWheelEventsHandler(wheelService WheelService, wheelEventsService WheelEventsService, sessionService SessionService, logger *zap.Logger) http.Handler {
	return &WheelEventsHandler{
		wheelService:       wheelService,
		wheelEventsService: wheelEventsService,
		sessionService:     sessionService,
		logger:             logger,
	}
}

func (h *WheelEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		components.StatusError("Invalid wheel ID").Render(r.Context(), w)
		return
	}

	if _, err := h.wheelService.GetWheelByID(r.Context(), id); err != nil {
		components.StatusError("Wheel not found").Render(r.Context(), w)
		return
	}

	// Write the headers for the SSE response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)

	eventChan, err := h.wheelEventsService.SubscribeToWheelEvents(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to subscribe to wheel events", zap.Int64("wheel_id", id), zap.Error(err))
		components.StatusError("Failed to subscribe to wheel events").Render(r.Context(), w)
		return
	}

	wheelLogger := h.logger.With(zap.Int64("wheel_id", id))
	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-eventChan:
			eventType, eventData, err := h.EventToView(r.Context(), event)
			if err != nil {
				wheelLogger.Error("Failed to convert event to view", zap.Any("event", event), zap.Error(err))
				continue
			}

			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, strings.ReplaceAll(eventData, "\n", "\ndata: ")); err != nil {
				wheelLogger.Error("Failed to write event to response", zap.Any("event", event), zap.Error(err))
				return
			}

			if err := rc.Flush(); err != nil {
				wheelLogger.Error("Failed to flush response", zap.Any("event", event), zap.Error(err))
				return
			}
		}
	}

}

func (s *WheelEventsHandler) EventToView(ctx context.Context, event interface{}) (string, string, error) {

	switch e := event.(type) {
	case services.NewCandidateAddedToWheelEvent:
		sessionId, _ := s.sessionService.GetSessionIdFromContext(ctx)
		content, err := templateToString(ctx, components.CandidateListItem(e.Candidate, sessionId))
		if err != nil {
			return "", "", err
		}
		return services.NewCandidateAddedToWheelEventType, content, nil
	default:
		return "", "", fmt.Errorf("unknown event type: %T", e)
	}
}

func templateToString(ctx context.Context, c templ.Component) (string, error) {
	var rendered bytes.Buffer
	context := context.WithoutCancel(ctx)
	if err := c.Render(context, &rendered); err != nil {
		return "", err
	}
	return rendered.String(), nil
}
