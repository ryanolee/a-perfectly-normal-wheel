package server

import (
	"net/http"

	"go.uber.org/zap"
)

func NewServerMux(
	logger *zap.Logger,
	homeHandler http.Handler,
	viteHandler http.Handler,
	imgHandler http.Handler,
	wheelHandler http.Handler,
	wheelEventsHandler http.Handler,
	adminHandler http.Handler,
	adminWheelHandler http.Handler,
	adminApiHandler http.Handler,
) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/assets/", viteHandler)
	mux.Handle("/img/", imgHandler)
	mux.Handle("/wheel/{id}", wheelHandler)
	mux.Handle("/api/wheel/{path...}", wheelEventsHandler)
	mux.Handle("/admin", adminHandler)
	mux.Handle("/admin/wheel/{id}", adminWheelHandler)
	mux.Handle("/admin/api/{path...}", adminApiHandler)
	mux.Handle("/", homeHandler)

	return mux
}
