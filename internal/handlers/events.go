package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"go.uber.org/zap"
)

type watermillEventToViewFunc func(ctx context.Context, event interface{}) (string, string, error)

func handleSSEEvents(w http.ResponseWriter, r *http.Request, logger *zap.Logger, eventChan <-chan interface{}, eventToView watermillEventToViewFunc) {
	// Write the headers for the SSE response
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	rc := http.NewResponseController(w)

	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-eventChan:
			eventType, eventData, err := eventToView(r.Context(), event)
			if err != nil {
				logger.Error("Failed to convert event to view", zap.Any("event", event), zap.Error(err))
				continue
			}

			// If there is no event but no error we can just skip this event for the next one
			if eventType == "" {
				continue
			}

			if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventType, strings.ReplaceAll(eventData, "\n", "\ndata: ")); err != nil {
				logger.Error("Failed to write event to response", zap.Any("event", event), zap.Error(err))
				return
			}

			if err := rc.Flush(); err != nil {
				logger.Error("Failed to flush response", zap.Any("event", event), zap.Error(err))
				return
			}
		}
	}

}

func renderComponent(ctx context.Context, eventType string, c templ.Component) (string, string, error) {
	content, err := templateToString(ctx, c)
	if err != nil {
		return "", "", err
	}
	return eventType, content, nil
}

func templateToString(ctx context.Context, c templ.Component) (string, error) {
	var rendered bytes.Buffer
	context := context.WithoutCancel(ctx)
	if err := c.Render(context, &rendered); err != nil {
		return "", err
	}
	return rendered.String(), nil
}
