package services

import (
	"context"
	"errors"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/ryanolee/a-perfectly-normal-wheel/internal/db"
)

type (
	CandidateQueries interface {
		ListCandidatesByWheel(context.Context, int64) ([]db.Candidate, error)
		AddCandidateToWheel(context.Context, db.AddCandidateToWheelParams) error
		GetDuplicateCandidatesForWheel(context.Context, db.GetDuplicateCandidatesForWheelParams) (db.Candidate, error)
		GetCandidateByCreatorIdAndWheelId(context.Context, db.GetCandidateByCreatorIdAndWheelIdParams) (db.Candidate, error)
	}

	CandidateService struct {
		dbQueries CandidateQueries
		events    *WheelEventsService
		session   *SessionService
	}
)

func NewCandidateService(dbQueries CandidateQueries, events *WheelEventsService, session *SessionService) *CandidateService {
	return &CandidateService{
		dbQueries: dbQueries,
		events:    events,
		session:   session,
	}
}

func (s *CandidateService) AddCandidateToWheel(ctx context.Context, username string, wheelID int64) error {
	sessionId, ok := s.session.GetSessionIdFromContext(ctx)
	if !ok {
		return errors.New("no session ID found in context")
	}

	if goaway.IsProfane(username) {
		return errors.New("username contains inappropriate language")
	}

	// Constraint
	candidate, err := s.dbQueries.GetDuplicateCandidatesForWheel(ctx, db.GetDuplicateCandidatesForWheelParams{
		WheelID:   wheelID,
		Username:  username,
		CreatorID: sessionId,
	})

	if err == nil && candidate.Username == username {
		return errors.New("duplicate candidate username for this wheel")
	}

	if err == nil && candidate.CreatorID == sessionId {
		return errors.New("you have already added a candidate to this wheel")
	}

	err = s.dbQueries.AddCandidateToWheel(ctx, db.AddCandidateToWheelParams{
		Username:  username,
		WheelID:   wheelID,
		CreatorID: sessionId,
	})

	if err != nil {
		return err
	}

	// Rehydrate the candidate to get the ID for the event
	dbCandidate, err := s.dbQueries.GetCandidateByCreatorIdAndWheelId(ctx, db.GetCandidateByCreatorIdAndWheelIdParams{
		CreatorID: sessionId,
		WheelID:   wheelID,
	})

	if err != nil {
		return err
	}

	return s.events.PublishNewCandidateAddedToWheelEvent(wheelID, CandidateFromDB(dbCandidate))
}

func (s *CandidateService) CandidateInCandidateList(creatorId string, candidates []Candidate) bool {
	for _, candidate := range candidates {
		if candidate.CreatorID == creatorId {
			return true
		}
	}
	return false
}

func (s *CandidateService) ListCandidatesByWheel(ctx context.Context, wheelID int64) ([]Candidate, error) {
	dbCandidates, err := s.dbQueries.ListCandidatesByWheel(ctx, wheelID)
	if err != nil {
		return nil, err
	}

	return CandidateListFromDB(dbCandidates), nil
}

type Candidate struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatorID string    `json:"creator_id"`
	CreatedAt time.Time `json:"created_at"`
}

func CandidateListFromDB(dbCandidates []db.Candidate) []Candidate {
	Candidates := make([]Candidate, len(dbCandidates))
	for i, dbCandidate := range dbCandidates {
		Candidates[i] = CandidateFromDB(dbCandidate)
	}
	return Candidates
}

func CandidateFromDB(u db.Candidate) Candidate {
	return Candidate{
		ID:        u.ID,
		Username:  u.Username,
		CreatorID: u.CreatorID,
		CreatedAt: u.CreatedAt.Time,
	}
}
