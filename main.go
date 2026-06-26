package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"

	dbcmd "github.com/ryanolee/a-perfectly-normal-wheel/cmd/db"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/handlers"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/logging"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/middleware"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/server"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/services"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

//go:embed all:frontend/dist
var distFS embed.FS

//go:embed all:frontend/img
var imgFS embed.FS

func startServer(dbPath string) {
	// Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	watermillLogger := logging.NewZapLoggerAdapter(logger)

	// Database
	dbConfig := db.DBConfig{
		FilePath: dbPath,
	}
	dbConnection := db.NewDBConnection(dbConfig)

	// Services
	viteService := services.NewViteService(&distFS, logger)
	wheelEventsService := services.NewWheelEventsService(logger, watermillLogger)
	wheelService := services.NewWheelService(dbConnection, wheelEventsService)
	sessionService := services.NewSessionService("your-secret-key")
	candidateService := services.NewCandidateService(dbConnection, wheelService, wheelEventsService, sessionService)

	// HTTP Handlers
	homeHandler := handlers.NewHomeHandler(viteService, wheelService, logger)
	viteHandler := handlers.NewViteHandler(viteService)
	imgFs, err := fs.Sub(&imgFS, "frontend")
	if err != nil {
		log.Fatalf("failed to sub img fs: %v", err)
	}
	imgHandler := http.FileServer(http.FS(imgFs))
	wheelHandler := handlers.NewWheelHandler(viteService, wheelService, candidateService, sessionService, logger)
	wheelEventsHandler := handlers.NewWheelEventsHandler(wheelService, wheelEventsService, sessionService, candidateService, logger)
	adminHandler := handlers.NewAdminHandler(viteService, wheelService)
	adminWheelHandler := handlers.NewAdminWheelHandler(viteService, wheelService, candidateService)
	adminApiHandler := handlers.NewAdminApiHandler(wheelService, candidateService, wheelEventsService, logger)

	// Server
	serverMux := server.NewServerMux(logger, homeHandler, viteHandler, imgHandler, wheelHandler, wheelEventsHandler, adminHandler, adminWheelHandler, adminApiHandler)

	// Middleware
	serverMux = middleware.LogRequests(serverMux, logger)
	serverMux = middleware.SessionMiddleware(serverMux, sessionService, logger)
	serverMux = middleware.BasicAuthMiddleware(serverMux, "admin", "password")

	log.Println("listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", serverMux); err != nil {
		logger.Fatal("failed to start server", zap.Error(err))
	}
}

var startDBPath string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the web server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer(startDBPath)
	},
}

func main() {
	startCmd.Flags().StringVar(&startDBPath, "db", "data.db", "path to the SQLite database file")
	dbcmd.RootCmd.AddCommand(startCmd)

	if err := dbcmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
