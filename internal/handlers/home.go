package handlers

import (
	"fmt"
	"net/http"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/components"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/repository"
	"go.uber.org/zap"
)

type (
	HomeHandler struct {
		wheelService WheelService
		viteService  ViteService
		logger       *zap.Logger
	}
)

func NewHomeHandler(viteService ViteService, wheelService WheelService, logger *zap.Logger) *HomeHandler {
	return &HomeHandler{
		viteService:  viteService,
		wheelService: wheelService,
		logger:       logger,
	}
}

func (h *HomeHandler) Pattern() string { return "/" }

func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		components.ErrorPage404(h.viteService.Tags()).Render(r.Context(), w)
		return
	}

	wheels, err := h.wheelService.ListWheels(r.Context())
	if err != nil {
		h.logger.Error("Failed to list wheels", zap.Error(err))
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
	wheels []repository.Wheel
}

func (h *HomeHandler) View(w http.ResponseWriter, r *http.Request, props HomeViewProps) {
	components.Home(h.viteService.Tags(), props.wheels).Render(r.Context(), w)
}
