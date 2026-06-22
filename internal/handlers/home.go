package handlers

import (
	"fmt"
	"net/http"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/components"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/services"
)

type (
	HomeHandler struct {
		wheelService WheelService
		viteService  ViteService
	}
)

func NewHomeHandler(viteService ViteService, wheelService WheelService) (http.Handler, error) {
	return &HomeHandler{
		viteService:  viteService,
		wheelService: wheelService,
	}, nil
}

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	wheels, err := h.wheelService.ListWheels(r.Context())
	if err != nil {
		http.Error(w, "Failed to list wheels", http.StatusInternalServerError)
		return
	}

	if len(wheels) == 1 {
		http.Redirect(w, r, fmt.Sprintf("/wheel/%d", wheels[0].ID), http.StatusTemporaryRedirect)
		return
	}

	h.View(w, r, HomeViewProps{
		wheels: wheels,
	})
}

type HomeViewProps struct {
	wheels []services.Wheel
}

func (h *HomeHandler) View(w http.ResponseWriter, r *http.Request, props HomeViewProps) {
	components.Home(h.viteService.Tags(), props.wheels).Render(r.Context(), w)
}
