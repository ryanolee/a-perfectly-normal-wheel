package server

import (
	"context"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Route interface {
	http.Handler
	Pattern() string
}

type Middleware func(http.Handler) http.Handler

func NewServerMux(routes []Route, middlewares []Middleware) http.Handler {
	mux := http.NewServeMux()
	for _, r := range routes {
		mux.Handle(r.Pattern(), r)
	}

	var handler http.Handler = mux
	for _, mw := range middlewares {
		handler = mw(handler)
	}
	return handler
}

func StartServer(lc fx.Lifecycle, handler http.Handler, logger *zap.Logger) {
	srv := &http.Server{Addr: ":8080", Handler: handler}
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			logger.Info("Starting server", zap.String("addr", srv.Addr))
			go func() {
				if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
					logger.Error("Server failed", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}

func AsRoute(handler interface{}, annotations ...fx.Annotation) fx.Option {
	annotations = append(annotations, fx.As(new(Route)), fx.ResultTags(`group:"routes"`))
	return fx.Provide(fx.Annotate(handler, annotations...))
}
