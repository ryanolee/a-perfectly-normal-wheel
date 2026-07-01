package handlers

import (
	"embed"
	"io/fs"
	"net/http"
)

type (
	ImgHandler struct {
		inner http.Handler
	}
)

func NewImageHandler(imageFs *embed.FS) (*ImgHandler, error) {
	filesystem, err := fs.Sub(imageFs, "frontend")
	if err != nil {
		return nil, err
	}

	return &ImgHandler{
		inner: http.FileServer(http.FS(filesystem)),
	}, nil
}

func (h *ImgHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.inner.ServeHTTP(w, r)
}

func (h *ImgHandler) Pattern() string { return "GET /img/" }
