package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type SessionService interface {
	HasSession(r *http.Request) bool
	GetSessionIdFromRequest(r *http.Request) (*string, error)
	GetSessionIdFromContext(ctx context.Context) (string, bool)
	RequestWithSessionId(r *http.Request, sessionId string) *http.Request
	SetSessionId(w http.ResponseWriter, r *http.Request, sessionId string) error
}

func SessionMiddleware(next http.Handler, session SessionService, logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionId, err := session.GetSessionIdFromRequest(r)
		if err != nil {
			logger.Error("failed to get session ID", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		if sessionId == nil {
			newSessionId := uuid.New().String()
			err := session.SetSessionId(w, r, newSessionId)
			if err != nil {
				logger.Error("failed to set session ID", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			sessionId = &newSessionId
		}
		r = session.RequestWithSessionId(r, *sessionId)
		next.ServeHTTP(w, r)
	})
}
