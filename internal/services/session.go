package services

import (
	"context"
	"net/http"

	jwt "github.com/golang-jwt/jwt/v5"
)

type (
	SessionService struct {
		signingKey []byte
	}
)

const (
	SessionCookieName  = "PHPSESSID" // Cursed PHP session name cuz hue hue hue
	ClaimsSessionIdKey = "session_id"
)

//type SessionService interface {
//	HasSession(r *http.Request) bool
//	GetSessionIdFromRequest(r *http.Request) (*string, error)
//	GetSessionIdFromContext(r *http.Request) (*string, error)
//	RequestWithSessionId(r *http.Request, sessionId string) *http.Request
//	SetSessionId(w http.ResponseWriter, r *http.Request, sessionId string) error
//}

func NewSessionService(signingKey string) *SessionService {
	return &SessionService{
		signingKey: []byte(signingKey),
	}
}

func (s *SessionService) HasSession(r *http.Request) bool {
	sessionId, err := s.GetSessionIdFromRequest(r)
	if err != nil {
		return false
	}
	return sessionId != nil
}

func (s *SessionService) GetSessionIdFromRequest(r *http.Request) (*string, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err == http.ErrNoCookie {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if sessionId, ok := claims[ClaimsSessionIdKey].(string); ok {
			return &sessionId, nil
		}
	}

	return nil, nil
}

func (s *SessionService) GetSessionIdFromContext(ctx context.Context) (string, bool) {
	sessionId, ok := ctx.Value(ClaimsSessionIdKey).(string)
	return sessionId, ok
}

func (s *SessionService) RequestWithSessionId(r *http.Request, sessionId string) *http.Request {
	ctx := context.WithValue(r.Context(), ClaimsSessionIdKey, sessionId)
	return r.WithContext(ctx)
}

func (s *SessionService) SetSessionId(w http.ResponseWriter, r *http.Request, sessionId string) error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		ClaimsSessionIdKey: sessionId,
	})

	tokenString, err := token.SignedString(s.signingKey)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:  SessionCookieName,
		Value: tokenString,
		Path:  "/",
	})

	return nil
}
