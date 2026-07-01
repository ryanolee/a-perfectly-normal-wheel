package handlers

import "net/http"

type (
	ViteHandler struct {
		inner http.Handler
	}
)

func NewViteHandler(viteService ViteService) *ViteHandler {

	return &ViteHandler{
		inner: http.FileServer(http.FS(viteService.AssetsFS())),
	}
}

func (h *ViteHandler) Pattern() string { return "GET /assets/" }

func (h *ViteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.inner.ServeHTTP(w, r)
}
