package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/components"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/repository"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/services"
	"go.uber.org/zap"
)

type (
	AdminApiHandler struct {
		wheelService       WheelService
		candidateService   CandidateService
		wheelEventsService WheelEventsService
		logger             *zap.Logger
		mux                *http.ServeMux
	}
)

func NewAdminApiHandler(wheelService WheelService, candidateService CandidateService, wheelEventsService WheelEventsService, logger *zap.Logger) http.Handler {
	h := &AdminApiHandler{
		wheelService:       wheelService,
		wheelEventsService: wheelEventsService,
		candidateService:   candidateService,
		logger:             logger,
		mux:                http.NewServeMux(),
	}
	h.mux.HandleFunc("DELETE /admin/api/wheel/{wheelId}/candidate/{candidateId}", h.DeleteCandidate)
	h.mux.HandleFunc("DELETE /admin/api/wheel/{wheelId}", h.DeleteWheel)
	h.mux.HandleFunc("PUT /admin/api/wheel/{wheelId}/{status}", h.SetWheelStatus)
	h.mux.HandleFunc("PUT /admin/api/wheel/{wheelId}/candidate/{candidateId}/declareWinner", h.DeclareWinner)
	h.mux.HandleFunc("POST /admin/api/wheel", h.CreateWheel)
	h.mux.HandleFunc("GET /admin/api/wheel/{wheelId}/events", h.Events)

	return h
}

func parsePathInt(w http.ResponseWriter, r *http.Request, key string, errMsg string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil {
		http.Error(w, errMsg, http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func (h *AdminApiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *AdminApiHandler) Events(w http.ResponseWriter, r *http.Request) {
	wheelId, ok := parsePathInt(w, r, "wheelId", "Invalid wheel ID")
	if !ok {
		return
	}

	eventChan, err := h.wheelEventsService.SubscribeToWheelEvents(r.Context(), wheelId)
	if err != nil {
		h.logger.Error("Failed to subscribe to global wheel events", zap.Error(err))
		components.StatusError("Failed to subscribe to global wheel events").Render(r.Context(), w)
		return
	}

	handleSSEEvents(w, r, h.logger, eventChan, func(ctx context.Context, event interface{}) (string, string, error) {
		switch e := event.(type) {
		case services.NewCandidateAddedToWheelEvent:
			componentData, err := templateToString(ctx, components.CandidatePill(wheelId, e.Candidate))
			return services.NewCandidateAddedToWheelEventType, componentData, err
		default:
			return "", "", nil
		}
	})

}

func (h *AdminApiHandler) DeclareWinner(w http.ResponseWriter, r *http.Request) {
	wheelId, ok := parsePathInt(w, r, "wheelId", "Invalid wheel ID")
	if !ok {
		return
	}

	trueRandomSeedSource, ok := parsePathInt(w, r, "candidateId", "Invalid candidate ID")
	if !ok {
		return
	}

	winningCandidate, err := h.candidateService.GetRandomCandidateForWheel(r.Context(), wheelId, trueRandomSeedSource)
	if err != nil {
		h.logger.Error("Failed to declare winner", zap.Int64("wheel_id", wheelId), zap.Int64("rand_seed", trueRandomSeedSource), zap.Error(err))
		http.Error(w, "Failed to declare winner", http.StatusInternalServerError)
		return
	}

	err = h.wheelService.DeclareWinnerForWheel(r.Context(), wheelId, winningCandidate)
	if err != nil {
		h.logger.Error("Failed to declare winner", zap.Int64("wheel_id", wheelId), zap.Int64("rand_seed", trueRandomSeedSource), zap.Error(err))
		http.Error(w, "Failed to declare winner", http.StatusInternalServerError)
		return
	}

	components.AdminWheelWinnerDeclaration(wheelId, *winningCandidate).Render(r.Context(), w)
}

func (h *AdminApiHandler) CreateWheel(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	description := r.FormValue("description")

	wheel, err := h.wheelService.CreateWheel(r.Context(), name, description)
	if err != nil {
		h.logger.Error("Failed to create wheel", zap.Error(err))
		http.Error(w, "Failed to create wheel", http.StatusInternalServerError)
		return
	}

	components.AdminWheelCard(*wheel).Render(r.Context(), w)

}

func (h *AdminApiHandler) DeleteCandidate(w http.ResponseWriter, r *http.Request) {
	wheelId, ok := parsePathInt(w, r, "wheelId", "Invalid wheel ID")
	if !ok {
		return
	}

	candidateId, ok := parsePathInt(w, r, "candidateId", "Invalid candidate ID")
	if !ok {
		return
	}

	err := h.candidateService.DeleteCandidateById(r.Context(), wheelId, candidateId)
	if err != nil {
		h.logger.Error("Failed to delete candidate", zap.Int64("wheel_id", wheelId), zap.Int64("candidate_id", candidateId), zap.Error(err))
		http.Error(w, "Failed to delete candidate", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *AdminApiHandler) DeleteWheel(w http.ResponseWriter, r *http.Request) {
	wheelId, ok := parsePathInt(w, r, "wheelId", "Invalid wheel ID")
	if !ok {
		return
	}

	err := h.wheelService.DeleteWheelByID(r.Context(), wheelId)
	if err != nil {
		h.logger.Error("Failed to delete wheel", zap.Int64("wheel_id", wheelId), zap.Error(err))
		http.Error(w, "Failed to delete wheel", http.StatusInternalServerError)
		return
	}

	components.StatusSuccess("Successfully deleted wheel", 0).Render(r.Context(), w)
}

func (h *AdminApiHandler) SetWheelStatus(w http.ResponseWriter, r *http.Request) {
	wheelId, ok := parsePathInt(w, r, "wheelId", "Invalid wheel ID")
	if !ok {
		return
	}

	status := repository.ParseWheelStatus(r.PathValue("status"))

	if status != repository.WheelStatusActive && status != repository.WheelStatusLocked {
		http.Error(w, "Cannot change to status", http.StatusBadRequest)
		return
	}

	err := h.wheelService.SetWheelStatus(r.Context(), wheelId, status)
	if err != nil {
		h.logger.Error("Failed to lock wheel", zap.Int64("wheel_id", wheelId), zap.Error(err))
		http.Error(w, "Failed to lock wheel", http.StatusInternalServerError)
		return
	}

	components.WheelLockButton(wheelId, status == repository.WheelStatusActive).Render(r.Context(), w)
}
