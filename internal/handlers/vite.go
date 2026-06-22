package handlers

import "net/http"

type (
	ViteHandler struct {
		inner http.Handler
	}
)

func NewViteHandler(viteService ViteService) (http.Handler, error) {

	return &ViteHandler{
		inner: http.FileServer(http.FS(viteService.AssetsFS())),
	}, nil
}

func (h *ViteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.inner.ServeHTTP(w, r)
}
