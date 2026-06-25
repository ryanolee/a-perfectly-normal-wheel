package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/components"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/services"
	"go.uber.org/zap"
)

type (
	WheelEventsHandler struct {
		wheelService       WheelService
		wheelEventsService WheelEventsService
		sessionService     SessionService
		candidateService   CandidateService
		mux                *http.ServeMux
		logger             *zap.Logger
	}
)

func NewWheelEventsHandler(wheelService WheelService, wheelEventsService WheelEventsService, sessionService SessionService, candidateService CandidateService, logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()
	we := &WheelEventsHandler{
		wheelService:       wheelService,
		wheelEventsService: wheelEventsService,
		candidateService:   candidateService,
		sessionService:     sessionService,
		mux:                mux,
		logger:             logger,
	}

	we.mux.HandleFunc("GET /api/wheel/{id}/events", we.HandleSpecificWheelEvents)
	we.mux.HandleFunc("GET /api/wheel/events", we.HandleGlobalWheelEvents)
	return we
}

func (h *WheelEventsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *WheelEventsHandler) HandleGlobalWheelEvents(w http.ResponseWriter, r *http.Request) {
	eventChan, err := h.wheelEventsService.SubscribeToGlobalWheelEvents(r.Context())
	if err != nil {
		h.logger.Error("Failed to subscribe to global wheel events", zap.Error(err))
		components.StatusError("Failed to subscribe to global wheel events").Render(r.Context(), w)
		return
	}

	handleSSEEvents(w, r, h.logger, eventChan, h.EventToView)
}

func (h *WheelEventsHandler) HandleSpecificWheelEvents(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		components.StatusError("Invalid wheel ID").Render(r.Context(), w)
		return
	}

	if _, err := h.wheelService.GetWheelByID(r.Context(), id); err != nil {
		components.StatusError("Wheel not found").Render(r.Context(), w)
		return
	}

	eventChan, err := h.wheelEventsService.SubscribeToWheelEvents(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to subscribe to wheel events", zap.Int64("wheel_id", id), zap.Error(err))
		components.StatusError("Failed to subscribe to wheel events").Render(r.Context(), w)
		return
	}

	logger := h.logger.With(zap.Int64("wheel_id", id))
	handleSSEEvents(w, r, logger, eventChan, h.EventToView)
}

func (s *WheelEventsHandler) EventToView(ctx context.Context, event interface{}) (string, string, error) {
	switch e := event.(type) {
	case services.NewCandidateAddedToWheelEvent:
		return s.renderCandidateAdded(ctx, e)
	case services.CandidateRemovedFromWheelEvent:
		return s.renderCandidateRemoved(ctx, e)
	case services.StatusChangedEvent:
		return s.renderStatusChanged(ctx, e)
	case services.WheelDeletedEvent:
		return s.renderWheelDeleted(ctx, e)
	case services.WheelAddedEvent:
		return s.renderWheelAdded(ctx, e)
	case services.WinnerDeclaredEvent:
		return s.renderWinnerDeclared(ctx, e)
	default:
		return "", "", fmt.Errorf("unknown event type: %T", e)
	}
}

func (s *WheelEventsHandler) renderWinnerDeclared(ctx context.Context, e services.WinnerDeclaredEvent) (string, string, error) {
	candidates, err := s.candidateService.ListCandidatesByWheel(ctx, e.WheelID)
	if err != nil {
		return "", "", err
	}

	candidateNames := make([]string, len(candidates))
	for i, candidate := range candidates {
		candidateNames[i] = candidate.Username
	}
	return renderComponent(ctx, services.WinnerDeclaredEventType, components.SpinningWheel(candidateNames, e.Winner.Username))
}

func (s *WheelEventsHandler) renderWheelAdded(ctx context.Context, e services.WheelAddedEvent) (string, string, error) {
	return renderComponent(ctx, services.WheelAddedEventType, components.WheelCard(e.Wheel))
}

func (s *WheelEventsHandler) renderCandidateAdded(ctx context.Context, e services.NewCandidateAddedToWheelEvent) (string, string, error) {
	sessionId, _ := s.sessionService.GetSessionIdFromContext(ctx)
	return renderComponent(ctx, services.NewCandidateAddedToWheelEventType, components.CandidateListItem(e.Candidate, sessionId))
}

func (s *WheelEventsHandler) renderCandidateRemoved(_ context.Context, e services.CandidateRemovedFromWheelEvent) (string, string, error) {
	return fmt.Sprintf("%s:%d", services.CandidateRemovedFromWheelEventType, e.CandidateID), "", nil
}

func (s *WheelEventsHandler) renderStatusChanged(ctx context.Context, e services.StatusChangedEvent) (string, string, error) {
	if services.ParseWheelStatus(e.Status) == services.WheelStatusLocked {
		return renderComponent(ctx, services.WheelStatusChangedEventType, components.WheelLockedNotice())
	}

	sessionId, _ := s.sessionService.GetSessionIdFromContext(ctx)
	candidates, err := s.candidateService.ListCandidatesByWheel(ctx, e.WheelID)
	if err != nil {
		return "", "", err
	}

	if !s.candidateService.CandidateInCandidateList(sessionId, candidates) {
		return renderComponent(ctx, services.WheelStatusChangedEventType, components.CandidateSubmissionForm(e.WheelID, ""))
	}

	return renderComponent(ctx, services.WheelStatusChangedEventType, components.WheelStatusStub())
}

func (s *WheelEventsHandler) renderWheelDeleted(ctx context.Context, e services.WheelDeletedEvent) (string, string, error) {
	return renderComponent(ctx, fmt.Sprintf("%s:%d", services.WheelDeletedEventType, e.WheelID), components.StatusError("Sorry, someone rudely decided to delete this wheel."))
}
