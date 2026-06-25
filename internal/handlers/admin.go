package handlers

import (
	"net/http"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/components"
)

type (
	AdminHandler struct {
		viteService  ViteService
		wheelService WheelService
	}
)

func NewAdminHandler(viteService ViteService, wheelService WheelService) http.Handler {
	return &AdminHandler{
		viteService:  viteService,
		wheelService: wheelService,
	}
}

func (h *AdminHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wheels, err := h.wheelService.ListWheels(r.Context())
	if err != nil {
		http.Error(w, "Failed to list wheels", http.StatusInternalServerError)
		return
	}

	components.AdminDashboard(h.viteService.Tags(), wheels).Render(r.Context(), w)
}
