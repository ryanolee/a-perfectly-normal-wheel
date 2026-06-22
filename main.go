package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/handlers"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/logging"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/middleware"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/server"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/services"
	"go.uber.org/zap"
)

//go:embed all:frontend/dist
var distFS embed.FS

func main() {
	// Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	watermillLogger := logging.NewZapLoggerAdapter(logger)

	// Database
	dbConfig := db.DBConfig{
		FilePath: "data.db",
	}

	dbConnection, err := db.NewDBConnection(dbConfig)
	if err != nil {
		logger.Fatal("failed to create database connection", zap.Error(err))
	}

	// Services
	viteService, err := services.NewViteService(&distFS, logger)
	if err != nil {
		logger.Fatal("failed to create Vite service", zap.Error(err))
	}

	wheelService := services.NewWheelService(dbConnection)
	if wheelService == nil {
		logger.Fatal("failed to create wheel service")
	}

	wheelEventsService := services.NewWheelEventsService(logger, watermillLogger)

	sessionService := services.NewSessionService("your-secret-key")

	candidateService := services.NewCandidateService(dbConnection, wheelEventsService, sessionService)
	if candidateService == nil {
		logger.Fatal("failed to create candidate service")
	}

	// HTTP Handlers
	homeHandler, err := handlers.NewHomeHandler(viteService, wheelService)
	if err != nil {
		logger.Fatal("failed to create home handler: %v", zap.Error(err))
	}

	viteHandler, err := handlers.NewViteHandler(viteService)
	if err != nil {
		logger.Fatal("failed to create Vite handler: %v", zap.Error(err))
	}

	wheelHandler, err := handlers.NewWheelHandler(viteService, wheelService, candidateService, sessionService, logger)
	if err != nil {
		logger.Fatal("failed to create wheel handler: %v", zap.Error(err))
	}

	wheelEventsHandler := handlers.NewWheelEventsHandler(wheelService, wheelEventsService, sessionService, logger)
	if err != nil {
		logger.Fatal("failed to create wheel events handler: %v", zap.Error(err))
	}

	// Server
	serverMux := server.NewServerMux(logger, homeHandler, viteHandler, wheelHandler, wheelEventsHandler)

	// Middleware
	serverMux = middleware.LogRequests(logger, serverMux)
	serverMux = middleware.SessionMiddleware(serverMux, sessionService, logger)

	log.Println("listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", serverMux); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}

}
