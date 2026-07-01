package main

import (
	"embed"
	"log"
	"net/http"

	"github.com/ThreeDotsLabs/watermill"

	dbcmd "github.com/ryanolee/a-perfectly-normal-wheel/cmd/db"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/config"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/handlers"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/logging"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/middleware"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/repository"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/server"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/services"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

//go:embed all:frontend/dist
var distFS embed.FS

//go:embed all:frontend/img
var imgFS embed.FS

func startServer(dbPath string, secretsPath string) {

	fx.New(
		// Secrets
		fx.Provide(func() (*config.Secrets, error) {
			return config.LoadSecretsFromSopsFile(secretsPath)
		}),

		// Logging
		fx.Provide(zap.NewProduction),
		//fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
		//	return &fxevent.ZapLogger{Logger: logger}
		//}),
		fx.Provide(fx.Annotate(logging.NewZapWatermillLoggerAdapter, fx.As(new(watermill.LoggerAdapter)))),

		// Embedded Filesystems
		fx.Provide(fx.Annotate(func() *embed.FS { return &distFS }, fx.ResultTags(`name:"dist"`))),
		fx.Provide(fx.Annotate(func() *embed.FS { return &imgFS }, fx.ResultTags(`name:"img"`))),

		// Database
		fx.Supply(db.DBConfig{
			FilePath: dbPath,
		}),
		fx.Provide(db.NewDBConnection),

		// Repositories
		fx.Provide(fx.Annotate(repository.NewWheelRepository, fx.As(new(services.WheelRepository)))),
		fx.Provide(fx.Annotate(repository.NewCandidateRepository, fx.As(new(services.CandidateRepository)))),

		// Services
		fx.Provide(fx.Annotate(services.NewViteService, fx.ParamTags(`name:"dist"`), fx.As(new(handlers.ViteService)))),
		fx.Provide(fx.Annotate(services.NewWheelEventsService, fx.As(fx.Self()), fx.As(new(handlers.WheelEventsService)))),
		fx.Provide(fx.Annotate(services.NewWheelService, fx.As(fx.Self()), fx.As(new(handlers.WheelService)))),
		fx.Provide(fx.Annotate(services.NewSessionService, fx.As(fx.Self()), fx.As(new(handlers.SessionService)))),
		fx.Provide(fx.Annotate(services.NewCandidateService, fx.As(fx.Self()), fx.As(new(handlers.CandidateService)))),

		// HTTP Handlers
		server.AsRoute(handlers.NewHomeHandler),
		server.AsRoute(handlers.NewViteHandler),
		server.AsRoute(handlers.NewWheelHandler),
		server.AsRoute(handlers.NewWheelEventsHandler),
		server.AsRoute(handlers.NewAdminHandler),
		server.AsRoute(handlers.NewAdminWheelHandler),
		server.AsRoute(handlers.NewAdminApiHandler),
		server.AsRoute(handlers.NewImageHandler, fx.ParamTags(`name:"img"`)),

		// Middleware
		fx.Provide(func(logger *zap.Logger, session *services.SessionService, secrets *config.Secrets) []server.Middleware {
			return []server.Middleware{
				func(next http.Handler) http.Handler {
					return middleware.SessionMiddleware(next, session, logger)
				},
				func(next http.Handler) http.Handler {
					return middleware.BasicAuthMiddleware(next, "admin", secrets.AdminPassword)
				},
				func(next http.Handler) http.Handler {
					return middleware.LogRequests(next, logger)
				},
			}
		}),

		// Server
		fx.Provide(fx.Annotate(server.NewServerMux, fx.ParamTags(`group:"routes"`))),
		fx.Invoke(server.StartServer),
	).Run()
}

var startDBPath string
var startSecretsPath string

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the web server",
	Run: func(cmd *cobra.Command, args []string) {
		startServer(startDBPath, startSecretsPath)
	},
}

func main() {
	startCmd.Flags().StringVar(&startDBPath, "db", "data.db", "path to the SQLite database file")
	startCmd.Flags().StringVar(&startSecretsPath, "secrets", "config/secrets.yml", "path to the SOPS-encrypted secrets file")
	dbcmd.RootCmd.AddCommand(startCmd)

	if err := dbcmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
