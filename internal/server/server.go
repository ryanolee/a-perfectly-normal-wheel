package server

import (
	"net/http"

	"go.uber.org/zap"
)

func NewServerMux(logger *zap.Logger, homeHandler http.Handler, viteHandler http.Handler, wheelHandler http.Handler, wheelEventsHandler http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/assets/", viteHandler)
	mux.Handle("/", homeHandler)
	mux.Handle("/wheel/{id}", wheelHandler)
	mux.Handle("/wheel/{id}/events", wheelEventsHandler)

	return mux
}
