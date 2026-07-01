package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/components"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/repository"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/services"
	"go.uber.org/zap"
)

var usernameRegex = regexp.MustCompile(`^([a-zA-Z0-9_-]|\s){1,50}$`)

type (
	WheelHandler struct {
		wheelService     WheelService
		viteService      ViteService
		candidateService CandidateService
		sessionService   SessionService
		logger           *zap.Logger
	}
)

func NewWheelHandler(viteService ViteService, wheelService WheelService, candidateService CandidateService, sessionService SessionService, logger *zap.Logger) *WheelHandler {
	return &WheelHandler{
		viteService:      viteService,
		wheelService:     wheelService,
		candidateService: candidateService,
		sessionService:   sessionService,
		logger:           logger,
	}
}

func (h *WheelHandler) Pattern() string { return "/wheel/{id}" }

func (h *WheelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Get(w, r)
	case http.MethodPost:
		h.Post(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

}

func (h *WheelHandler) Post(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		components.CandidateSubmissionForm(id, "Invalid wheel ID").Render(r.Context(), w)
		return
	}

	username := r.FormValue("username")
	if !usernameRegex.MatchString(username) {
		components.CandidateSubmissionForm(id, "Invalid username. Usernames must be 1-50 characters long and can only contain letters, numbers, underscores and spaces").Render(r.Context(), w)
		return
	}

	err = h.candidateService.AddCandidateToWheel(r.Context(), username, id)
	if err != nil {
		h.logger.Error("Failed to add candidate to wheel", zap.Int64("wheelID", id), zap.String("username", username), zap.Error(err))
		components.CandidateSubmissionForm(id, "Failed to add candidate to wheel").Render(r.Context(), w)
		return
	}

	components.StatusSuccess(fmt.Sprintf("Successfully added %s to the wheel!", username), 5).Render(r.Context(), w)
}

func (h *WheelHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid wheel ID", http.StatusBadRequest)
		return
	}

	logger := h.logger.With(zap.Int64("wheel_id", id))

	wheel, err := h.wheelService.GetWheelByID(r.Context(), id)
	if err != nil {
		logger.Error("Failed to get wheel", zap.Int64("wheel_id", id), zap.Error(err))
		http.Error(w, "Failed to get wheel", http.StatusInternalServerError)
		return
	}

	wheels, err := h.wheelService.CountWheels(r.Context())
	if err != nil {
		logger.Error("Failed to count wheels")
		http.Error(w, "Failed to count wheels", http.StatusInternalServerError)
		return
	}

	candidates, err := h.candidateService.ListCandidatesByWheel(r.Context(), wheel.ID)
	if err != nil {
		logger.Error("Failed to list candidates", zap.Int64("wheel_id", id), zap.Error(err))
		http.Error(w, "Failed to list candidates", http.StatusInternalServerError)
		return
	}

	sessionId, _ := h.sessionService.GetSessionIdFromContext(r.Context())

	var winner *repository.Candidate
	if wheel.WinnerID != nil {
		winner = services.GetCandidateFromListById(*wheel.WinnerID, candidates)
	}

	h.View(w, r, WheelPageViewProps{
		wheel:                *wheel,
		candidates:           candidates,
		sessionId:            sessionId,
		winner:               winner,
		renderSubmissionForm: !h.candidateService.CandidateInCandidateList(sessionId, candidates),
		renderBackButton:     wheels > 1,
	})
}

type WheelPageViewProps struct {
	wheel                repository.Wheel
	candidates           []repository.Candidate
	sessionId            string
	winner               *repository.Candidate
	renderSubmissionForm bool
	renderBackButton     bool
}

func (h *WheelHandler) View(w http.ResponseWriter, r *http.Request, props WheelPageViewProps) {
	components.WheelPage(h.viteService.Tags(), props.wheel, props.candidates, props.sessionId, props.winner, props.renderSubmissionForm, props.renderBackButton).Render(r.Context(), w)
}
