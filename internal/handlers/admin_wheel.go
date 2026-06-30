package handlers

import (
	"net/http"
	"strconv"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/components"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/repository"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/services"
)

type (
	AdminWheelHandler struct {
		viteService      ViteService
		wheelService     WheelService
		candidateService CandidateService
	}
)

func NewAdminWheelHandler(viteService ViteService, wheelService WheelService, candidateService CandidateService) http.Handler {
	return &AdminWheelHandler{
		viteService:      viteService,
		wheelService:     wheelService,
		candidateService: candidateService,
	}
}

func (h *AdminWheelHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	wheelId, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid wheel ID", http.StatusBadRequest)
		return
	}

	wheel, err := h.wheelService.GetWheelByID(r.Context(), wheelId)
	if err != nil {
		http.Error(w, "Failed to get wheel", http.StatusInternalServerError)
		return
	}

	candidates, err := h.candidateService.ListCandidatesByWheel(r.Context(), wheelId)
	if err != nil {
		http.Error(w, "Failed to list candidates", http.StatusInternalServerError)
		return
	}

	var winningCandidate *repository.Candidate
	if wheel.WinnerID != nil {
		winningCandidate = services.GetCandidateFromListById(*wheel.WinnerID, candidates)
	}

	h.View(w, r, AdminWheelViewProps{
		wheel:            *wheel,
		candidates:       candidates,
		winningCandidate: winningCandidate,
	})
}

type AdminWheelViewProps struct {
	wheel            repository.Wheel
	candidates       []repository.Candidate
	winningCandidate *repository.Candidate
}

func (h *AdminWheelHandler) View(w http.ResponseWriter, r *http.Request, props AdminWheelViewProps) {
	components.AdminWheelPage(h.viteService.Tags(), props.wheel, props.candidates, props.winningCandidate).Render(r.Context(), w)
}
